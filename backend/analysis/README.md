# Contact Center Analysis Package

This package provides Go implementations of contact center analysis functionality, ported from the Python library. It leverages Google's Gemini LLM API to analyze customer service conversations, extract attributes, identify intents, and generate insights.

## Features

- **Text Analysis**
  - Extract attributes from conversations
  - Identify customer intents
  - Generate required attributes for specific research questions

- **Data Analysis**
  - Analyze trends in conversation data
  - Identify conversation patterns
  - Analyze findings from attribute extraction

- **Recommendations & Planning**
  - Generate actionable recommendations based on analysis results
  - Prioritize recommendations based on custom criteria
  - Create comprehensive action plans from recommendations
  - Generate retention strategies for customer retention
  - Create implementation timelines with dependencies and milestones

- **API Integration**
  - RESTful API for all analysis functions
  - Rate-limited Gemini API client
  - Database storage for analysis results

## API Endpoints

### Standardized API (Recommended)

- `/api/analysis` - Unified endpoint for all analysis operations using a standardized request/response format

### Legacy Endpoints

- `/api/analysis/trends` - Analyze trends in conversation data
- `/api/analysis/patterns` - Identify patterns in conversation data
- `/api/analysis/findings` - Analyze findings from attribute extraction
- `/api/analysis/attributes` - Extract or generate attributes from text
- `/api/analysis/intent` - Identify the primary intent in a conversation
- `/api/analysis/results` - Retrieve stored analysis results

## Standardized API Usage

The standardized API uses a unified request/response format to make analysis operations more consistent and chainable.

### Request Format

```json
{
  "workflow_id": "optional-workflow-id",
  "text": "Conversation text if applicable",
  "analysis_type": "trends|patterns|findings|attributes|intent|recommendations|plan",
  "parameters": {
    // Analysis-specific parameters
  },
  "data": {
    // Input data for analysis (often results from previous steps)
  }
}
```

### Response Format

```json
{
  "analysis_type": "trends|patterns|findings|attributes|intent|recommendations|plan",
  "workflow_id": "workflow-id-if-provided",
  "timestamp": "2023-06-15T10:30:45Z",
  "results": {
    // Analysis results (structure depends on analysis_type)
  },
  "confidence": 0.95,
  "data_quality": {
    "assessment": "Good quality data with minor issues",
    "limitations": ["Limited sample size", "Missing customer demographic data"]
  },
  "error": null // Present only if there was an error
}
```

### Example: Chaining Analysis Operations

```go
// Step 1: Extract attributes
attributesReq := StandardAnalysisRequest{
    AnalysisType: "attributes",
    Text:         conversationText,
    Parameters: map[string]interface{}{
        "attributes": []map[string]string{
            {
                "field_name":  "sentiment",
                "title":       "Customer Sentiment",
                "description": "The sentiment expressed by the customer",
            },
            // More attributes...
        },
    },
}
attrResp, err := client.PerformAnalysis(attributesReq)

// Step 2: Analyze trends using extracted attributes
trendsReq := StandardAnalysisRequest{
    AnalysisType: "trends",
    Parameters: map[string]interface{}{
        "focus_areas": []string{"sentiment", "issue_resolution"},
    },
    Data: attrResp.Results,
}
trendsResp, err := client.PerformAnalysis(trendsReq)
```

### Recommendations & Planning Examples

```go
// Generate recommendations based on analysis results
recReq := StandardAnalysisRequest{
    AnalysisType: "recommendations",
    Parameters: map[string]interface{}{
        "focus_area": "customer retention",
        "criteria": map[string]float64{
            "impact": 0.6,
            "implementation_ease": 0.4,
        },
    },
    Data: analysisResults,
}
recResp, err := client.PerformAnalysis(recReq)

// Create an action plan from recommendations
planReq := StandardAnalysisRequest{
    AnalysisType: "plan",
    Parameters: map[string]interface{}{
        "constraints": map[string]interface{}{
            "budget": 50000,
            "timeline": "6 months",
            "resources": ["customer_support", "engineering", "marketing"],
        },
    },
    Data: map[string]interface{}{
        "recommendations": recResp.Results,
    },
}
planResp, err := client.PerformAnalysis(planReq)

// Generate a timeline for the action plan
timelineReq := StandardAnalysisRequest{
    AnalysisType: "plan",
    Parameters: map[string]interface{}{
        "generate_timeline": true,
    },
    Data: map[string]interface{}{
        "action_plan": planResp.Results,
        "resources": map[string]interface{}{
            "staff": 5,
            "start_date": "2023-07-01",
        },
    },
}
timelineResp, err := client.PerformAnalysis(timelineReq)
```

## Requirements

- Go 1.21+
- Google Gemini API key (set as `GEMINI_API_KEY` environment variable)
- SQLite database

## Usage

### Command Line Testing

Use the included test clients to test the API:

```bash
# Test the standardized API
go run cmd/examples/standardized_client.go -text "I want to cancel my subscription" -debug

# Test legacy endpoints
go run cmd/testclient/main.go -endpoint intent -text "I want to cancel my subscription"

# Process a file
go run cmd/testclient/main.go -endpoint attributes -file ./sample.txt

# View results for a workflow
go run cmd/testclient/main.go -endpoint results -workflow abc123
```

### Library Usage

```go
import "agenticflows/backend/analysis"

// Create a text generator
textGen, err := analysis.NewTextGenerator(apiKey, false)
if err != nil {
    log.Fatalf("Failed to create text generator: %v", err)
}

// Generate intent from text
intent, err := textGen.GenerateIntent(ctx, conversationText)
if err != nil {
    log.Fatalf("Failed to generate intent: %v", err)
}
fmt.Printf("Intent: %s (%s)\n", intent.LabelName, intent.Label)
fmt.Printf("Description: %s\n", intent.Description)

// Create a recommendation engine
recEngine, err := analysis.NewRecommendationEngine(apiKey, false)
if err != nil {
    log.Fatalf("Failed to create recommendation engine: %v", err)
}

// Generate recommendations
recs, err := recEngine.GenerateRecommendations(ctx, analysisResults, "customer satisfaction")
if err != nil {
    log.Fatalf("Failed to generate recommendations: %v", err)
}

// Create a planner
planner, err := analysis.NewPlanner(apiKey, false)
if err != nil {
    log.Fatalf("Failed to create planner: %v", err)
}

// Create an action plan
actionPlan, err := planner.CreateActionPlan(ctx, recs, constraints)
if err != nil {
    log.Fatalf("Failed to create action plan: %v", err)
}
```

## Architecture

The package is organized into several components:

- `models.go` - Data structures for requests and responses
- `llm.go` - Gemini API integration
- `text.go` - Text analysis functions
- `analyzer.go` - Core analysis functions
- `recommend.go` - Recommendation generation functions
- `plan.go` - Action planning and timeline functions
- `ratelimiter.go` - API rate limiting

## Database Integration

Analysis results are stored in the SQLite database using the `analysis_results` table. Results can be retrieved by ID or by workflow ID. 