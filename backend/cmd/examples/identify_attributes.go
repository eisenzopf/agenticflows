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
	debugFlag := flag.Bool("debug", true, "Enable debug output")
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
	
	// Print debug information if debug flag is enabled
	if *debugFlag {
		fmt.Println("Debug mode enabled: LLM inputs and outputs will be printed")
	}

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
	var attributeDefs []AttributeDefinition
	
	resp, err := apiClient.IdentifyAttributes(questions, conversationSamples)
	if err != nil {
		fmt.Printf("Warning: Error identifying attributes: %v\n", err)
		fmt.Println("Using fallback attribute definitions instead...")
		
		// Fallback: Define some generic banking conversation attributes
		attributeDefs = []AttributeDefinition{
			{
				FieldName:   "issue_type",
				Title:       "Issue Type",
				Description: "The primary type of issue being discussed in the conversation",
				Type:        "string",
				EnumValues:  []string{"fee_dispute", "account_inquiry", "technical_issue", "fraud_concern", "service_complaint", "other"},
			},
			{
				FieldName:   "resolution_status",
				Title:       "Resolution Status",
				Description: "Whether and how the customer's issue was resolved",
				Type:        "string",
				EnumValues:  []string{"resolved", "partially_resolved", "escalated", "unresolved"},
			},
			{
				FieldName:   "customer_sentiment",
				Title:       "Customer Sentiment",
				Description: "The overall emotional tone of the customer during the conversation",
				Type:        "string",
				EnumValues:  []string{"positive", "neutral", "negative", "very_negative"},
			},
			{
				FieldName:   "fee_amount",
				Title:       "Fee Amount",
				Description: "The monetary amount of the fee in dispute",
				Type:        "number",
			},
			{
				FieldName:   "fee_waived",
				Title:       "Fee Waived",
				Description: "Whether the fee was waived or reduced",
				Type:        "boolean",
			},
		}
	} else {
		// Get the intent from the response
		var intent string
		if label, ok := resp["label"].(string); ok {
			intent = label
		} else if labelName, ok := resp["label_name"].(string); ok {
			intent = labelName
		}

		// Generate attributes based on the intent and questions
		attributeDefs = generateAttributesFromIntent(intent, questions)
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

// generateAttributesFromIntent creates appropriate attribute definitions based on the intent and questions
func generateAttributesFromIntent(intent string, questions []string) []AttributeDefinition {
	// Start with common attributes for all banking conversations
	attributeDefs := []AttributeDefinition{
		{
			FieldName:   "customer_sentiment",
			Title:       "Customer Sentiment",
			Description: "The overall emotional tone of the customer during the conversation",
			Type:        "string",
			EnumValues:  []string{"positive", "neutral", "negative", "very_negative"},
		},
		{
			FieldName:   "resolution_status",
			Title:       "Resolution Status",
			Description: "Whether and how the customer's issue was resolved",
			Type:        "string",
			EnumValues:  []string{"resolved", "partially_resolved", "escalated", "unresolved"},
		},
	}

	// Add intent-specific attributes
	switch intent {
	case "Account Fee Dispute":
		attributeDefs = append(attributeDefs, []AttributeDefinition{
			{
				FieldName:   "fee_type",
				Title:       "Fee Type",
				Description: "The type of fee being disputed",
				Type:        "string",
				EnumValues:  []string{"overdraft", "monthly_service", "atm", "wire_transfer", "other"},
			},
			{
				FieldName:   "fee_amount",
				Title:       "Fee Amount",
				Description: "The monetary amount of the fee in dispute",
				Type:        "number",
			},
			{
				FieldName:   "fee_waived",
				Title:       "Fee Waived",
				Description: "Whether the fee was waived or reduced",
				Type:        "boolean",
			},
			{
				FieldName:   "dispute_reason",
				Title:       "Dispute Reason",
				Description: "The customer's stated reason for disputing the fee",
				Type:        "string",
			},
		}...)
	case "Account Inquiry":
		attributeDefs = append(attributeDefs, []AttributeDefinition{
			{
				FieldName:   "inquiry_type",
				Title:       "Inquiry Type",
				Description: "The type of account information being requested",
				Type:        "string",
				EnumValues:  []string{"balance", "transaction_history", "account_status", "other"},
			},
			{
				FieldName:   "account_type",
				Title:       "Account Type",
				Description: "The type of account being inquired about",
				Type:        "string",
				EnumValues:  []string{"checking", "savings", "credit_card", "loan", "other"},
			},
		}...)
	case "Technical Issue":
		attributeDefs = append(attributeDefs, []AttributeDefinition{
			{
				FieldName:   "issue_platform",
				Title:       "Issue Platform",
				Description: "The platform or channel where the technical issue occurred",
				Type:        "string",
				EnumValues:  []string{"mobile_app", "web_browser", "atm", "phone_system", "other"},
			},
			{
				FieldName:   "error_message",
				Title:       "Error Message",
				Description: "Any specific error message or code encountered",
				Type:        "string",
			},
			{
				FieldName:   "issue_resolved",
				Title:       "Issue Resolved",
				Description: "Whether the technical issue was resolved during the conversation",
				Type:        "boolean",
			},
		}...)
	case "Fraud Concern":
		attributeDefs = append(attributeDefs, []AttributeDefinition{
			{
				FieldName:   "fraud_type",
				Title:       "Fraud Type",
				Description: "The type of fraud concern reported",
				Type:        "string",
				EnumValues:  []string{"unauthorized_transaction", "suspicious_activity", "identity_theft", "other"},
			},
			{
				FieldName:   "transaction_amount",
				Title:       "Transaction Amount",
				Description: "The amount of the suspicious transaction",
				Type:        "number",
			},
			{
				FieldName:   "account_frozen",
				Title:       "Account Frozen",
				Description: "Whether the account was frozen as a precaution",
				Type:        "boolean",
			},
		}...)
	case "Service Complaint":
		attributeDefs = append(attributeDefs, []AttributeDefinition{
			{
				FieldName:   "complaint_type",
				Title:       "Complaint Type",
				Description: "The type of service complaint",
				Type:        "string",
				EnumValues:  []string{"wait_time", "staff_behavior", "process_issue", "other"},
			},
			{
				FieldName:   "resolution_offered",
				Title:       "Resolution Offered",
				Description: "Whether a specific resolution was offered to the customer",
				Type:        "boolean",
			},
			{
				FieldName:   "compensation_offered",
				Title:       "Compensation Offered",
				Description: "Whether any compensation was offered to the customer",
				Type:        "boolean",
			},
		}...)
	default:
		// For unknown intents, add generic attributes based on the questions
		for _, question := range questions {
			// Extract key terms from questions to create relevant attributes
			lowerQuestion := strings.ToLower(question)
			if strings.Contains(lowerQuestion, "amount") || strings.Contains(lowerQuestion, "cost") {
				attributeDefs = append(attributeDefs, AttributeDefinition{
					FieldName:   "amount",
					Title:       "Amount",
					Description: "The monetary amount discussed in the conversation",
					Type:        "number",
				})
			}
			if strings.Contains(lowerQuestion, "date") || strings.Contains(lowerQuestion, "when") {
				attributeDefs = append(attributeDefs, AttributeDefinition{
					FieldName:   "date",
					Title:       "Date",
					Description: "The date discussed in the conversation",
					Type:        "string",
				})
			}
			if strings.Contains(lowerQuestion, "reason") || strings.Contains(lowerQuestion, "why") {
				attributeDefs = append(attributeDefs, AttributeDefinition{
					FieldName:   "reason",
					Title:       "Reason",
					Description: "The reason or explanation discussed in the conversation",
					Type:        "string",
				})
			}
		}
	}

	return attributeDefs
} 