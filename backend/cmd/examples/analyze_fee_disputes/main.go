package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"agenticflows/backend/cmd/examples/client"
	"agenticflows/backend/cmd/examples/utils"

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

func main() {
	// Parse command-line flags
	dbPath := flag.String("db", "", "Path to the SQLite database")
	limit := flag.Int("limit", 10, "Number of disputes to analyze")
	debugFlag := flag.Bool("debug", false, "Enable debug mode")
	workflowID := flag.String("workflow", "", "Workflow ID for persisting results")
	flag.Parse()

	// Validate required flags
	if *dbPath == "" {
		fmt.Println("Error: --db flag is required")
		flag.Usage()
		os.Exit(1)
	}

	startTime := time.Now()

	// Create API client using the standardized client package
	apiClient := client.NewClient("http://localhost:8080", *workflowID, *debugFlag)

	// Print debug information if debug flag is enabled
	if *debugFlag {
		fmt.Println("Debug mode enabled: LLM inputs and outputs will be printed")
	}

	// Step 1: Fetch fee disputes from database
	fmt.Printf("Fetching %d fee disputes...\n", *limit)
	disputes, err := fetchFeeDisputes(*dbPath, *limit, apiClient)
	if err != nil {
		fmt.Printf("Error fetching disputes: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d disputes\n", len(disputes))

	// Step 2: Enrich disputes with sentiment analysis
	fmt.Println("\nAnalyzing sentiment for disputes...")
	for i, dispute := range disputes {
		fmt.Printf("Analyzing sentiment for dispute %s...\n", dispute.ID)

		// Use standardized API to analyze sentiment
		req := client.StandardAnalysisRequest{
			AnalysisType: "attributes",
			Text:         dispute.Text,
			Parameters: map[string]interface{}{
				"attributes": []map[string]interface{}{
					{
						"field_name":  "sentiment",
						"title":       "Sentiment",
						"description": "The overall sentiment of the text (positive, negative, neutral)",
					},
					{
						"field_name":  "sentiment_desc",
						"title":       "Sentiment Description",
						"description": "A brief explanation of why this sentiment was chosen",
					},
				},
			},
		}

		resp, err := apiClient.PerformAnalysis(req)
		if err != nil {
			fmt.Printf("Error generating attributes: %v\n", err)
			continue
		}

		if results, ok := resp.Results.(map[string]interface{}); ok {
			if attrValues, ok := results["attribute_values"].(map[string]interface{}); ok {
				disputes[i].Sentiment = utils.GetString(attrValues, "sentiment")
				disputes[i].SentimentDesc = utils.GetString(attrValues, "sentiment_desc")
			}
		}
	}

	// Step 3: Analyze trends
	fmt.Println("\nAnalyzing trends in fee disputes...")
	req := client.StandardAnalysisRequest{
		AnalysisType: "trends",
		Parameters: map[string]interface{}{
			"focus_areas": []string{
				"fee_types",
				"resolution_outcomes",
				"customer_sentiment",
				"dispute_patterns",
			},
		},
		Data: map[string]interface{}{
			"disputes": disputes,
		},
	}

	resp, err := apiClient.PerformAnalysis(req)
	if err != nil {
		fmt.Printf("Error analyzing trends: %v\n", err)
	}

	// Extract trends from response
	var analysis TrendAnalysis
	if results, ok := resp.Results.(map[string]interface{}); ok {
		if trends, ok := results["trends"].([]interface{}); ok {
			for _, trend := range trends {
				if trendMap, ok := trend.(map[string]interface{}); ok {
					focusArea := utils.GetString(trendMap, "focus_area")
					trendDesc := utils.GetString(trendMap, "trend")
					switch focusArea {
					case "fee_types":
						analysis.TrendDescription = trendDesc
					case "resolution_outcomes":
						analysis.RecommendedAction = trendDesc
					case "customer_sentiment":
						if sentiments, ok := trendMap["supporting_data"].([]interface{}); ok {
							for _, s := range sentiments {
								if str, ok := s.(string); ok {
									analysis.CommonSentiments = append(analysis.CommonSentiments, str)
								}
							}
						}
					}
				}
			}
		}
	}

	// Step 4: Identify patterns
	fmt.Println("\nIdentifying patterns in fee disputes...")
	var patterns []string
	var findings string

	// Convert disputes to format expected by API
	disputeData := make([]map[string]interface{}, len(disputes))
	for i, dispute := range disputes {
		disputeData[i] = map[string]interface{}{
			"id":          dispute.ID,
			"text":        dispute.Text,
			"amount":      dispute.Amount,
			"created_at":  dispute.CreatedAt.Format(time.RFC3339),
			"sentiment":   dispute.Sentiment,
			"description": dispute.SentimentDesc,
			"attributes": map[string]interface{}{
				"dispute_type":             "Service fee",
				"resolution_offered":       "Yes",
				"resolution_type":          "Full refund",
				"resolution_outcome":       "Resolved in customer's favor",
				"agent_explanation":        "The fee was charged due to account terms but was refunded as a courtesy",
				"de_escalation_techniques": "Empathy statements, offering solutions",
				"customer_sentiment_start": "Negative",
				"customer_sentiment_end":   "Positive",
				"call_id":                  dispute.ID,
				"call_duration":            "5 minutes",
				"agent_id":                 "EMPL_001",
				"call_timestamp":           dispute.CreatedAt.Format(time.RFC3339),
			},
		}
	}

	req = client.StandardAnalysisRequest{
		AnalysisType: "patterns",
		Parameters: map[string]interface{}{
			"pattern_types": []string{
				"fee_dispute_patterns",
				"resolution_patterns",
				"customer_behavior_patterns",
			},
		},
		Data: map[string]interface{}{
			"disputes":      disputeData,
			"conversations": disputeData,
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

	resp, err = apiClient.PerformAnalysis(req)
	if err != nil {
		fmt.Printf("Error identifying patterns: %v\n", err)
		patterns = []string{"No patterns identified due to error"}
	} else {
		// Extract patterns from response
		patterns = []string{}
		if results, ok := resp.Results.(map[string]interface{}); ok {
			if patternList, ok := results["patterns"].([]interface{}); ok {
				for _, pattern := range patternList {
					if patternMap, ok := pattern.(map[string]interface{}); ok {
						if desc, ok := patternMap["pattern_description"].(string); ok {
							patterns = append(patterns, desc)
						}
					}
				}
			}
		}
		if len(patterns) == 0 {
			patterns = []string{"No patterns identified"}
		}
	}

	// Step 5: Generate findings and recommendations
	fmt.Println("\nGenerating findings and recommendations...")
	req = client.StandardAnalysisRequest{
		AnalysisType: "findings",
		Parameters: map[string]interface{}{
			"questions": []string{
				"What are the most common types of fee disputes?",
				"What is the average resolution time?",
				"What are the most effective resolution strategies?",
				"What are the key areas for improvement?",
			},
		},
		Data: map[string]interface{}{
			"disputes":      disputeData,
			"conversations": disputeData,
			"trends":        analysis,
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

	resp, err = apiClient.PerformAnalysis(req)
	if err != nil {
		fmt.Printf("Error analyzing findings: %v\n", err)
		findings = "No findings generated due to error"
	} else {
		// Extract findings from response
		findings = "No findings available"
		if results, ok := resp.Results.(map[string]interface{}); ok {
			if answers, ok := results["answers"].([]interface{}); ok {
				var findingsText string
				for _, answer := range answers {
					if answerMap, ok := answer.(map[string]interface{}); ok {
						question := utils.GetString(answerMap, "question")
						answerText := utils.GetString(answerMap, "answer")
						confidence := utils.GetString(answerMap, "confidence")
						findingsText += fmt.Sprintf("\nQ: %s\nA: %s\nConfidence: %s\n", question, answerText, confidence)
					}
				}
				if findingsText != "" {
					findings = findingsText
				}
			}
		}
	}

	// Print results
	fmt.Println("\n=== Analysis Results ===")
	fmt.Printf("\nTotal Disputes: %d\n", analysis.TotalDisputes)
	fmt.Printf("Average Amount: $%.2f\n", analysis.AverageAmount)

	fmt.Println("\nCommon Statuses:")
	for _, status := range analysis.CommonStatuses {
		fmt.Printf("- %s\n", status)
	}

	fmt.Println("\nCommon Sentiments:")
	for _, sentiment := range analysis.CommonSentiments {
		fmt.Printf("- %s\n", sentiment)
	}

	fmt.Println("\nIdentified Patterns:")
	for _, pattern := range patterns {
		fmt.Printf("- %s\n", pattern)
	}

	fmt.Println("\nFindings and Recommendations:")
	fmt.Println(findings)

	utils.PrintTimeTaken(startTime, "Analyze fee disputes")
}

// fetchFeeDisputes fetches fee disputes from the database
func fetchFeeDisputes(dbPath string, limit int, apiClient *client.Client) ([]Dispute, error) {
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
				if attrs, ok := results["attributes"].([]interface{}); ok && len(attrs) > 0 {
					if attr, ok := attrs[0].(map[string]interface{}); ok {
						if val, ok := attr["value"].(string); ok {
							// Try to parse the amount from the value
							fmt.Sscanf(val, "$%f", &dispute.Amount)
						}
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
