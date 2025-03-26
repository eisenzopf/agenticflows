package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Dispute represents a fee dispute conversation with attributes
type Dispute struct {
	ConversationID string
	Text           string
	Intent         string
	DisputeType    string
	Resolution     string
	Amount         string
	DateOccurred   string
	CustomerSentiment  string
	JustificationRating string
}

// TrendAnalysis represents trend analysis results
type TrendAnalysis struct {
	DisputeTypeCounts       map[string]int    `json:"dispute_type_counts"`
	ResolutionCounts        map[string]int    `json:"resolution_counts"`
	AmountRanges            map[string]int    `json:"amount_ranges"`
	SentimentDistribution   map[string]int    `json:"sentiment_distribution"`
	JustificationEffectiveness map[string]int `json:"justification_effectiveness"`
	TopPatterns             []string          `json:"top_patterns"`
	OverallInsights         string            `json:"overall_insights"`
}

func main() {
	// Parse command-line arguments
	dbPath := flag.String("db", "", "Path to the SQLite database")
	outputPath := flag.String("output", "", "Path to save results as JSON (optional)")
	minCount := flag.Int("min-count", 10, "Minimum count for patterns to be considered significant")
	workflowID := flag.String("workflow", "", "Workflow ID for persisting results")
	debugFlag := flag.Bool("debug", false, "Enable debug output")
	limit := flag.Int("limit", 100, "Maximum number of fee disputes to analyze")
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

	// Step 1: Fetch fee dispute conversations
	disputes, err := fetchFeeDisputes(*dbPath, *minCount, *limit)
	if err != nil {
		fmt.Printf("Error fetching fee disputes: %v\n", err)
		os.Exit(1)
	}

	if len(disputes) == 0 {
		fmt.Println("No fee disputes found in the database")
		os.Exit(1)
	}

	fmt.Printf("Found %d fee dispute conversations\n", len(disputes))

	// Step 2: Enrich with sentiment analysis
	enrichDisputesWithSentiment(&disputes, apiClient)

	// Step 3: Analyze trends
	trendAnalysis, err := analyzeTrends(disputes, apiClient)
	if err != nil {
		fmt.Printf("Error analyzing trends: %v\n", err)
		os.Exit(1)
	}

	// Step 4: Identify key patterns
	patterns, err := identifyPatterns(disputes, apiClient)
	if err != nil {
		fmt.Printf("Error identifying patterns: %v\n", err)
		trendAnalysis.TopPatterns = []string{"Error identifying patterns"}
	} else {
		trendAnalysis.TopPatterns = patterns
	}

	// Step 5: Generate overall insights
	insights, err := analyzeFindings(trendAnalysis, apiClient)
	if err != nil {
		fmt.Printf("Error analyzing findings: %v\n", err)
		trendAnalysis.OverallInsights = "Error generating insights"
	} else {
		trendAnalysis.OverallInsights = insights
	}

	// Step 6: Print report
	printAnalysisReport(trendAnalysis)

	// Step 7: Save to output file if requested
	if *outputPath != "" {
		reportData := map[string]interface{}{
			"workflow_id":     *workflowID,
			"trend_analysis":  trendAnalysis,
			"total_disputes":  len(disputes),
			"timestamp":       time.Now().Format(time.RFC3339),
		}
		
		jsonData, err := json.MarshalIndent(reportData, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
			os.Exit(1)
		}
		
		err = os.WriteFile(*outputPath, jsonData, 0644)
		if err != nil {
			fmt.Printf("Error writing output file: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("\nAnalysis results written to %s\n", *outputPath)
	}

	PrintTimeTaken(startTime, "Analyze fee disputes")
}

// fetchFeeDisputes fetches fee dispute conversations from the database
func fetchFeeDisputes(dbPath string, minCount int, limit int) ([]Dispute, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Query for fee dispute conversations
	// This query joins conversations with their intent attributes and looks for fee-related intents
	query := `
	SELECT 
		c.conversation_id, 
		c.text, 
		ca_intent.value AS intent,
		ca_type.value AS dispute_type,
		ca_resolution.value AS resolution,
		ca_amount.value AS amount,
		ca_date.value AS date_occurred
	FROM 
		conversations c
	JOIN 
		conversation_attributes ca_intent 
		ON c.conversation_id = ca_intent.conversation_id AND ca_intent.type = 'intent'
	LEFT JOIN 
		conversation_attributes ca_type 
		ON c.conversation_id = ca_type.conversation_id AND ca_type.name = 'dispute_type'
	LEFT JOIN 
		conversation_attributes ca_resolution 
		ON c.conversation_id = ca_resolution.conversation_id AND ca_resolution.name = 'resolution'
	LEFT JOIN 
		conversation_attributes ca_amount 
		ON c.conversation_id = ca_amount.conversation_id AND ca_amount.name = 'disputed_amount'
	LEFT JOIN 
		conversation_attributes ca_date 
		ON c.conversation_id = ca_date.conversation_id AND ca_date.name = 'date_occurred'
	WHERE 
		(ca_intent.value LIKE '%fee%' OR ca_intent.value LIKE '%charge%' OR ca_intent.value LIKE '%billing%' OR ca_intent.value LIKE '%payment%')
		AND ca_intent.value LIKE '%dispute%'
		AND c.text IS NOT NULL 
		AND LENGTH(c.text) > 100
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	// Format disputes as objects
	disputes := make([]Dispute, 0)
	count := 0
	for rows.Next() {
		var dispute Dispute
		var intentNullable, disputeTypeNullable, resolutionNullable, amountNullable, dateNullable sql.NullString
		
		if err := rows.Scan(
			&dispute.ConversationID, 
			&dispute.Text, 
			&intentNullable,
			&disputeTypeNullable,
			&resolutionNullable,
			&amountNullable,
			&dateNullable,
		); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		
		// Handle nullable values
		if intentNullable.Valid {
			dispute.Intent = intentNullable.String
		}
		if disputeTypeNullable.Valid {
			dispute.DisputeType = disputeTypeNullable.String
		}
		if resolutionNullable.Valid {
			dispute.Resolution = resolutionNullable.String
		}
		if amountNullable.Valid {
			dispute.Amount = amountNullable.String
		}
		if dateNullable.Valid {
			dispute.DateOccurred = dateNullable.String
		}
		
		disputes = append(disputes, dispute)
		
		// Respect the limit exactly as specified
		count++
		if limit > 0 && count >= limit {
			break
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return disputes, nil
}

// enrichDisputesWithSentiment enriches disputes with sentiment analysis
func enrichDisputesWithSentiment(disputes *[]Dispute, apiClient *ApiClient) {
	fmt.Println("Enriching disputes with sentiment analysis...")
	
	for i := range *disputes {
		dispute := &(*disputes)[i]
		
		// Skip if text is empty
		if dispute.Text == "" {
			continue
		}
		
		// Define sentiment attributes to generate
		attributes := []map[string]string{
			{
				"field_name":  "customer_sentiment",
				"title":       "Customer Sentiment",
				"description": "The sentiment of the customer in the conversation, ranging from very negative to very positive.",
				"type":        "enum",
				"enum_values": "very negative, negative, neutral, positive, very positive",
			},
			{
				"field_name":  "justification_rating",
				"title":       "Agent Justification Rating",
				"description": "How well the agent justified the disputed fee in their explanation to the customer.",
				"type":        "enum",
				"enum_values": "poor, fair, good, excellent",
			},
		}
		
		// Call API to generate attribute values
		fmt.Printf("Analyzing sentiment for conversation %s...\n", dispute.ConversationID)
		resp, err := apiClient.GenerateAttributes(dispute.Text, attributes)
		if err != nil {
			fmt.Printf("Error generating attributes: %v\n", err)
			continue
		}
		
		// Extract attribute values from response
		if attrValues, ok := resp["attribute_values"].([]interface{}); ok {
			for _, av := range attrValues {
				if attrValue, ok := av.(map[string]interface{}); ok {
					fieldName, ok := attrValue["field_name"].(string)
					if !ok {
						continue
					}
					
					value, ok := attrValue["value"].(string)
					if !ok {
						continue
					}
					
					// Set attribute value in dispute
					switch fieldName {
					case "customer_sentiment":
						dispute.CustomerSentiment = value
					case "justification_rating":
						dispute.JustificationRating = value
					}
				}
			}
		}
	}
}

// analyzeTrends analyzes trends in fee disputes
func analyzeTrends(disputes []Dispute, apiClient *ApiClient) (TrendAnalysis, error) {
	fmt.Println("Analyzing trends in fee disputes...")
	
	// Initialize trend analysis
	analysis := TrendAnalysis{
		DisputeTypeCounts:       make(map[string]int),
		ResolutionCounts:        make(map[string]int),
		AmountRanges:            make(map[string]int),
		SentimentDistribution:   make(map[string]int),
		JustificationEffectiveness: make(map[string]int),
	}
	
	// Count dispute types
	for _, dispute := range disputes {
		if dispute.DisputeType != "" {
			analysis.DisputeTypeCounts[dispute.DisputeType]++
		}
		
		if dispute.Resolution != "" {
			analysis.ResolutionCounts[dispute.Resolution]++
		}
		
		// Categorize amounts into ranges
		if dispute.Amount != "" {
			// This is simplified; in a real implementation we would parse the amount
			// and categorize it into ranges like "Less than $10", "$10-$50", etc.
			analysis.AmountRanges[dispute.Amount]++
		}
		
		if dispute.CustomerSentiment != "" {
			analysis.SentimentDistribution[dispute.CustomerSentiment]++
		}
		
		if dispute.JustificationRating != "" {
			analysis.JustificationEffectiveness[dispute.JustificationRating]++
		}
	}
	
	// Call API for trend analysis if available
	if apiClient != nil {
		// Convert disputes to format expected by API
		apiData := make([]map[string]interface{}, len(disputes))
		for i, dispute := range disputes {
			apiData[i] = map[string]interface{}{
				"conversation_id":      dispute.ConversationID,
				"intent":               dispute.Intent,
				"dispute_type":         dispute.DisputeType,
				"resolution":           dispute.Resolution,
				"amount":               dispute.Amount,
				"customer_sentiment":   dispute.CustomerSentiment,
				"justification_rating": dispute.JustificationRating,
			}
		}
		
		// Call API to analyze trends
		_, err := apiClient.AnalyzeTrends(apiData)
		if err != nil {
			fmt.Printf("Error analyzing trends via API: %v\n", err)
		} else {
			// We could update our analysis with additional insights from the API response
			// But we'll keep our simplified version for now
		}
	}
	
	return analysis, nil
}

// identifyPatterns identifies patterns in fee disputes
func identifyPatterns(disputes []Dispute, apiClient *ApiClient) ([]string, error) {
	fmt.Println("Identifying patterns in fee disputes...")
	
	// Call API to identify patterns
	patternTypes := []string{
		"Customer escalation triggers",
		"Most effective explanation approaches",
		"Common misunderstandings",
		"Resolution strategies that work best",
	}
	
	// Convert disputes to format expected by API
	apiData := make([]map[string]interface{}, len(disputes))
	for i, dispute := range disputes {
		apiData[i] = map[string]interface{}{
			"conversation_id":      dispute.ConversationID,
			"text":                 dispute.Text,
			"intent":               dispute.Intent,
			"dispute_type":         dispute.DisputeType,
			"resolution":           dispute.Resolution,
			"amount":               dispute.Amount,
			"customer_sentiment":   dispute.CustomerSentiment,
			"justification_rating": dispute.JustificationRating,
		}
	}
	
	resp, err := apiClient.IdentifyPatterns(apiData, patternTypes)
	if err != nil {
		return nil, fmt.Errorf("error identifying patterns: %w", err)
	}
	
	// Extract patterns from response
	var patterns []string
	if patternsArr, ok := resp["patterns"].([]interface{}); ok {
		for _, p := range patternsArr {
			if pattern, ok := p.(string); ok {
				patterns = append(patterns, pattern)
			} else if patternMap, ok := p.(map[string]interface{}); ok {
				if patternText, ok := patternMap["pattern"].(string); ok {
					patterns = append(patterns, patternText)
				}
			}
		}
	}
	
	return patterns, nil
}

// analyzeFindings analyzes findings from trend analysis
func analyzeFindings(analysis TrendAnalysis, apiClient *ApiClient) (string, error) {
	fmt.Println("Analyzing findings...")
	
	// Call API to analyze findings
	analysisData := map[string]interface{}{
		"dispute_type_counts":         analysis.DisputeTypeCounts,
		"resolution_counts":           analysis.ResolutionCounts,
		"amount_ranges":               analysis.AmountRanges,
		"sentiment_distribution":      analysis.SentimentDistribution,
		"justification_effectiveness": analysis.JustificationEffectiveness,
		"top_patterns":                analysis.TopPatterns,
	}
	
	resp, err := apiClient.AnalyzeFindings(analysisData)
	if err != nil {
		return "", fmt.Errorf("error analyzing findings: %w", err)
	}
	
	// Extract insights from response
	insights, ok := resp["insights"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format")
	}
	
	return insights, nil
}

// printAnalysisReport prints the analysis report
func printAnalysisReport(analysis TrendAnalysis) {
	fmt.Println("\n===== Fee Dispute Analysis Report =====")
	
	// Print dispute type counts
	fmt.Println("\nDispute Types:")
	printSortedMap(analysis.DisputeTypeCounts)
	
	// Print resolution counts
	fmt.Println("\nResolutions:")
	printSortedMap(analysis.ResolutionCounts)
	
	// Print amount ranges
	fmt.Println("\nDisputed Amounts:")
	printSortedMap(analysis.AmountRanges)
	
	// Print sentiment distribution
	fmt.Println("\nCustomer Sentiment:")
	printSortedMap(analysis.SentimentDistribution)
	
	// Print justification effectiveness
	fmt.Println("\nAgent Justification Rating:")
	printSortedMap(analysis.JustificationEffectiveness)
	
	// Print top patterns
	fmt.Println("\nTop Patterns Identified:")
	for i, pattern := range analysis.TopPatterns {
		fmt.Printf("%d. %s\n", i+1, pattern)
	}
	
	// Print overall insights
	fmt.Println("\nOverall Insights:")
	fmt.Println(analysis.OverallInsights)
}

// printSortedMap prints a map sorted by keys
func printSortedMap(m map[string]int) {
	// Get keys
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	
	// Sort keys
	sort.Strings(keys)
	
	// Print map
	for _, k := range keys {
		fmt.Printf("  %s: %d\n", k, m[k])
	}
} 