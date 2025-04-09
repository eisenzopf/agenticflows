package processors

import (
	"context"
	"encoding/json"
	"fmt"

	"agenticflows/backend/analysis/core"
	"agenticflows/backend/analysis/models"
)

// TrendsAnalyzer handles analysis of trends in conversation data
type TrendsAnalyzer struct {
	analyzer *core.Analyzer
}

// NewTrendsAnalyzer creates a new TrendsAnalyzer
func NewTrendsAnalyzer(analyzer *core.Analyzer) *TrendsAnalyzer {
	return &TrendsAnalyzer{
		analyzer: analyzer,
	}
}

// AnalyzeTrends analyzes trends in conversation data for specified focus areas
func (t *TrendsAnalyzer) AnalyzeTrends(ctx context.Context, req models.AnalysisRequest) (*models.AnalysisResponse, error) {
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

	result, err := t.analyzer.LLMClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	return &models.AnalysisResponse{
		Results:    result,
		Confidence: 0.8, // Default confidence
	}, nil
}

// ExtractTrendsOutput extracts the most relevant information from trends analysis
func (t *TrendsAnalyzer) ExtractTrendsOutput(resp *models.AnalysisResponse) (map[string]interface{}, error) {
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
