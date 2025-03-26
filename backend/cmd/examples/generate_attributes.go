package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Attribute represents a database attribute
type Attribute struct {
	Name        string
	Type        string
	Value       string
	Description string
	Count       int
}

// We don't need to redeclare Conversation since it's in utils.go
// Keeping the comment for clarity

func main() {
	// Define command-line flags
	dbPath := flag.String("db", "", "Path to the SQLite database")
	minCount := flag.Int("min-count", 20, "Minimum count for attributes to be considered")
	sampleSizeFlag := flag.Int("sample-size", 0, "DEPRECATED: Use --limit instead")
	limit := flag.Int("limit", 3, "Number of sample conversations to analyze")
	targetClass := flag.String("target-class", "fee dispute", "Target class for intent matching")
	debugFlag := flag.Bool("debug", false, "Enable debug mode")
	// Removing unused variables but keeping the flag definitions
	_ = flag.String("output", "", "Optional path to save results as JSON")
	_ = flag.Float64("threshold", 0.7, "Confidence threshold for attribute matching")
	workflowID := flag.String("workflow", "", "Workflow ID for persisting results")
	flag.Parse()

	// Validate required flags
	if *dbPath == "" {
		fmt.Println("Error: --db flag is required")
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

	// Step 1: Get required attributes
	questions := []string{
		"What are the most common types of fee disputes customers call about?",
		"How often do agents offer refunds or credits to resolve fee disputes?",
		"What percentage of fee disputes are resolved in the customer's favor?",
		"What explanations do agents provide for disputed fees?",
		"How do agents de-escalate conversations when customers are upset about fees?",
	}

	fmt.Println("Generating required attributes for fee dispute analysis...")
	resp, err := apiClient.GenerateRequiredAttributes(questions, "")
	if err != nil {
		fmt.Printf("Error generating required attributes: %v\n", err)
		os.Exit(1)
	}

	// Extract attributes from response
	var attributes []map[string]interface{}
	
	// Try to extract as []interface{} first
	if attrsInterface, ok := resp["attributes"].([]interface{}); ok {
		attributes = make([]map[string]interface{}, 0, len(attrsInterface))
		for _, attr := range attrsInterface {
			if attrMap, ok := attr.(map[string]interface{}); ok {
				attributes = append(attributes, attrMap)
			}
		}
	} else if attrsMaps, ok := resp["attributes"].([]map[string]interface{}); ok {
		// Try to extract directly as []map[string]interface{}
		attributes = attrsMaps
	} else {
		fmt.Println("Error: Unexpected response format")
		os.Exit(1)
	}
	
	// Deduplicate attributes by field_name
	requiredAttributes := make(map[string]map[string]string)
	for _, attrMap := range attributes {
		// Extract field_name and convert to string
		fieldName, ok := attrMap["field_name"].(string)
		if !ok || fieldName == "" {
			continue
		}
		
		// Convert to string map for easier handling
		strAttr := make(map[string]string)
		for k, v := range attrMap {
			if str, ok := v.(string); ok {
				strAttr[k] = str
			}
		}
		
		requiredAttributes[fieldName] = strAttr
	}

	fmt.Printf("\nIdentified %d required attributes:\n", len(requiredAttributes))
	for _, attr := range requiredAttributes {
		fmt.Printf("  - %s (%s): %s\n", attr["title"], attr["field_name"], attr["description"])
	}

	// Step 2: Fetch existing attributes from database
	existingAttributes, err := fetchExistingAttributes(*dbPath, *minCount)
	if err != nil {
		fmt.Printf("Error fetching existing attributes: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\nFound %d existing attributes in the database\n", len(existingAttributes))

	// Step 3: Match required attributes against existing ones
	// For this Go version, we'll do a simple matching based on name similarity
	// since we don't have the actual API endpoint for this operation yet
	matches := make(map[string]map[string]interface{})
	missing := make([]map[string]string, 0)

	for fieldName, reqAttr := range requiredAttributes {
		// Check for exact match
		found := false
		for _, existingAttr := range existingAttributes {
			if existingAttr.Name == fieldName {
				matches[fieldName] = map[string]interface{}{
					"field":      existingAttr.Name,
					"confidence": 1.0,
				}
				found = true
				break
			}
		}

		if !found {
			missing = append(missing, reqAttr)
		}
	}

	// Process matching results
	fmt.Println("\n=== Attribute Analysis Summary ===")
	fmt.Printf("Total required attributes: %d\n", len(requiredAttributes))
	fmt.Printf("Existing attributes: %d\n", len(matches))
	fmt.Printf("Missing attributes: %d\n", len(missing))

	// Print existing attributes with matches
	if len(matches) > 0 {
		fmt.Println("\n=== Existing Attributes ===")
		for reqField, matchInfo := range matches {
			reqAttr := requiredAttributes[reqField]
			fmt.Printf("\n✓ Found: %s (%s)\n", reqAttr["title"], reqField)
			fmt.Printf("  - Matched to database field: %s\n", matchInfo["field"])
			fmt.Printf("  - Confidence: %.2f\n", matchInfo["confidence"])
		}
	}

	// Print missing attributes
	if len(missing) > 0 {
		fmt.Println("\n=== Missing Attributes ===")
		for _, attr := range missing {
			fmt.Printf("\n✗ Missing: %s (%s)\n", attr["title"], attr["field_name"])
			fmt.Printf("  - Description: %s\n", attr["description"])
		}

		// Step 4: Find matching intents for the target class
		fmt.Printf("\nFinding intents related to '%s'...\n", *targetClass)
		matchingIntents, err := findMatchingIntents(*dbPath, *targetClass, *minCount)
		if err != nil {
			fmt.Printf("Error finding matching intents: %v\n", err)
			os.Exit(1)
		}

		// If no matching intents, use sample conversations
		conversations := make([]Conversation, 0)
		if len(matchingIntents) == 0 {
			fmt.Printf("No intents matching '%s' were found. Using random conversations instead.\n", *targetClass)
			conversations, err = fetchSampleConversations(*dbPath, *limit)
			if err != nil {
				fmt.Printf("Error fetching sample conversations: %v\n", err)
				os.Exit(1)
			}
		} else {
			// Step 5: Fetch conversations with matching intents
			fmt.Printf("\nFetching %d conversations with '%s' intents...\n", *limit, *targetClass)
			conversations, err = fetchConversationsByIntents(*dbPath, matchingIntents, *limit)
			if err != nil {
				fmt.Printf("Error fetching conversations by intents: %v\n", err)
				os.Exit(1)
			}

			if len(conversations) == 0 {
				fmt.Println("No conversations with matching intents found. Using random conversations instead.")
				conversations, err = fetchSampleConversations(*dbPath, *limit)
				if err != nil {
					fmt.Printf("Error fetching sample conversations: %v\n", err)
					os.Exit(1)
				}
			}
		}

		if len(conversations) == 0 {
			fmt.Println("No conversations found in the database.")
			os.Exit(1)
		}

		// Step 6: Generate values for missing attributes
		fmt.Println("\nGenerating values for missing attributes...")
		results := make([]map[string]interface{}, 0)

		for _, conv := range conversations {
			fmt.Printf("\nAnalyzing conversation %s...\n", conv.ID)

			// Convert missing attributes to the format expected by the API
			apiAttrs := make([]map[string]string, len(missing))
			for i, attr := range missing {
				apiAttrs[i] = attr
			}

			resp, err := apiClient.GenerateAttributes(conv.Text, apiAttrs)
			if err != nil {
				fmt.Printf("Error generating attributes: %v\n", err)
				continue
			}

			// Format the result
			result := map[string]interface{}{
				"conversation_id":  conv.ID,
				"attribute_values": resp["attribute_values"],
			}
			results = append(results, result)
		}

		// Print generated attribute values
		fmt.Println("\n=== Generated Attribute Values ===")
		for _, result := range results {
			fmt.Printf("\nConversation ID: %s\n", result["conversation_id"])
			if attrValues, ok := result["attribute_values"].([]interface{}); ok {
				for _, av := range attrValues {
					if attrValue, ok := av.(map[string]interface{}); ok {
						fmt.Printf("\n  Attribute: %s\n", attrValue["field_name"])
						fmt.Printf("  Value: %s\n", attrValue["value"])
						fmt.Printf("  Confidence: %.2f\n", attrValue["confidence"])
						fmt.Printf("  Explanation: %s\n", attrValue["explanation"])
					}
				}
			}
		}
	}

	PrintTimeTaken(startTime, "Generate attributes")
}

// fetchExistingAttributes fetches attributes from the database
func fetchExistingAttributes(dbPath string, minCount int) ([]Attribute, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Query for unique attributes that appear at least minCount times
	query := `
	SELECT name, type, COUNT(*) as count
	FROM conversation_attributes
	WHERE type = 'attribute'
	GROUP BY name
	HAVING COUNT(*) >= ?
	`

	rows, err := db.Query(query, minCount)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	// Format attributes as objects
	attributes := make([]Attribute, 0)
	for rows.Next() {
		var attr Attribute
		if err := rows.Scan(&attr.Name, &attr.Type, &attr.Count); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// For each unique attribute name, get a sample value
		valueQuery := `
		SELECT value, description
		FROM conversation_attributes
		WHERE name = ? AND type = 'attribute'
		LIMIT 1
		`
		var description sql.NullString
		var value sql.NullString
		if err := db.QueryRow(valueQuery, attr.Name).Scan(&value, &description); err != nil {
			if err != sql.ErrNoRows {
				return nil, fmt.Errorf("error querying value: %w", err)
			}
		}

		attr.Value = value.String
		attr.Description = description.String
		attributes = append(attributes, attr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return attributes, nil
}

// findMatchingIntents finds intents matching the target class
func findMatchingIntents(dbPath, targetClass string, minCount int) ([]string, error) {
	// For this simplified version, we'll just return intents containing the target class
	// In a real implementation, this would use the API to classify intents
	
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Query for intents that contain the target class
	query := `
	SELECT value as intent_text
	FROM conversation_attributes
	WHERE type = 'intent'
	AND lower(value) LIKE ?
	GROUP BY value
	HAVING COUNT(*) >= ?
	`

	rows, err := db.Query(query, "%"+targetClass+"%", minCount)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	// Format intents as strings
	intents := make([]string, 0)
	for rows.Next() {
		var intent string
		if err := rows.Scan(&intent); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		intents = append(intents, intent)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	fmt.Printf("Found %d intents matching '%s'\n", len(intents), targetClass)
	if len(intents) > 0 {
		fmt.Println("Examples of matching intents:")
		for i, intent := range intents {
			if i >= 5 {
				break
			}
			fmt.Printf("  - %s\n", intent)
		}
	}

	return intents, nil
}

// fetchConversationsByIntents fetches conversations with matching intents
func fetchConversationsByIntents(dbPath string, matchingIntents []string, limit int) ([]Conversation, error) {
	if len(matchingIntents) == 0 {
		return nil, nil
	}

	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Create placeholders for SQL IN clause
	placeholders := "?"
	for i := 1; i < len(matchingIntents); i++ {
		placeholders += ",?"
	}

	// Query for conversations with matching intents
	query := fmt.Sprintf(`
	SELECT c.conversation_id, c.text
	FROM conversations c
	JOIN conversation_attributes ca ON c.conversation_id = ca.conversation_id
	WHERE ca.type = 'intent' AND ca.value IN (%s)
	AND c.text IS NOT NULL AND LENGTH(c.text) > 100
	ORDER BY RANDOM()
	LIMIT ?
	`, placeholders)

	// Create args slice
	args := make([]interface{}, len(matchingIntents)+1)
	for i, intent := range matchingIntents {
		args[i] = intent
	}
	args[len(matchingIntents)] = limit

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

// fetchSampleConversations fetches random sample conversations
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