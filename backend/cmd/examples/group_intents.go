package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// IntentGroup represents a group of similar intents
type IntentGroup struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
	Count       int      `json:"count"`
}

// IntentResult represents a conversation intent
type IntentResult struct {
	ConversationID string  `json:"conversation_id"`
	Intent         string  `json:"intent"`
	Confidence     float64 `json:"confidence"`
}

func main() {
	// Parse command-line arguments
	dbPath := flag.String("db", "", "Path to the SQLite database")
	outputPath := flag.String("output", "", "Path to save results as JSON (optional)")
	minCount := flag.Int("min-count", 100, "Minimum count for intent groups")
	maxGroups := flag.Int("max-groups", 20, "Maximum number of intent groups to generate")
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

	// Step 1: Fetch intents from database
	intents, err := fetchIntents(*dbPath, *minCount)
	if err != nil {
		fmt.Printf("Error fetching intents: %v\n", err)
		os.Exit(1)
	}

	if len(intents) == 0 {
		fmt.Println("No intents found in the database")
		os.Exit(1)
	}

	fmt.Printf("Found %d intents with minimum count of %d\n", len(intents), *minCount)

	// Step 2: Group similar intents
	groups, err := groupIntents(intents, *maxGroups, apiClient)
	if err != nil {
		fmt.Printf("Error grouping intents: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Generate descriptions for groups
	describeGroups(groups, apiClient)

	// Step 4: Print results
	printGroups(groups)

	// Step 5: Save to output file if requested
	if *outputPath != "" {
		resultData := map[string]interface{}{
			"workflow_id":  *workflowID,
			"intent_groups": groups,
			"timestamp":    time.Now().Format(time.RFC3339),
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

	// Step 6: Save to database if requested
	if *workflowID != "" {
		if saveGroupsToDatabase(*dbPath, groups, *workflowID) {
			fmt.Printf("\nIntent groups saved to database %s\n", *dbPath)
		}
	}

	PrintTimeTaken(startTime, "Group intents")
}

// fetchIntents fetches intents from the database
func fetchIntents(dbPath string, minCount int) (map[string]int, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Query for unique intents that appear at least minCount times
	query := `
	SELECT value as intent, COUNT(*) as count
	FROM conversation_attributes
	WHERE type = 'intent'
	GROUP BY value
	HAVING COUNT(*) >= ?
	ORDER BY COUNT(*) DESC
	`

	rows, err := db.Query(query, minCount)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	// Format intents as a map of intent -> count
	intents := make(map[string]int)
	for rows.Next() {
		var intent string
		var intentCount int
		if err := rows.Scan(&intent, &intentCount); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		intents[intent] = intentCount
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return intents, nil
}

// groupIntents groups similar intents
func groupIntents(intents map[string]int, maxGroups int, apiClient *ApiClient) ([]IntentGroup, error) {
	fmt.Println("Grouping similar intents...")
	
	// Check if we have the API client for advanced grouping
	if apiClient == nil {
		// Fallback to simple grouping by common words
		return groupIntentsByCommonWords(intents, maxGroups), nil
	}
	
	// Prepare the list of intents for the API
	intentsList := make([]map[string]interface{}, 0, len(intents))
	for intent, count := range intents {
		intentsList = append(intentsList, map[string]interface{}{
			"intent": intent,
			"count":  count,
		})
	}
	
	// Call the API to group intents
	resp, err := apiClient.GroupIntents(intentsList, maxGroups)
	if err != nil {
		fmt.Printf("Error calling API to group intents: %v\n", err)
		fmt.Println("Falling back to simple grouping by common words...")
		return groupIntentsByCommonWords(intents, maxGroups), nil
	}
	
	// Extract groups from response
	groupsData, ok := resp["groups"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}
	
	// Convert response to IntentGroup structs
	groups := make([]IntentGroup, 0, len(groupsData))
	for _, g := range groupsData {
		groupMap, ok := g.(map[string]interface{})
		if !ok {
			continue
		}
		
		// Extract name
		name, ok := groupMap["name"].(string)
		if !ok || name == "" {
			continue
		}
		
		// Extract examples
		examplesData, ok := groupMap["examples"].([]interface{})
		if !ok {
			continue
		}
		
		examples := make([]string, 0, len(examplesData))
		for _, ex := range examplesData {
			if exStr, ok := ex.(string); ok {
				examples = append(examples, exStr)
			}
		}
		
		// Calculate total count for this group
		count := 0
		for _, ex := range examples {
			count += intents[ex]
		}
		
		// Create group
		group := IntentGroup{
			Name:     name,
			Examples: examples,
			Count:    count,
		}
		
		groups = append(groups, group)
	}
	
	// Sort groups by count in descending order
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Count > groups[j].Count
	})
	
	return groups, nil
}

// groupIntentsByCommonWords groups intents by common words (fallback method)
func groupIntentsByCommonWords(intents map[string]int, maxGroups int) []IntentGroup {
	// Simplified grouping approach:
	// 1. Extract significant words from each intent
	// 2. Group intents that share significant words
	
	// Create a map of word -> intents containing that word
	wordToIntents := make(map[string][]string)
	
	// Stopwords that shouldn't be used for grouping
	stopwords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"in": true, "on": true, "at": true, "to": true, "for": true,
		"of": true, "with": true, "about": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "but": true, "by": true, "from": true,
	}
	
	// Extract words from each intent
	for intent := range intents {
		words := strings.Fields(strings.ToLower(intent))
		
		// Filter out stopwords and very short words
		for _, word := range words {
			if len(word) > 3 && !stopwords[word] {
				// Remove trailing punctuation
				word = strings.TrimRight(word, ".,;:!?")
				
				if wordToIntents[word] == nil {
					wordToIntents[word] = make([]string, 0)
				}
				wordToIntents[word] = append(wordToIntents[word], intent)
			}
		}
	}
	
	// Find the most common words that appear in multiple intents
	type WordCount struct {
		Word  string
		Count int
	}
	
	wordCounts := make([]WordCount, 0, len(wordToIntents))
	for word, wordIntents := range wordToIntents {
		if len(wordIntents) > 1 {
			wordCounts = append(wordCounts, WordCount{
				Word:  word,
				Count: len(wordIntents),
			})
		}
	}
	
	// Sort words by frequency
	sort.Slice(wordCounts, func(i, j int) bool {
		return wordCounts[i].Count > wordCounts[j].Count
	})
	
	// Take the top N words as group labels (where N = maxGroups)
	numGroups := int(math.Min(float64(maxGroups), float64(len(wordCounts))))
	
	// Create groups based on top words
	groups := make([]IntentGroup, 0, numGroups)
	processedIntents := make(map[string]bool)
	
	for i := 0; i < numGroups; i++ {
		if i >= len(wordCounts) {
			break
		}
		
		word := wordCounts[i].Word
		relatedIntents := wordToIntents[word]
		
		// Filter out already processed intents
		uniqueIntents := make([]string, 0)
		for _, intent := range relatedIntents {
			if !processedIntents[intent] {
				uniqueIntents = append(uniqueIntents, intent)
				processedIntents[intent] = true
			}
		}
		
		if len(uniqueIntents) == 0 {
			continue
		}
		
		// Calculate total count for this group
		count := 0
		for _, intent := range uniqueIntents {
			count += intents[intent]
		}
		
		// Create a more readable group name by capitalizing the key word
		groupName := strings.Title(word) + "-related inquiries"
		
		group := IntentGroup{
			Name:     groupName,
			Examples: uniqueIntents,
			Count:    count,
		}
		
		groups = append(groups, group)
	}
	
	// Handle any remaining intents
	var remainingIntents []string
	for intent := range intents {
		if !processedIntents[intent] {
			remainingIntents = append(remainingIntents, intent)
		}
	}
	
	if len(remainingIntents) > 0 {
		// Calculate total count for remaining intents
		count := 0
		for _, intent := range remainingIntents {
			count += intents[intent]
		}
		
		// Add a miscellaneous group
		group := IntentGroup{
			Name:     "Miscellaneous inquiries",
			Examples: remainingIntents,
			Count:    count,
		}
		
		groups = append(groups, group)
	}
	
	// Sort groups by count in descending order
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Count > groups[j].Count
	})
	
	return groups
}

