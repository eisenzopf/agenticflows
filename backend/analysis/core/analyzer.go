package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
)

// Analyzer provides methods for analyzing conversation data
type Analyzer struct {
	LLMClient *LLMClient
	Debug     bool
}

// NewAnalyzer creates a new Analyzer instance
func NewAnalyzer(apiKey string, debug bool) (*Analyzer, error) {
	llmClient, err := NewLLMClient(apiKey, debug)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}
	return &Analyzer{
		LLMClient: llmClient,
		Debug:     debug,
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

// TransformForTrends prepares data for trend analysis by standardizing format
func (a *Analyzer) TransformForTrends(data interface{}) (map[string]interface{}, error) {
	// Convert input data to the format expected by AnalyzeTrends
	result := make(map[string]interface{})

	// Handle different input types
	switch v := data.(type) {
	case map[string]interface{}:
		// If it's already a map, check if it has the right structure
		result = v
	case []map[string]interface{}:
		// If it's an array of maps, convert to attribute_values
		result["attribute_values"] = v
	case string:
		// If it's a string, try to parse as JSON
		if err := json.Unmarshal([]byte(v), &result); err != nil {
			// If parsing fails, treat as raw text
			result["text"] = v
		}
	default:
		// For other types, try to marshal and unmarshal
		bytes, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to transform data: %w", err)
		}
		if err := json.Unmarshal(bytes, &result); err != nil {
			return nil, fmt.Errorf("failed to transform data: %w", err)
		}
	}

	return result, nil
}

// TransformForPatterns prepares data for pattern identification
func (a *Analyzer) TransformForPatterns(data interface{}, patternTypes []string) (map[string]interface{}, error) {
	// Convert input data to the format expected by IdentifyPatterns
	result := make(map[string]interface{})

	// Handle different input types
	switch v := data.(type) {
	case map[string]interface{}:
		// If it's already a map, use it directly
		result = v
	case []map[string]interface{}:
		// If it's an array of maps, convert to attribute_values
		result["attribute_values"] = v
	default:
		// For other types, try to marshal and unmarshal
		bytes, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to transform data: %w", err)
		}
		if err := json.Unmarshal(bytes, &result); err != nil {
			return nil, fmt.Errorf("failed to transform data: %w", err)
		}
	}

	// Add pattern types if not already present
	if _, ok := result["pattern_types"]; !ok && len(patternTypes) > 0 {
		result["pattern_types"] = patternTypes
	}

	return result, nil
}

// ChainAnalysis performs a chain of analyses
func (a *Analyzer) ChainAnalysis(ctx context.Context, inputData interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	if a.Debug {
		log.Printf("Starting chain analysis with config: %+v", config)
	}

	// Extract steps from config
	stepsVal, ok := config["steps"]
	if !ok {
		return nil, fmt.Errorf("steps configuration is required for chain analysis")
	}

	stepsArray, ok := stepsVal.([]interface{})
	if !ok {
		return nil, fmt.Errorf("steps must be an array")
	}

	// Convert steps to strings
	steps := make([]string, 0, len(stepsArray))
	for _, step := range stepsArray {
		if stepStr, ok := step.(string); ok {
			steps = append(steps, stepStr)
		} else {
			return nil, fmt.Errorf("steps must be strings")
		}
	}

	if len(steps) == 0 {
		return nil, fmt.Errorf("at least one step is required")
	}

	// Initialize results with the input data
	results := make(map[string]interface{})
	currentData := inputData

	// Process each step in sequence
	for i, step := range steps {
		if a.Debug {
			log.Printf("Processing step %d: %s", i+1, step)
		}

		// Extract step-specific configuration
		stepConfig := make(map[string]interface{})
		if stepsConfigVal, ok := config["step_config"]; ok {
			if stepsConfigMap, ok := stepsConfigVal.(map[string]interface{}); ok {
				if stepConfigVal, ok := stepsConfigMap[step]; ok {
					if stepCfg, ok := stepConfigVal.(map[string]interface{}); ok {
						stepConfig = stepCfg
					}
				}
			}
		}

		// Include the current data in the step configuration
		stepConfig["input_data"] = currentData

		// Process the step
		var stepResult interface{}
		var err error

		// For actual implementation, call the appropriate analysis function here
		// This is a simplified placeholder
		stepResult = map[string]interface{}{
			"step":     step,
			"step_num": i + 1,
			"processed_data": fmt.Sprintf("Processed %s with config: %v",
				step, stepConfig),
		}

		if err != nil {
			return results, fmt.Errorf("error in step %d (%s): %w", i+1, step, err)
		}

		// Add this step's result to the results map
		results[step] = stepResult

		// Update current data for the next step
		currentData = stepResult
	}

	return results, nil
}

// Helper functions for extraction
func extractStringSlice(config map[string]interface{}, key string) ([]string, error) {
	if val, ok := config[key]; ok {
		if slice, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(slice))
			for _, item := range slice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result, nil
		}
		return nil, fmt.Errorf("%s must be an array of strings", key)
	}
	return []string{}, nil
}
