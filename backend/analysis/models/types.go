package models

import "time"

// AnalysisRequest represents the data needed for various analysis functions
type AnalysisRequest struct {
	Text            string                 `json:"text"`
	Questions       []string               `json:"questions,omitempty"`
	FocusAreas      []string               `json:"focus_areas,omitempty"`
	PatternTypes    []string               `json:"pattern_types,omitempty"`
	AttributeValues map[string]interface{} `json:"attribute_values,omitempty"`
	BatchSize       *int                   `json:"batch_size,omitempty"`
}

// StandardAnalysisRequest represents a unified request structure for all analysis endpoints
type StandardAnalysisRequest struct {
	// Common fields
	WorkflowID string `json:"workflow_id,omitempty"`
	Text       string `json:"text,omitempty"`

	// Analysis-specific fields
	AnalysisType string                 `json:"analysis_type"`  // "trends", "patterns", "findings", "attributes", "intent", "recommendations", "plan"
	Parameters   map[string]interface{} `json:"parameters"`     // Analysis-specific parameters
	Data         map[string]interface{} `json:"data,omitempty"` // Input data for analysis
}

// AnalysisResponse represents a generic response from analysis methods
type AnalysisResponse struct {
	Results     interface{} `json:"results"`
	Confidence  float64     `json:"confidence,omitempty"`
	Explanation string      `json:"explanation,omitempty"`
	DataGaps    []string    `json:"data_gaps,omitempty"`
}

// StandardAnalysisResponse represents a unified response structure
type StandardAnalysisResponse struct {
	// Common fields
	AnalysisType string    `json:"analysis_type"`
	WorkflowID   string    `json:"workflow_id,omitempty"`
	Timestamp    time.Time `json:"timestamp"`

	// Results
	Results    interface{} `json:"results"`
	Confidence float64     `json:"confidence,omitempty"`

	// Metadata
	DataQuality struct {
		Assessment  string   `json:"assessment,omitempty"`
		Limitations []string `json:"limitations,omitempty"`
	} `json:"data_quality,omitempty"`

	// Error handling
	Error *AnalysisError `json:"error,omitempty"`
}

// AnalysisError represents error information
type AnalysisError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// AttributeDefinition represents a required data attribute
type AttributeDefinition struct {
	FieldName   string `json:"field_name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Rationale   string `json:"rationale,omitempty"`
}

// AttributeValue represents an extracted value for an attribute
type AttributeValue struct {
	FieldName   string  `json:"field_name"`
	Value       string  `json:"value"`
	Confidence  float64 `json:"confidence"`
	Explanation string  `json:"explanation,omitempty"`
	Label       string  `json:"label,omitempty"`
}

// IntentClassification represents intent classification results
type IntentClassification struct {
	LabelName   string `json:"label_name"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// AnalysisResult represents a persisted analysis result
type AnalysisResult struct {
	ID           string    `json:"id"`
	WorkflowID   string    `json:"workflow_id"`
	AnalysisType string    `json:"analysis_type"`
	Results      string    `json:"results"` // JSON string
	CreatedAt    time.Time `json:"created_at"`
}

// Recommendation represents a specific action recommendation
type Recommendation struct {
	Action         string `json:"action"`
	Rationale      string `json:"rationale"`
	ExpectedImpact string `json:"expected_impact"`
	Priority       int    `json:"priority"`
}

// RecommendationResponse represents a full set of recommendations
type RecommendationResponse struct {
	ImmediateActions    []Recommendation `json:"immediate_actions"`
	ImplementationNotes []string         `json:"implementation_notes"`
	SuccessMetrics      []string         `json:"success_metrics"`
}

// CriterionScore represents an evaluation score for a specific criterion
type CriterionScore struct {
	Criterion         string  `json:"criterion"`
	Score             float64 `json:"score"`
	Assessment        string  `json:"assessment"`
	ImprovementNeeded bool    `json:"improvement_needed"`
}

// ReviewResponse represents a complete review of analysis results
type ReviewResponse struct {
	CriteriaScores      []CriterionScore `json:"criteria_scores"`
	OverallQuality      OverallQuality   `json:"overall_quality"`
	PromptEffectiveness PromptFeedback   `json:"prompt_effectiveness,omitempty"`
}

// OverallQuality represents the overall assessment of analysis quality
type OverallQuality struct {
	Score      float64  `json:"score"`
	Strengths  []string `json:"strengths"`
	Weaknesses []string `json:"weaknesses"`
}

// PromptFeedback represents feedback on the prompt that generated the analysis
type PromptFeedback struct {
	Assessment            string   `json:"assessment"`
	SuggestedImprovements []string `json:"suggested_improvements"`
}

// ActionPlan represents a plan to implement recommendations
type ActionPlan struct {
	Goals              []string        `json:"goals"`
	ImmediateActions   []ActionItem    `json:"immediate_actions"`
	ShortTermActions   []ActionItem    `json:"short_term_actions"`
	LongTermActions    []ActionItem    `json:"long_term_actions"`
	ResponsibleParties []string        `json:"responsible_parties"`
	Timeline           []TimelineEvent `json:"timeline"`
	SuccessMetrics     []string        `json:"success_metrics"`
	RisksMitigations   []RiskItem      `json:"risks_mitigations"`
}

// ActionItem represents a specific action to be taken
type ActionItem struct {
	Action          string   `json:"action"`
	Description     string   `json:"description"`
	Priority        int      `json:"priority"`
	EstimatedEffort string   `json:"estimated_effort"`
	Dependencies    []string `json:"dependencies,omitempty"`
	ResponsibleRole string   `json:"responsible_role,omitempty"`
}

// TimelineEvent represents an event in the implementation timeline
type TimelineEvent struct {
	Phase       string   `json:"phase"`
	Description string   `json:"description"`
	Duration    string   `json:"duration"`
	Milestones  []string `json:"milestones"`
}

// RiskItem represents a risk and its mitigation strategy
type RiskItem struct {
	Risk             string `json:"risk"`
	Impact           string `json:"impact"`
	Probability      string `json:"probability"`
	MitigationPlan   string `json:"mitigation_plan"`
	ContingencyPlan  string `json:"contingency_plan,omitempty"`
	ResponsibleParty string `json:"responsible_party,omitempty"`
}

// RetentionStrategy represents strategies to improve customer retention
type RetentionStrategy struct {
	TargetSegment    string           `json:"target_segment,omitempty"`
	ImmediateActions []Recommendation `json:"immediate_actions"`
	ProcessChanges   []string         `json:"process_changes"`
	TrainingNeeds    []string         `json:"training_needs"`
	SuccessMetrics   []string         `json:"success_metrics"`
}
