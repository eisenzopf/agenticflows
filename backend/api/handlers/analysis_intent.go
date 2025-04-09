package handlers

import (
	"context"
	"fmt"
	"time"

	"agenticflows/backend/analysis/models"
)

// handleIntentAnalysisImpl implements the actual intent analysis logic
func (h *AnalysisHandler) handleIntentAnalysisImpl(ctx context.Context, req models.StandardAnalysisRequest) (*models.StandardAnalysisResponse, error) {
	// Validate request
	if req.Text == "" {
		return nil, fmt.Errorf("text is required for intent analysis")
	}

	// Process the intent generation
	intent, err := h.textGenerator.GenerateIntent(ctx, req.Text)
	if err != nil {
		return nil, err
	}

	// Return generated intent in standard response
	resp := &models.StandardAnalysisResponse{
		AnalysisType: "intent",
		WorkflowID:   req.WorkflowID,
		Timestamp:    time.Now(),
		Results:      intent,
		Confidence:   0.85,
	}

	return resp, nil
}

// handleIntentAnalysis is kept for backward compatibility - delegates to the actual implementation
func (h *AnalysisHandler) handleIntentAnalysis(ctx context.Context, req models.StandardAnalysisRequest) (*models.StandardAnalysisResponse, error) {
	// This method is required to be compatible with the handler framework in analysis_base.go
	return h.handleIntentAnalysisImpl(ctx, req)
}