// describeGroups generates descriptions for intent groups
func describeGroups(groups []IntentGroup, apiClient *ApiClient) {
	fmt.Println("Generating descriptions for intent groups...")
	
	for i := range groups {
		// Skip already described groups
		if groups[i].Description != "" {
			continue
		}
		
		// Prepare examples for the description
		examples := make([]string, 0)
		if len(groups[i].Examples) > 5 {
			examples = groups[i].Examples[:5]
		} else {
			examples = groups[i].Examples
		}
		
		// Skip if no examples
		if len(examples) == 0 {
			continue
		}
		
		// Generate description
		fmt.Printf("Request to describe_group:\n%s\n", groups[i].Name)
		resp, err := apiClient.DescribeGroup(groups[i].Name, examples)
		if err != nil {
			fmt.Printf("Error generating description for group %s: %v\n", groups[i].Name, err)
			groups[i].Description = fmt.Sprintf("This group contains conversations related to %s.", strings.ToLower(strings.TrimSuffix(groups[i].Name, " inquiries")))
			continue
		}
		
		// Extract description from response
		description, ok := resp["description"].(string)
		if !ok || description == "" {
			groups[i].Description = fmt.Sprintf("This group contains conversations related to %s.", strings.ToLower(strings.TrimSuffix(groups[i].Name, " inquiries")))
		} else {
			groups[i].Description = description
		}
	}
}

