package analysis

import (
	"context"
	"fmt"

	"agenticflows/backend/analysis/core"
	"agenticflows/backend/analysis/models"
	"agenticflows/backend/analysis/processors"
)

// AnalysisFacade provides a unified interface to the analysis capabilities
type AnalysisFacade struct {
	Analyzer                 *core.Analyzer
	TrendsAnalyzer           *processors.TrendsAnalyzer
	PatternsAnalyzer         *processors.PatternsAnalyzer
	TextProcessor            *processors.TextProcessor
	RecommendationsProcessor *processors.RecommendationsProcessor
	PlannerProcessor         *processors.PlannerProcessor
}

// NewAnalysisFacade creates a new AnalysisFacade
func NewAnalysisFacade(apiKey string, debug bool) (*AnalysisFacade, error) {
	// Create the base analyzer
	analyzer, err := core.NewAnalyzer(apiKey, debug)
	if err != nil {
		return nil, fmt.Errorf("failed to create analyzer: %w", err)
	}

	// Create specialized processors
	trendsAnalyzer := processors.NewTrendsAnalyzer(analyzer)
	patternsAnalyzer := processors.NewPatternsAnalyzer(analyzer)
	textProcessor := processors.NewTextProcessor(analyzer)
	recommendationsProcessor := processors.NewRecommendationsProcessor(analyzer)
	plannerProcessor := processors.NewPlannerProcessor(analyzer)

	return &AnalysisFacade{
		Analyzer:                 analyzer,
		TrendsAnalyzer:           trendsAnalyzer,
		PatternsAnalyzer:         patternsAnalyzer,
		TextProcessor:            textProcessor,
		RecommendationsProcessor: recommendationsProcessor,
		PlannerProcessor:         plannerProcessor,
	}, nil
}

// AnalyzeTrends analyzes trends in conversation data
func (f *AnalysisFacade) AnalyzeTrends(ctx context.Context, req models.AnalysisRequest) (*models.AnalysisResponse, error) {
	return f.TrendsAnalyzer.AnalyzeTrends(ctx, req)
}

// IdentifyPatterns identifies patterns in conversation data
func (f *AnalysisFacade) IdentifyPatterns(ctx context.Context, req models.AnalysisRequest) (*models.AnalysisResponse, error) {
	return f.PatternsAnalyzer.IdentifyPatterns(ctx, req)
}

// GenerateRequiredAttributes generates required attributes for answering questions
func (f *AnalysisFacade) GenerateRequiredAttributes(ctx context.Context, questions []string, existingAttributes []string) ([]models.AttributeDefinition, error) {
	return f.TextProcessor.GenerateRequiredAttributes(ctx, questions, existingAttributes)
}

// GenerateAttributes extracts attribute values from text
func (f *AnalysisFacade) GenerateAttributes(ctx context.Context, text string, attributes []models.AttributeDefinition) ([]models.AttributeValue, error) {
	return f.TextProcessor.GenerateAttributes(ctx, text, attributes)
}

// GenerateIntent generates the intent classification for a conversation
func (f *AnalysisFacade) GenerateIntent(ctx context.Context, text string) (*models.IntentClassification, error) {
	return f.TextProcessor.GenerateIntent(ctx, text)
}

// GenerateRecommendations generates recommendations based on analysis results
func (f *AnalysisFacade) GenerateRecommendations(ctx context.Context, analysisResults map[string]interface{}, focusArea string) (*models.RecommendationResponse, error) {
	return f.RecommendationsProcessor.GenerateRecommendations(ctx, analysisResults, focusArea)
}

// PrioritizeRecommendations prioritizes recommendations based on criteria
func (f *AnalysisFacade) PrioritizeRecommendations(ctx context.Context, recommendations []models.Recommendation, criteria map[string]float64) ([]models.Recommendation, error) {
	return f.RecommendationsProcessor.PrioritizeRecommendations(ctx, recommendations, criteria)
}

