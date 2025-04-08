package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// StandardAnalysisRequest represents a request to the standardized analysis API
type StandardAnalysisRequest struct {
	// Common fields
	WorkflowID string `json:"workflow_id,omitempty"`
	Text       string `json:"text,omitempty"`

	// Analysis-specific fields
	AnalysisType string                 `json:"analysis_type"`  // "trends", "patterns", "findings", "attributes", "intent", "recommendations", "action_plan", "timeline"
	Parameters   map[string]interface{} `json:"parameters"`     // Analysis-specific parameters
	Data         map[string]interface{} `json:"data,omitempty"` // Input data for analysis

	// Note: For recommendations, action_plan, and timeline analysis types,
	// you can include "use_mock_data": true in the Parameters map to get
	// predefined mock responses instead of making actual LLM API calls.
	// This is useful for testing and demonstrations.
}

// StandardAnalysisResponse represents a response from the standardized analysis API
type StandardAnalysisResponse struct {
	AnalysisType string      `json:"analysis_type"`
	Results      interface{} `json:"results"`
	Confidence   float64     `json:"confidence,omitempty"`
	Error        string      `json:"error,omitempty"`
}

// Client represents a client for the standardized analysis API
type Client struct {
	baseURL    string
	httpClient *http.Client
	workflowID string
	debug      bool
}

// NewClient creates a new standardized API client
func NewClient(baseURL string, workflowID string, debug bool) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		workflowID: workflowID,
		debug:      debug,
	}
}

// EnableDebug enables debug mode for the client
func (c *Client) EnableDebug() {
	c.debug = true
}

// DisableDebug disables debug mode for the client
func (c *Client) DisableDebug() {
	c.debug = false
}

// PerformAnalysis performs an analysis using the standardized API
func (c *Client) PerformAnalysis(req StandardAnalysisRequest) (*StandardAnalysisResponse, error) {
	// Add workflow ID to request if provided
	requestData := map[string]interface{}{
		"workflow_id":   c.workflowID,
		"analysis_type": req.AnalysisType,
		"parameters":    req.Parameters,
	}

	if req.Text != "" {
		requestData["text"] = req.Text
	}

	// Include Data field if provided
	if req.Data != nil && len(req.Data) > 0 {
		requestData["data"] = req.Data
	}

	// Convert request to JSON
	reqBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Debug output - print full request details in debug mode
	if c.debug {
		fmt.Printf("\n=== API REQUEST ===\n")
		fmt.Printf("URL: %s/api/analysis\n", c.baseURL)
		fmt.Printf("Type: %s\n", req.AnalysisType)
		fmt.Printf("Request Payload:\n%s\n", prettyJSON(reqBody))
		fmt.Printf("==================\n\n")
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/api/analysis", c.baseURL), bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Debug output - print full response details in debug mode
	if c.debug {
		fmt.Printf("\n=== API RESPONSE ===\n")
		fmt.Printf("Status: %s\n", resp.Status)
		fmt.Printf("Response Payload:\n%s\n", prettyJSON(respBody))
		fmt.Printf("===================\n\n")
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s, body: %s", resp.Status, string(respBody))
	}

	// Parse response
	var result StandardAnalysisResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		// If standard unmarshal fails, try to parse as a general map
		var rawResponse map[string]interface{}
		if jsonErr := json.Unmarshal(respBody, &rawResponse); jsonErr != nil {
			return nil, fmt.Errorf("error parsing response: %w (raw error: %v)", err, jsonErr)
		}

		// Try to extract fields from raw response
		analysisType, _ := rawResponse["analysis_type"].(string)
		results := rawResponse["results"]
		confidence, _ := rawResponse["confidence"].(float64)
		errorMsg, _ := rawResponse["error"].(string)

		// Create a response from the raw data
		result = StandardAnalysisResponse{
			AnalysisType: analysisType,
			Results:      results,
			Confidence:   confidence,
			Error:        errorMsg,
		}
	}

	// Check for API-level errors
	if result.Error != "" {
		return nil, fmt.Errorf("API returned error: %s", result.Error)
	}

	// If the Results field is null or missing, try to establish a default result structure
	if result.Results == nil {
		switch req.AnalysisType {
		case "trends":
			result.Results = map[string]interface{}{
				"trend_descriptions":  []interface{}{},
				"recommended_actions": []interface{}{},
			}
		case "patterns":
			result.Results = map[string]interface{}{
				"patterns": []interface{}{},
			}
		case "findings":
			result.Results = map[string]interface{}{
				"findings":        []interface{}{},
				"recommendations": []interface{}{},
			}
		default:
			result.Results = map[string]interface{}{}
		}
	}

	return &result, nil
}

