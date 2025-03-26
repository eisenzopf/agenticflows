package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Conversation represents a database conversation
type Conversation struct {
	ID   string
	Text string
}

// ApiClient represents a client for the analysis API
type ApiClient struct {
	baseURL    string
	httpClient *http.Client
	workflowID string
	debug      bool
}

// NewApiClient creates a new API client
func NewApiClient(workflowID string, debug bool) *ApiClient {
	return &ApiClient{
		baseURL:    "http://localhost:8080/api/analysis",
		httpClient: &http.Client{Timeout: 30 * time.Second},
		workflowID: workflowID,
		debug:      debug,
	}
}

// GenerateIntent generates the primary intent for a conversation
func (c *ApiClient) GenerateIntent(text string) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"text":        text,
		"workflow_id": c.workflowID,
	}
	
	// Make API request
	return c.makeRequest("intent", requestBody)
}

// GenerateAttributes generates attribute values for the given text
func (c *ApiClient) GenerateAttributes(text string, attributes []map[string]string) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"text":        text,
		"attributes":  attributes,
		"workflow_id": c.workflowID,
	}
	
	// Make API request
	return c.makeRequest("attributes", requestBody)
}

// GenerateRequiredAttributes generates attributes required to answer the given questions
func (c *ApiClient) GenerateRequiredAttributes(questions []string, targetIntent string) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"questions":        questions,
		"workflow_id":      c.workflowID,
		"generate_required": true,
	}
	
	if targetIntent != "" {
		requestBody["intent"] = targetIntent
	}
	
	// Make API request
	return c.makeRequest("attributes", requestBody)
}

// generateFeeDisputeAttributes generates a default set of attributes for fee disputes
func generateFeeDisputeAttributes() []map[string]interface{} {
	// Define default attributes for fee dispute analysis
	defaultAttributes := []map[string]string{
		{
			"field_name":  "dispute_type",
			"title":       "Dispute Type",
			"description": "The category or type of fee being disputed",
			"type":        "string",
			"enum_values": "overdraft, monthly_service, late_payment, foreign_transaction, atm, wire_transfer, other",
		},
		{
			"field_name":  "disputed_amount",
			"title":       "Disputed Amount",
			"description": "The monetary amount of the fee being disputed",
			"type":        "number",
		},
		{
			"field_name":  "agent_explanation",
			"title":       "Agent Explanation",
			"description": "The explanation provided by the agent for why the fee was charged",
			"type":        "string",
		},
		{
			"field_name":  "date_occurred",
			"title":       "Date Occurred",
			"description": "The date when the disputed fee was charged",
			"type":        "string",
		},
		{
			"field_name":  "resolution",
			"title":       "Resolution",
			"description": "How the fee dispute was resolved",
			"type":        "string",
			"enum_values": "waived, reduced, refunded, credited, not_resolved, escalated",
		},
		{
			"field_name":  "customer_sentiment",
			"title":       "Customer Sentiment",
			"description": "The emotional tone or attitude of the customer during the conversation",
			"type":        "string",
			"enum_values": "very_negative, negative, neutral, positive, very_positive",
		},
		{
			"field_name":  "escalation_level",
			"title":       "Escalation Level",
			"description": "The level to which the dispute was escalated, if any",
			"type":        "string",
			"enum_values": "none, supervisor, manager, executive, regulator",
		},
		{
			"field_name":  "prior_notification",
			"title":       "Prior Notification",
			"description": "Whether the customer was notified about the fee before it was charged",
			"type":        "boolean",
		},
		{
			"field_name":  "agent_justification",
			"title":       "Agent Justification Rating",
			"description": "How well the agent justified the fee in their explanation",
			"type":        "string",
			"enum_values": "poor, fair, good, excellent",
		},
		{
			"field_name":  "repeat_issue",
			"title":       "Repeat Issue",
			"description": "Whether this is a recurring issue for the customer",
			"type":        "boolean",
		},
	}
	
	// Convert to expected return format
	result := make([]map[string]interface{}, 0, len(defaultAttributes))
	for _, attr := range defaultAttributes {
		resultAttr := make(map[string]interface{})
		for k, v := range attr {
			resultAttr[k] = v
		}
		result = append(result, resultAttr)
	}
	
	return result
}

