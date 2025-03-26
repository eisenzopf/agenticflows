package analysis

import (
	"context"
	"encoding/json"
	"fmt"
)

// Analyzer provides methods for analyzing conversation data
type Analyzer struct {
	llmClient *LLMClient
	debug     bool
}

// NewAnalyzer creates a new Analyzer instance
func NewAnalyzer(apiKey string, debug bool) (*Analyzer, error) {
	llmClient, err := NewLLMClient(apiKey, debug)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}
	return &Analyzer{
		llmClient: llmClient,
		debug:     debug,
	}, nil
}

// AnalyzeTrends analyzes trends in conversation data for specified focus areas
func (a *Analyzer) AnalyzeTrends(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	// Validate request
	if len(req.FocusAreas) == 0 {
		return nil, fmt.Errorf("focus areas are required")
	}

	focusAreasStr, err := json.Marshal(req.FocusAreas)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal focus areas: %w", err)
	}

	// Format data for the prompt
	dataStr := "No data provided"
	if req.AttributeValues != nil {
		dataBytes, err := json.Marshal(req.AttributeValues)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal attribute values: %w", err)
		}
		dataStr = string(dataBytes)
	}

	prompt := fmt.Sprintf(`Analyze trends in the following conversation data for these focus areas:

Focus Areas:
%s

Data:
%s

Identify notable trends, patterns, and insights related to the specified focus areas.
Format your response as JSON with these fields:
{
  "trends": [
    {
      "focus_area": str,
      "trend": str,
      "supporting_data": str,
      "confidence": float
    }
  ],
  "overall_insights": [str],
  "data_quality": {
    "assessment": str,
    "limitations": [str]
  }
}`, string(focusAreasStr), dataStr)

	expectedFormat := map[string]interface{}{
		"trends": []interface{}{},
		"overall_insights": []interface{}{},
		"data_quality": map[string]interface{}{},
	}

	result, err := a.llmClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return &AnalysisResponse{
		Results:    result,
		Confidence: 0.8, // Default confidence
	}, nil
}

// IdentifyPatterns identifies specific patterns in conversation data
func (a *Analyzer) IdentifyPatterns(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	// Validate request
	if len(req.PatternTypes) == 0 {
		return nil, fmt.Errorf("pattern types are required")
	}

	patternTypesStr, err := json.Marshal(req.PatternTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pattern types: %w", err)
	}

	// Format data for the prompt
	dataStr := "No data provided"
	if req.AttributeValues != nil {
		dataBytes, err := json.Marshal(req.AttributeValues)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal attribute values: %w", err)
		}
		dataStr = string(dataBytes)
	}

	prompt := fmt.Sprintf(`Identify patterns in the following conversation data for these pattern types:

Pattern Types:
%s

Data:
%s

Identify specific patterns in the conversation data related to the specified pattern types.
Format your response as JSON with these fields:
{
  "patterns": [
    {
      "pattern_type": str,
      "pattern_description": str,
      "occurrences": int,
      "examples": [str],
      "significance": str
    }
  ],
  "unexpected_patterns": [
    {
      "description": str,
      "potential_causes": [str]
    }
  ]
}`, string(patternTypesStr), dataStr)

	expectedFormat := map[string]interface{}{
		"patterns": []interface{}{},
		"unexpected_patterns": []interface{}{},
	}

	result, err := a.llmClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return &AnalysisResponse{
		Results:    result,
		Confidence: 0.8, // Default confidence
	}, nil
}

// AnalyzeFindings analyzes findings from attribute extraction
func (a *Analyzer) AnalyzeFindings(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	// Validate request
	if len(req.Questions) == 0 {
		return nil, fmt.Errorf("questions are required")
	}
	if req.AttributeValues == nil {
		return nil, fmt.Errorf("attribute values are required")
	}

	// Format questions for the prompt
	questionsStr := ""
	for i, q := range req.Questions {
		questionsStr += fmt.Sprintf("%d. %s\n", i+1, q)
	}

	// Format attribute values for the prompt
	attributeValuesBytes, err := json.Marshal(req.AttributeValues)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal attribute values: %w", err)
	}

	prompt := fmt.Sprintf(`Based on the analysis of customer service conversations, help answer these questions:

Questions:
%s

Analysis Data:
%s

Please provide:
1. Specific answers to each question, citing the data
2. Key metrics (1-2 words or numbers) that quantify the answer when applicable
3. Confidence level (High/Medium/Low) for each answer
4. Identification of any data gaps

Format as JSON:
{
  "answers": [
    {
      "question": str,
      "answer": str,
      "key_metrics": [str],
      "confidence": str,
      "supporting_data": str
    }
  ],
  "data_gaps": [str]
}`, questionsStr, string(attributeValuesBytes))

	expectedFormat := map[string]interface{}{
		"answers": []interface{}{},
		"data_gaps": []interface{}{},
	}

	result, err := a.llmClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return &AnalysisResponse{
		Results:    result,
		Confidence: 0.9, // Higher confidence for this analysis
	}, nil
}

// ProcessInBatches processes items in batches with parallelism
func (a *Analyzer) ProcessInBatches(ctx context.Context, items []interface{}, batchSize int, processFunc func(interface{}) (interface{}, error)) ([]interface{}, error) {
	if len(items) == 0 {
		return []interface{}{}, nil
	}

	if batchSize <= 0 {
		batchSize = 10 // Default batch size
	}

	results := make([]interface{}, len(items))
	errChan := make(chan error, len(items))
	
	// Create a new context that can be cancelled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Process items in batches
	for i := 0; i < len(items); i += batchSize {
		// Calculate end index for this batch
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}

		// Process this batch
		for j := i; j < end; j++ {
			go func(idx int, item interface{}) {
				result, err := processFunc(item)
				if err != nil {
					errChan <- err
					return
				}
				results[idx] = result
				errChan <- nil
			}(j, items[j])
		}

		// Wait for this batch to complete
		for j := i; j < end; j++ {
			if err := <-errChan; err != nil {
				return nil, err
			}
		}
	}

	return results, nil
} 