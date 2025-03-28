package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"agenticflows/backend/analysis"
	"agenticflows/backend/db"

	"github.com/google/uuid"
)

// AnalysisHandler handles analysis API requests
type AnalysisHandler struct {
	analyzer             *analysis.Analyzer
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

	// Create analyzer and text generator
	analyzer, err := analysis.NewAnalyzer(apiKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create analyzer: %w", err)
	}

	textGenerator, err := analysis.NewTextGenerator(apiKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create text generator: %w", err)
	}

	// Create recommendation engine and planner
	recommendationEngine, err := analysis.NewRecommendationEngine(apiKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create recommendation engine: %w", err)
	}

	planner, err := analysis.NewPlanner(apiKey, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create planner: %w", err)
	}

	return &AnalysisHandler{
		analyzer:             analyzer,
		textGenerator:        textGenerator,
		recommendationEngine: recommendationEngine,
		planner:              planner,
		apiKey:               apiKey,
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

// handleAnalysis handles the unified /api/analysis endpoint
func (h *AnalysisHandler) handleAnalysis(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse standard request
	var req analysis.StandardAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendAnalysisError(w, "invalid_request", fmt.Sprintf("Invalid request format: %s", err), http.StatusBadRequest)
		return
	}

	// Route to appropriate analysis function based on type
	var resp *analysis.StandardAnalysisResponse
	var err error

	switch req.AnalysisType {
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

// Helper function to send standardized error responses
func sendAnalysisError(w http.ResponseWriter, code string, message string, statusCode int) {
	resp := analysis.StandardAnalysisResponse{
		Timestamp: time.Now(),
		Error: &analysis.AnalysisError{
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

// handleTrendsAnalysis processes a trends analysis request
func (h *AnalysisHandler) handleTrendsAnalysis(ctx context.Context, req analysis.StandardAnalysisRequest) (*analysis.StandardAnalysisResponse, error) {
	// Extract parameters
	focusAreas, err := extractStringSlice(req.Parameters, "focus_areas")
	if err != nil {
		return nil, fmt.Errorf("invalid focus_areas parameter: %w", err)
	}

	// Convert to legacy request format
	legacyReq := analysis.AnalysisRequest{
		FocusAreas:      focusAreas,
		AttributeValues: req.Data,
		Text:            req.Text,
	}

	// Process the request
	legacyResp, err := h.analyzer.AnalyzeTrends(ctx, legacyReq)
	if err != nil {
		return nil, err
	}

	// Convert to standard response format
	resp := &analysis.StandardAnalysisResponse{
		AnalysisType: "trends",
		WorkflowID:   req.WorkflowID,
		Timestamp:    time.Now(),
		Results:      legacyResp.Results,
		Confidence:   legacyResp.Confidence,
	}

	// Extract data quality if available
	if results, ok := legacyResp.Results.(map[string]interface{}); ok {
		if dataQuality, ok := results["data_quality"].(map[string]interface{}); ok {
			assessment, _ := dataQuality["assessment"].(string)
			limitations, _ := dataQuality["limitations"].([]string)

			// If limitations is an []interface{}, convert to []string
			if limitIface, ok := dataQuality["limitations"].([]interface{}); ok {
				limitations = make([]string, 0, len(limitIface))
				for _, v := range limitIface {
					if str, ok := v.(string); ok {
						limitations = append(limitations, str)
					}
				}
			}

			resp.DataQuality.Assessment = assessment
			resp.DataQuality.Limitations = limitations
		}
	}

	return resp, nil
}

// handlePatternsAnalysis processes a patterns analysis request
func (h *AnalysisHandler) handlePatternsAnalysis(ctx context.Context, req analysis.StandardAnalysisRequest) (*analysis.StandardAnalysisResponse, error) {
	// Extract parameters
	patternTypes, err := extractStringSlice(req.Parameters, "pattern_types")
	if err != nil {
		return nil, fmt.Errorf("invalid pattern_types parameter: %w", err)
	}

	// Convert to legacy request format
	legacyReq := analysis.AnalysisRequest{
		PatternTypes:    patternTypes,
		AttributeValues: req.Data,
		Text:            req.Text,
	}

	// Process the request
	legacyResp, err := h.analyzer.IdentifyPatterns(ctx, legacyReq)
	if err != nil {
		return nil, err
	}

	// Convert to standard response format
	resp := &analysis.StandardAnalysisResponse{
		AnalysisType: "patterns",
		WorkflowID:   req.WorkflowID,
		Timestamp:    time.Now(),
		Results:      legacyResp.Results,
		Confidence:   legacyResp.Confidence,
	}

	return resp, nil
}

// handleFindingsAnalysis processes a findings analysis request
func (h *AnalysisHandler) handleFindingsAnalysis(ctx context.Context, req analysis.StandardAnalysisRequest) (*analysis.StandardAnalysisResponse, error) {
	// Extract parameters
	questions, err := extractStringSlice(req.Parameters, "questions")
	if err != nil {
		return nil, fmt.Errorf("invalid questions parameter: %w", err)
	}

	// Convert to legacy request format
	legacyReq := analysis.AnalysisRequest{
		Questions:       questions,
		AttributeValues: req.Data,
		Text:            req.Text,
	}

	// Process the request
	legacyResp, err := h.analyzer.AnalyzeFindings(ctx, legacyReq)
	if err != nil {
		return nil, err
	}

	// Convert to standard response format
	resp := &analysis.StandardAnalysisResponse{
		AnalysisType: "findings",
		WorkflowID:   req.WorkflowID,
		Timestamp:    time.Now(),
		Results:      legacyResp.Results,
		Confidence:   legacyResp.Confidence,
	}

	// Extract data gaps if available
	if results, ok := legacyResp.Results.(map[string]interface{}); ok {
		if dataGaps, ok := results["data_gaps"].([]interface{}); ok {
			gaps := make([]string, 0, len(dataGaps))
			for _, gap := range dataGaps {
				if gapStr, ok := gap.(string); ok {
					gaps = append(gaps, gapStr)
				}
			}

			resp.DataQuality.Limitations = gaps
			resp.DataQuality.Assessment = "Based on identified data gaps"
		}
	}

	return resp, nil
}

// handleAttributesAnalysis processes an attributes analysis request
func (h *AnalysisHandler) handleAttributesAnalysis(ctx context.Context, req analysis.StandardAnalysisRequest) (*analysis.StandardAnalysisResponse, error) {
	// Check if this is a generate_required request
	generateRequired, _ := req.Parameters["generate_required"].(bool)

	if generateRequired {
		// Extract questions for generating required attributes
		questions, err := extractStringSlice(req.Parameters, "questions")
		if err != nil {
			return nil, fmt.Errorf("invalid questions parameter: %w", err)
		}

		// Get existing attributes for reference
		var existingAttributes []string
		if existingAttrs, ok := req.Parameters["existing_attributes"].([]interface{}); ok {
			for _, attr := range existingAttrs {
				if attrStr, ok := attr.(string); ok {
					existingAttributes = append(existingAttributes, attrStr)
				}
			}
		}

		// Generate required attributes
		attributes, err := h.textGenerator.GenerateRequiredAttributes(ctx, questions, existingAttributes)
		if err != nil {
			return nil, err
		}

		// Return generated attributes in standard response
		resp := &analysis.StandardAnalysisResponse{
			AnalysisType: "attributes",
			WorkflowID:   req.WorkflowID,
			Timestamp:    time.Now(),
			Results:      map[string]interface{}{"attributes": attributes},
			Confidence:   0.9,
		}

		return resp, nil
	} else {
		// Handle attribute extraction
		// Extract attributes definition
		attributesRaw, ok := req.Parameters["attributes"].([]interface{})
		if !ok || len(attributesRaw) == 0 {
			return nil, fmt.Errorf("attributes parameter is required and must be an array")
		}

		// Convert to AttributeDefinition array
		attributes := make([]analysis.AttributeDefinition, 0, len(attributesRaw))
		for _, attrRaw := range attributesRaw {
			if attrMap, ok := attrRaw.(map[string]interface{}); ok {
				attr := analysis.AttributeDefinition{
					FieldName:   getString(attrMap, "field_name"),
					Title:       getString(attrMap, "title"),
					Description: getString(attrMap, "description"),
					Rationale:   getString(attrMap, "rationale"),
				}

				if attr.FieldName != "" {
					attributes = append(attributes, attr)
				}
			}
		}

		if len(attributes) == 0 {
			return nil, fmt.Errorf("at least one valid attribute definition is required")
		}

		// Process attribute extraction
		attributeValues, err := h.textGenerator.GenerateAttributes(ctx, req.Text, attributes)
		if err != nil {
			return nil, err
		}

		// Return extracted attributes in standard response
		resp := &analysis.StandardAnalysisResponse{
			AnalysisType: "attributes",
			WorkflowID:   req.WorkflowID,
			Timestamp:    time.Now(),
			Results:      map[string]interface{}{"attribute_values": attributeValues},
			Confidence:   0.9,
		}

		return resp, nil
	}
}

// handleIntentAnalysis processes an intent analysis request
func (h *AnalysisHandler) handleIntentAnalysis(ctx context.Context, req analysis.StandardAnalysisRequest) (*analysis.StandardAnalysisResponse, error) {
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
	resp := &analysis.StandardAnalysisResponse{
		AnalysisType: "intent",
		WorkflowID:   req.WorkflowID,
		Timestamp:    time.Now(),
		Results:      intent,
		Confidence:   0.9,
	}

	return resp, nil
}

// handleRecommendationsAnalysis processes a recommendations analysis request
func (h *AnalysisHandler) handleRecommendationsAnalysis(ctx context.Context, req analysis.StandardAnalysisRequest) (*analysis.StandardAnalysisResponse, error) {
	// Extract parameters
	focusArea, _ := req.Parameters["focus_area"].(string)
	if focusArea == "" {
		// Default to a general focus area if not specified
		focusArea = "improving customer experience"
	}

	// Extract prioritization criteria if provided
	criteriaRaw, hasCriteria := req.Parameters["criteria"].(map[string]interface{})

	// Convert to the format expected by the recommendation engine
	var data map[string]interface{}
	if req.Data != nil {
		data = req.Data
	} else {
		// Create an empty data map if none provided
		data = make(map[string]interface{})
	}

	// Generate recommendations
	recs, err := h.recommendationEngine.GenerateRecommendations(ctx, data, focusArea)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// If prioritization criteria are provided, prioritize the recommendations
	if hasCriteria && len(recs.ImmediateActions) > 0 {
		// Convert criteria to the expected format
		criteria := make(map[string]float64)
		for k, v := range criteriaRaw {
			if fv, ok := v.(float64); ok {
				criteria[k] = fv
			} else if iv, ok := v.(int); ok {
				criteria[k] = float64(iv)
			}
		}

		// Prioritize recommendations if we have valid criteria
		if len(criteria) > 0 {
			prioritizedRecs, err := h.recommendationEngine.PrioritizeRecommendations(ctx, recs.ImmediateActions, criteria)
			if err == nil {
				recs.ImmediateActions = prioritizedRecs
			} else {
				log.Printf("Warning: Failed to prioritize recommendations: %v", err)
			}
		}
	}

	// Return standardized response
	resp := &analysis.StandardAnalysisResponse{
		AnalysisType: "recommendations",
		WorkflowID:   req.WorkflowID,
		Timestamp:    time.Now(),
		Results:      recs,
		Confidence:   0.85,
	}

	return resp, nil
}

// handlePlanAnalysis processes a plan generation request
func (h *AnalysisHandler) handlePlanAnalysis(ctx context.Context, req analysis.StandardAnalysisRequest) (*analysis.StandardAnalysisResponse, error) {
	// Check if we're generating a timeline or a full action plan
	generateTimeline, _ := req.Parameters["generate_timeline"].(bool)

	if generateTimeline {
		// Extract action plan and resources
		actionPlanRaw, ok := req.Data["action_plan"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("action_plan is required in data field")
		}

		// Convert to ActionPlan
		actionPlanBytes, err := json.Marshal(actionPlanRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal action plan: %w", err)
		}

		var actionPlan analysis.ActionPlan
		if err := json.Unmarshal(actionPlanBytes, &actionPlan); err != nil {
			return nil, fmt.Errorf("failed to unmarshal action plan: %w", err)
		}

		// Extract resources
		resources, ok := req.Data["resources"].(map[string]interface{})
		if !ok {
			resources = make(map[string]interface{})
		}

		// Generate timeline
		timeline, err := h.planner.GenerateTimeline(ctx, &actionPlan, resources)
		if err != nil {
			return nil, fmt.Errorf("failed to generate timeline: %w", err)
		}

		// Return standardized response
		resp := &analysis.StandardAnalysisResponse{
			AnalysisType: "plan",
			WorkflowID:   req.WorkflowID,
			Timestamp:    time.Now(),
			Results:      map[string]interface{}{"timeline": timeline},
			Confidence:   0.8,
		}

		return resp, nil
	} else {
		// Extract recommendations and constraints
		recsRaw, ok := req.Data["recommendations"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("recommendations are required in data field")
		}

		// Convert to RecommendationResponse
		recsBytes, err := json.Marshal(recsRaw)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal recommendations: %w", err)
		}

		var recommendations analysis.RecommendationResponse
		if err := json.Unmarshal(recsBytes, &recommendations); err != nil {
			return nil, fmt.Errorf("failed to unmarshal recommendations: %w", err)
		}

		// Extract constraints
		constraints, ok := req.Parameters["constraints"].(map[string]interface{})
		if !ok {
			constraints = make(map[string]interface{})
		}

		// Create action plan
		actionPlan, err := h.planner.CreateActionPlan(ctx, &recommendations, constraints)
		if err != nil {
			return nil, fmt.Errorf("failed to create action plan: %w", err)
		}

		// Return standardized response
		resp := &analysis.StandardAnalysisResponse{
			AnalysisType: "plan",
			WorkflowID:   req.WorkflowID,
			Timestamp:    time.Now(),
			Results:      actionPlan,
			Confidence:   0.85,
		}

		return resp, nil
	}
}

// Helper functions to extract values from parameters
func extractStringSlice(params map[string]interface{}, key string) ([]string, error) {
	if params == nil {
		return nil, fmt.Errorf("%s is required", key)
	}

	raw, ok := params[key]
	if !ok {
		return nil, fmt.Errorf("%s is required", key)
	}

	// Handle slice already in string format
	if strSlice, ok := raw.([]string); ok {
		return strSlice, nil
	}

	// Handle slice as interface array
	if ifaceSlice, ok := raw.([]interface{}); ok {
		result := make([]string, 0, len(ifaceSlice))
		for _, v := range ifaceSlice {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}

		if len(result) == 0 {
			return nil, fmt.Errorf("%s must contain strings", key)
		}

		return result, nil
	}

	return nil, fmt.Errorf("%s must be an array", key)
}

// Helper function to safely get string from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// handleChainAnalysis handles a request to perform a chained analysis workflow
func (h *AnalysisHandler) handleChainAnalysis(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		sendAnalysisError(w, "invalid_method", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request
	var req struct {
		WorkflowID string                 `json:"workflow_id,omitempty"`
		InputData  interface{}            `json:"input_data"`
		Config     map[string]interface{} `json:"config"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendAnalysisError(w, "invalid_request", fmt.Sprintf("Invalid request format: %s", err), http.StatusBadRequest)
		return
	}

	// Add timestamps to ensure we can trace the workflow
	timestamps := map[string]time.Time{
		"start_time": time.Now(),
	}

	// Perform the chained analysis
	result, err := h.analyzer.ChainAnalysis(r.Context(), req.InputData, req.Config)
	if err != nil {
		log.Printf("Error in chain analysis: %v", err)
		sendAnalysisError(w, "chain_analysis_error", err.Error(), http.StatusInternalServerError)
		return
	}

	// Add end timestamp
	timestamps["end_time"] = time.Now()

	// Create the response
	response := map[string]interface{}{
		"workflow_id": req.WorkflowID,
		"results":     result,
		"timestamps":  timestamps,
		"duration":    timestamps["end_time"].Sub(timestamps["start_time"]).String(),
	}

	// Save result to database if workflow ID is provided
	if req.WorkflowID != "" {
		resultID := uuid.New().String()
		resultsJSON, err := json.Marshal(result)
		if err != nil {
			log.Printf("Error marshaling results for storage: %v", err)
		} else {
			if err := db.SaveAnalysisResult(resultID, req.WorkflowID, "chain_analysis", string(resultsJSON)); err != nil {
				log.Printf("Error saving analysis result: %v", err)
			}
		}
	}

	// Return the response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
