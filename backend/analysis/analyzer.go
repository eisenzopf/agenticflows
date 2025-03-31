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

// ExtractTrendsOutput extracts the most relevant information from trends analysis
func (a *Analyzer) ExtractTrendsOutput(resp *AnalysisResponse) (map[string]interface{}, error) {
	if resp == nil || resp.Results == nil {
		return nil, fmt.Errorf("no results to extract")
	}

	// Extract results as map
	resultsMap, ok := resp.Results.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected results format")
	}

	// Extract trends and overall insights
	output := make(map[string]interface{})

	// Extract trends
	if trends, ok := resultsMap["trends"].([]interface{}); ok {
		trendDescriptions := make([]string, 0)

		for _, t := range trends {
			if trend, ok := t.(map[string]interface{}); ok {
				// Extract trend description
				trendStr, _ := trend["trend"].(string)
				if trendStr != "" {
					trendDescriptions = append(trendDescriptions, trendStr)
				}
			}
		}

		output["trend_descriptions"] = trendDescriptions
	}

	// Extract overall insights
	if insights, ok := resultsMap["overall_insights"].([]interface{}); ok {
		insightsList := make([]string, 0)

		for _, i := range insights {
			if insight, ok := i.(string); ok && insight != "" {
				insightsList = append(insightsList, insight)
			}
		}

		output["recommended_actions"] = insightsList
	}

	// Include confidence from the response
	output["confidence"] = resp.Confidence

	return output, nil
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

// ExtractPatternsOutput extracts and simplifies patterns from the analysis
func (a *Analyzer) ExtractPatternsOutput(resp *AnalysisResponse) ([]string, error) {
	if resp == nil || resp.Results == nil {
		return nil, fmt.Errorf("no results to extract")
	}

	// Extract results as map
	resultsMap, ok := resp.Results.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected results format")
	}

	// Extract patterns
	patterns := make([]string, 0)

	if patternList, ok := resultsMap["patterns"].([]interface{}); ok {
		for _, p := range patternList {
			if pattern, ok := p.(map[string]interface{}); ok {
				// Extract pattern description
				if desc, ok := pattern["pattern_description"].(string); ok && desc != "" {
					patterns = append(patterns, desc)
				}
			}
		}
	}

	return patterns, nil
}