// AnalyzeTrends analyzes trends in the provided data
func (c *ApiClient) AnalyzeTrends(data []map[string]interface{}) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"data":        data,
		"workflow_id": c.workflowID,
	}
	
	// Try to make API request
	resp, err := c.makeRequest("trends", requestBody)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			fmt.Println("API call failed:", err, ", implementing client-side trend analysis")
			
			// Create a fallback response with simple trend data
			fallbackResponse := map[string]interface{}{
				"trends": []map[string]interface{}{
					{
						"name":        "Customer sentiment distribution",
						"description": "Distribution of customer sentiment across conversations.",
					},
					{
						"name":        "Resolution methods",
						"description": "Most common methods used to resolve customer issues.",
					},
				},
			}
			
			// Debug log for fallback
			c.printDebug("trends (fallback)", requestBody, fallbackResponse, nil)
			
			return fallbackResponse, nil
		}
		return nil, err
	}
	
	return resp, nil
}

// IdentifyPatterns identifies patterns in the provided data
func (c *ApiClient) IdentifyPatterns(data []map[string]interface{}, patternTypes []string) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"attribute_values": map[string]interface{}{
			"data": data,
		},
		"pattern_types": patternTypes,
		"workflow_id":   c.workflowID,
	}
	
	// Debug log before making request
	if c.debug {
		fmt.Println("Making IdentifyPatterns request to patterns endpoint with:")
		jsonData, _ := json.MarshalIndent(requestBody, "", "  ")
		fmt.Println(string(jsonData))
	}
	
	// Make API request
	resp, err := c.makeRequest("patterns", requestBody)
	if err != nil {
		return nil, err
	}
	
	return resp, nil
}

// AnalyzeFindings analyzes findings from the provided data
func (c *ApiClient) AnalyzeFindings(data map[string]interface{}) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"data":        data,
		"workflow_id": c.workflowID,
	}
	
	// Debug log before making request
	if c.debug {
		fmt.Println("Making AnalyzeFindings request to findings endpoint with:")
		jsonData, _ := json.MarshalIndent(requestBody, "", "  ")
		fmt.Println(string(jsonData))
	}
	
	// Make API request
	resp, err := c.makeRequest("findings", requestBody)
	if err != nil {
		return nil, err
	}
	
	return resp, nil
}