// GenerateIntent generates intent for the given text
func (c *Client) GenerateIntent(text string) (map[string]interface{}, error) {
	req := StandardAnalysisRequest{
		AnalysisType: "intent",
		Text:         text,
		Parameters:   map[string]interface{}{},
	}

	resp, err := c.PerformAnalysis(req)
	if err != nil {
		return nil, err
	}

	if results, ok := resp.Results.(map[string]interface{}); ok {
		return results, nil
	}

	return nil, fmt.Errorf("unexpected response format")
}

// GenerateAttributes generates attribute values for the given text
func (c *Client) GenerateAttributes(text string, attributes []map[string]string) (map[string]interface{}, error) {
	req := StandardAnalysisRequest{
		AnalysisType: "attributes",
		Text:         text,
		Parameters: map[string]interface{}{
			"attributes": attributes,
		},
	}

	resp, err := c.PerformAnalysis(req)
	if err != nil {
		return nil, err
	}

	if results, ok := resp.Results.(map[string]interface{}); ok {
		return results, nil
	}

	return nil, fmt.Errorf("unexpected response format")
}

// AnalyzeTrends analyzes trends in the provided data
func (c *Client) AnalyzeTrends(data []map[string]interface{}) (map[string]interface{}, error) {
	req := StandardAnalysisRequest{
		AnalysisType: "trends",
		Parameters: map[string]interface{}{
			"focus_areas": []string{
				"disputed_charges",
				"resolution_effectiveness",
				"customer_satisfaction",
				"dispute_reasons",
			},
		},
		Data: map[string]interface{}{
			"attribute_values": data,
		},
	}

	resp, err := c.PerformAnalysis(req)
	if err != nil {
		return nil, err
	}

	if results, ok := resp.Results.(map[string]interface{}); ok {
		return results, nil
	}

	return nil, fmt.Errorf("unexpected response format")
}

// IdentifyPatterns identifies patterns in the provided data
func (c *Client) IdentifyPatterns(data []map[string]interface{}, patternTypes []string) (map[string]interface{}, error) {
	req := StandardAnalysisRequest{
		AnalysisType: "patterns",
		Parameters: map[string]interface{}{
			"pattern_types": patternTypes,
		},
		Data: map[string]interface{}{
			"attribute_values": data,
		},
	}

	resp, err := c.PerformAnalysis(req)
	if err != nil {
		return nil, err
	}

	if results, ok := resp.Results.(map[string]interface{}); ok {
		return results, nil
	}

	return nil, fmt.Errorf("unexpected response format")
}

// AnalyzeFindings analyzes findings from the provided data
func (c *Client) AnalyzeFindings(data map[string]interface{}) (map[string]interface{}, error) {
	req := StandardAnalysisRequest{
		AnalysisType: "findings",
		Parameters: map[string]interface{}{
			"questions": []string{
				"What are the most common types of fee disputes?",
				"How effective are the current resolution approaches?",
				"What patterns exist in customer sentiment regarding fee disputes?",
				"What are the key opportunities for improving fee dispute handling?",
			},
		},
		Data: data,
	}

	resp, err := c.PerformAnalysis(req)
	if err != nil {
		return nil, err
	}

	if results, ok := resp.Results.(map[string]interface{}); ok {
		return results, nil
	}

	return nil, fmt.Errorf("unexpected response format")
}

// Example usage:
//
// func ExampleWithMockData() {
//     // Create client
//     client := NewClient("http://localhost:8080", "workflow123", true)
//
//     // Create request with mock data enabled
//     request := &StandardAnalysisRequest{
//         AnalysisType: "recommendations",
//         Parameters: map[string]interface{}{
//             "focus_area": "customer_retention",
//             "use_mock_data": true,  // This will return predefined mock data
//         },
//         Data: map[string]interface{}{
//             "conversations": []interface{}{
//                 map[string]interface{}{"id": "1", "text": "Sample conversation"},
//             },
//         },
//     }
//
//     // Perform analysis with mock data
//     response, err := client.PerformAnalysis(context.Background(), request)
//     if err != nil {
//         log.Fatalf("Error: %v", err)
//     }
//
//     // Process the mock response
//     fmt.Printf("Got %d mock recommendations\n", len(response.Results.(map[string]interface{})["recommendations"].([]interface{})))
// }

// prettyJSON formats a JSON byte array for better readability
func prettyJSON(data []byte) string {
	var out bytes.Buffer
	err := json.Indent(&out, data, "", "  ")
	if err != nil {
		return string(data) // Return raw data if prettifying fails
	}
	return out.String()
}
