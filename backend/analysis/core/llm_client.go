package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// LLMClient provides methods for generating text using a language model
type LLMClient struct {
	apiKey    string
	debug     bool
	modelName string
}

// NewLLMClient creates a new LLMClient instance
func NewLLMClient(apiKey string, debug bool) (*LLMClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	return &LLMClient{
		apiKey:    apiKey,
		debug:     debug,
		modelName: "gemini-pro", // Default model
	}, nil
}

// GenerateContent generates content using the language model
func (c *LLMClient) GenerateContent(ctx context.Context, prompt string, expectedFormat interface{}) (interface{}, error) {
	// Log prompt in debug mode
	if c.debug {
		log.Printf("LLM Prompt: %s", prompt)
	}

	// In a real implementation, this would call the LLM API
	// For now, we'll just return a mock response that matches the expected format

	// Parse the expected format to determine what to return
	var result interface{}

	// Check if expectedFormat is nil
	if expectedFormat == nil {
		// Return the prompt as is if no format is expected
		return prompt, nil
	}

	// Use the expectedFormat to guide the response structure
	switch format := expectedFormat.(type) {
	case map[string]interface{}:
		// If we expect a map, create a default map with empty values for each key
		resultMap := make(map[string]interface{})
		for k, v := range format {
			resultMap[k] = v // Use the provided default values
		}
		result = resultMap
	case []interface{}:
		// If we expect an array, return an empty array
		result = format
	case string:
		// If we expect a string, return a mock string
		result = "Generated content based on: " + format
	default:
		// For other types, try to encode and decode to get the structure
		encoded, err := json.Marshal(expectedFormat)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal expected format: %w", err)
		}

		err = json.Unmarshal(encoded, &result)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal expected format: %w", err)
		}
	}

	// Log the result in debug mode
	if c.debug {
		resultJSON, _ := json.MarshalIndent(result, "", "  ")
		log.Printf("LLM Response: %s", string(resultJSON))
	}

	return result, nil
}

// SummarizeText summarizes text using the language model
func (c *LLMClient) SummarizeText(ctx context.Context, text string, maxLength int) (string, error) {
	if text == "" {
		return "", fmt.Errorf("text is required")
	}

	// Truncate text if too long for the prompt
	maxInputLen := 10000
	if len(text) > maxInputLen {
		text = text[:maxInputLen] + "..."
	}

	prompt := fmt.Sprintf(`Summarize the following text in %d words or less:

%s

Provide only the summary, without any introductory text or explanations.`, maxLength, text)

	result, err := c.GenerateContent(ctx, prompt, "")
	if err != nil {
		return "", fmt.Errorf("failed to generate summary: %w", err)
	}

	summary, ok := result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected result type: %T", result)
	}

	return strings.TrimSpace(summary), nil
}

// ExtractKeypoints extracts key points from text
func (c *LLMClient) ExtractKeypoints(ctx context.Context, text string, maxPoints int) ([]string, error) {
	if text == "" {
		return nil, fmt.Errorf("text is required")
	}

	// Truncate text if too long for the prompt
	maxInputLen := 10000
	if len(text) > maxInputLen {
		text = text[:maxInputLen] + "..."
	}

	prompt := fmt.Sprintf(`Extract up to %d key points from the following text:

%s

Format your response as a JSON array of strings, with each string representing one key point.
Each key point should be concise (1-2 sentences) and capture an important idea from the text.`, maxPoints, text)

	expectedFormat := []string{}
	result, err := c.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to extract key points: %w", err)
	}

	// Convert result to []string
	points := make([]string, 0)
	resultArray, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	for _, point := range resultArray {
		if pointStr, ok := point.(string); ok {
			points = append(points, pointStr)
		}
	}

	return points, nil
}

// AnalyzeText performs a basic analysis of text
func (c *LLMClient) AnalyzeText(ctx context.Context, text string, analysisType string) (map[string]interface{}, error) {
	if text == "" {
		return nil, fmt.Errorf("text is required")
	}

	// Validate analysis type
	validTypes := map[string]bool{
		"sentiment":   true,
		"topics":      true,
		"entities":    true,
		"intent":      true,
		"complexity":  true,
		"readability": true,
	}

	if !validTypes[analysisType] {
		return nil, fmt.Errorf("invalid analysis type: %s", analysisType)
	}

	// Truncate text if too long for the prompt
	maxInputLen := 10000
	if len(text) > maxInputLen {
		text = text[:maxInputLen] + "..."
	}

	prompt := fmt.Sprintf(`Analyze the following text for %s:

%s

Format your response as a JSON object with relevant fields for %s analysis.`, analysisType, text, analysisType)

	expectedFormat := map[string]interface{}{}
	result, err := c.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze text: %w", err)
	}

	// Convert result to map[string]interface{}
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}

	return resultMap, nil
}