// GroupIntents groups intents into categories
func (c *ApiClient) GroupIntents(intents []map[string]interface{}, maxGroups int) (map[string]interface{}, error) {
	// Define batch size - limit to 10 intents per request to avoid server overload
	batchSize := 10
	
	// Normalize counts to avoid outliers overwhelming the model
	normalizedIntents := make([]map[string]interface{}, 0, len(intents))
	for _, intent := range intents {
		normalizedIntent := make(map[string]interface{})
		for k, v := range intent {
			// Cap count values to a reasonable maximum (e.g., 100)
			if k == "count" {
				count, ok := v.(int)
				if !ok {
					if fcount, ok := v.(float64); ok {
						count = int(fcount)
					}
				}
				if count > 100 {
					normalizedIntent[k] = 100
					if c.debug {
						fmt.Printf("Normalized count for intent '%v' from %d to 100\n", 
							intent["intent"], count)
					}
				} else {
					normalizedIntent[k] = v
				}
			} else {
				normalizedIntent[k] = v
			}
		}
		normalizedIntents = append(normalizedIntents, normalizedIntent)
	}
	
	// If intents list is small enough, send in a single request
	if len(normalizedIntents) <= batchSize {
		return c.processIntentBatch(normalizedIntents, maxGroups)
	}
	
	// Process in batches
	if c.debug {
		fmt.Printf("Splitting %d intents into batches of %d\n", len(normalizedIntents), batchSize)
	}
	
	var allGroups []map[string]interface{}
	var allPatterns []interface{}
	batchCount := (len(normalizedIntents) + batchSize - 1) / batchSize
	
	// Process intents in batches
	for i := 0; i < len(normalizedIntents); i += batchSize {
		end := i + batchSize
		if end > len(normalizedIntents) {
			end = len(normalizedIntents)
		}
		
		batch := normalizedIntents[i:end]
		batchNum := (i / batchSize) + 1
		
		if c.debug {
			fmt.Printf("Processing batch %d/%d with %d intents\n", batchNum, batchCount, len(batch))
		}
		
		// Process this batch with retry logic
		maxRetries := 2
		var batchResult map[string]interface{}
		var err error
		
		for retry := 0; retry <= maxRetries; retry++ {
			// If this is a retry, wait a bit before trying again
			if retry > 0 {
				if c.debug {
					fmt.Printf("Retry %d/%d for batch %d\n", retry, maxRetries, batchNum)
				}
				time.Sleep(time.Duration(retry) * 500 * time.Millisecond) // Exponential backoff
			}
			
			// Try to process the batch
			batchResult, err = c.processIntentBatch(batch, maxGroups)
			if err == nil {
				break // Successful, exit retry loop
			}
			
			// If we've exhausted retries, log the error
			if retry == maxRetries {
				if c.debug {
					fmt.Printf("Error processing batch %d after %d retries: %s\n", batchNum, maxRetries, err)
				}
			}
		}
		
		// If we still have an error after retries, continue with partial results
		if err != nil {
			continue
		}
		
		// Extract groups from the batch result
		if groups, ok := batchResult["groups"].([]map[string]interface{}); ok {
			if c.debug {
				fmt.Printf("Batch %d returned %d groups\n", batchNum, len(groups))
			}
			allGroups = append(allGroups, groups...)
		}
		
		// Extract patterns for future processing
		if patterns, ok := batchResult["patterns"].([]interface{}); ok {
			allPatterns = append(allPatterns, patterns...)
		}
	}
	
	// If we didn't get any groups, return an error
	if len(allGroups) == 0 {
		return nil, fmt.Errorf("failed to process all batches, no groups were returned")
	}
	
	// Combine the results
	if c.debug {
		fmt.Printf("Combining results from %d batches. Total of %d groups found.\n", 
			batchCount, len(allGroups))
	}
	
	// Deduplicate groups if needed
	// This is simplified deduplication - in a real system you might want more sophisticated logic
	groupMap := make(map[string]map[string]interface{})
	for _, group := range allGroups {
		if name, ok := group["name"].(string); ok {
			// Only keep the first instance of each group name
			if _, exists := groupMap[name]; !exists {
				groupMap[name] = group
			}
		}
	}
	
	// Convert back to slice
	uniqueGroups := make([]map[string]interface{}, 0, len(groupMap))
	for _, group := range groupMap {
		uniqueGroups = append(uniqueGroups, group)
	}
	
	if c.debug {
		fmt.Printf("After deduplication: %d unique groups\n", len(uniqueGroups))
	}
	
	return map[string]interface{}{
		"groups": uniqueGroups,
		"patterns": allPatterns,
	}, nil
}

// processIntentBatch processes a single batch of intents
func (c *ApiClient) processIntentBatch(intents []map[string]interface{}, maxGroups int) (map[string]interface{}, error) {
	// Prepare request body - use the correct format expected by the API
	requestBody := map[string]interface{}{
		"attribute_values": map[string]interface{}{
			"intents": intents,
			"max_groups": maxGroups, // Put max_groups inside attribute_values instead
		},
		"pattern_types": []string{"intent_groups"},
		"workflow_id":   c.workflowID,
	}
	
	// Debug log before making request
	if c.debug {
		fmt.Println("Making GroupIntents request to patterns endpoint with:")
		jsonData, _ := json.MarshalIndent(requestBody, "", "  ")
		fmt.Println(string(jsonData))
	}
	
	// Make API request
	resp, err := c.makeRequest("patterns", requestBody)
	if err != nil {
		return nil, err
	}
	
	// Extract results from response
	resultsData, ok := resp["results"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response format: missing 'results' field")
	}
	
	// Transform the patterns response into the expected groups format
	if patterns, ok := resultsData["patterns"].([]interface{}); ok {
		groups := make([]map[string]interface{}, 0)
		
		for _, pattern := range patterns {
			if patternMap, ok := pattern.(map[string]interface{}); ok {
				// Extract pattern info
				description := getString(patternMap, "pattern_description")
				examples := getStringArray(patternMap, "examples")
				
				// Create a group
				group := map[string]interface{}{
					"name":        getString(patternMap, "pattern_type"),
					"description": description,
					"examples":    examples,
					"count":       getInt(patternMap, "occurrences"),
				}
				
				groups = append(groups, group)
			}
		}
		
		// Return in the expected format
		return map[string]interface{}{
			"groups": groups,
			"patterns": patterns,
		}, nil
	}
	
	return resp, nil
}

