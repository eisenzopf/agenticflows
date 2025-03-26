package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"agenticflows/backend/analysis"
	"agenticflows/backend/db"

	"github.com/google/uuid"
)

// AnalysisHandler handles analysis API requests
type AnalysisHandler struct {
	analyzer      *analysis.Analyzer
	textGenerator *analysis.TextGenerator
	apiKey        string
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

	// Create analyzer and text generator
	analyzer, err := analysis.NewAnalyzer(apiKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create analyzer: %w", err)
	}

	textGenerator, err := analysis.NewTextGenerator(apiKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create text generator: %w", err)
	}

	return &AnalysisHandler{
		analyzer:      analyzer,
		textGenerator: textGenerator,
		apiKey:        apiKey,
	}, nil
}

// handleAnalysisTrends handles /api/analysis/trends
func (h *AnalysisHandler) handleAnalysisTrends(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req analysis.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.FocusAreas) == 0 {
		http.Error(w, "Focus areas are required", http.StatusBadRequest)
		return
	}

	// Extract workflow ID from the request header or query parameter
	workflowID := r.Header.Get("X-Workflow-ID")
	if workflowID == "" {
		workflowID = r.URL.Query().Get("workflow_id")
	}

	// Process the request
	resp, err := h.analyzer.AnalyzeTrends(r.Context(), req)
	if err != nil {
		log.Printf("Error analyzing trends: %v", err)
		http.Error(w, "Failed to analyze trends", http.StatusInternalServerError)
		return
	}

	// Save result to database if workflow ID is provided
	if workflowID != "" {
		resultID := uuid.New().String()
		if err := db.SaveAnalysisResult(resultID, workflowID, "trends", resp.Results); err != nil {
			log.Printf("Error saving analysis result: %v", err)
		}
	}

	// Return response
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleAnalysisPatterns handles /api/analysis/patterns
func (h *AnalysisHandler) handleAnalysisPatterns(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req analysis.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.PatternTypes) == 0 {
		http.Error(w, "Pattern types are required", http.StatusBadRequest)
		return
	}

	// Extract workflow ID
	workflowID := r.Header.Get("X-Workflow-ID")
	if workflowID == "" {
		workflowID = r.URL.Query().Get("workflow_id")
	}

	// Process the request
	resp, err := h.analyzer.IdentifyPatterns(r.Context(), req)
	if err != nil {
		log.Printf("Error identifying patterns: %v", err)
		http.Error(w, "Failed to identify patterns", http.StatusInternalServerError)
		return
	}

	// Save result to database if workflow ID is provided
	if workflowID != "" {
		resultID := uuid.New().String()
		if err := db.SaveAnalysisResult(resultID, workflowID, "patterns", resp.Results); err != nil {
			log.Printf("Error saving analysis result: %v", err)
		}
	}

	// Return response
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleAnalysisFindings handles /api/analysis/findings
func (h *AnalysisHandler) handleAnalysisFindings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req analysis.AnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.Questions) == 0 {
		http.Error(w, "Questions are required", http.StatusBadRequest)
		return
	}
	if req.AttributeValues == nil {
		http.Error(w, "Attribute values are required", http.StatusBadRequest)
		return
	}

	// Extract workflow ID
	workflowID := r.Header.Get("X-Workflow-ID")
	if workflowID == "" {
		workflowID = r.URL.Query().Get("workflow_id")
	}

	// Process the request
	resp, err := h.analyzer.AnalyzeFindings(r.Context(), req)
	if err != nil {
		log.Printf("Error analyzing findings: %v", err)
		http.Error(w, "Failed to analyze findings", http.StatusInternalServerError)
		return
	}

	// Save result to database if workflow ID is provided
	if workflowID != "" {
		resultID := uuid.New().String()
		if err := db.SaveAnalysisResult(resultID, workflowID, "findings", resp.Results); err != nil {
			log.Printf("Error saving analysis result: %v", err)
		}
	}

	// Return response
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// handleTextAttributes handles /api/analysis/attributes
func (h *AnalysisHandler) handleTextAttributes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Text             string                         `json:"text"`
		Attributes       []analysis.AttributeDefinition `json:"attributes"`
		WorkflowID       string                         `json:"workflow_id,omitempty"`
		GenerateRequired bool                           `json:"generate_required,omitempty"`
		Questions        []string                       `json:"questions,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
		return
	}

	// Generate required attributes if needed
	if req.GenerateRequired {
		if len(req.Questions) == 0 {
			http.Error(w, "Questions are required when generating attributes", http.StatusBadRequest)
			return
		}

		// Get existing attributes for reference
		var existingAttributes []string
		for _, attr := range req.Attributes {
			existingAttributes = append(existingAttributes, attr.FieldName)
		}

		// Generate required attributes
		attributes, err := h.textGenerator.GenerateRequiredAttributes(
			r.Context(),
			req.Questions,
			existingAttributes,
		)
		if err != nil {
			log.Printf("Error generating required attributes: %v", err)
			http.Error(w, "Failed to generate required attributes", http.StatusInternalServerError)
			return
		}

		// Return generated attributes
		resp := map[string]interface{}{
			"attributes": attributes,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}

	// Extract values for provided attributes
	if len(req.Attributes) == 0 {
		http.Error(w, "Attributes are required", http.StatusBadRequest)
		return
	}

	// Process the attribute extraction
	attributeValues, err := h.textGenerator.GenerateAttributes(r.Context(), req.Text, req.Attributes)
	if err != nil {
		log.Printf("Error generating attribute values: %v", err)
		http.Error(w, "Failed to generate attribute values", http.StatusInternalServerError)
		return
	}

	// Save result to database if workflow ID is provided
	if req.WorkflowID != "" {
		resultID := uuid.New().String()
		if err := db.SaveAnalysisResult(resultID, req.WorkflowID, "attributes", attributeValues); err != nil {
			log.Printf("Error saving analysis result: %v", err)
		}
	}

	// Return response
	resp := map[string]interface{}{
		"attribute_values": attributeValues,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// handleTextIntent handles /api/analysis/intent
func (h *AnalysisHandler) handleTextIntent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Text       string `json:"text"`
		WorkflowID string `json:"workflow_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Text == "" {
		http.Error(w, "Text is required", http.StatusBadRequest)
		return
	}

	// Process the intent generation
	intent, err := h.textGenerator.GenerateIntent(r.Context(), req.Text)
	if err != nil {
		log.Printf("Error generating intent: %v", err)
		http.Error(w, "Failed to generate intent", http.StatusInternalServerError)
		return
	}

	// Save result to database if workflow ID is provided
	if req.WorkflowID != "" {
		resultID := uuid.New().String()
		if err := db.SaveAnalysisResult(resultID, req.WorkflowID, "intent", intent); err != nil {
			log.Printf("Error saving analysis result: %v", err)
		}
	}

	// Return response
	if err := json.NewEncoder(w).Encode(intent); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// handleAnalysisResults handles /api/analysis/results
func (h *AnalysisHandler) handleAnalysisResults(w http.ResponseWriter, r *http.Request) {
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