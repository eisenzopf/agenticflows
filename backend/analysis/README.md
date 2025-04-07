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

### Example: Intent Analysis

```go
// Get intent from a customer conversation
intentReq := StandardAnalysisRequest{
    AnalysisType: "intent",
    Text:         "I've been charged twice for my last payment and need this fixed immediately.",
    Parameters:   map[string]interface{}{},
}
intentResp, err := client.PerformAnalysis(intentReq)

// Access intent analysis results
if intentResults, ok := intentResp.Results.(map[string]interface{}); ok {
    fmt.Printf("Intent: %s\n", intentResults["intent"])
    fmt.Printf("Confidence: %.2f\n", intentResp.Confidence)
    fmt.Printf("Explanation: %s\n", intentResults["explanation"])
}
```

```bash
# Test intent analysis with curl (simple version for copy-paste)
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-intent-123","analysis_type":"intent","text":"I have been charged twice for my last payment and need this fixed immediately.","parameters":{}}'
```

### Example: Attribute Extraction

```go
// Extract specific attributes from a conversation
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
            {
                "field_name":  "issue_type",
                "title":       "Issue Type",
                "description": "The category of the customer's issue",
            },
            {
                "field_name":  "resolution_status",
                "title":       "Resolution Status",
                "description": "Whether the issue was resolved during the interaction",
            },
        },
    },
}
attrResp, err := client.PerformAnalysis(attributesReq)

// Access attribute values
if attrResults, ok := attrResp.Results.(map[string]interface{}); ok {
    if values, ok := attrResults["attribute_values"].([]interface{}); ok {
        for _, v := range values {
            if attr, ok := v.(map[string]interface{}); ok {
                fmt.Printf("Attribute: %s\n", attr["field_name"])
                fmt.Printf("Value: %s\n", attr["value"])
                fmt.Printf("Confidence: %.2f\n", attr["confidence"])
            }
        }
    }
}
```

```bash
# Test attribute extraction with curl (simple version for copy-paste)
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-attrs-123","analysis_type":"attributes","text":"Customer: I am having issues with my latest bill. The charges are wrong.\nAgent: I understand your frustration. Let me help resolve this billing issue.","parameters":{"attributes":[{"field_name":"sentiment","title":"Customer Sentiment","description":"The sentiment expressed by the customer"},{"field_name":"issue_type","title":"Issue Type","description":"The category of the customer issue"},{"field_name":"resolution_status","title":"Resolution Status","description":"Whether the issue was resolved during the interaction"}]}}'
```

### Example: Trend Analysis

```go
// Analyze trends in conversation data
trendsReq := StandardAnalysisRequest{
    AnalysisType: "trends",
    Parameters: map[string]interface{}{
        "focus_areas": []string{"sentiment", "issue_resolution", "response_time"},
        "time_period": "last_quarter",
    },
    Data: conversationData,
}
trendsResp, err := client.PerformAnalysis(trendsReq)

// Access trend analysis results
if trendResults, ok := trendsResp.Results.(map[string]interface{}); ok {
    if trends, ok := trendResults["trends"].([]interface{}); ok {
        for _, t := range trends {
            if trend, ok := t.(map[string]interface{}); ok {
                fmt.Printf("Trend: %s\n", trend["name"])
                fmt.Printf("Description: %s\n", trend["description"])
                fmt.Printf("Significance: %.2f\n", trend["significance"])
            }
        }
    }
}
```

```bash
# Test trend analysis with curl (simple version for copy-paste)
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-trends-123","analysis_type":"trends","parameters":{"focus_areas":["sentiment","issue_resolution","response_time"],"time_period":"last_quarter"},"data":{"conversation_metrics":[{"id":"conv1","timestamp":"2023-01-15T10:30:00Z","sentiment":"negative","issue_resolution":"unresolved","response_time":480},{"id":"conv2","timestamp":"2023-01-16T11:45:00Z","sentiment":"neutral","issue_resolution":"resolved","response_time":320},{"id":"conv3","timestamp":"2023-01-18T14:20:00Z","sentiment":"positive","issue_resolution":"resolved","response_time":180}]}}'
```

### Example: Pattern Identification

```go
// Identify patterns in conversation data
patternsReq := StandardAnalysisRequest{
    AnalysisType: "patterns",
    Parameters: map[string]interface{}{
        "pattern_types": []string{"agent_behavior", "escalation_triggers"},
        "min_confidence": 0.7,
    },
    Data: conversationData,
}
patternsResp, err := client.PerformAnalysis(patternsReq)

// Access identified patterns
if patternResults, ok := patternsResp.Results.(map[string]interface{}); ok {
    if patterns, ok := patternResults["patterns"].([]interface{}); ok {
        for _, p := range patterns {
            if pattern, ok := p.(map[string]interface{}); ok {
                fmt.Printf("Pattern: %s\n", pattern["name"])
                fmt.Printf("Description: %s\n", pattern["description"])
                fmt.Printf("Frequency: %.2f\n", pattern["frequency"])
                fmt.Printf("Impact: %s\n", pattern["impact"])
            }
        }
    }
}
```