// TransformForFindings prepares data for findings analysis
func (a *Analyzer) TransformForFindings(data interface{}, questions []string, trendsData map[string]interface{}, patternsData []string) (map[string]interface{}, error) {
	// Convert input data to the format expected by AnalyzeFindings
	result := make(map[string]interface{})

	// Handle different input types for the base data
	switch v := data.(type) {
	case map[string]interface{}:
		// If it's already a map, use it directly
		result = v
	case []map[string]interface{}:
		// If it's an array of maps, add as disputes
		result["disputes"] = v
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

	// Add questions if provided
	if len(questions) > 0 {
		result["questions"] = questions
	}

	// Add trends data if provided
	if trendsData != nil && len(trendsData) > 0 {
		result["trends"] = trendsData
	}

	// Add patterns data if provided
	if len(patternsData) > 0 {
		result["patterns"] = patternsData
	}

	return result, nil
}

// ExtractFindingsOutput extracts findings and recommendations
func (a *Analyzer) ExtractFindingsOutput(resp *AnalysisResponse) ([]string, []string, error) {
	if resp == nil || resp.Results == nil {
		return nil, nil, fmt.Errorf("no results to extract")
	}

	// Extract results as map
	resultsMap, ok := resp.Results.(map[string]interface{})
	if !ok {
		return nil, nil, fmt.Errorf("unexpected results format")
	}

	// Extract findings and recommendations
	findings := make([]string, 0)
	recommendations := make([]string, 0)

	// Try to extract from "answers" field (standard format)
	if answers, ok := resultsMap["answers"].([]interface{}); ok {
		for _, a := range answers {
			if answer, ok := a.(map[string]interface{}); ok {
				if answerText, ok := answer["answer"].(string); ok && answerText != "" {
					findings = append(findings, answerText)
				}
			}
		}
	}

	// Try to extract from "findings" field (alternate format)
	if findingsData, ok := resultsMap["findings"].([]interface{}); ok {
		for _, f := range findingsData {
			if finding, ok := f.(string); ok && finding != "" {
				findings = append(findings, finding)
			}
		}
	}

	// Extract recommendations
	if recsData, ok := resultsMap["recommendations"].([]interface{}); ok {
		for _, r := range recsData {
			if rec, ok := r.(string); ok && rec != "" {
				recommendations = append(recommendations, rec)
			}
		}
	}

	return findings, recommendations, nil
}

// ChainAnalysis performs a complete analysis chain combining multiple analysis steps
func (a *Analyzer) ChainAnalysis(ctx context.Context, inputData interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Extract configuration options
	focusAreas, _ := extractStringSlice(config, "focus_areas")
	patternTypes, _ := extractStringSlice(config, "pattern_types")
	questions, _ := extractStringSlice(config, "questions")
	useAttributes, _ := config["use_attributes"].(bool)

	// Step 1: If attribute extraction is enabled and we have text input
	if useAttributes {
		if textInput, ok := inputData.(string); ok && textInput != "" {
			// Create attribute extraction request
			attributesReq := AnalysisRequest{
				Text: textInput,
			}

			// Extract attributes
			attributesResp, err := a.ExtractAttributes(ctx, attributesReq)
			if err != nil {
				return nil, fmt.Errorf("attribute extraction failed: %w", err)
			}

			// Use attributes result as input to next step
			inputData = attributesResp.Results
			result["attributes"] = attributesResp.Results
		}
	}

	// Step 2: Trends Analysis
	if len(focusAreas) > 0 {
		// Transform data for trends analysis
		trendsInput, err := a.TransformForTrends(inputData)
		if err != nil {
			return nil, fmt.Errorf("data transformation for trends failed: %w", err)
		}

		// Create trends request
		trendsReq := AnalysisRequest{
			FocusAreas:      focusAreas,
			AttributeValues: trendsInput,
		}

		// Perform trends analysis
		trendsResp, err := a.AnalyzeTrends(ctx, trendsReq)
		if err != nil {
			return nil, fmt.Errorf("trends analysis failed: %w", err)
		}

		// Extract and format trends output
		trendsOutput, err := a.ExtractTrendsOutput(trendsResp)
		if err != nil {
			return nil, fmt.Errorf("extracting trends output failed: %w", err)
		}

		result["trends"] = trendsOutput
	}

	// Step 3: Pattern Analysis
	if len(patternTypes) > 0 {
		// Transform data for pattern analysis
		patternsInput, err := a.TransformForPatterns(inputData, patternTypes)
		if err != nil {
			return nil, fmt.Errorf("data transformation for patterns failed: %w", err)
		}

		// Create patterns request
		patternsReq := AnalysisRequest{
			PatternTypes:    patternTypes,
			AttributeValues: patternsInput,
		}

		// Perform pattern analysis
		patternsResp, err := a.IdentifyPatterns(ctx, patternsReq)
		if err != nil {
			return nil, fmt.Errorf("pattern analysis failed: %w", err)
		}

		// Extract and format patterns output
		patternsOutput, err := a.ExtractPatternsOutput(patternsResp)
		if err != nil {
			return nil, fmt.Errorf("extracting patterns output failed: %w", err)
		}

		result["patterns"] = patternsOutput
	}

	// Step 4: Findings Analysis
	if len(questions) > 0 {
		// Get trends data if available
		var trendsData map[string]interface{}
		if trends, ok := result["trends"].(map[string]interface{}); ok {
			trendsData = trends
		}

		// Get patterns data if available
		var patternsData []string
		if patterns, ok := result["patterns"].([]string); ok {
			patternsData = patterns
		}

		// Transform data for findings analysis
		findingsInput, err := a.TransformForFindings(inputData, questions, trendsData, patternsData)
		if err != nil {
			return nil, fmt.Errorf("data transformation for findings failed: %w", err)
		}

		// Create findings request
		findingsReq := AnalysisRequest{
			Questions:       questions,
			AttributeValues: findingsInput,
		}

		// Perform findings analysis
		findingsResp, err := a.AnalyzeFindings(ctx, findingsReq)
		if err != nil {
			return nil, fmt.Errorf("findings analysis failed: %w", err)
		}

		// Extract and format findings output
		findingsOutput, recommendationsOutput, err := a.ExtractFindingsOutput(findingsResp)
		if err != nil {
			return nil, fmt.Errorf("extracting findings output failed: %w", err)
		}

		result["findings"] = findingsOutput
		result["recommendations"] = recommendationsOutput
	}

	return result, nil
}

// Helper function to extract string slice from config map
func extractStringSlice(config map[string]interface{}, key string) ([]string, error) {
	if val, ok := config[key]; ok {
		switch v := val.(type) {
		case []string:
			return v, nil
		case []interface{}:
			result := make([]string, 0, len(v))
			for _, item := range v {
				if s, ok := item.(string); ok {
					result = append(result, s)
				}
			}
			return result, nil
		default:
			return nil, fmt.Errorf("invalid type for %s", key)
		}
	}
	return []string{}, nil
}

// ExtractAttributes is a placeholder for attribute extraction functionality
// In a complete implementation, this would extract attributes from text
func (a *Analyzer) ExtractAttributes(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	// This is a placeholder - in a real implementation, this would call
	// the appropriate attribute extraction logic
	return &AnalysisResponse{
		Results:    map[string]interface{}{"attribute_values": map[string]string{}},
		Confidence: 0.8,
	}, nil
}

// TransformForIntent prepares attribute data for intent analysis
func (a *Analyzer) TransformForIntent(data interface{}) (string, error) {
	// Intent analysis primarily expects text, so we need to extract or convert to text
	var text string

	switch v := data.(type) {
	case string:
		// If it's already a string, use it directly
		text = v
	case map[string]interface{}:
		// If it's a map, check if it has text or attribute values we can use
		if textVal, ok := v["text"].(string); ok && textVal != "" {
			text = textVal
		} else if rawContent, ok := v["raw_content"].(string); ok && rawContent != "" {
			text = rawContent
		} else if attrValues, ok := v["attribute_values"].(map[string]interface{}); ok {
			// Try to find conversation text in attributes
			// Common field names for conversation text
			textFields := []string{"transcript", "conversation_text", "content", "message", "text"}

			for _, field := range textFields {
				if fieldVal, ok := attrValues[field].(string); ok && fieldVal != "" {
					text = fieldVal
					break
				}
			}

			// If no text found but we have attributes, create a simplified text representation
			if text == "" {
				var sb strings.Builder
				sb.WriteString("Conversation attributes:\n")

				for k, v := range attrValues {
					sb.WriteString(fmt.Sprintf("%s: %v\n", k, v))
				}

				text = sb.String()
			}
		}
	default:
		// For other types, try to marshal to string
		bytes, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("failed to transform data for intent: %w", err)
		}
		text = string(bytes)
	}

	return text, nil
}

