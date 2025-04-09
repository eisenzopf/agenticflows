package handlers

import (
	"context"
	"fmt"
	"time"

	"agenticflows/backend/analysis/models"
)

// handlePatternsAnalysis handles patterns analysis requests
func (h *AnalysisHandler) handlePatternsAnalysis(ctx context.Context, req models.StandardAnalysisRequest) (*models.StandardAnalysisResponse, error) {
	// Extract pattern types from parameters
	patternTypes := []string{}
	if typesParam, ok := req.Parameters["pattern_types"]; ok {
		if types, ok := typesParam.([]interface{}); ok {
			for _, t := range types {
				if typeStr, ok := t.(string); ok {
					patternTypes = append(patternTypes, typeStr)
				}
			}
		}
	}

	// If no pattern types were specified, use defaults
	if len(patternTypes) == 0 {
		patternTypes = []string{
			"communication_patterns",
			"recurring_issues",
			"customer_behavior",
		}
	}

	// Create a request object for the patterns analyzer
	analysisReq := models.AnalysisRequest{
		Text:         req.Text,
		PatternTypes: patternTypes,
	}

	// If data was provided, add it to the request
	if req.Data != nil {
		analysisReq.AttributeValues = req.Data
	}

	// Perform the patterns analysis using the facade
	result, err := h.analysisFacade.IdentifyPatterns(ctx, analysisReq)
	if err != nil {
		return nil, fmt.Errorf("failed to identify patterns: %w", err)
	}

	// Return the results in the standard response format
	return &models.StandardAnalysisResponse{
		AnalysisType: "patterns",
		WorkflowID:   req.WorkflowID,
		Timestamp:    time.Now(),
		Results:      result.Results,
		Confidence:   result.Confidence,
	}, nil
}
