package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"agenticflows/backend/cmd/examples/client"
	"agenticflows/backend/cmd/examples/utils"

	_ "github.com/mattn/go-sqlite3"
)

// IntentGroup represents a group of similar intents
type IntentGroup struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
	Count       int      `json:"count"`
}

// Custom handling for potentially truncated JSON responses
func extractPatternsFromPartialJSON(jsonStr string) ([]map[string]interface{}, error) {
	// Extract the patterns array from the JSON string
	// The patterns are in the format: "patterns": [ { ... }, { ... }, ... ]

	// Look for pattern objects
	patternRegex := regexp.MustCompile(`\{\s*"pattern_type":\s*"([^"]+)",\s*"pattern_description":\s*"([^"]+)",\s*"occurrences":\s*(\d+),\s*"examples":\s*\[(.*?)\],\s*"significance":\s*"([^"]+)"\s*\}`)

	matches := patternRegex.FindAllStringSubmatch(jsonStr, -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("could not find pattern objects in the JSON response")
	}

	// Extract the pattern objects
	patterns := make([]map[string]interface{}, 0, len(matches))

	for _, match := range matches {
		if len(match) < 6 {
			continue
		}

		// Extract the examples
		examplesStr := match[4]
		var examples []string

		// Split examples by comma and clean up
		examplesList := strings.Split(examplesStr, ",")
		for _, ex := range examplesList {
			ex = strings.TrimSpace(ex)
			ex = strings.Trim(ex, "\"")
			if ex != "" {
				examples = append(examples, ex)
			}
		}

		// Create a pattern object
		pattern := map[string]interface{}{
			"pattern_type":        match[1],
			"pattern_description": match[2],
			"occurrences":         match[3],
			"examples":            examples,
			"significance":        match[5],
		}

		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

func main() {
	// Parse command-line flags
	dbPath := flag.String("db", "", "Path to the SQLite database")
	minCount := flag.Int("min-count", 5, "Minimum count for intents to be considered")
	maxGroups := flag.Int("max-groups", 10, "Maximum number of intent groups to create")
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

	// Step 1: Fetch intents from database
	fmt.Println("Fetching intents from database...")
	intents, err := fetchIntents(*dbPath, *minCount)
	if err != nil {
		fmt.Printf("Error fetching intents: %v\n", err)
		os.Exit(1)
	}

	if len(intents) == 0 {
		fmt.Println("No intents found in database")
		os.Exit(1)
	}

	fmt.Printf("Found %d unique intents\n", len(intents))

	// Step 2: Group intents using the standardized API
	fmt.Printf("\nGrouping intents into maximum %d groups...\n", *maxGroups)

	// Convert intents to format expected by API
	intentsList := make([]map[string]interface{}, 0, len(intents))
	conversations := make([]map[string]interface{}, 0)

	// Limit the number of intents to reduce data volume
	count := 0
	maxIntents := 100 // Limit to a maximum of 100 intents

	for intent, intentCount := range intents {
		// Add intent to list
		intentsList = append(intentsList, map[string]interface{}{
			"intent": intent,
			"count":  intentCount,
			"examples": []string{
				fmt.Sprintf("Customer: Hi, I need help with %s", intent),
				fmt.Sprintf("Agent: I understand you need assistance with %s. How can I help?", intent),
			},
		})

		// Add example conversation for this intent (only for the first few)
		if count < 10 { // Limit example conversations to just 10
			conversations = append(conversations, map[string]interface{}{
				"id": fmt.Sprintf("conv-%s", intent),
				"text": fmt.Sprintf(`
Customer: Hi, I need help with %s
Agent: Hello! I understand you need assistance with %s. How can I help?
Customer: Yes, I'm having trouble with it
Agent: I'll be happy to help you with that. Could you provide more details?
Customer: [Details about %s]
Agent: Thank you for explaining. Let me help you resolve this.
`, intent, intent, intent),
				"intent": intent,
				"metadata": map[string]interface{}{
					"timestamp": time.Now().Format(time.RFC3339),
					"channel":   "chat",
				},
			})
		}

		count++
		if count >= maxIntents {
			break
		}
	}

	// Use standardized API to group intents
	req := client.StandardAnalysisRequest{
		AnalysisType: "patterns",
		Parameters: map[string]interface{}{
			"pattern_types":          []string{"intent_groups"},
			"max_groups":             *maxGroups,
			"min_count":              *minCount,
			"max_examples_per_group": 3,
			"concise_response":       true,
		},
		Data: map[string]interface{}{
			"intents":       intentsList,
			"conversations": conversations,
			"constraints": map[string]interface{}{
				"max_examples_per_group": 3,
				"max_patterns":           *maxGroups,
				"examples_format":        "compact",
			},
		},
	}

	resp, err := apiClient.PerformAnalysis(req)
	if err != nil {
		// Try to extract patterns from partial response if it's a JSON parsing error
		if strings.Contains(err.Error(), "failed to parse JSON") {
			errMsg := err.Error()
			// Extract the text from the error message
			textStart := strings.Index(errMsg, "text: ")
			if textStart != -1 {
				jsonStr := errMsg[textStart+6:]
				patterns, extractErr := extractPatternsFromPartialJSON(jsonStr)
				if extractErr == nil && len(patterns) > 0 {
					fmt.Println("\nSuccessfully extracted patterns from partial response:")
					for i, pattern := range patterns {
						fmt.Printf("\nPattern %d: %s\n", i+1, pattern["pattern_type"])
						fmt.Printf("Description: %s\n", pattern["pattern_description"])
						fmt.Printf("Examples: %v\n", pattern["examples"])
						fmt.Printf("Significance: %s\n", pattern["significance"])
					}
					return
				}
			}
		}
		log.Fatalf("Error grouping intents: %v", err)
	}

	// Extract groups from response
	var groups []IntentGroup
	if results, ok := resp.Results.(map[string]interface{}); ok {
		if patterns, ok := results["patterns"].([]interface{}); ok {
			for _, pattern := range patterns {
				if patternMap, ok := pattern.(map[string]interface{}); ok {
					group := IntentGroup{
						Name:        utils.GetString(patternMap, "pattern_type"),
						Description: utils.GetString(patternMap, "pattern_description"),
						Examples:    utils.GetStringArray(patternMap, "examples"),
						Count:       utils.GetInt(patternMap, "occurrences"),
					}
					groups = append(groups, group)
				}
			}
		}
	}

	// Print results
	fmt.Printf("\nFound %d intent groups:\n", len(groups))
	for i, group := range groups {
		fmt.Printf("\nGroup %d: %s\n", i+1, group.Name)
		fmt.Printf("Description: %s\n", group.Description)
		fmt.Printf("Count: %d\n", group.Count)
		fmt.Println("Examples:")
		for _, example := range group.Examples {
			fmt.Printf("  - %s\n", example)
		}
	}

	utils.PrintTimeTaken(startTime, "Group intents")
}

// fetchIntents fetches intents from the database with their counts
func fetchIntents(dbPath string, minCount int) (map[string]int, error) {
	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}
	defer db.Close()

	// Query for intents with their counts
	query := `
	SELECT value, COUNT(*) as count
	FROM conversation_attributes
	WHERE type = 'intent'
	GROUP BY value
	HAVING COUNT(*) >= ?
	ORDER BY count DESC
	`

	rows, err := db.Query(query, minCount)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	// Format intents as map
	intents := make(map[string]int)
	for rows.Next() {
		var intent string
		var count int
		if err := rows.Scan(&intent, &count); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		intents[intent] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return intents, nil
}
