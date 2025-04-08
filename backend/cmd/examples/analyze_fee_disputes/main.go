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

		// Enable debug for the first batch only if debug flag is on
		if i == 0 {
			apiClient.EnableDebug()
		} else {
			apiClient.DisableDebug()
		}

		// Create request for this batch
		req := client.StandardAnalysisRequest{
			AnalysisType: "trends",
			Parameters: map[string]interface{}{
				"focus_areas": []string{
					"fee_dispute_trends",
					"customer_impact",
					"financial_impact",
				},
				"concise_response": true,
			},
			Data: map[string]interface{}{
				"attribute_values": batch,
				"conversations":    getLimitedConversations(conversations, 2),
				"metadata": map[string]interface{}{
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

		// Extract trends from response
		if results, ok := resp.Results.(map[string]interface{}); ok {
			// Try to extract trends
			extractedTrends := extractArrayFromMap(results, []string{
				"trend_descriptions", "trends",
			})

			for _, t := range extractedTrends {
				if trend, ok := t.(string); ok {
					allTrends = append(allTrends, trend)
				} else if trendMap, ok := t.(map[string]interface{}); ok {
					if desc, ok := trendMap["trend"].(string); ok {
						allTrends = append(allTrends, desc)
					}
				}
			}

			// Try to extract insights/recommended actions
			extractedActions := extractArrayFromMap(results, []string{
				"recommended_actions", "overall_insights",
			})

			for _, a := range extractedActions {
				if action, ok := a.(string); ok {
					allActions = append(allActions, action)
				}
			}
		}
	}

	// If we didn't get any trends, return the default
	if len(allTrends) == 0 {
		return defaultAnalysis, fmt.Errorf("no trends extracted from analysis results")
	}

	return Analysis{
		TrendDescriptions:  allTrends,
		RecommendedActions: allActions,
	}, nil
}

// processBatchedPatterns processes disputes in batches for pattern analysis
func processBatchedPatterns(apiClient *client.Client, disputeData []map[string]interface{}, conversations []map[string]interface{}, batchSize int) ([]string, error) {
	// If there are no disputes, return empty
	if len(disputeData) == 0 {
		return []string{"No patterns identified due to empty dataset"}, nil
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

		// Only enable debug for first batch
		if i == 0 {
			apiClient.EnableDebug()
		} else {
			apiClient.DisableDebug()
		}

		// Create request for this batch
		req := client.StandardAnalysisRequest{
			AnalysisType: "patterns",
			Parameters: map[string]interface{}{
				"pattern_types": []string{
					"fee_disputes",
					"customer_behavior",
					"resolution_patterns",
				},
				"concise_response": true,
			},
			Data: map[string]interface{}{
				"attribute_values": batch,
				"conversations":    getLimitedConversations(conversations, 2),
			},
		}

		// Make API request
		resp, err := apiClient.PerformAnalysis(req)
		if err != nil {
			fmt.Printf("Error identifying patterns in batch %d: %v\n", (i/batchSize)+1, err)
			continue
		}

		// Extract patterns from response
		if results, ok := resp.Results.(map[string]interface{}); ok {
			// Try to extract patterns directly
			if patterns, ok := results["patterns"].([]interface{}); ok {
				for _, p := range patterns {
					if pattern, ok := p.(string); ok {
						allPatterns = append(allPatterns, pattern)
					}
				}
			}

			// Try alternative field names that might contain patterns
			for _, field := range []string{"identified_patterns", "key_patterns"} {
				if patterns, ok := results[field].([]interface{}); ok {
					for _, p := range patterns {
						if pattern, ok := p.(string); ok {
							allPatterns = append(allPatterns, pattern)
						}
					}
				}
			}
		}
	}

	// If we didn't get any patterns, return a default
	if len(allPatterns) == 0 {
		return []string{"No significant patterns identified in the data"}, nil
	}

	return allPatterns, nil
}

// processBatchedFindings generates findings and recommendations from the analysis results
func processBatchedFindings(apiClient *client.Client, disputeData []map[string]interface{}, conversations []map[string]interface{},
	analysis Analysis, patterns []string, batchSize int) ([]string, []string, error) {

	// If there are no disputes, return empty
	if len(disputeData) == 0 {
		return []string{"No findings available due to empty dataset"},
			[]string{"Collect more data to generate meaningful recommendations"}, nil
	}

	// Use a smaller batch size for findings as they are more complex
	findingsBatchSize := batchSize / 2
	if findingsBatchSize < 5 {
		findingsBatchSize = 5
	}

	batchCount := (len(disputeData) + findingsBatchSize - 1) / findingsBatchSize
	fmt.Printf("Processing %d disputes in %d batches for findings analysis\n", len(disputeData), batchCount)

	// Combine results from all batches
	var allFindings []string
	var allRecommendations []string

	// Process each batch
	for i := 0; i < len(disputeData); i += findingsBatchSize {
		end := i + findingsBatchSize
		if end > len(disputeData) {
			end = len(disputeData)
		}
		batch := disputeData[i:end]

		fmt.Printf("Processing findings batch %d/%d (%d disputes)...\n", (i/findingsBatchSize)+1, batchCount, len(batch))

		// Only enable debug for first batch
		if i == 0 {
			apiClient.EnableDebug()
		} else {
			apiClient.DisableDebug()
		}

		// Prepare trends and patterns to include in this request
		// We need to include the overall analysis to provide context
		trendsData := map[string]interface{}{
			"trend_descriptions":  analysis.TrendDescriptions,
			"recommended_actions": analysis.RecommendedActions,
		}

		// Create request for this batch
		req := client.StandardAnalysisRequest{
			AnalysisType: "findings",
			Parameters: map[string]interface{}{
				"questions": []string{
					"What are the key issues in fee disputes?",
					"How can customer satisfaction be improved?",
					"What are the financial implications of these disputes?",
					"How effective are current dispute resolution processes?",
				},
				"concise_response": true,
			},
			Data: map[string]interface{}{
				"attribute_values": batch,
				"conversations":    getLimitedConversations(conversations, 2),
				"trends_data":      trendsData,
				"patterns_data":    patterns,
				"metadata": map[string]interface{}{
					"avg_amount":     calculateAverageAmount(disputeData),
					"total_disputes": len(disputeData),
				},
			},
		}

		// Make API request
		resp, err := apiClient.PerformAnalysis(req)
		if err != nil {
			fmt.Printf("Error generating findings in batch %d: %v\n", (i/findingsBatchSize)+1, err)
			continue
		}

		// Extract findings from response
		if results, ok := resp.Results.(map[string]interface{}); ok {
			// Try to extract findings directly
			if findings, ok := results["findings"].([]interface{}); ok {
				for _, f := range findings {
					if finding, ok := f.(string); ok {
						allFindings = append(allFindings, finding)
					}
				}
			}

			// Try to extract recommendations directly
			if recs, ok := results["recommendations"].([]interface{}); ok {
				for _, r := range recs {
					if rec, ok := r.(string); ok {
						allRecommendations = append(allRecommendations, rec)
					}
				}
			}
		}
	}

	// If we didn't get any findings, return defaults
	if len(allFindings) == 0 {
		allFindings = []string{"No significant findings could be derived from the analysis"}
	}

	if len(allRecommendations) == 0 {
		allRecommendations = []string{"Consider more detailed analysis to generate specific recommendations"}
	}

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

// Helper function to extract arrays from map using multiple possible field names
func extractArrayFromMap(data map[string]interface{}, possibleFields []string) []interface{} {
	for _, field := range possibleFields {
		if arr, ok := data[field].([]interface{}); ok && len(arr) > 0 {
			return arr
		}
	}
	return []interface{}{}
}