// TransformIntentForFindings prepares intent data for findings analysis
func (a *Analyzer) TransformIntentForFindings(intentData interface{}, questions []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Extract intent data
	switch v := intentData.(type) {
	case map[string]interface{}:
		// If it's already a map, check if it has the intent structure
		if v["label"] != nil || v["label_name"] != nil {
			// It appears to be intent data, add it to intents array
			result["intents"] = []interface{}{v}
		} else {
			// Just use the map directly
			result = v
		}
	case *IntentClassification:
		// Convert IntentClassification to map
		intentMap := map[string]interface{}{
			"label":       v.Label,
			"label_name":  v.LabelName,
			"description": v.Description,
		}
		result["intents"] = []interface{}{intentMap}
	default:
		// For other types, try to marshal and unmarshal
		bytes, err := json.Marshal(intentData)
		if err != nil {
			return nil, fmt.Errorf("failed to transform intent data: %w", err)
		}

		// Try to unmarshal as intent classification
		var intent IntentClassification
		if err := json.Unmarshal(bytes, &intent); err == nil && intent.Label != "" {
			intentMap := map[string]interface{}{
				"label":       intent.Label,
				"label_name":  intent.LabelName,
				"description": intent.Description,
			}
			result["intents"] = []interface{}{intentMap}
		} else {
			// Try as generic map
			var dataMap map[string]interface{}
			if err := json.Unmarshal(bytes, &dataMap); err == nil {
				result = dataMap
			} else {
				return nil, fmt.Errorf("failed to transform intent data to findings format: %w", err)
			}
		}
	}

	// Add questions if provided
	if len(questions) > 0 {
		result["questions"] = questions
	}

	return result, nil
}

