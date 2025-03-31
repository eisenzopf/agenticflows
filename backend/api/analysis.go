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

	// Create and return the analysis handler
	// Note: The API supports mock data mode through the "use_mock_data" parameter
	// When this parameter is set to true in API requests to /api/analysis,
	// predefined mock responses will be returned instead of making LLM API calls.
	// This is useful for testing, demonstrations, and development environments.
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

	// Log the analysis type for debugging
	log.Printf("Received analysis request with type: %s", req.AnalysisType)

	// Convert analysis type to lowercase for case-insensitive matching
	analysisType := strings.ToLower(req.AnalysisType)
	log.Printf("Using normalized analysis type: %s", analysisType)

	// Route to appropriate analysis function based on type
	var resp *analysis.StandardAnalysisResponse
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
	log.Printf("Processing recommendations analysis request")

	// Check if recommendationEngine is properly initialized
	if h.recommendationEngine == nil {
		return nil, fmt.Errorf("recommendation engine is not initialized")
	}

	// Extract parameters
	focusArea, _ := req.Parameters["focus_area"].(string)
	if focusArea == "" {
		// Default to a general focus area if not specified
		focusArea = "improving customer experience"
	}

	// Check for prioritization criteria
	hasCriteria := false
	if _, ok := req.Parameters["criteria"].(map[string]interface{}); ok {
		hasCriteria = true
	}

	// Check if mock data should be used
	// 'use_mock_data' is an optional boolean parameter that, when set to true,
	// will return predefined mock data instead of making an actual API call.
	// This is useful for testing, demonstrations, and environments where the LLM API is unavailable.
	useMockData := false
	if mockParam, ok := req.Parameters["use_mock_data"].(bool); ok {
		useMockData = mockParam
	}

	var recs *analysis.RecommendationResponse

	if useMockData {
		// Generate recommendations - use mock data for testing
		log.Printf("Using mock recommendations for testing (focus area: %s)", focusArea)
		recs = &analysis.RecommendationResponse{
			ImmediateActions: []analysis.Recommendation{
				{
					Action:         "Implement callback option",
					Rationale:      "Reduces customer frustration during peak times",
					ExpectedImpact: "15% reduction in call abandonment rate",
					Priority:       5,
				},
				{
					Action:         "Add self-service order tracking",
					Rationale:      "Customers frequently check order status",
					ExpectedImpact: "25% reduction in status-related calls",
					Priority:       4,
				},
				{
					Action:         "Improve post-purchase email communication",
					Rationale:      "Customers need clearer delivery information",
					ExpectedImpact: "10% reduction in delivery-related inquiries",
					Priority:       3,
				},
			},
			ImplementationNotes: []string{
				"Focus on mobile-friendly interfaces",
				"Ensure integration with existing CRM system",
				"Provide comprehensive training for support staff",
			},
			SuccessMetrics: []string{
				"Reduction in call volume for routine inquiries",
				"Improvement in customer satisfaction scores",
				"Increase in first-call resolution rate",
			},
		}

		// If prioritization criteria are provided, prioritize the recommendations with mock data
		if hasCriteria && len(recs.ImmediateActions) > 0 {
			log.Printf("Using mock prioritized recommendations for testing")
			// We'll just pretend to reprioritize by making slight adjustments
			recs.ImmediateActions[0].Priority = 5
			recs.ImmediateActions[1].Priority = 4
			recs.ImmediateActions[2].Priority = 3
		}
	} else {
		// Use real API for production
		// Convert to the format expected by the recommendation engine
		var data map[string]interface{}
		if req.Data != nil {
			data = req.Data
		} else {
			data = make(map[string]interface{})
		}

		// Generate recommendations
		var err error
		recs, err = h.recommendationEngine.GenerateRecommendations(ctx, data, focusArea)
		if err != nil {
			return nil, fmt.Errorf("failed to generate recommendations: %w", err)
		}

		// If prioritization criteria are provided, prioritize the recommendations
		if hasCriteria && len(recs.ImmediateActions) > 0 {
			// Extract prioritization criteria
			criteriaRaw, _ := req.Parameters["criteria"].(map[string]interface{})

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
				prioritizedRecs, prErr := h.recommendationEngine.PrioritizeRecommendations(ctx, recs.ImmediateActions, criteria)
				if prErr == nil {
					recs.ImmediateActions = prioritizedRecs
				} else {
					log.Printf("Warning: Failed to prioritize recommendations: %v", prErr)
				}
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
	log.Printf("Processing plan analysis request")

	// Check if planner is properly initialized
	if h.planner == nil {
		return nil, fmt.Errorf("planner is not initialized")
	}

	// Check if we're generating a timeline or a full action plan
	generateTimeline, _ := req.Parameters["generate_timeline"].(bool)

	// Check if mock data should be used
	// 'use_mock_data' is an optional boolean parameter that, when set to true,
	// will return predefined mock data instead of making an actual API call.
	// This is useful for testing, demonstrations, and environments where the LLM API is unavailable.
	useMockData := false
	if mockParam, ok := req.Parameters["use_mock_data"].(bool); ok {
		useMockData = mockParam
	}

	if generateTimeline {
		log.Printf("Generating timeline for action plan")

		// Verify that action plan exists in the request data
		if _, ok := req.Data["action_plan"]; !ok {
			return nil, fmt.Errorf("action_plan is required in data field")
		}

		var timeline []analysis.TimelineEvent

		if useMockData {
			// Use mock timeline for testing
			log.Printf("Using mock timeline for testing")
			timeline = []analysis.TimelineEvent{
				{
					Phase:       "Phase 1: Initial Implementation",
					Description: "Set up the basic infrastructure for the callback system",
					Duration:    "2 weeks",
					Milestones:  []string{"Backend API setup", "Database schema design", "Basic UI mockups"},
				},
				{
					Phase:       "Phase 2: Development",
					Description: "Develop the callback functionality and integrate with existing systems",
					Duration:    "4 weeks",
					Milestones:  []string{"Backend development", "Frontend integration", "Unit testing"},
				},
				{
					Phase:       "Phase 3: Testing and Deployment",
					Description: "Test the system and roll out to production",
					Duration:    "2 weeks",
					Milestones:  []string{"QA testing", "User acceptance testing", "Production deployment"},
				},
			}
		} else {
			// Use real API for production
			// Extract action plan and resources
			actionPlanRaw, _ := req.Data["action_plan"].(map[string]interface{})

			// Convert to ActionPlan
			actionPlanBytes, marshalErr := json.Marshal(actionPlanRaw)
			if marshalErr != nil {
				return nil, fmt.Errorf("failed to marshal action plan: %w", marshalErr)
			}

			var actionPlan analysis.ActionPlan
			if unmarshalErr := json.Unmarshal(actionPlanBytes, &actionPlan); unmarshalErr != nil {
				return nil, fmt.Errorf("failed to unmarshal action plan: %w", unmarshalErr)
			}

			// Extract resources
			resources, ok := req.Data["resources"].(map[string]interface{})
			if !ok {
				resources = make(map[string]interface{})
			}

			// Generate timeline
			var timelineErr error
			timeline, timelineErr = h.planner.GenerateTimeline(ctx, &actionPlan, resources)
			if timelineErr != nil {
				return nil, fmt.Errorf("failed to generate timeline: %w", timelineErr)
			}
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
		// Extract recommendations
		if _, ok := req.Data["recommendations"]; !ok {
			return nil, fmt.Errorf("recommendations are required in data field")
		}

		var actionPlan *analysis.ActionPlan

		if useMockData {
			// Use mock action plan for testing
			log.Printf("Using mock action plan for testing")
			actionPlan = &analysis.ActionPlan{
				Goals: []string{
					"Improve customer retention rates",
					"Reduce call center wait times",
					"Increase customer satisfaction scores",
				},
				ImmediateActions: []analysis.ActionItem{
					{
						Action:          "Implement callback option",
						Description:     "Add callback feature for customers on hold",
						Priority:        5,
						EstimatedEffort: "2 weeks",
						ResponsibleRole: "Engineering",
					},
					{
						Action:          "Train agents on new retention offers",
						Description:     "Provide comprehensive training on new retention policies",
						Priority:        4,
						EstimatedEffort: "1 week",
						ResponsibleRole: "Training",
					},
				},
				ShortTermActions: []analysis.ActionItem{
					{
						Action:          "Develop self-service order tracking",
						Description:     "Create web and mobile interfaces for order status tracking",
						Priority:        4,
						EstimatedEffort: "1 month",
						ResponsibleRole: "Engineering",
						Dependencies:    []string{"Upgrade backend API"},
					},
				},
				LongTermActions: []analysis.ActionItem{
					{
						Action:          "Implement AI-powered assistance",
						Description:     "Develop AI chatbot for common inquiries",
						Priority:        3,
						EstimatedEffort: "3 months",
						ResponsibleRole: "Engineering",
					},
				},
				ResponsibleParties: []string{
					"Customer Support",
					"Engineering",
					"Training",
					"Marketing",
				},
				Timeline: []analysis.TimelineEvent{
					{
						Phase:       "Phase 1: Immediate Improvements",
						Description: "Focus on quick wins with high impact",
						Duration:    "1 month",
						Milestones:  []string{"Callback feature launch", "Agent training complete"},
					},
					{
						Phase:       "Phase 2: System Enhancements",
						Description: "Roll out system improvements and self-service features",
						Duration:    "2 months",
						Milestones:  []string{"Self-service tracking launch", "Knowledge base update"},
					},
				},
				SuccessMetrics: []string{
					"15% reduction in call abandonment rate",
					"10% increase in customer satisfaction scores",
					"20% reduction in routine inquiry calls",
				},
				RisksMitigations: []analysis.RiskItem{
					{
						Risk:           "Technical integration issues",
						Impact:         "High",
						Probability:    "Medium",
						MitigationPlan: "Comprehensive testing and phased rollout",
					},
					{
						Risk:           "Agent adoption resistance",
						Impact:         "Medium",
						Probability:    "Low",
						MitigationPlan: "Early involvement and feedback collection",
					},
				},
			}
		} else {
			// Use real API for production
			// Convert to RecommendationResponse
			recsRaw, _ := req.Data["recommendations"].(map[string]interface{})
			recsBytes, marshalErr := json.Marshal(recsRaw)
			if marshalErr != nil {
				return nil, fmt.Errorf("failed to marshal recommendations: %w", marshalErr)
			}

			var recommendations analysis.RecommendationResponse
			if unmarshalErr := json.Unmarshal(recsBytes, &recommendations); unmarshalErr != nil {
				return nil, fmt.Errorf("failed to unmarshal recommendations: %w", unmarshalErr)
			}

			// Extract constraints
			constraints, ok := req.Parameters["constraints"].(map[string]interface{})
			if !ok {
				constraints = make(map[string]interface{})
			}

			// Create action plan
			var planErr error
			actionPlan, planErr = h.planner.CreateActionPlan(ctx, &recommendations, constraints)
			if planErr != nil {
				return nil, fmt.Errorf("failed to create action plan: %w", planErr)
			}
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

// FunctionMetadata represents the metadata for an analysis function
type FunctionMetadata struct {
	ID          string                 `json:"id"`
	Label       string                 `json:"label"`
	Description string                 `json:"description"`
	Inputs      []ParameterDefinition  `json:"inputs"`
	Outputs     []OutputDefinition     `json:"outputs"`
	Example     map[string]interface{} `json:"example,omitempty"`
}

type ParameterDefinition struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
}

type OutputDefinition struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

func (h *AnalysisHandler) handleGetFunctionMetadata(w http.ResponseWriter, r *http.Request) {
	metadata := map[string]FunctionMetadata{
		"trends": {
			ID:          "analysis-trends",
			Label:       "Analyze Trends",
			Description: "Analyze trends in conversation data",
			Inputs: []ParameterDefinition{
				{
					Name:        "Focus Areas",
					Path:        "parameters.focus_areas",
					Description: "Areas to focus trend analysis on",
					Required:    true,
					Type:        "string[]",
				},
				{
					Name:        "Historical Data",
					Path:        "data.historical_data",
					Description: "Historical data for trend analysis",
					Required:    true,
					Type:        "object",
				},
			},
			Outputs: []OutputDefinition{
				{
					Name:        "Trends",
					Path:        "results.trends",
					Description: "Identified trends and patterns",
					Type:        "object[]",
				},
				{
					Name:        "Metrics",
					Path:        "results.metrics",
					Description: "Trend metrics and statistics",
					Type:        "object",
				},
			},
		},
		"patterns": {
			ID:          "analysis-patterns",
			Label:       "Identify Patterns",
			Description: "Identify patterns in conversation data",
			Inputs: []ParameterDefinition{
				{
					Name:        "Pattern Types",
					Path:        "parameters.pattern_types",
					Description: "Types of patterns to identify",
					Required:    true,
					Type:        "string[]",
				},
				{
					Name:        "Sample Data",
					Path:        "data.sample_data",
					Description: "Data samples to analyze",
					Required:    true,
					Type:        "object[]",
				},
			},
			Outputs: []OutputDefinition{
				{
					Name:        "Patterns",
					Path:        "results.patterns",
					Description: "Identified patterns",
					Type:        "object[]",
				},
				{
					Name:        "Categories",
					Path:        "results.categories",
					Description: "Pattern categories",
					Type:        "string[]",
				},
			},
		},
		// Add other function metadata...
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}
