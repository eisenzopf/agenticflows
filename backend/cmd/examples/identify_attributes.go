package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// AttributeDefinition represents a structural definition of an attribute
type AttributeDefinition struct {
	FieldName   string   `json:"field_name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	EnumValues  []string `json:"enum_values,omitempty"`
}

func main() {
	// Parse command-line arguments
	dbPath := flag.String("db", "", "Path to the SQLite database")
	outputPath := flag.String("output", "", "Path to save results as JSON (optional)")
	intentType := flag.String("intent", "all", "Type of intent to analyze (all, cancel, upgrade, etc.)")
	workflowID := flag.String("workflow", "", "Workflow ID for persisting results")
	debugFlag := flag.Bool("debug", false, "Enable debug output")
	limit := flag.Int("limit", 100, "Maximum number of conversations to analyze")
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

	// Step 1: Define analysis questions
	questions := []string{
		"What are the key data points we need to track for customer conversations?",
		"What structured attributes would be most useful for analyzing customer issues?",
		"What fields would help us categorize and route customer requests effectively?",
		"What metrics would help us measure conversation quality and outcomes?",
	}

	// If a specific intent was provided, adjust the questions
	if *intentType != "all" {
		intentSpecificQuestions := []string{
			fmt.Sprintf("What are the key data points we need to track for '%s' conversations?", *intentType),
			fmt.Sprintf("What structured attributes would be most useful for analyzing '%s' issues?", *intentType),
			fmt.Sprintf("What fields would help us categorize and route '%s' requests effectively?", *intentType),
			fmt.Sprintf("What metrics would help us measure '%s' conversation quality and outcomes?", *intentType),
		}
		questions = intentSpecificQuestions
	}

	// Step 2: Fetch sample conversations
	conversations, err := fetchSampleConversations(*dbPath, *intentType, *limit)
	if err != nil {
		fmt.Printf("Error fetching conversations: %v\n", err)
		os.Exit(1)
	}

	if len(conversations) == 0 {
		fmt.Println("No conversations found in the database")
		os.Exit(1)
	}

	fmt.Printf("Found %d conversations for analysis\n", len(conversations))

	// Step 3: Generate attribute recommendations
	fmt.Println("Generating attribute recommendations...")
	
	// Combine sample conversations for context
	var sampleTexts []string
	for i, conv := range conversations {
		if i >= 5 { // Use at most 5 samples
			break
		}
		sampleTexts = append(sampleTexts, conv.Text)
	}
	
	conversationSamples := strings.Join(sampleTexts, "\n\n---\n\n")

	// Call API to generate attribute definitions
	resp, err := apiClient.IdentifyAttributes(questions, conversationSamples)
	if err != nil {
		fmt.Printf("Error identifying attributes: %v\n", err)
		os.Exit(1)
	}

	// Extract attribute definitions from response
	var attributeDefs []AttributeDefinition
	if attrsData, ok := resp["attributes"].([]interface{}); ok {
		for _, attr := range attrsData {
			if attrMap, ok := attr.(map[string]interface{}); ok {
				// Extract field values
				fieldName, _ := attrMap["field_name"].(string)
				title, _ := attrMap["title"].(string)
				description, _ := attrMap["description"].(string)
				attrType, _ := attrMap["type"].(string)
				
				// Skip if missing required fields
				if fieldName == "" || title == "" {
					continue
				}
				
				// Extract enum values if applicable
				var enumValues []string
				if enumStr, ok := attrMap["enum_values"].(string); ok && enumStr != "" {
					// Split by commas and trim spaces
					for _, val := range strings.Split(enumStr, ",") {
						enumValues = append(enumValues, strings.TrimSpace(val))
					}
				}
				
				// Create attribute definition
				attrDef := AttributeDefinition{
					FieldName:   fieldName,
					Title:       title,
					Description: description,
					Type:        attrType,
					EnumValues:  enumValues,
				}
				
				attributeDefs = append(attributeDefs, attrDef)
			}
		}
	}

	// Step 4: Validate attribute definitions
	fmt.Println("Validating attribute definitions...")
	
	// For now, we'll simply check that all attributes have required fields
	validAttrs := make([]AttributeDefinition, 0)
	for _, attr := range attributeDefs {
		if attr.FieldName != "" && attr.Title != "" && attr.Type != "" {
			validAttrs = append(validAttrs, attr)
		}
	}

	// Step 5: Print attribute definitions
	fmt.Printf("\n===== Identified Attributes (%d) =====\n", len(validAttrs))
	for i, attr := range validAttrs {
		fmt.Printf("\n%d. %s (%s)\n", i+1, attr.Title, attr.FieldName)
		fmt.Printf("   Type: %s\n", attr.Type)
		fmt.Printf("   Description: %s\n", attr.Description)
		
		if len(attr.EnumValues) > 0 {
			fmt.Printf("   Values: %s\n", strings.Join(attr.EnumValues, ", "))
		}
	}

	// Step 6: Save to database if requested
	if *workflowID != "" {
		if saveAttributesToDatabase(*dbPath, validAttrs, *workflowID, *intentType) {
			fmt.Printf("\nAttribute definitions saved to database %s\n", *dbPath)
		}
	}

	// Step 7: Save to output file if requested
	if *outputPath != "" {
		resultData := map[string]interface{}{
			"workflow_id":   *workflowID,
			"intent_type":   *intentType,
			"attributes":    validAttrs,
			"timestamp":     time.Now().Format(time.RFC3339),
			"conversation_count": len(conversations),
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

	PrintTimeTaken(startTime, "Identify attributes")
}

// fetchSampleConversations fetches sample conversations from the database
func fetchSampleConversations(dbPath, intentType string, limit int) ([]Conversation, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	var query string
	var args []interface{}

	if intentType == "all" {
		// Query for random conversations
		query = `
		SELECT conversation_id, text
		FROM conversations
		WHERE text IS NOT NULL AND LENGTH(text) > 100
		ORDER BY RANDOM()
		LIMIT ?
		`
		args = []interface{}{limit}
	} else {
		// Query for conversations with a specific intent
		query = `
		SELECT c.conversation_id, c.text
		FROM conversations c
		JOIN conversation_attributes ca ON c.conversation_id = ca.conversation_id
		WHERE ca.type = 'intent'
		AND ca.value LIKE ?
		AND c.text IS NOT NULL
		AND LENGTH(c.text) > 100
		ORDER BY RANDOM()
		LIMIT ?
		`
		args = []interface{}{"%" + intentType + "%", limit}
	}

	rows, err := db.Query(query, args...)
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

// saveAttributesToDatabase saves attribute definitions to the database
func saveAttributesToDatabase(dbPath string, attributes []AttributeDefinition, workflowID, intentType string) bool {
	if len(attributes) == 0 {
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
	
	// Check if attribute_definitions table exists, create if not
	_, err = tx.Exec(`
		CREATE TABLE IF NOT EXISTS attribute_definitions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			field_name TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			type TEXT NOT NULL,
			enum_values TEXT,
			intent_type TEXT,
			workflow_id TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		fmt.Printf("Error creating attribute_definitions table: %v\n", err)
		tx.Rollback()
		return false
	}
	
	// Prepare insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO attribute_definitions 
		(field_name, title, description, type, enum_values, intent_type, workflow_id) 
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		fmt.Printf("Error preparing statement: %v\n", err)
		tx.Rollback()
		return false
	}
	defer stmt.Close()
	
	// Insert each attribute
	for _, attr := range attributes {
		// Convert enum values to string
		enumValuesStr := ""
		if len(attr.EnumValues) > 0 {
			enumValuesStr = strings.Join(attr.EnumValues, ", ")
		}
		
		// Execute the insert
		_, err = stmt.Exec(
			attr.FieldName,
			attr.Title,
			attr.Description,
			attr.Type,
			enumValuesStr,
			intentType,
			workflowID,
		)
		
		if err != nil {
			fmt.Printf("Error inserting attribute: %v\n", err)
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