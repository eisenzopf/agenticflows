package analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// LLMClient handles interactions with Gemini or other LLM APIs
type LLMClient struct {
	apiKey      string
	modelName   string
	maxRetries  int
	retryDelay  time.Duration
	debug       bool
	httpClient  *http.Client
	rateLimiter *RateLimiter
}

// LLMRequest represents a request to the Gemini API
type LLMRequest struct {
	Contents         []Content        `json:"contents"`
	GenerationConfig GenerationConfig `json:"generationConfig"`
}

// Content represents content in an LLM request
type Content struct {
	Parts []Part `json:"parts"`
}

// Part represents a part in an LLM request content
type Part struct {
	Text string `json:"text"`
}

// GenerationConfig contains settings for text generation
type GenerationConfig struct {
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxOutputTokens,omitempty"`
	TopP        float64 `json:"topP,omitempty"`
	TopK        int     `json:"topK,omitempty"`
}

// LLMResponse represents a response from the Gemini API
type LLMResponse struct {
	Candidates     []Candidate    `json:"candidates"`
	PromptFeedback PromptFeedback `json:"promptFeedback,omitempty"`
}

// Candidate represents a generated response candidate
type Candidate struct {
	Content      Content `json:"content"`
	FinishReason string  `json:"finishReason"`
}

// NewLLMClient creates a new LLM client
func NewLLMClient(apiKey string, debug bool) (*LLMClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	return &LLMClient{
		apiKey:      apiKey,
		modelName:   "gemini-2.0-flash",
		maxRetries:  3,
		retryDelay:  1 * time.Second,
		debug:       debug,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		rateLimiter: NewRateLimiter(1900), // 1900 requests per minute
	}, nil
}

// GenerateContent sends a text generation request to the Gemini API
func (c *LLMClient) GenerateContent(ctx context.Context, prompt string, expectedFormat interface{}) (interface{}, error) {
	var result interface{}

	// Try the request with retries
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		// Check rate limit
		if err := c.rateLimiter.Acquire(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter error: %w", err)
		}

		resp, err := c.makeRequest(ctx, prompt)
		if err != nil {
			if attempt < c.maxRetries-1 {
				// Wait before retrying
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(c.retryDelay * time.Duration(attempt+1)):
					// Continue and retry
					continue
				}
			}
			return nil, err
		}

		// Parse the response
		parsedResp, err := c.parseResponse(resp)
		if err != nil {
			if attempt < c.maxRetries-1 {
				continue
			}
			return nil, err
		}

		// Validate against expected format if provided
		if expectedFormat != nil {
			if err := c.validateResponse(parsedResp, expectedFormat); err != nil {
				if attempt < c.maxRetries-1 {
					continue
				}
				return nil, err
			}
		}

		result = parsedResp
		break
	}

	return result, nil
}

// makeRequest sends the HTTP request to the LLM API
func (c *LLMClient) makeRequest(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.modelName, c.apiKey)

	req := LLMRequest{
		Contents: []Content{
			{
				Parts: []Part{
					{
						Text: prompt,
					},
				},
			},
		},
		GenerationConfig: GenerationConfig{
			Temperature: 0.0,
			MaxTokens:   8192,
			TopP:        0.95,
			TopK:        40,
		},
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Log the request if debug is enabled
	if c.debug {
		fmt.Printf("LLM Request: %s\n", string(reqBody))
	}

	// Send the request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error: %s, body: %s", resp.Status, string(body))
	}

	// Log the response if debug is enabled
	if c.debug {
		fmt.Printf("LLM Response: %s\n", string(body))
	}

	var llmResp LLMResponse
	if err := json.Unmarshal(body, &llmResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(llmResp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	// Get the text content
	if len(llmResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no parts in candidate content")
	}

	return llmResp.Candidates[0].Content.Parts[0].Text, nil
}

// parseResponse extracts and parses the JSON content from the LLM response
func (c *LLMClient) parseResponse(responseText string) (interface{}, error) {
	// Clean the response text
	text := responseText
	if strings.HasPrefix(text, "```") {
		// Extract JSON from code blocks
		lines := strings.Split(text, "\n")
		if len(lines) > 2 {
			text = strings.Join(lines[1:len(lines)-1], "\n")
		}
		text = strings.TrimPrefix(text, "```json\n")
		text = strings.TrimSuffix(text, "```")
	}

	text = strings.TrimSpace(text)

	// Parse as generic JSON
	var result interface{}
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w, text: %s", err, text)
	}

	return result, nil
}

// validateResponse checks if the response matches the expected format
func (c *LLMClient) validateResponse(response interface{}, expectedFormat interface{}) error {
	// Simple validation for now - more complex validation can be added later
	respMap, ok := response.(map[string]interface{})
	if !ok {
		return fmt.Errorf("response is not a JSON object")
	}

	expMap, ok := expectedFormat.(map[string]interface{})
	if !ok {
		return nil // Not a map, skip validation
	}

	// Check if required fields exist
	for key := range expMap {
		if _, exists := respMap[key]; !exists {
			return fmt.Errorf("missing required field: %s", key)
		}
	}

	return nil
}

// GenerateResponses processes multiple prompts in parallel
func (c *LLMClient) GenerateResponses(ctx context.Context, prompts []string, expectedFormat interface{}) ([]interface{}, error) {
	results := make([]interface{}, len(prompts))
	errChan := make(chan error, len(prompts))

	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Process all prompts in parallel
	for i, prompt := range prompts {
		go func(idx int, p string) {
			result, err := c.GenerateContent(ctx, p, expectedFormat)
			if err != nil {
				errChan <- err
				return
			}
			results[idx] = result
			errChan <- nil
		}(i, prompt)
	}

	// Collect results
	var lastErr error
	for i := 0; i < len(prompts); i++ {
		if err := <-errChan; err != nil {
			lastErr = err
		}
	}

	return results, lastErr
}
