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
	// Extract pattern objects from the JSON string
	var patterns []map[string]interface{}

	// Simple extraction of intent groups - less strict regex to match the pattern structure we're seeing
	patternTypeRegex := regexp.MustCompile(`"pattern_type":\s*"([^"]+)"`)
	patternDescRegex := regexp.MustCompile(`"pattern_description":\s*"([^"]+)"`)
	occurrencesRegex := regexp.MustCompile(`"occurrences":\s*(\d+)`)
	examplesRegex := regexp.MustCompile(`"examples":\s*\[(.*?)\]`)
	significanceRegex := regexp.MustCompile(`"significance":\s*"([^"]+)"`)

	// Find all pattern objects by splitting on the pattern delimiter
	patternDelimiter := "\"pattern_type\":"
	patternBlocks := strings.Split(jsonStr, patternDelimiter)

	// Skip the first element which is just the JSON opening
	for i := 1; i < len(patternBlocks); i++ {
		patternBlock := patternDelimiter + patternBlocks[i]

		// Extract pattern type
		typeMatch := patternTypeRegex.FindStringSubmatch(patternBlock)
		if len(typeMatch) < 2 {
			continue
		}
		patternType := typeMatch[1]

		// Extract pattern description
		descMatch := patternDescRegex.FindStringSubmatch(patternBlock)
		var patternDescription string
		if len(descMatch) >= 2 {
			patternDescription = descMatch[1]
		} else {
			patternDescription = "Description not available"
		}

		// Extract occurrences
		occMatch := occurrencesRegex.FindStringSubmatch(patternBlock)
		var occurrences string
		if len(occMatch) >= 2 {
			occurrences = occMatch[1]
		} else {
			occurrences = "0"
		}

		// Extract examples
		examplesMatch := examplesRegex.FindStringSubmatch(patternBlock)
		var examples []string
		if len(examplesMatch) >= 2 {
			examplesStr := examplesMatch[1]
			exampleMatches := regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(examplesStr, -1)
			for _, example := range exampleMatches {
				if len(example) > 1 {
					examples = append(examples, example[1])
				}
			}
		}

		// Extract significance
		sigMatch := significanceRegex.FindStringSubmatch(patternBlock)
		var significance string
		if len(sigMatch) >= 2 {
			significance = sigMatch[1]
		} else {
			significance = "Significance not available"
		}

		// Create a pattern object with the extracted data
		pattern := map[string]interface{}{
			"pattern_type":        patternType,
			"pattern_description": patternDescription,
			"occurrences":         occurrences,
			"examples":            examples,
			"significance":        significance,
		}

		patterns = append(patterns, pattern)
	}

	if len(patterns) == 0 {
		return nil, fmt.Errorf("could not extract pattern objects from JSON")
	}

	return patterns, nil
}

// Process intents in batches and consolidate results
func processBatchedIntents(apiClient *client.Client, intentsList []map[string]interface{}, conversations []map[string]interface{}, maxGroups int, minCount int, debugFlag bool) ([]IntentGroup, error) {
	// Configuration
	batchSize := 15 // Small batch size to avoid token limits

	// Split intents into batches
	var batches [][]map[string]interface{}
	for i := 0; i < len(intentsList); i += batchSize {
		end := i + batchSize
		if end > len(intentsList) {
			end = len(intentsList)
		}
		batches = append(batches, intentsList[i:end])
	}

	fmt.Printf("Processing %d intents in %d batches (batch size: %d)\n", len(intentsList), len(batches), batchSize)

	// Process each batch
	allGroups := []map[string]interface{}{}
	for i, batch := range batches {
		fmt.Printf("Processing batch %d/%d (%d intents)...\n", i+1, len(batches), len(batch))

		// Process this batch
		req := client.StandardAnalysisRequest{
			AnalysisType: "patterns",
			Parameters: map[string]interface{}{
				"pattern_types":          []string{"intent_groups"},
				"max_groups":             3, // Smaller number of groups per batch
				"min_count":              minCount,
				"max_examples_per_group": 3,
				"concise_response":       true,
			},
			Data: map[string]interface{}{
				"intents":       batch,
				"conversations": conversations[:min(len(conversations), 3)], // Limit example conversations
				"constraints": map[string]interface{}{
					"max_examples_per_group": 3,
					"max_patterns":           3, // Smaller number of groups per batch
					"examples_format":        "compact",
				},
			},
		}

		resp, err := apiClient.PerformAnalysis(req)
		if err != nil {
			fmt.Printf("Error processing batch %d: %v\n", i+1, err)
			continue
		}

		// Extract groups from response
		if results, ok := resp.Results.(map[string]interface{}); ok {
			if patterns, ok := results["patterns"].([]interface{}); ok {
				for _, pattern := range patterns {
					if patternMap, ok := pattern.(map[string]interface{}); ok {
						allGroups = append(allGroups, patternMap)
					}
				}
			}
		}
	}

	if len(allGroups) == 0 {
		return nil, fmt.Errorf("no groups found in any batch")
	}

	// If we have too many groups, just take the top maxGroups
	// Instead of using the LLM for consolidation, which is giving inconsistent results
	if len(allGroups) > maxGroups {
		fmt.Printf("Taking top %d groups from %d available groups...\n", maxGroups, len(allGroups))

		// Sort groups by occurrences (if available) or use the order they were found
		// For simplicity in this example, we'll just take the first maxGroups
		if len(allGroups) > maxGroups {
			allGroups = allGroups[:maxGroups]
		}
	}

	// Convert to IntentGroup format
	var groups []IntentGroup
	for _, patternMap := range allGroups {
		group := IntentGroup{
			Name:        utils.GetString(patternMap, "pattern_type"),
			Description: utils.GetString(patternMap, "pattern_description"),
			Examples:    utils.GetStringArray(patternMap, "examples"),
			Count:       utils.GetInt(patternMap, "occurrences"),
		}
		groups = append(groups, group)
	}

	return groups, nil
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
	maxIntents := 100 // Increase to 100 intents

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
		if count < 5 { // Keep at 5 example conversations to avoid token limits
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

	// Use our batched processing function instead of a single API call
	groups, err := processBatchedIntents(apiClient, intentsList, conversations, *maxGroups, *minCount, *debugFlag)
	if err != nil {
		log.Fatalf("Error grouping intents: %v", err)
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