```bash
# Test pattern identification with curl (simple version for copy-paste)
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-patterns-123","analysis_type":"patterns","parameters":{"pattern_types":["agent_behavior","escalation_triggers"],"min_confidence":0.7},"data":{"conversations":[{"id":"c1","text":"Customer: I am really frustrated with this service.\nAgent: I understand your frustration. Let me escalate this to my supervisor.","attributes":{"escalated":true,"customer_sentiment":"negative"}},{"id":"c2","text":"Customer: This charge seems incorrect.\nAgent: I will look into that for you right away.","attributes":{"escalated":false,"customer_sentiment":"neutral"}},{"id":"c3","text":"Customer: I have been waiting for weeks!\nAgent: Let me transfer you to my manager to help resolve this.","attributes":{"escalated":true,"customer_sentiment":"negative"}}]}}'
```

### Example: Findings Analysis

```go
// Analyze findings from conversation data
findingsReq := StandardAnalysisRequest{
    AnalysisType: "findings",
    Parameters: map[string]interface{}{
        "focus_area": "customer_satisfaction",
        "data_points": []string{"sentiment", "resolution_time", "follow_up_required"},
    },
    Data: attributeData,
}
findingsResp, err := client.PerformAnalysis(findingsReq)

// Access analysis findings
if findingResults, ok := findingsResp.Results.(map[string]interface{}); ok {
    if findings, ok := findingResults["findings"].([]interface{}); ok {
        for _, f := range findings {
            if finding, ok := f.(map[string]interface{}); ok {
                fmt.Printf("Finding: %s\n", finding["title"])
                fmt.Printf("Description: %s\n", finding["description"])
                fmt.Printf("Evidence: %s\n", finding["evidence"])
                fmt.Printf("Confidence: %.2f\n", finding["confidence"])
            }
        }
    }
}
```

```bash
# Test findings analysis with curl (simple version for copy-paste)
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-findings-123","analysis_type":"findings","parameters":{"focus_area":"customer_satisfaction","data_points":["sentiment","resolution_time","follow_up_required"]},"data":{"conversation_attributes":[{"id":"c1","sentiment":"negative","resolution_time":25,"follow_up_required":true},{"id":"c2","sentiment":"positive","resolution_time":5,"follow_up_required":false},{"id":"c3","sentiment":"negative","resolution_time":15,"follow_up_required":true}]}}'
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
```

```bash
# Test recommendations generation with curl (simple version for copy-paste)
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-recs-123","analysis_type":"recommendations","parameters":{"focus_area":"customer retention","criteria":{"impact":0.6,"implementation_ease":0.4}},"data":{"findings":[{"title":"High cancellation rate due to billing issues","description":"20% of customers who cancel cite billing problems as the primary reason","evidence":"Exit survey data from Q1-Q3 2023","confidence":0.85},{"title":"Long wait times leading to customer frustration","description":"Average hold time of 15 minutes correlates with 30% lower CSAT scores","evidence":"Call metrics and CSAT correlation analysis","confidence":0.92},{"title":"Lack of mobile app features driving competitive disadvantage","description":"Competitors offer 5-8 more core features in their mobile experience","evidence":"Competitive analysis conducted in August 2023","confidence":0.78}]}}'
```

```go
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
```

```bash
# Test action plan generation with curl (simple version for copy-paste)
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-plan-123","analysis_type":"plan","parameters":{"constraints":{"budget":50000,"timeline":"6 months","resources":["customer_support","engineering","marketing"]}},"data":{"recommendations":[{"id":"rec1","title":"Implement automated billing verification system","description":"Create an automated system to verify billing accuracy before charges are processed","impact":0.8,"implementation_ease":0.5,"priority_score":0.68},{"id":"rec2","title":"Add chat support to reduce call wait times","description":"Implement chat support option to divert 30% of calls to faster text-based resolution","impact":0.6,"implementation_ease":0.7,"priority_score":0.64},{"id":"rec3","title":"Develop mobile app account management features","description":"Add self-service account management features to the mobile app","impact":0.7,"implementation_ease":0.4,"priority_score":0.58}]}}'
```

```go
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

```bash
# Test timeline generation with curl (simple version for copy-paste)
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-timeline-123","analysis_type":"plan","parameters":{"generate_timeline":true},"data":{"action_plan":{"initiatives":[{"id":"init1","title":"Billing System Enhancement","description":"Improve billing accuracy and automation","tasks":[{"id":"task1","title":"Audit current billing process","duration":"2 weeks","dependencies":[]},{"id":"task2","title":"Implement verification checks","duration":"4 weeks","dependencies":["task1"]}]},{"id":"init2","title":"Support Channel Expansion","description":"Add chat support capabilities","tasks":[{"id":"task3","title":"Select chat platform vendor","duration":"3 weeks","dependencies":[]},{"id":"task4","title":"Train support staff on chat system","duration":"2 weeks","dependencies":["task3"]}]}]},"resources":{"staff":5,"start_date":"2023-07-01"}}}'
```

## Requirements

- Go 1.21+
- Google Gemini API key (set as `GEMINI_API_KEY` environment variable)
- SQLite database

## Usage

### Command Line Testing

Use the included test clients to test the API:

```bash
# Test intent analysis with standardized API
go run cmd/testclient/main.go -type intent -text "I want to cancel my subscription"

# Test attribute extraction 
go run cmd/testclient/main.go -type attributes -text "I'm having issues with my latest bill"

# Test recommendations generation
go run cmd/testclient/main.go -type recommendations -text "Our customers frequently complain about long wait times"

# Process a file for trend analysis
go run cmd/testclient/main.go -type trends -file ./sample.txt

# View results for a workflow
go run cmd/testclient/main.go -results -workflow abc123
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