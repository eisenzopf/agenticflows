package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"agenticflows/backend/analysis"
	"agenticflows/backend/analysis/models"
	"agenticflows/backend/db"

	"github.com/google/uuid"
)

// AnalysisHandler handles analysis API requests
type AnalysisHandler struct {
	analysisFacade       *analysis.AnalysisFacade
	textGenerator        *analysis.TextGenerator
	recommendationEngine *analysis.RecommendationEngine
	planner              *analysis.Planner
	apiKey               string
}

// NewAnalysisHandler creates a new handler for analysis endpoints
func NewAnalysisHandler() (*AnalysisHandler, error) {
	// Initialize database table
	if err := db.AddTableForAnalysis(); err != nil {
		return nil, fmt.Errorf("failed to initialize analysis table: %w", err)
	}

	// Get API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	// Create analyzer facade
	analysisFacade, err := analysis.NewAnalysisFacade(apiKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create analysis facade: %w", err)
	}

	// Create text generator, recommendation engine, and planner
	// TODO: These will be migrated to the facade in the future
	textGenerator, err := analysis.NewTextGenerator(apiKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create text generator: %w", err)
	}

	recommendationEngine, err := analysis.NewRecommendationEngine(apiKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create recommendation engine: %w", err)
	}

	planner, err := analysis.NewPlanner(apiKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create planner: %w", err)
	}

	// Create and return the analysis handler
	return &AnalysisHandler{
		analysisFacade:       analysisFacade,
		textGenerator:        textGenerator,
		recommendationEngine: recommendationEngine,
		planner:              planner,
		apiKey:               apiKey,
	}, nil
}

// HandleAnalysis handles the unified /api/analysis endpoint
func (h *AnalysisHandler) HandleAnalysis(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse standard request
	var req models.StandardAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendAnalysisError(w, "invalid_request", fmt.Sprintf("Invalid request format: %s", err), http.StatusBadRequest)
		return
	}

	// Log the analysis type for debugging
	log.Printf("Received analysis request with type: %s", req.AnalysisType)

	// Convert analysis type to lowercase for case-insensitive matching
	analysisType := strings.ToLower(req.AnalysisType)
	log.Printf("Using normalized analysis type: %s", analysisType)

	// Route to appropriate analysis function based on type
	var resp *models.StandardAnalysisResponse
	var err error

	switch analysisType {
	case "trends":
		resp, err = h.handleTrendsAnalysis(r.Context(), req)
	case "patterns":
		resp, err = h.handlePatternsAnalysis(r.Context(), req)
	case "findings":
		resp, err = h.handleFindingsAnalysis(r.Context(), req)
	case "attributes":
		resp, err = h.handleAttributesAnalysis(r.Context(), req)
	case "intent":
		resp, err = h.handleIntentAnalysis(r.Context(), req)
	case "recommendations":
		resp, err = h.handleRecommendationsAnalysis(r.Context(), req)
	case "plan":
		resp, err = h.handlePlanAnalysis(r.Context(), req)
	default:
		log.Printf("Invalid analysis type: %s (original: %s)", analysisType, req.AnalysisType)
		sendAnalysisError(w, "invalid_analysis_type", "Invalid analysis type", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Printf("Error processing %s analysis: %v", req.AnalysisType, err)
		sendAnalysisError(w, "analysis_error", err.Error(), http.StatusInternalServerError)
		return
	}

	// Save result to database if workflow ID is provided
	if req.WorkflowID != "" && resp != nil && resp.Error == nil {
		resultID := uuid.New().String()
		resultsJSON, err := json.Marshal(resp.Results)
		if err != nil {
			log.Printf("Error marshaling results for storage: %v", err)
		} else {
			if err := db.SaveAnalysisResult(resultID, req.WorkflowID, req.AnalysisType, string(resultsJSON)); err != nil {
				log.Printf("Error saving analysis result: %v", err)
			}
		}
	}

	// Return standard response
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// HandleAnalysisResults handles /api/analysis/results endpoint
func (h *AnalysisHandler) HandleAnalysisResults(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		// Get analysis results for a workflow
		workflowID := r.URL.Query().Get("workflow_id")
		if workflowID == "" {
			http.Error(w, "workflow_id is required", http.StatusBadRequest)
			return
		}

		results, err := db.GetAnalysisResultsByWorkflow(workflowID)
		if err != nil {
			log.Printf("Error getting analysis results: %v", err)
			http.Error(w, "Failed to get analysis results", http.StatusInternalServerError)
			return
		}

		if err := json.NewEncoder(w).Encode(results); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}

	case http.MethodDelete:
		// Delete an analysis result
		id := strings.TrimPrefix(r.URL.Path, "/api/analysis/results/")
		if id == "" {
			http.Error(w, "Result ID is required", http.StatusBadRequest)
			return
		}

		if err := db.DeleteAnalysisResult(id); err != nil {
			log.Printf("Error deleting analysis result: %v", err)
			http.Error(w, "Failed to delete analysis result", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleChainAnalysis handles the chain analysis endpoint for workflows
func (h *AnalysisHandler) HandleChainAnalysis(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req struct {
		WorkflowID string                 `json:"workflow_id"`
		Steps      []string               `json:"steps"`
		Text       string                 `json:"text"`
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if req.WorkflowID == "" {
		http.Error(w, "workflow_id is required", http.StatusBadRequest)
		return
	}
	if len(req.Steps) == 0 {
		http.Error(w, "steps are required", http.StatusBadRequest)
		return
	}

	// Initialize chain analysis config
	config := map[string]interface{}{
		"steps": req.Steps,
	}
	if req.Parameters != nil {
		config["step_config"] = req.Parameters
	}

	// Create input data with the text if provided
	inputData := map[string]interface{}{}
	if req.Text != "" {
		inputData["text"] = req.Text
	}

	// Perform chain analysis
	results, err := h.analysisFacade.ChainAnalysis(r.Context(), inputData, config)
	if err != nil {
		log.Printf("Error in chain analysis: %v", err)
		http.Error(w, fmt.Sprintf("Error in chain analysis: %v", err), http.StatusInternalServerError)
		return
	}

	// Return chain analysis response
	chainResp := struct {
		WorkflowID string                 `json:"workflow_id"`
		Timestamp  time.Time              `json:"timestamp"`
		Results    map[string]interface{} `json:"results"`
	}{
		WorkflowID: req.WorkflowID,
		Timestamp:  time.Now(),
		Results:    results,
	}

	if err := json.NewEncoder(w).Encode(chainResp); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// HandleGetFunctionMetadata handles metadata requests for analysis functions
func (h *AnalysisHandler) HandleGetFunctionMetadata(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Define function metadata
	metadata := getFunctionMetadata()

	// Return metadata
	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Helper function to send standardized error responses
func sendAnalysisError(w http.ResponseWriter, code string, message string, statusCode int) {
	resp := models.StandardAnalysisResponse{
		Timestamp: time.Now(),
		Error: &models.AnalysisError{
			Code:    code,
			Message: message,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding error response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
