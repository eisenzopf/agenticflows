package handlers

import (
	"context"
	"fmt"
	"time"

	"agenticflows/backend/analysis/models"
)

// handleTrendsAnalysis handles trends analysis requests
func (h *AnalysisHandler) handleTrendsAnalysis(ctx context.Context, req models.StandardAnalysisRequest) (*models.StandardAnalysisResponse, error) {
	// Extract focus areas from parameters
	focusAreas := []string{}
	if areasParam, ok := req.Parameters["focus_areas"]; ok {
		if areas, ok := areasParam.([]interface{}); ok {
			for _, area := range areas {
				if areaStr, ok := area.(string); ok {
					focusAreas = append(focusAreas, areaStr)
				}
			}
		}
	}

	// If no focus areas were specified, use defaults
	if len(focusAreas) == 0 {
		focusAreas = []string{
			"Customer Satisfaction",
			"Service Issues",
			"Product Quality",
			"Wait Times",
		}
	}

	// Create a request object for the trends analyzer
	analysisReq := models.AnalysisRequest{
		Text:       req.Text,
		FocusAreas: focusAreas,
	}

	// If data was provided, add it to the request
	if req.Data != nil {
		analysisReq.AttributeValues = req.Data
	}

	// Perform the trends analysis using the facade
	result, err := h.analysisFacade.AnalyzeTrends(ctx, analysisReq)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze trends: %w", err)
	}

	// Return the results in the standard response format
	return &models.StandardAnalysisResponse{
		AnalysisType: "trends",
		WorkflowID:   req.WorkflowID,
		Timestamp:    time.Now(),
		Results:      result.Results,
		Confidence:   result.Confidence,
	}, nil
}
