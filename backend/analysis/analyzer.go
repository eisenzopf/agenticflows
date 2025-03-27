package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
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
		"trends":           []interface{}{},
		"overall_insights": []interface{}{},
		"data_quality":     map[string]interface{}{},
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

	// Check if this is a request for intent groups
	isIntentGroupsRequest := false
	for _, patternType := range req.PatternTypes {
		if patternType == "intent_groups" {
			isIntentGroupsRequest = true
			break
		}
	}

	// Special handling for intent_groups pattern type
	if isIntentGroupsRequest && req.AttributeValues != nil {
		// Extract intents from the attribute values
		intents, ok := req.AttributeValues["intents"]
		if !ok {
			return nil, fmt.Errorf("'intents' field is required in attribute_values for intent_groups pattern type")
		}

		// Extract max_groups from the request if present
		maxGroups := 20 // Default value
		if maxGroupsVal, ok := req.AttributeValues["max_groups"]; ok {
			if maxGroupsInt, ok := maxGroupsVal.(float64); ok {
				maxGroups = int(maxGroupsInt)
			} else if maxGroupsInt, ok := maxGroupsVal.(int); ok {
				maxGroups = maxGroupsInt
			}
		}

		// Extract min_count from the request if present
		minCount := 5 // Default value
		if minCountVal, ok := req.AttributeValues["min_count"]; ok {
			if minCountInt, ok := minCountVal.(float64); ok {
				minCount = int(minCountInt)
			} else if minCountInt, ok := minCountVal.(int); ok {
				minCount = minCountInt
			}
		}

		// Process intents in batches and consolidate them iteratively
		result, err := a.processIntentsIteratively(ctx, intents, maxGroups, minCount)
		if err != nil {
			return nil, fmt.Errorf("failed to process intents iteratively: %w", err)
		}

		return &AnalysisResponse{
			Results:    result,
			Confidence: 0.8,
		}, nil
	}

	// Default pattern identification prompt (for non-intent_groups)
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
		"patterns":            []interface{}{},
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
		"answers":   []interface{}{},
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

// processIntentsIteratively processes intents in batches and consolidates them iteratively
// This is similar to the Python implementation's process_labels_iteratively function
func (a *Analyzer) processIntentsIteratively(
	ctx context.Context,
	intents interface{},
	maxGroups int,
	minCount int,
) (interface{}, error) {
	// Convert intents to a list of maps
	intentsList, ok := intents.([]interface{})
	if !ok {
		return nil, fmt.Errorf("intents must be an array")
	}

	// Filter intents based on min_count
	filteredIntents := make([]map[string]interface{}, 0)
	for _, intentObj := range intentsList {
		intent, ok := intentObj.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if count meets minimum requirement
		countVal, ok := intent["count"].(float64)
		if !ok {
			// Try to get count as int
			countIntVal, ok := intent["count"].(int)
			if !ok {
				continue
			}
			countVal = float64(countIntVal)
		}

		if int(countVal) >= minCount {
			filteredIntents = append(filteredIntents, intent)
		}
	}

	if len(filteredIntents) == 0 {
		return map[string]interface{}{
			"patterns":            []interface{}{},
			"unexpected_patterns": []interface{}{},
		}, nil
	}

	if a.debug {
		log.Printf("Processing %d intents (after filtering by min_count=%d)", len(filteredIntents), minCount)
	}

	// Determine batch size based on number of intents
	batchSize := 50
	if len(filteredIntents) <= 50 {
		batchSize = len(filteredIntents)
	}

	// Process in batches
	// For initial implementation, we'll do a simpler approach than the Python version
	// We'll just split the intents into batches and process each batch to get groups
	// Then we'll consolidate the groups into a final set

	// Step 1: Split intents into batches and process each batch
	var batches [][]map[string]interface{}
	for i := 0; i < len(filteredIntents); i += batchSize {
		end := i + batchSize
		if end > len(filteredIntents) {
			end = len(filteredIntents)
		}
		batches = append(batches, filteredIntents[i:end])
	}

	// Step 2: Process each batch to get initial groups
	batchResults := make([]map[string]interface{}, 0)
	for i, batch := range batches {
		if a.debug {
			log.Printf("Processing batch %d/%d with %d intents", i+1, len(batches), len(batch))
		}

		// Process this batch
		result, err := a.processIntentsBatch(ctx, batch, maxGroups/len(batches))
		if err != nil {
			log.Printf("Error processing batch %d: %v", i+1, err)
			continue
		}

		patterns, ok := result["patterns"].([]interface{})
		if !ok || len(patterns) == 0 {
			continue
		}

		// Add each pattern from this batch to batchResults
		for _, pattern := range patterns {
			patternMap, ok := pattern.(map[string]interface{})
			if !ok {
				continue
			}
			batchResults = append(batchResults, patternMap)
		}
	}

	// If we only have one batch or didn't get enough patterns, return the results directly
	if len(batches) == 1 || len(batchResults) <= maxGroups {
		return map[string]interface{}{
			"patterns":            batchResults[:min(len(batchResults), maxGroups)],
			"unexpected_patterns": []interface{}{},
		}, nil
	}

	// Step 3: Consolidate the groups from all batches into final groups
	finalGroups, err := a.consolidateIntentGroups(ctx, batchResults, maxGroups)
	if err != nil {
		return nil, fmt.Errorf("failed to consolidate intent groups: %w", err)
	}

	return map[string]interface{}{
		"patterns":            finalGroups,
		"unexpected_patterns": []interface{}{},
	}, nil
}

