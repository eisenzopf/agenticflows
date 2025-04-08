package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"agenticflows/backend/cmd/examples/client"

	_ "github.com/mattn/go-sqlite3"
)

// Dispute represents a fee dispute record
type Dispute struct {
	ID            string
	Text          string
	Amount        float64
	CreatedAt     time.Time
	Sentiment     string
	SentimentDesc string
}

// TrendAnalysis represents the analysis of fee dispute trends
type TrendAnalysis struct {
	TotalDisputes     int
	AverageAmount     float64
	CommonStatuses    []string
	CommonSentiments  []string
	TrendDescription  string
	RecommendedAction string
}

// Main function
func main() {
	// Parse command-line flags
	dbPath := flag.String("db", "", "Path to the SQLite database file")
	maxDisputes := flag.Int("max", 100, "Maximum number of disputes to analyze")
	batchSize := flag.Int("batch", 10, "Batch size for processing disputes")
	debug := flag.Bool("debug", false, "Enable debug mode")
	workflowID := flag.String("workflow", "", "Workflow ID for persisting results")
	flag.Parse()

	// Validate required flags
	if *dbPath == "" {
		fmt.Println("Error: --db flag is required")
		flag.Usage()
		os.Exit(1)
	}

	// Initialize API client
	apiClient := client.NewClient("http://localhost:8080", *workflowID, *debug)

	// Step 1: Fetch fee disputes
	fmt.Println("Fetching fee disputes from database...")
	disputes, err := fetchDisputes(*dbPath, *maxDisputes, apiClient)
	if err != nil {
		fmt.Printf("Error fetching disputes: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d fee disputes\n", len(disputes))

	// Step 2: Fetch example conversations
	fmt.Println("Fetching example conversations...")
	conversations, err := fetchConversations(*dbPath, 5) // Limit to 5 conversations
	if err != nil {
		fmt.Printf("Error fetching conversations: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d example conversations\n", len(conversations))

	// Step 3: Prepare structured data for the API
	disputeData := prepareDisputeData(disputes)

	// Step 4: Analyze trends in batches
	fmt.Println("\nAnalyzing trends in fee disputes...")
	trends, err := processBatchedTrends(apiClient, disputeData, conversations, *batchSize)
	if err != nil {
		fmt.Printf("Warning: Error analyzing trends: %v\n", err)
		fmt.Println("Continuing with partial or default trends...")
	}

	// Step 5: Identify patterns in batches
	fmt.Println("\nIdentifying patterns in fee disputes...")
	patterns, err := processBatchedPatterns(apiClient, disputeData, conversations, *batchSize)
	if err != nil {
		fmt.Printf("Warning: Error identifying patterns: %v\n", err)
		fmt.Println("Continuing with partial or default patterns...")
	}

	// Step 6: Generate findings and recommendations
	fmt.Println("\nGenerating findings and recommendations...")
	findings, recommendations, err := processBatchedFindings(apiClient, disputeData, conversations, trends, patterns, *batchSize)
	if err != nil {
		fmt.Printf("Warning: Error generating findings: %v\n", err)
		fmt.Println("Continuing with partial or default findings...")
	}

	// Step 7: Print results
	fmt.Println("\n=== Results ===")
	fmt.Println("\nTrends:")
	for _, trend := range trends.TrendDescriptions {
		fmt.Printf("- %s\n", trend)
	}

	fmt.Println("\nPatterns:")
	for _, pattern := range patterns {
		fmt.Printf("- %s\n", pattern)
	}

	fmt.Println("\nFindings:")
	for _, finding := range findings {
		fmt.Printf("- %s\n", finding)
	}

	fmt.Println("\nRecommendations:")
	for _, rec := range recommendations {
		fmt.Printf("- %s\n", rec)
	}

	fmt.Println("\nAnalysis complete!")
}

// prepareDisputeData converts disputes to a structured format for the API
func prepareDisputeData(disputes []Dispute) []map[string]interface{} {
	data := make([]map[string]interface{}, len(disputes))
	for i, dispute := range disputes {
		data[i] = map[string]interface{}{
			"id":          dispute.ID,
			"text":        dispute.Text,
			"amount":      dispute.Amount,
			"created_at":  dispute.CreatedAt.Format(time.RFC3339),
			"sentiment":   dispute.Sentiment,
			"description": dispute.SentimentDesc,
		}
	}
	return data
}

// Analysis represents trends analysis results
type Analysis struct {
	TrendDescriptions  []string `json:"trend_descriptions"`
	RecommendedActions []string `json:"recommended_actions"`
}

// processBatchedTrends processes disputes in batches for trend analysis
func processBatchedTrends(apiClient *client.Client, disputeData []map[string]interface{}, conversations []map[string]interface{}, batchSize int) (Analysis, error) {
	// Define default trends in case of failure
	defaultAnalysis := Analysis{
		TrendDescriptions:  []string{"No trends identified due to processing error"},
		RecommendedActions: []string{"Review the raw data manually to identify trends"},
	}

	// If there are no disputes, return default
	if len(disputeData) == 0 {
		return defaultAnalysis, fmt.Errorf("no dispute data to analyze")
	}

	// Process in smaller batches to avoid token limits
	batchCount := (len(disputeData) + batchSize - 1) / batchSize
	fmt.Printf("Processing %d disputes in %d batches for trend analysis\n", len(disputeData), batchCount)

	// Combine results from all batches
	var allTrends []string
	var allActions []string

	// Process each batch
	for i := 0; i < len(disputeData); i += batchSize {
		end := i + batchSize
		if end > len(disputeData) {
			end = len(disputeData)
		}
		batch := disputeData[i:end]

		fmt.Printf("Processing trends batch %d/%d (%d disputes)...\n", (i/batchSize)+1, batchCount, len(batch))

		// Create request for this batch - restructured to match API expectations
		req := client.StandardAnalysisRequest{
			AnalysisType: "trends",
			Text:         "", // Not using text field for this analysis
			Parameters: map[string]interface{}{
				"focus_areas": []string{
					"fee_dispute_trends",
					"customer_impact",
					"financial_impact",
				},
				"concise_response": true,
			},
			Data: map[string]interface{}{
				"disputes":      batch,
				"conversations": getLimitedConversations(conversations, 2),
				"attributes": map[string]interface{}{
					"avg_amount":       calculateAverageAmount(disputeData),
					"total_disputes":   len(disputeData),
					"dispute_timespan": "3 months",
				},
			},
		}

		// Make API request
		resp, err := apiClient.PerformAnalysis(req)
		if err != nil {
			fmt.Printf("Error analyzing trends in batch %d: %v\n", (i/batchSize)+1, err)
			continue
		}

		// Extract trends from response - updated to match API response format
		if results, ok := resp.Results.(map[string]interface{}); ok {
			// Extract trend descriptions
			if trendsData, ok := results["trend_descriptions"].([]interface{}); ok {
				for _, t := range trendsData {
					if trend, ok := t.(string); ok {
						allTrends = append(allTrends, trend)
					}
				}
			} else if trendsData, ok := results["trends"].([]interface{}); ok {
				// Try alternative field name
				for _, t := range trendsData {
					if trendMap, ok := t.(map[string]interface{}); ok {
						if desc, ok := trendMap["description"].(string); ok {
							allTrends = append(allTrends, desc)
						}
					} else if trend, ok := t.(string); ok {
						allTrends = append(allTrends, trend)
					}
				}
			}

			// Extract recommended actions
			if actionsData, ok := results["recommended_actions"].([]interface{}); ok {
				for _, a := range actionsData {
					if action, ok := a.(string); ok {
						allActions = append(allActions, action)
					}
				}
			} else if actionsData, ok := results["actions"].([]interface{}); ok {
				// Try alternative field name
				for _, a := range actionsData {
					if actionMap, ok := a.(map[string]interface{}); ok {
						if action, ok := actionMap["action"].(string); ok {
							allActions = append(allActions, action)
						}
					} else if action, ok := a.(string); ok {
						allActions = append(allActions, action)
					}
				}
			}
		}
	}

	// If we didn't get any trends, return the default
	if len(allTrends) == 0 {
		return defaultAnalysis, nil
	}

	// Return the combined results
	return Analysis{
		TrendDescriptions:  allTrends,
		RecommendedActions: allActions,
	}, nil
}

// processBatchedPatterns processes disputes in batches for pattern analysis
func processBatchedPatterns(apiClient *client.Client, disputeData []map[string]interface{}, conversations []map[string]interface{}, batchSize int) ([]string, error) {
	// Define default patterns in case of failure
	defaultPatterns := []string{"No patterns identified due to processing error"}

	// If there are no disputes, return default
	if len(disputeData) == 0 {
		return defaultPatterns, fmt.Errorf("no dispute data to analyze")
	}

	// Process in smaller batches to avoid token limits
	batchCount := (len(disputeData) + batchSize - 1) / batchSize
	fmt.Printf("Processing %d disputes in %d batches for pattern analysis\n", len(disputeData), batchCount)

	// Combine results from all batches
	var allPatterns []string

	// Process each batch
	for i := 0; i < len(disputeData); i += batchSize {
		end := i + batchSize
		if end > len(disputeData) {
			end = len(disputeData)
		}
		batch := disputeData[i:end]

		fmt.Printf("Processing patterns batch %d/%d (%d disputes)...\n", (i/batchSize)+1, batchCount, len(batch))

		// Create request for this batch - restructured to match API expectations
		req := client.StandardAnalysisRequest{
			AnalysisType: "patterns",
			Text:         "", // Not using text field for this analysis
			Parameters: map[string]interface{}{
				"pattern_types": []string{
					"fee_dispute_patterns",
					"resolution_patterns",
					"customer_behavior_patterns",
				},
				"concise_response": true,
			},
			Data: map[string]interface{}{
				"disputes":      batch,
				"conversations": getLimitedConversations(conversations, 2),
				"attributes": []map[string]interface{}{
					{
						"field_name":  "dispute_type",
						"title":       "Type of Fee Dispute",
						"description": "Categorization of the fee dispute (e.g., late fee, overdraft fee, service fee)",
					},
					{
						"field_name":  "resolution_offered",
						"title":       "Resolution Offered",
						"description": "Whether a resolution was offered to the customer",
					},
					{
						"field_name":  "resolution_type",
						"title":       "Type of Resolution",
						"description": "The type of resolution offered (e.g., full refund, partial refund)",
					},
					{
						"field_name":  "resolution_outcome",
						"title":       "Resolution Outcome",
						"description": "The final outcome of the dispute",
					},
					{
						"field_name":  "agent_explanation",
						"title":       "Agent Explanation",
						"description": "The agent's explanation for the fee",
					},
					{
						"field_name":  "de_escalation_techniques",
						"title":       "De-escalation Techniques Used",
						"description": "Techniques used by the agent to de-escalate the situation",
					},
					{
						"field_name":  "customer_sentiment_start",
						"title":       "Customer Sentiment (Start)",
						"description": "Customer's sentiment at the start of the conversation",
					},
					{
						"field_name":  "customer_sentiment_end",
						"title":       "Customer Sentiment (End)",
						"description": "Customer's sentiment at the end of the conversation",
					},
					{
						"field_name":  "call_id",
						"title":       "Call ID",
						"description": "Unique identifier for the call",
					},
					{
						"field_name":  "call_duration",
						"title":       "Call Duration",
						"description": "Length of the call",
					},
					{
						"field_name":  "agent_id",
						"title":       "Agent ID",
						"description": "Unique identifier for the agent",
					},
					{
						"field_name":  "call_timestamp",
						"title":       "Call Timestamp",
						"description": "When the call occurred",
					},
				},
			},
		}

		// Make API request
		resp, err := apiClient.PerformAnalysis(req)
		if err != nil {
			fmt.Printf("Error identifying patterns in batch %d: %v\n", (i/batchSize)+1, err)
			continue
		}

		// Extract patterns from response - updated to match API response format
		if results, ok := resp.Results.(map[string]interface{}); ok {
			// Try multiple possible field names for patterns
			if patternList, ok := results["patterns"].([]interface{}); ok {
				for _, pattern := range patternList {
					if patternMap, ok := pattern.(map[string]interface{}); ok {
						if desc, ok := patternMap["pattern_description"].(string); ok {
							allPatterns = append(allPatterns, desc)
						} else if desc, ok := patternMap["description"].(string); ok {
							allPatterns = append(allPatterns, desc)
						}
					} else if pattern, ok := pattern.(string); ok {
						allPatterns = append(allPatterns, pattern)
					}
				}
			}
		}
	}

	// If we didn't get any patterns, return the default
	if len(allPatterns) == 0 {
		return defaultPatterns, nil
	}

	// Return the combined results
	return allPatterns, nil
}

// processBatchedFindings processes disputes in batches for findings analysis
func processBatchedFindings(apiClient *client.Client, disputeData []map[string]interface{}, conversations []map[string]interface{},
	analysis Analysis, patterns []string, batchSize int) ([]string, []string, error) {
	// Define default findings in case of failure
	defaultFindings := []string{"No findings identified due to processing error"}
	defaultRecommendations := []string{"Review the raw data manually to identify recommendations"}

	// If there are no disputes, return defaults
	if len(disputeData) == 0 {
		return defaultFindings, defaultRecommendations, fmt.Errorf("no dispute data to analyze")
	}

	// Process in smaller batches to avoid token limits
	batchCount := (len(disputeData) + batchSize - 1) / batchSize
	fmt.Printf("Processing %d disputes in %d batches for findings analysis\n", len(disputeData), batchCount)

	// Combine results from all batches
	var allFindings []string
	var allRecommendations []string

	// Process each batch
	for i := 0; i < len(disputeData); i += batchSize {
		end := i + batchSize
		if end > len(disputeData) {
			end = len(disputeData)
		}
		batch := disputeData[i:end]

		fmt.Printf("Processing findings batch %d/%d (%d disputes)...\n", (i/batchSize)+1, batchCount, len(batch))

		// Create request for this batch - restructured to match API expectations
		req := client.StandardAnalysisRequest{
			AnalysisType: "findings",
			Text:         "", // Not using text field for this analysis
			Parameters: map[string]interface{}{
				"questions": []string{
					"What are the most common types of fee disputes?",
					"What is the average resolution time?",
					"What are the most effective resolution strategies?",
					"What are the key areas for improvement?",
				},
				"concise_response": true,
			},
			Data: map[string]interface{}{
				"disputes":      batch,
				"conversations": getLimitedConversations(conversations, 2),
				"trends":        analysis.TrendDescriptions, // Send just the trend descriptions
				"patterns":      patterns,
				"attributes": []map[string]interface{}{
					{
						"field_name":  "dispute_type",
						"title":       "Type of Fee Dispute",
						"description": "Categorization of the fee dispute (e.g., late fee, overdraft fee, service fee)",
					},
					{
						"field_name":  "resolution_offered",
						"title":       "Resolution Offered",
						"description": "Whether a resolution was offered to the customer",
					},
					{
						"field_name":  "resolution_type",
						"title":       "Type of Resolution",
						"description": "The type of resolution offered (e.g., full refund, partial refund)",
					},
					{
						"field_name":  "resolution_outcome",
						"title":       "Resolution Outcome",
						"description": "The final outcome of the dispute",
					},
					{
						"field_name":  "agent_explanation",
						"title":       "Agent Explanation",
						"description": "The agent's explanation for the fee",
					},
					{
						"field_name":  "de_escalation_techniques",
						"title":       "De-escalation Techniques Used",
						"description": "Techniques used by the agent to de-escalate the situation",
					},
					{
						"field_name":  "customer_sentiment_start",
						"title":       "Customer Sentiment (Start)",
						"description": "Customer's sentiment at the start of the conversation",
					},
					{
						"field_name":  "customer_sentiment_end",
						"title":       "Customer Sentiment (End)",
						"description": "Customer's sentiment at the end of the conversation",
					},
					{
						"field_name":  "call_id",
						"title":       "Call ID",
						"description": "Unique identifier for the call",
					},
					{
						"field_name":  "call_duration",
						"title":       "Call Duration",
						"description": "Length of the call",
					},
					{
						"field_name":  "agent_id",
						"title":       "Agent ID",
						"description": "Unique identifier for the agent",
					},
					{
						"field_name":  "call_timestamp",
						"title":       "Call Timestamp",
						"description": "When the call occurred",
					},
				},
			},
		}

		// Make API request
		resp, err := apiClient.PerformAnalysis(req)
		if err != nil {
			fmt.Printf("Error generating findings in batch %d: %v\n", (i/batchSize)+1, err)
			continue
		}

		// Extract findings from response - updated to match API response format
		if results, ok := resp.Results.(map[string]interface{}); ok {
			// Extract findings
			if findingsData, ok := results["findings"].([]interface{}); ok {
				for _, f := range findingsData {
					if finding, ok := f.(string); ok {
						allFindings = append(allFindings, finding)
					} else if findingMap, ok := f.(map[string]interface{}); ok {
						if finding, ok := findingMap["description"].(string); ok {
							allFindings = append(allFindings, finding)
						}
					}
				}
			}

			// Extract recommendations
			if recsData, ok := results["recommendations"].([]interface{}); ok {
				for _, r := range recsData {
					if rec, ok := r.(string); ok {
						allRecommendations = append(allRecommendations, rec)
					} else if recMap, ok := r.(map[string]interface{}); ok {
						if rec, ok := recMap["action"].(string); ok {
							allRecommendations = append(allRecommendations, rec)
						}
					}
				}
			}
		}
	}

	// If we didn't get any findings, return the defaults
	if len(allFindings) == 0 {
		return defaultFindings, defaultRecommendations, nil
	}

	// Return the combined results
	return allFindings, allRecommendations, nil
}

// getLimitedConversations returns a limited number of conversations
func getLimitedConversations(conversations []map[string]interface{}, limit int) []map[string]interface{} {
	if len(conversations) <= limit {
		return conversations
	}
	return conversations[:limit]
}

// Helper function to calculate the average amount of disputes
func calculateAverageAmount(disputes []map[string]interface{}) float64 {
	total := 0.0
	count := 0
	for _, dispute := range disputes {
		if amt, ok := dispute["amount"].(float64); ok {
			total += amt
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

// Helper function to find the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// fetchDisputes fetches fee disputes from the database
func fetchDisputes(dbPath string, limit int, apiClient *client.Client) ([]Dispute, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Query for fee disputes
	query := `
	SELECT 
		conversation_id,
		text,
		COALESCE(date_time, CURRENT_TIMESTAMP) as date_time
	FROM conversations
	WHERE text IS NOT NULL 
	AND LENGTH(text) > 100
	AND (
		text LIKE '%fee%'
		OR text LIKE '%charge%'
		OR text LIKE '%billing%'
		OR text LIKE '%refund%'
		OR text LIKE '%dispute%'
	)
	ORDER BY RANDOM()
	LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	// Format disputes as objects
	disputes := make([]Dispute, 0)
	for rows.Next() {
		var dispute Dispute
		var createdAtStr string
		if err := rows.Scan(&dispute.ID, &dispute.Text, &createdAtStr); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Parse created_at timestamp
		dispute.CreatedAt, _ = time.Parse("2006-01-02T15:04:05-07:00", createdAtStr)

		// Extract amount from text using the API
		req := client.StandardAnalysisRequest{
			AnalysisType: "attributes",
			Text:         dispute.Text,
			Parameters: map[string]interface{}{
				"attributes": []map[string]string{
					{
						"field_name":  "amount",
						"title":       "Disputed Amount",
						"description": "The amount of money being disputed",
					},
				},
			},
		}

		resp, err := apiClient.PerformAnalysis(req)
		if err == nil {
			if results, ok := resp.Results.(map[string]interface{}); ok {
				if attrValues, ok := results["attribute_values"].(map[string]interface{}); ok {
					if amountStr, ok := attrValues["amount"].(string); ok {
						// Try to parse the amount from the value
						fmt.Sscanf(amountStr, "$%f", &dispute.Amount)
					}
				}
			}
		}

		disputes = append(disputes, dispute)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return disputes, nil
}

// fetchConversations fetches example conversations from the database
func fetchConversations(dbPath string, limit int) ([]map[string]interface{}, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Query for representative conversations
	query := `
	SELECT 
		conversation_id,
		text,
		COALESCE(date_time, CURRENT_TIMESTAMP) as date_time
	FROM conversations
	WHERE text IS NOT NULL 
	AND LENGTH(text) > 200
	ORDER BY RANDOM()
	LIMIT ?
	`

	rows, err := db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	// Format conversations as objects
	conversations := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, text, createdAtStr string
		if err := rows.Scan(&id, &text, &createdAtStr); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Parse created_at timestamp
		createdAt, _ := time.Parse("2006-01-02T15:04:05-07:00", createdAtStr)

		conversations = append(conversations, map[string]interface{}{
			"id":         id,
			"text":       text,
			"created_at": createdAt.Format(time.RFC3339),
			"type":       "customer_service",
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return conversations, nil
}