// GenerateRetentionStrategies generates strategies for improving customer retention
func (f *AnalysisFacade) GenerateRetentionStrategies(ctx context.Context, analysisResults map[string]interface{}) (*models.RetentionStrategy, error) {
	return f.RecommendationsProcessor.GenerateRetentionStrategies(ctx, analysisResults)
}

// CreateActionPlan creates an implementation plan based on recommendations
func (f *AnalysisFacade) CreateActionPlan(ctx context.Context, recommendations *models.RecommendationResponse, constraints map[string]interface{}) (*models.ActionPlan, error) {
	return f.PlannerProcessor.CreateActionPlan(ctx, recommendations, constraints)
}

// GenerateTimeline generates an implementation timeline for an action plan
func (f *AnalysisFacade) GenerateTimeline(ctx context.Context, actionPlan *models.ActionPlan, resources map[string]interface{}) ([]models.TimelineEvent, error) {
	return f.PlannerProcessor.GenerateTimeline(ctx, actionPlan, resources)
}

// ChainAnalysis performs a chain of analyses
func (f *AnalysisFacade) ChainAnalysis(ctx context.Context, inputData interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return f.Analyzer.ChainAnalysis(ctx, inputData, config)
}

// TransformForTrends prepares data for trend analysis
func (f *AnalysisFacade) TransformForTrends(data interface{}) (map[string]interface{}, error) {
	return f.Analyzer.TransformForTrends(data)
}

// TransformForPatterns prepares data for pattern identification
func (f *AnalysisFacade) TransformForPatterns(data interface{}, patternTypes []string) (map[string]interface{}, error) {
	return f.Analyzer.TransformForPatterns(data, patternTypes)
}

// ExtractTrendsOutput extracts the most relevant information from trends analysis
func (f *AnalysisFacade) ExtractTrendsOutput(resp *models.AnalysisResponse) (map[string]interface{}, error) {
	return f.TrendsAnalyzer.ExtractTrendsOutput(resp)
}

// ExtractPatternsOutput extracts and simplifies patterns from the analysis
func (f *AnalysisFacade) ExtractPatternsOutput(resp *models.AnalysisResponse) ([]string, error) {
	return f.PatternsAnalyzer.ExtractPatternsOutput(resp)
}

// LegacyAnalyzer provides backward compatibility with existing code
// IMPORTANT: This replaces the original Analyzer struct after refactoring.
// Existing code should migrate from using Analyzer to using AnalysisFacade directly.
type LegacyAnalyzer struct {
	facade *AnalysisFacade
}

// NewLegacyAnalyzer creates a new LegacyAnalyzer for backward compatibility
// Note: For backward compatibility, handlers and other packages should be updated
// to use this method instead of NewAnalyzer (which is removed).
// For new development, use NewAnalysisFacade directly.
func NewLegacyAnalyzer(apiKey string, debug bool) (*LegacyAnalyzer, error) {
	facade, err := NewAnalysisFacade(apiKey, debug)
	if err != nil {
		return nil, err
	}
	return &LegacyAnalyzer{
		facade: facade,
	}, nil
}

// AnalyzeTrends analyzes trends (backward compatibility method)
func (a *LegacyAnalyzer) AnalyzeTrends(ctx context.Context, req models.AnalysisRequest) (*models.AnalysisResponse, error) {
	return a.facade.AnalyzeTrends(ctx, req)
}

// IdentifyPatterns identifies patterns (backward compatibility method)
func (a *LegacyAnalyzer) IdentifyPatterns(ctx context.Context, req models.AnalysisRequest) (*models.AnalysisResponse, error) {
	return a.facade.IdentifyPatterns(ctx, req)
}

// AnalyzeFindings analyzes findings from attribute extraction (backward compatibility)
func (a *LegacyAnalyzer) AnalyzeFindings(ctx context.Context, req models.AnalysisRequest) (*models.AnalysisResponse, error) {
	// This is a stub that needs to be implemented when migrating the findings analyzer
	return nil, fmt.Errorf("AnalyzeFindings not implemented in refactored version yet")
}

