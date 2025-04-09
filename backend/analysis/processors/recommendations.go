package processors

import (
	"context"
	"encoding/json"
	"fmt"

	"agenticflows/backend/analysis/core"
	"agenticflows/backend/analysis/models"
)

// RecommendationsProcessor handles generation of recommendations based on analysis results
type RecommendationsProcessor struct {
	analyzer *core.Analyzer
}

// NewRecommendationsProcessor creates a new RecommendationsProcessor instance
func NewRecommendationsProcessor(analyzer *core.Analyzer) *RecommendationsProcessor {
	return &RecommendationsProcessor{
		analyzer: analyzer,
	}
}

// GenerateRecommendations generates specific recommendations based on analysis results and focus area
func (r *RecommendationsProcessor) GenerateRecommendations(
	ctx context.Context,
	analysisResults map[string]interface{},
	focusArea string,
) (*models.RecommendationResponse, error) {
	// Validate input
	if len(analysisResults) == 0 {
		return nil, fmt.Errorf("analysis results are required")
	}
	if focusArea == "" {
		return nil, fmt.Errorf("focus area is required")
	}

	// Format analysis results for the prompt
	analysisBytes, err := json.Marshal(analysisResults)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal analysis results: %w", err)
	}

	prompt := fmt.Sprintf(`Based on this analysis focused on %s:

%s

Generate specific, actionable recommendations. Consider:
1. Immediate actions that can be taken
2. Rationale for each recommendation
3. Expected impact of implementation
4. Priority level (1-5, where 5 is highest)

Format your response as JSON with these fields:
{
  "immediate_actions": [
    {
      "action": str,
      "rationale": str,
      "expected_impact": str,
      "priority": int
    }
  ],
  "implementation_notes": [str],
  "success_metrics": [str]
}`, focusArea, string(analysisBytes))

	expectedFormat := map[string]interface{}{
		"immediate_actions": []interface{}{
			map[string]interface{}{
				"action":          "",
				"rationale":       "",
				"expected_impact": "",
				"priority":        0,
			},
		},
		"implementation_notes": []interface{}{},
		"success_metrics":      []interface{}{},
	}

	result, err := r.analyzer.LLMClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Parse the result into RecommendationResponse
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	// Extract recommendations
	response := &models.RecommendationResponse{}

	// Extract immediate actions
	if actionsRaw, ok := resultMap["immediate_actions"].([]interface{}); ok {
		for _, actionRaw := range actionsRaw {
			if actionMap, ok := actionRaw.(map[string]interface{}); ok {
				rec := models.Recommendation{
					Action:         getString(actionMap, "action"),
					Rationale:      getString(actionMap, "rationale"),
					ExpectedImpact: getString(actionMap, "expected_impact"),
					Priority:       int(getFloat(actionMap, "priority")),
				}
				response.ImmediateActions = append(response.ImmediateActions, rec)
			}
		}
	}

	// Extract implementation notes
	if notesRaw, ok := resultMap["implementation_notes"].([]interface{}); ok {
		for _, noteRaw := range notesRaw {
			if note, ok := noteRaw.(string); ok && note != "" {
				response.ImplementationNotes = append(response.ImplementationNotes, note)
			}
		}
	}

	// Extract success metrics
	if metricsRaw, ok := resultMap["success_metrics"].([]interface{}); ok {
		for _, metricRaw := range metricsRaw {
			if metric, ok := metricRaw.(string); ok && metric != "" {
				response.SuccessMetrics = append(response.SuccessMetrics, metric)
			}
		}
	}

	return response, nil
}

