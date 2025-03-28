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

// Conversation represents a conversation record from the database
type Conversation struct {
	ID        string
	Text      string
	CreatedAt time.Time
}

// Main function
func main() {
	// Parse command-line flags
	dbPath := flag.String("db", "", "Path to the SQLite database file")
	limit := flag.Int("limit", 10, "Limit number of conversations to analyze")
	focusArea := flag.String("focus", "customer_retention", "Focus area for recommendations")
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

	// Step 1: Fetch conversations
	fmt.Println("Fetching conversations from database...")
	conversations, err := fetchConversations(*dbPath, *limit)
	if err != nil {
		fmt.Printf("Error fetching conversations: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Found %d conversations\n", len(conversations))

	// Step 2: Prepare analysis data - first run trend and pattern analysis
	fmt.Println("\nAnalyzing conversation trends and patterns...")
	trends, patterns, err := analyzeConversations(apiClient, conversations, *focusArea)
	if err != nil {
		fmt.Printf("Warning: Error analyzing conversations: %v\n", err)
		fmt.Println("Continuing with limited analysis data...")
	}

	// Step 3: Generate recommendations based on analysis
	fmt.Println("\nGenerating recommendations...")
	recommendations, err := generateRecommendations(apiClient, conversations, trends, patterns, *focusArea)
	if err != nil {
		fmt.Printf("Error generating recommendations: %v\n", err)
		os.Exit(1)
	}

	// Step 4: Print results
	fmt.Println("\n=== Results ===")
	fmt.Println("\nImmediate Actions:")
	for i, action := range recommendations.ImmediateActions {
		fmt.Printf("%d. %s\n", i+1, action.Action)
		fmt.Printf("   Rationale: %s\n", action.Rationale)
		fmt.Printf("   Expected Impact: %s\n", action.ExpectedImpact)
		fmt.Printf("   Priority: %d\n\n", action.Priority)
	}

	fmt.Println("\nImplementation Notes:")
	for _, note := range recommendations.ImplementationNotes {
		fmt.Printf("- %s\n", note)
	}

	fmt.Println("\nSuccess Metrics:")
	for _, metric := range recommendations.SuccessMetrics {
		fmt.Printf("- %s\n", metric)
	}

	fmt.Println("\nRecommendation Generation complete!")
}

// TrendResult represents the trend analysis results
type TrendResult struct {
	Trends          []map[string]interface{} `json:"trends"`
	OverallInsights []string                 `json:"overall_insights"`
}

// PatternResult represents the pattern analysis results
type PatternResult struct {
	Patterns           []map[string]interface{} `json:"patterns"`
	UnexpectedPatterns []map[string]interface{} `json:"unexpected_patterns"`
}

// RecommendationAction represents a recommended action
type RecommendationAction struct {
	Action         string `json:"action"`
	Rationale      string `json:"rationale"`
	ExpectedImpact string `json:"expected_impact"`
	Priority       int    `json:"priority"`
}

// RecommendationResult represents the recommendation results
type RecommendationResult struct {
	ImmediateActions    []RecommendationAction `json:"immediate_actions"`
	ImplementationNotes []string               `json:"implementation_notes"`
	SuccessMetrics      []string               `json:"success_metrics"`
}

// fetchConversations fetches conversations from the database
func fetchConversations(dbPath string, limit int) ([]Conversation, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Query for conversations
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

	// Parse results into Conversation objects
	conversations := make([]Conversation, 0)
	for rows.Next() {
		var conv Conversation
		var createdAtStr string
		if err := rows.Scan(&conv.ID, &conv.Text, &createdAtStr); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// Parse created_at timestamp
		conv.CreatedAt, _ = time.Parse("2006-01-02T15:04:05-07:00", createdAtStr)
		conversations = append(conversations, conv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return conversations, nil
}

// analyzeConversations analyzes conversations for trends and patterns
func analyzeConversations(apiClient *client.Client, conversations []Conversation, focusArea string) (TrendResult, PatternResult, error) {
	emptyTrends := TrendResult{}
	emptyPatterns := PatternResult{}

	// Prepare conversation data
	convData := make([]map[string]interface{}, len(conversations))
	for i, conv := range conversations {
		convData[i] = map[string]interface{}{
			"id":         conv.ID,
			"text":       conv.Text,
			"created_at": conv.CreatedAt.Format(time.RFC3339),
		}
	}

	// Step 1: Analyze trends
	fmt.Println("Analyzing trends in conversations...")
	trendReq := client.StandardAnalysisRequest{
		AnalysisType: "trends",
		Parameters: map[string]interface{}{
			"focus_areas": []string{
				focusArea,
				"customer_satisfaction",
				"agent_effectiveness",
			},
		},
		Data: map[string]interface{}{
			"conversations": convData,
			"attributes": map[string]interface{}{
				"total_conversations":   len(conversations),
				"conversation_timespan": "30 days",
			},
		},
	}

	trendResp, err := apiClient.PerformAnalysis(trendReq)
	if err != nil {
		return emptyTrends, emptyPatterns, fmt.Errorf("error analyzing trends: %w", err)
	}

	// Extract trend results
	var trendResult TrendResult
	if results, ok := trendResp.Results.(map[string]interface{}); ok {
		// Extract trends
		if trends, ok := results["trends"].([]interface{}); ok {
			trendResult.Trends = make([]map[string]interface{}, 0, len(trends))
			for _, t := range trends {
				if trend, ok := t.(map[string]interface{}); ok {
					trendResult.Trends = append(trendResult.Trends, trend)
				}
			}
		}

		// Extract overall insights
		if insights, ok := results["overall_insights"].([]interface{}); ok {
			trendResult.OverallInsights = make([]string, 0, len(insights))
			for _, i := range insights {
				if insight, ok := i.(string); ok {
					trendResult.OverallInsights = append(trendResult.OverallInsights, insight)
				}
			}
		}
	}

	// Step 2: Identify patterns
	fmt.Println("Identifying patterns in conversations...")
	patternReq := client.StandardAnalysisRequest{
		AnalysisType: "patterns",
		Parameters: map[string]interface{}{
			"pattern_types": []string{
				"conversation_flow",
				"customer_behavior",
				"agent_response",
			},
		},
		Data: map[string]interface{}{
			"conversations": convData,
		},
	}

	patternResp, err := apiClient.PerformAnalysis(patternReq)
	if err != nil {
		return trendResult, emptyPatterns, fmt.Errorf("error identifying patterns: %w", err)
	}

	// Extract pattern results
	var patternResult PatternResult
	if results, ok := patternResp.Results.(map[string]interface{}); ok {
		// Extract patterns
		if patterns, ok := results["patterns"].([]interface{}); ok {
			patternResult.Patterns = make([]map[string]interface{}, 0, len(patterns))
			for _, p := range patterns {
				if pattern, ok := p.(map[string]interface{}); ok {
					patternResult.Patterns = append(patternResult.Patterns, pattern)
				}
			}
		}

		// Extract unexpected patterns
		if unexpected, ok := results["unexpected_patterns"].([]interface{}); ok {
			patternResult.UnexpectedPatterns = make([]map[string]interface{}, 0, len(unexpected))
			for _, u := range unexpected {
				if pattern, ok := u.(map[string]interface{}); ok {
					patternResult.UnexpectedPatterns = append(patternResult.UnexpectedPatterns, pattern)
				}
			}
		}
	}

	return trendResult, patternResult, nil
}

// generateRecommendations generates recommendations based on analysis
func generateRecommendations(apiClient *client.Client, conversations []Conversation, trends TrendResult, patterns PatternResult, focusArea string) (RecommendationResult, error) {
	emptyResult := RecommendationResult{}

	// Prepare data for recommendations request
	analysisData := map[string]interface{}{
		"trends":   trends,
		"patterns": patterns,
		"metrics": map[string]interface{}{
			"total_conversations": len(conversations),
			"timespan":            "30 days",
		},
	}

	// Create prioritization criteria
	criteria := map[string]interface{}{
		"impact":              0.6,
		"implementation_ease": 0.4,
	}

	// Request recommendations
	fmt.Println("Requesting recommendations from API...")
	req := client.StandardAnalysisRequest{
		AnalysisType: "recommendations",
		Parameters: map[string]interface{}{
			"focus_area": focusArea,
			"criteria":   criteria,
		},
		Data: analysisData,
	}

	resp, err := apiClient.PerformAnalysis(req)
	if err != nil {
		return emptyResult, fmt.Errorf("error generating recommendations: %w", err)
	}

	// Parse recommendations
	var result RecommendationResult
	if results, ok := resp.Results.(map[string]interface{}); ok {
		// Extract immediate actions
		if actions, ok := results["immediate_actions"].([]interface{}); ok {
			result.ImmediateActions = make([]RecommendationAction, 0, len(actions))
			for _, a := range actions {
				if actionMap, ok := a.(map[string]interface{}); ok {
					action := RecommendationAction{
						Action:         utils.GetString(actionMap, "action"),
						Rationale:      utils.GetString(actionMap, "rationale"),
						ExpectedImpact: utils.GetString(actionMap, "expected_impact"),
						Priority:       utils.GetInt(actionMap, "priority"),
					}
					result.ImmediateActions = append(result.ImmediateActions, action)
				}
			}
		}

		// Extract implementation notes
		if notes, ok := results["implementation_notes"].([]interface{}); ok {
			result.ImplementationNotes = make([]string, 0, len(notes))
			for _, n := range notes {
				if note, ok := n.(string); ok {
					result.ImplementationNotes = append(result.ImplementationNotes, note)
				}
			}
		}

		// Extract success metrics
		if metrics, ok := results["success_metrics"].([]interface{}); ok {
			result.SuccessMetrics = make([]string, 0, len(metrics))
			for _, m := range metrics {
				if metric, ok := m.(string); ok {
					result.SuccessMetrics = append(result.SuccessMetrics, metric)
				}
			}
		}
	}

	return result, nil
}