// processIntentsBatch processes a batch of intents and returns the groups
func (a *Analyzer) processIntentsBatch(
	ctx context.Context,
	intents []map[string]interface{},
	maxGroupsPerBatch int,
) (map[string]interface{}, error) {
	// Build a prompt for grouping this batch of intents
	intentsList, err := json.Marshal(intents)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal intents: %w", err)
	}

	prompt := fmt.Sprintf(`Group the following intents into semantic categories:

Intents:
%s

Your task is to group these intents into at most %d semantic categories based on their meaning and purpose.
For each group:
1. Assign a descriptive category name
2. Include relevant examples from the input list
3. Provide a brief description of the group

Format your response as JSON with these fields:
{
  "patterns": [
    {
      "pattern_type": str,        // This should be the category/group name
      "pattern_description": str,  // Description of what this group represents
      "occurrences": int,         // How many intents belong to this group
      "examples": [str],          // List of example intents in this group (limit to 5-7 examples)
      "significance": str         // Brief explanation of why this grouping is meaningful
    }
  ],
  "unexpected_patterns": []
}`, string(intentsList), maxGroupsPerBatch)

	expectedFormat := map[string]interface{}{
		"patterns": []interface{}{
			map[string]interface{}{
				"pattern_type":        "",
				"pattern_description": "",
				"occurrences":         0,
				"examples":            []interface{}{},
				"significance":        "",
			},
		},
		"unexpected_patterns": []interface{}{},
	}

	result, err := a.llmClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content for intent groups: %w", err)
	}

	return result.(map[string]interface{}), nil
}

// consolidateIntentGroups consolidates groups from multiple batches into a final set of groups
func (a *Analyzer) consolidateIntentGroups(
	ctx context.Context,
	groups []map[string]interface{},
	maxGroups int,
) ([]map[string]interface{}, error) {
	// If we already have fewer groups than the max, return them directly
	if len(groups) <= maxGroups {
		return groups, nil
	}

	// Create a description of each group
	groupDescriptions := make([]string, 0, len(groups))
	for _, group := range groups {
		patternType, _ := group["pattern_type"].(string)
		patternDesc, _ := group["pattern_description"].(string)
		examples, _ := group["examples"].([]interface{})

		// Format examples as strings
		examplesStr := ""
		if len(examples) > 0 {
			exampleTexts := make([]string, 0, len(examples))
			for _, ex := range examples {
				if exStr, ok := ex.(string); ok {
					exampleTexts = append(exampleTexts, exStr)
				}
			}
			if len(exampleTexts) > 0 {
				examplesStr = fmt.Sprintf(" Examples: %s", strings.Join(exampleTexts, ", "))
			}
		}

		groupDesc := fmt.Sprintf("%s: %s.%s", patternType, patternDesc, examplesStr)
		groupDescriptions = append(groupDescriptions, groupDesc)
	}

	// Build a prompt to consolidate the groups
	prompt := fmt.Sprintf(`You are a label clustering expert. Your task is to consolidate similar intent groups into higher-level categories.

INPUT GROUPS TO CONSOLIDATE:
%s

Rules:
1. Group similar intent categories together under a common, higher-level category
2. Maintain semantic meaning
3. Use consistent labeling style (Title Case)
4. Maximum number of consolidated groups: %d

Format your response as JSON with these fields:
{
  "consolidated_groups": [
    {
      "pattern_type": str,        // The higher-level category name
      "pattern_description": str,  // Description of what this group represents
      "occurrences": int,         // How many original groups belong to this category
      "examples": [str],          // List of example original groups in this category
      "significance": str         // Brief explanation of why this grouping is meaningful
    }
  ]
}`, strings.Join(groupDescriptions, "\n"), maxGroups)

	expectedFormat := map[string]interface{}{
		"consolidated_groups": []interface{}{
			map[string]interface{}{
				"pattern_type":        "",
				"pattern_description": "",
				"occurrences":         0,
				"examples":            []interface{}{},
				"significance":        "",
			},
		},
	}

	result, err := a.llmClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to consolidate groups: %w", err)
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	consolidatedGroups, ok := resultMap["consolidated_groups"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("consolidated_groups field is missing or not an array")
	}

	// Convert to the same format as the original groups
	finalGroups := make([]map[string]interface{}, 0, len(consolidatedGroups))
	for _, group := range consolidatedGroups {
		groupMap, ok := group.(map[string]interface{})
		if !ok {
			continue
		}
		finalGroups = append(finalGroups, groupMap)
	}

	return finalGroups, nil
}