// PrioritizeRecommendations prioritizes recommendations based on given criteria
func (r *RecommendationsProcessor) PrioritizeRecommendations(
	ctx context.Context,
	recommendations []models.Recommendation,
	criteria map[string]float64,
) ([]models.Recommendation, error) {
	// Validate input
	if len(recommendations) == 0 {
		return nil, fmt.Errorf("recommendations are required")
	}
	if len(criteria) == 0 {
		return nil, fmt.Errorf("prioritization criteria are required")
	}

	// Format recommendations for the prompt
	recsBytes, err := json.Marshal(recommendations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal recommendations: %w", err)
	}

	// Format criteria for the prompt
	criteriaBytes, err := json.Marshal(criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal criteria: %w", err)
	}

	prompt := fmt.Sprintf(`Prioritize these recommendations based on the given criteria:

Recommendations:
%s

Prioritization Criteria (with weights):
%s

Review each recommendation and re-prioritize them based on the weighted criteria.
Assign a new priority score (1-10) to each, where 10 is highest priority.

Return the reprioritized recommendations as JSON with the same structure as the input, but with updated priority values.
Include a brief explanation of why each recommendation received its new priority.
Format as a JSON array.`, string(recsBytes), string(criteriaBytes))

	result, err := r.analyzer.LLMClient.GenerateContent(ctx, prompt, []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Parse the result into a list of recommendations
	resultArray, ok := result.([]interface{})
	if !ok {
		resultMap, isMap := result.(map[string]interface{})
		if !isMap {
			return nil, fmt.Errorf("unexpected result format")
		}

		// Check if result is wrapped in a 'recommendations' field
		if recs, ok := resultMap["recommendations"].([]interface{}); ok {
			resultArray = recs
		} else {
			return nil, fmt.Errorf("unexpected result format, missing recommendations array")
		}
	}

	// Convert to Recommendation objects
	prioritizedRecs := make([]models.Recommendation, 0, len(resultArray))
	for _, recRaw := range resultArray {
		if recMap, ok := recRaw.(map[string]interface{}); ok {
			rec := models.Recommendation{
				Action:         getString(recMap, "action"),
				Rationale:      getString(recMap, "rationale"),
				ExpectedImpact: getString(recMap, "expected_impact"),
				Priority:       int(getFloat(recMap, "priority")),
			}
			prioritizedRecs = append(prioritizedRecs, rec)
		}
	}

	return prioritizedRecs, nil
}

// GenerateRetentionStrategies generates retention strategy recommendations
func (r *RecommendationsProcessor) GenerateRetentionStrategies(
	ctx context.Context,
	analysisResults map[string]interface{},
) (*models.RetentionStrategy, error) {
	// Validate input
	if len(analysisResults) == 0 {
		return nil, fmt.Errorf("analysis results are required")
	}

	// Format analysis results for the prompt
	analysisBytes, err := json.Marshal(analysisResults)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal analysis results: %w", err)
	}

	prompt := fmt.Sprintf(`Based on this analysis of customer cancellations and retention efforts:

%s

Recommend specific, actionable steps to improve customer retention. Consider:
1. Immediate changes to agent behavior
2. Process improvements
3. Most effective retention offers
4. Training opportunities

Format as JSON:
{
  "target_segment": str,
  "immediate_actions": [
    {
      "action": str,
      "rationale": str,
      "expected_impact": str,
      "priority": int
    }
  ],
  "process_changes": [str],
  "training_needs": [str],
  "success_metrics": [str]
}`, string(analysisBytes))

	expectedFormat := map[string]interface{}{
		"target_segment": "",
		"immediate_actions": []interface{}{
			map[string]interface{}{
				"action":          "",
				"rationale":       "",
				"expected_impact": "",
				"priority":        0,
			},
		},
		"process_changes": []interface{}{},
		"training_needs":  []interface{}{},
		"success_metrics": []interface{}{},
	}

	result, err := r.analyzer.LLMClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Parse the result into RetentionStrategy
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	strategy := &models.RetentionStrategy{
		TargetSegment: getString(resultMap, "target_segment"),
	}

	// Extract immediate actions
	if actionsRaw, ok := resultMap["immediate_actions"].([]interface{}); ok {
		for _, actionRaw := range actionsRaw {
			if actionMap, ok := actionRaw.(map[string]interface{}); ok {
				rec := models.Recommendation{
					Action:         getString(actionMap, "action"),
					Rationale:      getString(actionMap, "rationale"),
					ExpectedImpact: getString(actionMap, "expected_impact"),
					Priority:       int(getFloat(actionMap, "priority")),
				}
				strategy.ImmediateActions = append(strategy.ImmediateActions, rec)
			}
		}
	}

	// Extract process changes
	if changesRaw, ok := resultMap["process_changes"].([]interface{}); ok {
		for _, changeRaw := range changesRaw {
			if change, ok := changeRaw.(string); ok && change != "" {
				strategy.ProcessChanges = append(strategy.ProcessChanges, change)
			}
		}
	}

	// Extract training needs
	if trainingRaw, ok := resultMap["training_needs"].([]interface{}); ok {
		for _, trainRaw := range trainingRaw {
			if training, ok := trainRaw.(string); ok && training != "" {
				strategy.TrainingNeeds = append(strategy.TrainingNeeds, training)
			}
		}
	}

	// Extract success metrics
	if metricsRaw, ok := resultMap["success_metrics"].([]interface{}); ok {
		for _, metricRaw := range metricsRaw {
			if metric, ok := metricRaw.(string); ok && metric != "" {
				strategy.SuccessMetrics = append(strategy.SuccessMetrics, metric)
			}
		}
	}

	return strategy, nil
}

// Helper functions for type conversion

// getString safely extracts a string from a map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

// getFloat safely extracts a float from a map
func getFloat(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case string:
			f, err := json.Number(v).Float64()
			if err == nil {
				return f
			}
		}
	}
	return 0.0
}
