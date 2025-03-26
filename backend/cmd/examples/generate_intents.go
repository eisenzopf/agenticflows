package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Intent represents an intent classification result
type Intent struct {
	Text        string  `json:"text"`
	Intent      string  `json:"intent"`
	Confidence  float64 `json:"confidence"`
	Explanation string  `json:"explanation"`
}

func main() {
	// Parse command-line arguments
	dbPath := flag.String("db", "", "Path to the SQLite database")
	inputPath := flag.String("input", "", "Path to a file containing conversations to classify (optional)")
	outputPath := flag.String("output", "", "Path to save results as JSON (optional)")
	limit := flag.Int("limit", 10, "Number of conversations to process if no input file is provided")
	workflowID := flag.String("workflow", "", "Workflow ID for persisting results")
	debugFlag := flag.Bool("debug", false, "Enable debug output")
	flag.Parse()

	// Validate required flags
	if *dbPath == "" {
		fmt.Println("Error: --db flag is required")
		flag.Usage()
		os.Exit(1)
	}

	startTime := time.Now()

	// Create API client
	apiClient := NewApiClient(*workflowID, *debugFlag)
	
	// Print debug information if debug flag is enabled
	if *debugFlag {
		fmt.Println("Debug mode enabled: LLM inputs and outputs will be printed")
	}

	// Step 1: Get conversations to analyze
	conversations, err := getConversationsToAnalyze(*dbPath, *inputPath, *limit)
	if err != nil {
		fmt.Printf("Error getting conversations: %v\n", err)
		os.Exit(1)
	}

	if len(conversations) == 0 {
		fmt.Println("No conversations found to analyze")
		os.Exit(1)
	}

	fmt.Printf("Analyzing %d conversations for intents...\n", len(conversations))

	// Step 2: Generate intents for each conversation
	results := make([]Intent, 0, len(conversations))
	for i, conv := range conversations {
		fmt.Printf("\nAnalyzing conversation %d/%d (ID: %s)...\n", i+1, len(conversations), conv.ID)
		
		resp, err := apiClient.GenerateIntent(conv.Text)
		if err != nil {
			fmt.Printf("Error generating intent: %v\n", err)
			continue
		}
		
		// Extract intent data from response
		// The API returns 'label', 'label_name', and 'description' instead of 'intent'
		intent, ok := resp["label"].(string)
		if !ok {
			// Try label_name as fallback
			intent, ok = resp["label_name"].(string)
			if !ok {
				fmt.Println("Error: Unexpected response format - missing intent")
				continue
			}
		}
		
		confidence, ok := resp["confidence"].(float64)
		if !ok {
			confidence = 1.0  // If confidence is not provided, assume high confidence
		}
		
		explanation, ok := resp["description"].(string)
		if !ok {
			explanation = ""
		}
		
		// Create intent result
		intentResult := Intent{
			Text:        conv.Text,
			Intent:      intent,
			Confidence:  confidence,
			Explanation: explanation,
		}
		
		results = append(results, intentResult)
		
		fmt.Printf("Intent: %s\n", intent)
		fmt.Printf("Confidence: %.2f\n", confidence)
		fmt.Printf("Explanation: %s\n", explanation)
	}

	// Step 3: Write results to output file if requested
	if *outputPath != "" {
		resultData := map[string]interface{}{
			"workflow_id": *workflowID,
			"intents":     results,
			"timestamp":   time.Now().Format(time.RFC3339),
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
	
	// Step 4: Save to database if database path was provided
	if saveToDatabase(*dbPath, results, *workflowID) {
		fmt.Printf("\nResults saved to database %s\n", *dbPath)
	}

	PrintTimeTaken(startTime, "Generate intents")
}

// getConversationsToAnalyze fetches conversations from the database or input file
func getConversationsToAnalyze(dbPath, inputPath string, limit int) ([]Conversation, error) {
	// If input file is provided, read conversations from it
	if inputPath != "" {
		return readConversationsFromFile(inputPath)
	}
	
	// Otherwise, fetch random conversations from the database
	return fetchSampleConversations(dbPath, limit)
}

// readConversationsFromFile reads conversations from a JSON file
func readConversationsFromFile(inputPath string) ([]Conversation, error) {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return nil, fmt.Errorf("error reading input file: %w", err)
	}
	
	// Try to parse as array of conversations
	var conversations []struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	}
	
	err = json.Unmarshal(data, &conversations)
	if err != nil {
		// Try to parse as single conversation
		var conversation struct {
			ID   string `json:"id"`
			Text string `json:"text"`
		}
		
		err = json.Unmarshal(data, &conversation)
		if err != nil {
			return nil, fmt.Errorf("error parsing input file: %w", err)
		}
		
		return []Conversation{{ID: conversation.ID, Text: conversation.Text}}, nil
	}
	
	// Convert to Conversation structs
	result := make([]Conversation, len(conversations))
	for i, conv := range conversations {
		result[i] = Conversation{ID: conv.ID, Text: conv.Text}
	}
	
	return result, nil
}

// fetchSampleConversations fetches random sample conversations from the database
func fetchSampleConversations(dbPath string, limit int) ([]Conversation, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Query for random conversations
	query := `
	SELECT conversation_id, text
	FROM conversations
	WHERE text IS NOT NULL AND LENGTH(text) > 100
	ORDER BY RANDOM()
	LIMIT ?
	`

	rows, err := db.Query(query, limit)
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

// saveToDatabase saves the generated intents to the database
func saveToDatabase(dbPath string, intents []Intent, workflowID string) bool {
	if len(intents) == 0 {
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
	
	// Prepare insert statement for conversation_attributes
	// Remove confidence, explanation, and workflow_id since those columns don't exist
	stmt, err := tx.Prepare(`
		INSERT INTO conversation_attributes 
		(conversation_id, type, name, value) 
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		fmt.Printf("Error preparing statement: %v\n", err)
		tx.Rollback()
		return false
	}
	defer stmt.Close()
	
	// Insert each intent
	for i, intent := range intents {
		// Check if we have a conversation_id
		conversationID := fmt.Sprintf("autogen_%d", i+1)
		
		// Execute the insert - removed confidence, explanation, and workflow_id parameters
		_, err = stmt.Exec(conversationID, "intent", "primary_intent", intent.Intent)
		
		if err != nil {
			fmt.Printf("Error inserting intent: %v\n", err)
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