// TransformFindingsForRecommendations prepares findings data for recommendations generation
func (a *Analyzer) TransformFindingsForRecommendations(findingsData interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Extract findings data
	switch v := findingsData.(type) {
	case map[string]interface{}:
		// If it's already a map, use it directly
		result = v
	case []string:
		// If it's an array of string findings, format appropriately
		result["findings"] = v
	default:
		// For other types, try to marshal and unmarshal
		bytes, err := json.Marshal(findingsData)
		if err != nil {
			return nil, fmt.Errorf("failed to transform findings data: %w", err)
		}

		if err := json.Unmarshal(bytes, &result); err != nil {
			return nil, fmt.Errorf("failed to transform findings to recommendations format: %w", err)
		}
	}

	// Ensure the result has a findings array for the recommendation engine
	if _, hasFindings := result["findings"]; !hasFindings {
		// If we have answers field instead, convert it to findings
		if answers, ok := result["answers"].([]interface{}); ok {
			findings := make([]string, 0, len(answers))
			for _, ans := range answers {
				if answer, ok := ans.(map[string]interface{}); ok {
					if text, ok := answer["answer"].(string); ok && text != "" {
						findings = append(findings, text)
					}
				}
			}
			result["findings"] = findings
		}
	}

	return result, nil
}

// TransformRecommendationsForPlan prepares recommendations data for action plan generation
func (a *Analyzer) TransformRecommendationsForPlan(recsData interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Extract recommendations data
	switch v := recsData.(type) {
	case *RecommendationResponse:
		// If it's already a RecommendationResponse, convert to map
		bytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal recommendations: %w", err)
		}

		var recsMap map[string]interface{}
		if err := json.Unmarshal(bytes, &recsMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recommendations: %w", err)
		}
		result["recommendations"] = recsMap
	case map[string]interface{}:
		// Check if it has the expected structure
		if _, hasActions := v["immediate_actions"]; hasActions {
			// It seems to be a recommendation response
			result["recommendations"] = v
		} else if recs, hasRecs := v["recommendations"]; hasRecs {
			// It has a recommendations field
			result["recommendations"] = recs
		} else {
			// Just use the map and assume it has the needed data
			result = v
		}
	default:
		// For other types, try to marshal and unmarshal
		bytes, err := json.Marshal(recsData)
		if err != nil {
			return nil, fmt.Errorf("failed to transform recommendations data: %w", err)
		}

		if err := json.Unmarshal(bytes, &result); err != nil {
			return nil, fmt.Errorf("failed to transform recommendations to plan format: %w", err)
		}
	}

	return result, nil
}

// TransformPlanForTimeline prepares action plan data for timeline generation
func (a *Analyzer) TransformPlanForTimeline(planData interface{}, resources map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Extract action plan data
	switch v := planData.(type) {
	case *ActionPlan:
		// If it's already an ActionPlan, add it directly
		result["action_plan"] = v
	case map[string]interface{}:
		// Check if it has the expected action plan structure
		if _, hasGoals := v["goals"]; hasGoals {
			// It seems to be an action plan
			result["action_plan"] = v
		} else if plan, hasPlan := v["action_plan"]; hasPlan {
			// It has an action_plan field
			result["action_plan"] = plan
		} else {
			// Just use the map and assume it has the needed data
			result = v
		}
	default:
		// For other types, try to marshal and unmarshal
		bytes, err := json.Marshal(planData)
		if err != nil {
			return nil, fmt.Errorf("failed to transform plan data: %w", err)
		}

		// Try to unmarshal as ActionPlan
		var plan ActionPlan
		if err := json.Unmarshal(bytes, &plan); err == nil && len(plan.Goals) > 0 {
			result["action_plan"] = plan
		} else {
			// Try as generic map
			var dataMap map[string]interface{}
			if err := json.Unmarshal(bytes, &dataMap); err == nil {
				result = dataMap
			} else {
				return nil, fmt.Errorf("failed to transform plan data to timeline format: %w", err)
			}
		}
	}

	// Add resources if provided
	if resources != nil && len(resources) > 0 {
		result["resources"] = resources
	}

	return result, nil
}
