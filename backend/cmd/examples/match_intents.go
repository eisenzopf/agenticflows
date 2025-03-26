package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// IntentMatch represents a matched intent result
type IntentMatch struct {
	Intent     string  `json:"intent"`
	Confidence float64 `json:"confidence"`
	Matched    bool    `json:"matched"`
	Reason     string  `json:"reason"`
}

// MatchResult represents the result of intent matching
type MatchResult struct {
	ConversationID string        `json:"conversation_id"`
	TargetIntent   string        `json:"target_intent"`
	Matches        []IntentMatch `json:"matches"`
	ExactMatch     bool          `json:"exact_match"`
	TopMatch       string        `json:"top_match"`
	Confidence     float64       `json:"confidence"`
	Text           string        `json:"text"`
}

func main() {
	// Parse command-line arguments
	dbPath := flag.String("db", "", "Path to the SQLite database")
	outputPath := flag.String("output", "", "Path to save results as JSON (optional)")
	targetIntent := flag.String("target", "", "Target intent to match against")
	intentList := flag.String("intents", "", "List of intents to match (comma-separated)")
	limit := flag.Int("limit", 25, "Number of conversations to sample per intent")
	sampleSizeFlag := flag.Int("sample-size", 0, "DEPRECATED: Use --limit instead")
	threshold := flag.Float64("threshold", 0.7, "Confidence threshold for matching")
	workflowID := flag.String("workflow", "", "Workflow ID for persisting results")
	debugFlag := flag.Bool("debug", false, "Enable debug output")
	flag.Parse()

	// Validate required flags
	if *dbPath == "" {
		fmt.Println("Error: --db flag is required")
		flag.Usage()
		os.Exit(1)
	}

	if *targetIntent == "" && *intentList == "" {
		fmt.Println("Error: Either --target or --intents flag must be provided")
		flag.Usage()
		os.Exit(1)
	}
	
	// Handle deprecated sampleSize flag
	if *sampleSizeFlag > 0 {
		fmt.Println("Warning: --sample-size is deprecated, please use --limit instead")
		*limit = *sampleSizeFlag
	}

	startTime := time.Now()

	// Create API client
	apiClient := NewApiClient(*workflowID, *debugFlag)
	
	// Print debug information if debug flag is enabled
	if *debugFlag {
		fmt.Println("Debug mode enabled: LLM inputs and outputs will be printed")
	}

	// Step 1: Determine intents to match
	var intentsToMatch []string
	if *intentList != "" {
		// Split by commas and trim spaces
		for _, intent := range strings.Split(*intentList, ",") {
			trimmedIntent := strings.TrimSpace(intent)
			if trimmedIntent != "" {
				intentsToMatch = append(intentsToMatch, trimmedIntent)
			}
		}
	} else if *targetIntent != "" {
		intentsToMatch = []string{*targetIntent}
	}

	if len(intentsToMatch) == 0 {
		fmt.Println("Error: No valid intents provided")
		os.Exit(1)
	}

	fmt.Printf("Matching against %d intents: %s\n", len(intentsToMatch), strings.Join(intentsToMatch, ", "))

	// Step 2: Fetch sample conversations for each intent
	allResults := make([]MatchResult, 0)
	
	for _, intent := range intentsToMatch {
		// Fetch conversations with this intent
		conversations, err := fetchConversationsByIntent(*dbPath, intent, *limit)
		if err != nil {
			fmt.Printf("Error fetching conversations for intent '%s': %v\n", intent, err)
			continue
		}
		
		if len(conversations) == 0 {
			fmt.Printf("No conversations found for intent '%s'\n", intent)
			continue
		}
		
		fmt.Printf("\nAnalyzing %d conversations for intent '%s'...\n", len(conversations), intent)
		
		// Process each conversation
		for _, conv := range conversations {
			result, err := matchIntent(conv, intent, intentsToMatch, *threshold, apiClient)
			if err != nil {
				fmt.Printf("Error matching conversation %s: %v\n", conv.ID, err)
				continue
			}
			
			allResults = append(allResults, result)
			
			// Print result
			if result.ExactMatch {
				fmt.Printf("✓ Conversation %s matched to '%s' (confidence: %.2f)\n", 
					conv.ID, result.TopMatch, result.Confidence)
			} else {
				fmt.Printf("✗ Conversation %s matched to '%s' instead of '%s' (confidence: %.2f)\n", 
					conv.ID, result.TopMatch, result.TargetIntent, result.Confidence)
			}
		}
	}

	// Step 3: Calculate and print statistics
	printMatchStatistics(allResults, intentsToMatch)

	// Step 4: Save to output file if requested
	if *outputPath != "" {
		resultData := map[string]interface{}{
			"workflow_id":   *workflowID,
			"match_results": allResults,
			"timestamp":     time.Now().Format(time.RFC3339),
			"threshold":     *threshold,
		}
		
		jsonData, err := json.MarshalIndent(resultData, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}
		
		err = os.WriteFile(*outputPath, jsonData, 0644)
		if err != nil {
			fmt.Printf("Error writing output file: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("\nResults written to %s\n", *outputPath)
	}

	// Step 5: Save to database if requested
	if *workflowID != "" {
		if saveMatchResultsToDatabase(*dbPath, allResults, *workflowID) {
			fmt.Printf("\nMatch results saved to database %s\n", *dbPath)
		}
	}

	PrintTimeTaken(startTime, "Match intents")
}

// fetchConversationsByIntent fetches conversations with a specific intent from the database
func fetchConversationsByIntent(dbPath, intent string, limit int) ([]Conversation, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Query for conversations with the specified intent
	query := `
	SELECT c.conversation_id, c.text
	FROM conversations c
	JOIN conversation_attributes ca ON c.conversation_id = ca.conversation_id
	WHERE ca.type = 'intent'
	AND ca.value = ?
	AND c.text IS NOT NULL
	AND LENGTH(c.text) > 100
	ORDER BY RANDOM()
	LIMIT ?
	`

	rows, err := db.Query(query, intent, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	// Format conversations as objects
	conversations := make([]Conversation, 0)
	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.ID, &conv.Text); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		conversations = append(conversations, conv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return conversations, nil
}

// matchIntent matches a conversation against a set of intents
func matchIntent(conv Conversation, targetIntent string, allIntents []string, threshold float64, apiClient *ApiClient) (MatchResult, error) {
	// Initialize result
	result := MatchResult{
		ConversationID: conv.ID,
		TargetIntent:   targetIntent,
		Matches:        make([]IntentMatch, 0),
		ExactMatch:     false,
		Text:           conv.Text,
	}
	
	// Call API to generate intent
	resp, err := apiClient.GenerateIntent(conv.Text)
	if err != nil {
		return result, fmt.Errorf("error generating intent: %w", err)
	}
	
	// Extract primary intent
	primaryIntent, ok := resp["intent"].(string)
	if !ok {
		return result, fmt.Errorf("unexpected response format - missing intent")
	}
	
	confidence, ok := resp["confidence"].(float64)
	if !ok {
		confidence = 0.0
	}
	
	explanation, ok := resp["explanation"].(string)
	if !ok {
		explanation = ""
	}
	
	// Set top match and confidence
	result.TopMatch = primaryIntent
	result.Confidence = confidence
	
	// Check if it's an exact match
	exactMatch := strings.ToLower(primaryIntent) == strings.ToLower(targetIntent)
	result.ExactMatch = exactMatch
	
	// Create match for primary intent
	primaryMatch := IntentMatch{
		Intent:     primaryIntent,
		Confidence: confidence,
		Matched:    exactMatch,
		Reason:     explanation,
	}
	
	result.Matches = append(result.Matches, primaryMatch)
	
	// Add matches for other intents (with zero confidence since we don't have the data)
	for _, intent := range allIntents {
		if strings.ToLower(intent) != strings.ToLower(primaryIntent) {
			match := IntentMatch{
				Intent:     intent,
				Confidence: 0.0,
				Matched:    false,
				Reason:     "Not matched",
			}
			result.Matches = append(result.Matches, match)
		}
	}
	
	return result, nil
}

// printMatchStatistics prints statistics about intent matching
func printMatchStatistics(results []MatchResult, intents []string) {
	if len(results) == 0 {
		fmt.Println("\nNo results to analyze")
		return
	}
	
	fmt.Println("\n===== Intent Matching Statistics =====")
	
	// Calculate overall accuracy
	totalMatches := 0
	for _, result := range results {
		if result.ExactMatch {
			totalMatches++
		}
	}
	
	accuracy := float64(totalMatches) / float64(len(results))
	fmt.Printf("\nOverall accuracy: %.2f%% (%d/%d)\n", accuracy*100, totalMatches, len(results))
	
	// Calculate per-intent statistics
	intentStats := make(map[string]struct {
		Total   int
		Correct int
		Wrong   int
	})
	
	// Initialize stats for each intent
	for _, intent := range intents {
		intentStats[intent] = struct {
			Total   int
			Correct int
			Wrong   int
		}{0, 0, 0}
	}
	
	// Calculate stats
	for _, result := range results {
		stats := intentStats[result.TargetIntent]
		stats.Total++
		
		if result.ExactMatch {
			stats.Correct++
		} else {
			stats.Wrong++
		}
		
		intentStats[result.TargetIntent] = stats
	}
	
	// Print per-intent statistics
	fmt.Println("\nPer-intent statistics:")
	
	// Sort intents by accuracy
	type IntentAccuracy struct {
		Intent    string
		Accuracy  float64
		Total     int
		Correct   int
		Wrong     int
	}
	
	intentAccuracies := make([]IntentAccuracy, 0, len(intentStats))
	for intent, stats := range intentStats {
		var accuracy float64
		if stats.Total > 0 {
			accuracy = float64(stats.Correct) / float64(stats.Total)
		}
		
		intentAccuracies = append(intentAccuracies, IntentAccuracy{
			Intent:    intent,
			Accuracy:  accuracy,
			Total:     stats.Total,
			Correct:   stats.Correct,
			Wrong:     stats.Wrong,
		})
	}
	
	// Sort by accuracy (descending)
	sort.Slice(intentAccuracies, func(i, j int) bool {
		return intentAccuracies[i].Accuracy > intentAccuracies[j].Accuracy
	})
	
	// Print sorted intents
	for _, ia := range intentAccuracies {
		if ia.Total > 0 {
			fmt.Printf("  %s: %.2f%% (%d/%d)\n", ia.Intent, ia.Accuracy*100, ia.Correct, ia.Total)
		}
	}
	
	// Print confusion matrix
	fmt.Println("\nConfusion Matrix (Actual → Predicted):")
	
	// Calculate prediction counts
	confusionMatrix := make(map[string]map[string]int)
	for _, intent := range intents {
		confusionMatrix[intent] = make(map[string]int)
	}
	
	for _, result := range results {
		if confusionMatrix[result.TargetIntent] == nil {
			confusionMatrix[result.TargetIntent] = make(map[string]int)
		}
		confusionMatrix[result.TargetIntent][result.TopMatch]++
	}
	
	// Print matrix
	fmt.Printf("%-30s", "Actual \\ Predicted")
	for _, intent := range intents {
		fmt.Printf(" %-15s", intent)
	}
	fmt.Println()
	
	for _, actual := range intents {
		fmt.Printf("%-30s", actual)
		for _, predicted := range intents {
			count := confusionMatrix[actual][predicted]
			fmt.Printf(" %-15d", count)
		}
		fmt.Println()
	}
	
	// Print common error examples
	fmt.Println("\nCommon misclassifications:")
	
	// Find misclassified examples
	misclassified := make(map[string][]MatchResult)
	for _, result := range results {
		if !result.ExactMatch {
			key := fmt.Sprintf("%s → %s", result.TargetIntent, result.TopMatch)
			misclassified[key] = append(misclassified[key], result)
		}
	}
	
	// Sort misclassifications by frequency
	type MisclassEntry struct {
		Key    string
		Count  int
		Results []MatchResult
	}
	
	misclassEntries := make([]MisclassEntry, 0, len(misclassified))
	for key, results := range misclassified {
		misclassEntries = append(misclassEntries, MisclassEntry{
			Key:    key,
			Count:  len(results),
			Results: results,
		})
	}
	
	// Sort by count (descending)
	sort.Slice(misclassEntries, func(i, j int) bool {
		return misclassEntries[i].Count > misclassEntries[j].Count
	})
	
	// Print top misclassifications
	for i, entry := range misclassEntries {
		if i >= 5 { // Limit to top 5
			break
		}
		
		fmt.Printf("\n%s (%d occurrences):\n", entry.Key, entry.Count)
		
		// Print example conversations (up to 2)
		maxExamples := 2
		if len(entry.Results) < maxExamples {
			maxExamples = len(entry.Results)
		}
		
		for j := 0; j < maxExamples; j++ {
			result := entry.Results[j]
			fmt.Printf("  Example %d (Conversation %s):\n", j+1, result.ConversationID)
			
			// Truncate text if too long
			text := "N/A"
			for _, conv := range results {
				if conv.ConversationID == result.ConversationID {
					text = conv.Text
					break
				}
			}
			
			if len(text) > 100 {
				text = text[:100] + "..."
			}
			
			fmt.Printf("  %s\n", text)
		}
	}
}

// saveMatchResultsToDatabase saves match results to the database
func saveMatchResultsToDatabase(dbPath string, results []MatchResult, workflowID string) bool {
	if len(results) == 0 {
		return false
	}
	
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return false
	}
	defer db.Close()
	
	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("Error beginning transaction: %v\n", err)
		return false
	}
	
	// Check if intent_match_results table exists, create if not
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS intent_match_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			conversation_id TEXT NOT NULL,
			target_intent TEXT NOT NULL,
			top_match TEXT NOT NULL,
			confidence REAL,
			exact_match INTEGER,
			workflow_id TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Printf("Error creating intent_match_results table: %v\n", err)
		tx.Rollback()
		return false
	}
	
	// Prepare insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO intent_match_results 
		(conversation_id, target_intent, top_match, confidence, exact_match, workflow_id) 
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		fmt.Printf("Error preparing statement: %v\n", err)
		tx.Rollback()
		return false
	}
	defer stmt.Close()
	
	// Insert each result
	for _, result := range results {
		exactMatchInt := 0
		if result.ExactMatch {
			exactMatchInt = 1
		}
		
		// Execute the insert
		_, err = stmt.Exec(
			result.ConversationID,
			result.TargetIntent,
			result.TopMatch,
			result.Confidence,
			exactMatchInt,
			workflowID,
		)
		
		if err != nil {
			fmt.Printf("Error inserting match result: %v\n", err)
			tx.Rollback()
			return false
		}
	}
	
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		fmt.Printf("Error committing transaction: %v\n", err)
		return false
	}
	
	return true
} 