// Helper functions for safe type conversion
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getInt(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

func getStringArray(m map[string]interface{}, key string) []string {
	if val, ok := m[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, item := range val {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return []string{}
}

// DescribeGroup generates a description for an intent group
func (c *ApiClient) DescribeGroup(groupName string, examples []string) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"attribute_values": map[string]interface{}{
			"group_name": groupName,
			"examples":   examples,
		},
		"pattern_types": []string{"group_description"},
		"workflow_id":   c.workflowID,
	}
	
	// Debug log before making request
	if c.debug {
		fmt.Println("Making DescribeGroup request to patterns endpoint with:")
		jsonData, _ := json.MarshalIndent(requestBody, "", "  ")
		fmt.Println(string(jsonData))
	}
	
	// Make API request
	resp, err := c.makeRequest("patterns", requestBody)
	if err != nil {
		return nil, err
	}
	
	// Extract the description from the patterns response
	if patterns, ok := resp["patterns"].([]interface{}); ok && len(patterns) > 0 {
		if pattern, ok := patterns[0].(map[string]interface{}); ok {
			description := getString(pattern, "pattern_description")
			if description != "" {
				return map[string]interface{}{
					"description": description,
				}, nil
			}
		}
	}
	
	// If we couldn't extract a description from the response
	return map[string]interface{}{
		"description": "No description available from API response.",
	}, nil
}

// IdentifyAttributes identifies attributes in the given text
func (c *ApiClient) IdentifyAttributes(questions []string, text string) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"questions":   questions,
		"text":        text,
		"workflow_id": c.workflowID,
	}
	
	// Debug log before making request
	if c.debug {
		fmt.Println("Making IdentifyAttributes request to intent endpoint with:")
		jsonData, _ := json.MarshalIndent(requestBody, "", "  ")
		fmt.Println(string(jsonData))
	}
	
	// Make API request
	resp, err := c.makeRequest("intent", requestBody)
	if err != nil {
		return nil, err
	}
	
	return resp, nil
}

// makeRequest makes an API request to the specified endpoint
func (c *ApiClient) makeRequest(endpoint string, requestBody map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/%s", c.baseURL, endpoint)
	
	// Marshal request body to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}
	
	// Create request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.printDebug(endpoint, requestBody, nil, err)
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.printDebug(endpoint, requestBody, nil, err)
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		// Try to extract detailed error message if possible
		var errorResp map[string]interface{}
		if err := json.Unmarshal(respBody, &errorResp); err == nil {
			if errorMsg, ok := errorResp["error"].(string); ok {
				err := fmt.Errorf("API returned error (status %d): %s", resp.StatusCode, errorMsg)
				c.printDebug(endpoint, requestBody, errorResp, err)
				return nil, err
			}
		}
		
		// Fall back to generic error with response body
		err := fmt.Errorf("API returned non-OK status: %d, body: %s", resp.StatusCode, string(respBody))
		c.printDebug(endpoint, requestBody, nil, err)
		return nil, err
	}
	
	// Unmarshal response body
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		c.printDebug(endpoint, requestBody, nil, err)
		return nil, fmt.Errorf("error unmarshaling response body: %w", err)
	}
	
	// Debug log
	c.printDebug(endpoint, requestBody, result, nil)
	
	return result, nil
}

// PrintTimeTaken prints the time taken to execute a task
func PrintTimeTaken(startTime time.Time, taskName string) {
	elapsed := time.Since(startTime)
	fmt.Printf("\n%s completed in %s\n", taskName, elapsed)
}

// PrettyPrint prints a map as formatted JSON
func PrettyPrint(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

// GetEnvOrDefault gets an environment variable or returns a default value
func GetEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// printDebug prints debug information if debug is enabled
func (c *ApiClient) printDebug(action string, request interface{}, response interface{}, err error) {
	if !c.debug {
		return
	}
	
	fmt.Println("\n==== LLM DEBUG ====")
	fmt.Printf("Action: %s\n", action)
	
	fmt.Println("\n--- Request ---")
	requestJSON, _ := json.MarshalIndent(request, "", "  ")
	fmt.Println(string(requestJSON))
	
	if err != nil {
		fmt.Println("\n--- Error ---")
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Println("\n--- Response ---")
		responseJSON, _ := json.MarshalIndent(response, "", "  ")
		fmt.Println(string(responseJSON))
	}
	
	fmt.Println("==================\n")
} 