// printGroups prints intent groups
func printGroups(groups []IntentGroup) {
	fmt.Println("\n===== Intent Groups =====")
	
	totalIntents := 0
	for _, group := range groups {
		totalIntents += len(group.Examples)
	}
	
	fmt.Printf("Total groups: %d (containing %d unique intents)\n\n", len(groups), totalIntents)
	
	for i, group := range groups {
		fmt.Printf("%d. %s\n", i+1, group.Name)
		fmt.Printf("   Description: %s\n", group.Description)
		fmt.Printf("   Count: %d\n", group.Count)
		fmt.Printf("   Examples (%d):\n", len(group.Examples))
		
		// Print at most 5 examples
		maxExamples := int(math.Min(float64(len(group.Examples)), 5))
		for j := 0; j < maxExamples; j++ {
			fmt.Printf("     - %s\n", group.Examples[j])
		}
		
		if len(group.Examples) > maxExamples {
			fmt.Printf("     ... and %d more\n", len(group.Examples)-maxExamples)
		}
		
		fmt.Println()
	}
}

// saveGroupsToDatabase saves intent groups to the database
func saveGroupsToDatabase(dbPath string, groups []IntentGroup, workflowID string) bool {
	if len(groups) == 0 {
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
	
	// Prepare insert statement for intent_groups
	stmt, err := tx.Prepare(`
		INSERT INTO intent_groups 
		(name, description, example_count, total_count, workflow_id, created_at) 
		VALUES (?, ?, ?, ?, ?, datetime('now'))
	`)
	if err != nil {
		// Table might not exist, try to create it
		_, err = tx.Exec(`
			CREATE TABLE IF NOT EXISTS intent_groups (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				description TEXT,
				example_count INTEGER,
				total_count INTEGER,
				workflow_id TEXT,
				created_at TIMESTAMP
			)
		`)
		if err != nil {
			fmt.Printf("Error creating intent_groups table: %v\n", err)
			tx.Rollback()
			return false
		}
		
		// Try preparing statement again
		stmt, err = tx.Prepare(`
			INSERT INTO intent_groups 
			(name, description, example_count, total_count, workflow_id, created_at) 
			VALUES (?, ?, ?, ?, ?, datetime('now'))
		`)
		if err != nil {
			fmt.Printf("Error preparing statement: %v\n", err)
			tx.Rollback()
			return false
		}
	}
	defer stmt.Close()
	
	// Prepare insert statement for intent_group_examples
	exampleStmt, err := tx.Prepare(`
		INSERT INTO intent_group_examples 
		(group_id, intent, workflow_id) 
		VALUES (?, ?, ?)
	`)
	if err != nil {
		// Table might not exist, try to create it
		_, err = tx.Exec(`
			CREATE TABLE IF NOT EXISTS intent_group_examples (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				group_id INTEGER,
				intent TEXT NOT NULL,
				workflow_id TEXT,
				FOREIGN KEY (group_id) REFERENCES intent_groups (id)
			)
		`)
		if err != nil {
			fmt.Printf("Error creating intent_group_examples table: %v\n", err)
			tx.Rollback()
			return false
		}
		
		// Try preparing statement again
		exampleStmt, err = tx.Prepare(`
			INSERT INTO intent_group_examples 
			(group_id, intent, workflow_id) 
			VALUES (?, ?, ?)
		`)
		if err != nil {
			fmt.Printf("Error preparing example statement: %v\n", err)
			tx.Rollback()
			return false
		}
	}
	defer exampleStmt.Close()
	
	// Insert each group
	for _, group := range groups {
		// Insert group
		result, err := stmt.Exec(
			group.Name,
			group.Description,
			len(group.Examples),
			group.Count,
			workflowID,
		)
		if err != nil {
			fmt.Printf("Error inserting group: %v\n", err)
			tx.Rollback()
			return false
		}
		
		// Get the ID of the inserted group
		groupID, err := result.LastInsertId()
		if err != nil {
			fmt.Printf("Error getting last insert ID: %v\n", err)
			tx.Rollback()
			return false
		}
		
		// Insert examples
		for _, example := range group.Examples {
			_, err = exampleStmt.Exec(
				groupID,
				example,
				workflowID,
			)
			if err != nil {
				fmt.Printf("Error inserting example: %v\n", err)
				tx.Rollback()
				return false
			}
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