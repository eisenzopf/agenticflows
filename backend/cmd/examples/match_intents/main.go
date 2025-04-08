package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"agenticflows/backend/cmd/examples/client"
	"agenticflows/backend/cmd/examples/utils"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Parse command-line flags
	dbPath := flag.String("db", "", "Path to the SQLite database")
	intents := flag.String("intents", "fee dispute", "Comma-separated list of intents to match")
	limit := flag.Int("limit", 10, "Number of conversations to analyze")
	threshold := flag.Float64("threshold", 0.7, "Confidence threshold for intent matching")
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

	// Step 1: Parse intents list
	intentList := strings.Split(*intents, ",")
	fmt.Printf("Matching conversations against %d intents: %s\n", len(intentList), strings.Join(intentList, ", "))

	// Step 2: Fetch sample conversations from database
	fmt.Printf("Fetching %d sample conversations...\n", *limit)
	conversations, err := fetchSampleConversations(*dbPath, *limit)
	if err != nil {
		fmt.Printf("Error fetching conversations: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d conversations\n", len(conversations))

	// Step 3: Match intents for each conversation
	fmt.Println("\nMatching intents for conversations...")
	results := make([]map[string]interface{}, 0)

	for _, conv := range conversations {
		fmt.Printf("\nAnalyzing conversation %s...\n", conv.ID)

		// Use standardized API to match intents
		req := client.StandardAnalysisRequest{
			AnalysisType: "intent",
			Text:         conv.Text,
			Parameters: map[string]interface{}{
				"target_intents": intentList,
				"threshold":      *threshold,
			},
		}

		resp, err := apiClient.PerformAnalysis(req)
		if err != nil {
			fmt.Printf("Error matching intents: %v\n", err)
			continue
		}

		// Extract intent from response
		if intentResults, ok := resp.Results.(map[string]interface{}); ok {
			result := map[string]interface{}{
				"conversation_id": conv.ID,
				"intent":          intentResults["label_name"],
				"intent_label":    intentResults["label"],
				"confidence":      resp.Confidence,
				"explanation":     intentResults["description"],
			}
			results = append(results, result)
		}
	}

	// Print results
	fmt.Println("\n=== Intent Matching Results ===")
	for _, result := range results {
		fmt.Printf("\nConversation ID: %s\n", result["conversation_id"])
		fmt.Printf("Intent: %s\n", result["intent"])
		fmt.Printf("Intent Label: %s\n", result["intent_label"])
		fmt.Printf("Confidence: %.2f\n", result["confidence"])
		fmt.Printf("Explanation: %s\n", result["explanation"])
	}

	utils.PrintTimeTaken(startTime, "Match intents")
}

// fetchSampleConversations fetches random sample conversations
func fetchSampleConversations(dbPath string, limit int) ([]utils.Conversation, error) {
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
	conversations := make([]utils.Conversation, 0)
	for rows.Next() {
		var conv utils.Conversation
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
