package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
func (c *ApiClient) GenerateRequiredAttributes(questions []string, existingAttributes []map[string]string) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"questions":         questions,
		"existing_attributes": existingAttributes,
		"workflow_id":       c.workflowID,
	}
	
	// Make API request
	return c.makeRequest("required_attributes", requestBody)
}

// AnalyzeTrends analyzes trends in the provided data
func (c *ApiClient) AnalyzeTrends(data []map[string]interface{}) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"data":        data,
		"workflow_id": c.workflowID,
	}
	
	// Make API request
	return c.makeRequest("trends", requestBody)
}

// IdentifyPatterns identifies patterns in the data based on the specified types
func (c *ApiClient) IdentifyPatterns(data []map[string]interface{}, patternTypes []string) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"data":          data,
		"pattern_types": patternTypes,
		"workflow_id":   c.workflowID,
	}
	
	// Make API request
	return c.makeRequest("patterns", requestBody)
}

// AnalyzeFindings analyzes findings from analysis results
func (c *ApiClient) AnalyzeFindings(data map[string]interface{}) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"data":        data,
		"workflow_id": c.workflowID,
	}
	
	// Make API request
	return c.makeRequest("findings", requestBody)
}

// GroupIntents groups similar intents
func (c *ApiClient) GroupIntents(intents []map[string]interface{}, maxGroups int) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"intents":     intents,
		"max_groups":  maxGroups,
		"workflow_id": c.workflowID,
	}
	
	// Make API request
	return c.makeRequest("group_intents", requestBody)
}

// DescribeIntentGroup generates a description for an intent group
func (c *ApiClient) DescribeIntentGroup(groupName, examples string) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"group_name":  groupName,
		"examples":    examples,
		"workflow_id": c.workflowID,
	}
	
	// Make API request
	return c.makeRequest("describe_group", requestBody)
}

// IdentifyAttributes identifies attributes from conversations
func (c *ApiClient) IdentifyAttributes(questions []string, conversationSamples string) (map[string]interface{}, error) {
	// Prepare request body
	requestBody := map[string]interface{}{
		"questions":    questions,
		"conversations": conversationSamples,
		"workflow_id":  c.workflowID,
	}
	
	// Make API request
	return c.makeRequest("identify_attributes", requestBody)
}

// makeRequest makes a request to the API
func (c *ApiClient) makeRequest(endpoint string, requestBody map[string]interface{}) (map[string]interface{}, error) {
	// Convert request body to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}
	
	// Log request if debug is enabled
	if c.debug {
		fmt.Printf("Request to %s:\n%s\n", endpoint, string(jsonData))
	}
	
	// Prepare request
	url := fmt.Sprintf("%s/%s", c.baseURL, endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	
	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	
	// Log response if debug is enabled
	if c.debug {
		fmt.Printf("Response from %s:\n%s\n", endpoint, string(respBody))
	}
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-OK status: %d, body: %s", resp.StatusCode, string(respBody))
	}
	
	// Parse response
	var result map[string]interface{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}
	
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