// GenerateRequiredAttributes generates required attributes (backward compatibility)
func (a *LegacyAnalyzer) GenerateRequiredAttributes(ctx context.Context, questions []string, existingAttributes []string) ([]models.AttributeDefinition, error) {
	return a.facade.GenerateRequiredAttributes(ctx, questions, existingAttributes)
}

// GenerateAttributes extracts attribute values (backward compatibility)
func (a *LegacyAnalyzer) GenerateAttributes(ctx context.Context, text string, attributes []models.AttributeDefinition) ([]models.AttributeValue, error) {
	return a.facade.GenerateAttributes(ctx, text, attributes)
}

// GenerateIntent classifies intent (backward compatibility)
func (a *LegacyAnalyzer) GenerateIntent(ctx context.Context, text string) (*models.IntentClassification, error) {
	return a.facade.GenerateIntent(ctx, text)
}

// GenerateRecommendations generates recommendations (backward compatibility)
func (a *LegacyAnalyzer) GenerateRecommendations(ctx context.Context, analysisResults map[string]interface{}, focusArea string) (*models.RecommendationResponse, error) {
	return a.facade.GenerateRecommendations(ctx, analysisResults, focusArea)
}

// CreateActionPlan creates an action plan (backward compatibility)
func (a *LegacyAnalyzer) CreateActionPlan(ctx context.Context, recommendations *models.RecommendationResponse, constraints map[string]interface{}) (*models.ActionPlan, error) {
	return a.facade.CreateActionPlan(ctx, recommendations, constraints)
}

// TransformForTrends prepares data for trend analysis (backward compatibility)
func (a *LegacyAnalyzer) TransformForTrends(data interface{}) (map[string]interface{}, error) {
	return a.facade.TransformForTrends(data)
}

// TransformForPatterns prepares data for pattern identification (backward compatibility)
func (a *LegacyAnalyzer) TransformForPatterns(data interface{}, patternTypes []string) (map[string]interface{}, error) {
	return a.facade.TransformForPatterns(data, patternTypes)
}

// ExtractTrendsOutput extracts the most relevant information from trends analysis (backward compatibility)
func (a *LegacyAnalyzer) ExtractTrendsOutput(resp *models.AnalysisResponse) (map[string]interface{}, error) {
	return a.facade.ExtractTrendsOutput(resp)
}

// ExtractPatternsOutput extracts and simplifies patterns from the analysis (backward compatibility)
func (a *LegacyAnalyzer) ExtractPatternsOutput(resp *models.AnalysisResponse) ([]string, error) {
	return a.facade.ExtractPatternsOutput(resp)
}

// ChainAnalysis performs a chain of analyses (backward compatibility)
func (a *LegacyAnalyzer) ChainAnalysis(ctx context.Context, inputData interface{}, config map[string]interface{}) (map[string]interface{}, error) {
	return a.facade.ChainAnalysis(ctx, inputData, config)
}

// ProcessInBatches processes items in batches (backward compatibility)
func (a *LegacyAnalyzer) ProcessInBatches(ctx context.Context, items []interface{}, batchSize int, processFunc func(interface{}) (interface{}, error)) ([]interface{}, error) {
	return a.facade.Analyzer.ProcessInBatches(ctx, items, batchSize, processFunc)
}

// IMPORTANT: To maintain backward compatibility, the original NewAnalyzer function has been
// replaced with NewLegacyAnalyzer. Code that currently uses the original NewAnalyzer function
// should be updated to use the new function with the same signature.
//
// For example, in backend/api/handlers/analysis_base.go:
//   Change:
//     analyzer, err := analysis.NewAnalyzer(apiKey, false)
//   To:
//     analyzer, err := analysis.NewLegacyAnalyzer(apiKey, false)
