# Contact Center Analysis Package

This package provides Go implementations of contact center analysis functionality, ported from the Python library. It leverages Google's Gemini LLM API to analyze customer service conversations, extract attributes, identify intents, and generate insights.

## Features

- **Text Analysis**
  - [Extract attributes from conversations](#attribute-extraction)
  - [Identify customer intents](#intent-analysis)
  - [Generate required attributes for specific research questions](#generate-required-attributes)

- **Data Analysis**
  - [Analyze trends in conversation data](#trend-analysis)
  - [Identify conversation patterns](#pattern-identification)
  - [Analyze findings from attribute extraction](#findings-analysis)

- **Recommendations & Planning**
  - [Generate actionable recommendations based on analysis results](#recommendations-generation)
  - [Prioritize recommendations based on custom criteria](#recommendation-prioritization)
  - [Create comprehensive action plans from recommendations](#action-plan-generation)
  - [Generate retention strategies for customer retention](#retention-strategy-generation)
  - [Create implementation timelines with dependencies and milestones](#timeline-generation)

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

## Examples

### Text Analysis

#### Intent Analysis

> Identifies the primary intent or purpose behind a customer service conversation. The endpoint analyzes the conversation text and determines why the customer contacted support.
>
> **Inputs:** Conversation text between customer and agent.
>
> **Outputs:** A classification of the customer's intent including a human-readable label, machine-readable code, and a brief description of the customer's request or issue.

##### curl Example
```bash
# Test intent analysis with curl
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-intent-123","analysis_type":"intent","text":"I have been charged twice for my last payment and need this fixed immediately.","parameters":{}}'
```

##### Go Example
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

##### JSON Request
```json
{
  "workflow_id": "test-intent-123",
  "analysis_type": "intent",
  "text": "I've been charged twice for my last payment and need this fixed immediately.",
  "parameters": {}
}
```

#### Generate Required Attributes

> Determines what data attributes are needed to answer specific research questions. This endpoint helps in designing data collection schemas by suggesting what information needs to be captured.
>
> **Inputs:** A list of research questions and optionally any existing attributes already available.
>
> **Outputs:** A list of suggested attribute definitions, each with a field name, title, description, and rationale explaining why the attribute is needed to answer the questions.

##### curl Example
```bash
# Test generating required attributes with curl
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-attrs-gen-123","analysis_type":"attributes","parameters":{"generate_required":true,"questions":["What is the customer'\''s main problem?","How urgent is the customer request?","What was the resolution?"],"existing_attributes":["customer_name","order_number"]}}'
```

##### Go Example
```go
// Generate required attributes based on research questions
attrGenReq := StandardAnalysisRequest{
    AnalysisType: "attributes",
    Parameters: map[string]interface{}{
        "generate_required": true,
        "questions": []string{
            "What is the customer's main problem?",
            "How urgent is the customer request?",
            "What was the resolution?",
        },
        "existing_attributes": []string{
            "customer_name",
            "order_number",
        },
    },
}
attrGenResp, err := client.PerformAnalysis(attrGenReq)

// Access generated attributes
if attrGenResults, ok := attrGenResp.Results.(map[string]interface{}); ok {
    if attributes, ok := attrGenResults["attributes"].([]interface{}); ok {
        for _, a := range attributes {
            if attr, ok := a.(map[string]interface{}); ok {
                fmt.Printf("Field Name: %s\n", attr["field_name"])
                fmt.Printf("Title: %s\n", attr["title"])
                fmt.Printf("Description: %s\n", attr["description"])
                fmt.Printf("Rationale: %s\n", attr["rationale"])
            }
        }
    }
}
```

##### JSON Request
```json
{
  "workflow_id": "test-attrs-gen-123",
  "analysis_type": "attributes",
  "parameters": {
    "generate_required": true,
    "questions": [
      "What is the customer's main problem?",
      "How urgent is the customer request?",
      "What was the resolution?"
    ],
    "existing_attributes": [
      "customer_name",
      "order_number"
    ]
  }
}
```

#### Attribute Extraction

> Extracts specific attributes or data points from conversation text. This endpoint analyzes text to find values for predefined attributes.
>
> **Inputs:** Conversation text and a list of attribute definitions to extract values for.
>
> **Outputs:** Extracted values for each attribute, along with confidence scores and explanations of how each value was determined.

##### curl Example
```bash
# Test attribute extraction with curl
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-attrs-123","analysis_type":"attributes","text":"Customer: I am having issues with my latest bill. The charges are wrong.\nAgent: I understand your frustration. Let me help resolve this billing issue.","parameters":{"attributes":[{"field_name":"sentiment","title":"Customer Sentiment","description":"The sentiment expressed by the customer"},{"field_name":"issue_type","title":"Issue Type","description":"The category of the customer issue"},{"field_name":"resolution_status","title":"Resolution Status","description":"Whether the issue was resolved during the interaction"}]}}'
```

##### Go Example
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

##### JSON Request
```json
{
  "workflow_id": "test-attrs-123",
  "analysis_type": "attributes",
  "text": "Customer: I'm having issues with my latest bill. The charges are wrong.\nAgent: I understand your frustration. Let me help resolve this billing issue.",
  "parameters": {
    "attributes": [
      {
        "field_name": "sentiment",
        "title": "Customer Sentiment",
        "description": "The sentiment expressed by the customer"
      },
      {
        "field_name": "issue_type",
        "title": "Issue Type", 
        "description": "The category of the customer issue"
      },
      {
        "field_name": "resolution_status",
        "title": "Resolution Status",
        "description": "Whether the issue was resolved during the interaction"
      }
    ]
  }
}
```

### Data Analysis

#### Trend Analysis

> Identifies patterns and trends across multiple conversations over time. This endpoint analyzes aggregated conversation data to detect emerging trends.
>
> **Inputs:** Conversation metrics data with timestamps and various data points like sentiment, resolution status, and response times.
>
> **Outputs:** Identified trends with descriptions, significance scores, and insights about changes in customer behaviors or service metrics.

##### curl Example
```bash
# Test trend analysis with curl
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-trends-123","analysis_type":"trends","parameters":{"focus_areas":["sentiment","issue_resolution","response_time"],"time_period":"last_quarter"},"data":{"conversation_metrics":[{"id":"conv1","timestamp":"2023-01-15T10:30:00Z","sentiment":"negative","issue_resolution":"unresolved","response_time":480},{"id":"conv2","timestamp":"2023-01-16T11:45:00Z","sentiment":"neutral","issue_resolution":"resolved","response_time":320},{"id":"conv3","timestamp":"2023-01-18T14:20:00Z","sentiment":"positive","issue_resolution":"resolved","response_time":180}]}}'
```

##### Go Example
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

##### JSON Request
```json
{
  "workflow_id": "test-trends-123",
  "analysis_type": "trends",
  "parameters": {
    "focus_areas": ["sentiment", "issue_resolution", "response_time"],
    "time_period": "last_quarter"
  },
  "data": {
    "conversation_metrics": [
      {
        "id": "conv1",
        "timestamp": "2023-01-15T10:30:00Z",
        "sentiment": "negative",
        "issue_resolution": "unresolved",
        "response_time": 480
      },
      {
        "id": "conv2",
        "timestamp": "2023-01-16T11:45:00Z",
        "sentiment": "neutral",
        "issue_resolution": "resolved",
        "response_time": 320
      },
      {
        "id": "conv3",
        "timestamp": "2023-01-18T14:20:00Z",
        "sentiment": "positive",
        "issue_resolution": "resolved",
        "response_time": 180
      }
    ]
  }
}
```

#### Pattern Identification

> Detects recurring patterns in conversation data such as common agent behaviors or escalation triggers. This endpoint helps identify systematic issues or successful interaction patterns.
>
> **Inputs:** A collection of conversations with associated attributes and pattern types to look for.
>
> **Outputs:** Identified patterns with descriptions, frequency measurements, and their impact on customer experience or operational metrics.

##### curl Example
```bash
# Test pattern identification with curl
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-patterns-123","analysis_type":"patterns","parameters":{"pattern_types":["agent_behavior","escalation_triggers"],"min_confidence":0.7},"data":{"conversations":[{"id":"c1","text":"Customer: I am really frustrated with this service.\nAgent: I understand your frustration. Let me escalate this to my supervisor.","attributes":{"escalated":true,"customer_sentiment":"negative"}},{"id":"c2","text":"Customer: This charge seems incorrect.\nAgent: I will look into that for you right away.","attributes":{"escalated":false,"customer_sentiment":"neutral"}},{"id":"c3","text":"Customer: I have been waiting for weeks!\nAgent: Let me transfer you to my manager to help resolve this.","attributes":{"escalated":true,"customer_sentiment":"negative"}}]}}'
```

##### Go Example
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

##### JSON Request
```json
{
  "workflow_id": "test-patterns-123",
  "analysis_type": "patterns",
  "parameters": {
    "pattern_types": ["agent_behavior", "escalation_triggers"],
    "min_confidence": 0.7
  },
  "data": {
    "conversations": [
      {
        "id": "c1",
        "text": "Customer: I'm really frustrated with this service.\nAgent: I understand your frustration. Let me escalate this to my supervisor.",
        "attributes": {
          "escalated": true,
          "customer_sentiment": "negative"
        }
      },
      {
        "id": "c2",
        "text": "Customer: This charge seems incorrect.\nAgent: I'll look into that for you right away.",
        "attributes": {
          "escalated": false,
          "customer_sentiment": "neutral"
        }
      },
      {
        "id": "c3",
        "text": "Customer: I've been waiting for weeks!\nAgent: Let me transfer you to my manager to help resolve this.",
        "attributes": {
          "escalated": true,
          "customer_sentiment": "negative"
        }
      }
    ]
  }
}
```

#### Findings Analysis

> Analyzes conversation attributes to generate insights about specific focus areas like customer satisfaction. This endpoint synthesizes data points into actionable findings.
>
> **Inputs:** Conversation attributes data and specific data points to analyze, along with a focus area for the analysis.
>
> **Outputs:** Key findings with titles, descriptions, supporting evidence, and confidence scores.

##### curl Example
```bash
# Test findings analysis with curl
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-findings-123","analysis_type":"findings","parameters":{"focus_area":"customer_satisfaction","data_points":["sentiment","resolution_time","follow_up_required"]},"data":{"conversation_attributes":[{"id":"c1","sentiment":"negative","resolution_time":25,"follow_up_required":true},{"id":"c2","sentiment":"positive","resolution_time":5,"follow_up_required":false},{"id":"c3","sentiment":"negative","resolution_time":15,"follow_up_required":true}]}}'
```

##### Go Example
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

##### JSON Request
```json
{
  "workflow_id": "test-findings-123",
  "analysis_type": "findings",
  "parameters": {
    "focus_area": "customer_satisfaction",
    "data_points": ["sentiment", "resolution_time", "follow_up_required"]
  },
  "data": {
    "conversation_attributes": [
      {
        "id": "c1",
        "sentiment": "negative",
        "resolution_time": 25,
        "follow_up_required": true
      },
      {
        "id": "c2",
        "sentiment": "positive",
        "resolution_time": 5,
        "follow_up_required": false
      },
      {
        "id": "c3",
        "sentiment": "negative",
        "resolution_time": 15,
        "follow_up_required": true
      }
    ]
  }
}
```

### Recommendations & Planning

#### Recommendations Generation

> Generates actionable recommendations based on analysis findings. This endpoint transforms insights into specific actions that can be implemented.
>
> **Inputs:** Analysis findings data and a focus area for recommendations.
>
> **Outputs:** A set of recommendations including immediate actions to take, each with rationale, expected impact, and priority level, plus implementation notes and success metrics.

##### curl Example
```bash
# Test recommendations generation with curl
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-recs-123","analysis_type":"recommendations","parameters":{"focus_area":"customer retention","criteria":{"impact":0.6,"implementation_ease":0.4}},"data":{"findings":[{"title":"High cancellation rate due to billing issues","description":"20% of customers who cancel cite billing problems as the primary reason","evidence":"Exit survey data from Q1-Q3 2023","confidence":0.85},{"title":"Long wait times leading to customer frustration","description":"Average hold time of 15 minutes correlates with 30% lower CSAT scores","evidence":"Call metrics and CSAT correlation analysis","confidence":0.92},{"title":"Lack of mobile app features driving competitive disadvantage","description":"Competitors offer 5-8 more core features in their mobile experience","evidence":"Competitive analysis conducted in August 2023","confidence":0.78}]}}'
```

##### Go Example
```go
// Generate recommendations based on analysis results
recReq := StandardAnalysisRequest{
    AnalysisType: "recommendations",
    Parameters: map[string]interface{}{
        "focus_area": "customer retention",
    },
    Data: analysisResults,
}
recResp, err := client.PerformAnalysis(recReq)

// Access recommendations
if recResults, ok := recResp.Results.(map[string]interface{}); ok {
    if actions, ok := recResults["immediate_actions"].([]interface{}); ok {
        for _, a := range actions {
            if action, ok := a.(map[string]interface{}); ok {
                fmt.Printf("Action: %s\n", action["action"])
                fmt.Printf("Rationale: %s\n", action["rationale"])
                fmt.Printf("Expected Impact: %s\n", action["expected_impact"])
                fmt.Printf("Priority: %d\n", int(action["priority"].(float64)))
            }
        }
    }
}
```

##### JSON Request
```json
{
  "workflow_id": "test-recs-123",
  "analysis_type": "recommendations",
  "parameters": {
    "focus_area": "customer retention"
  },
  "data": {
    "findings": [
      {
        "title": "High cancellation rate due to billing issues",
        "description": "20% of customers who cancel cite billing problems as the primary reason",
        "evidence": "Exit survey data from Q1-Q3 2023",
        "confidence": 0.85
      },
      {
        "title": "Long wait times leading to customer frustration",
        "description": "Average hold time of 15 minutes correlates with 30% lower CSAT scores",
        "evidence": "Call metrics and CSAT correlation analysis",
        "confidence": 0.92
      },
      {
        "title": "Lack of mobile app features driving competitive disadvantage",
        "description": "Competitors offer 5-8 more core features in their mobile experience",
        "evidence": "Competitive analysis conducted in August 2023",
        "confidence": 0.78
      }
    ]
  }
}
```

#### Recommendation Prioritization

> Prioritizes a list of recommendations based on custom criteria. This endpoint helps organizations determine which recommendations to implement first.
>
> **Inputs:** A list of recommendations and a set of prioritization criteria with weights (like impact, implementation ease, cost efficiency).
>
> **Outputs:** The same recommendations with updated priority scores based on the provided criteria.

##### curl Example
```bash
# Test recommendation prioritization with curl
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-priority-123","analysis_type":"recommendations","parameters":{"focus_area":"customer retention","criteria":{"impact":0.6,"implementation_ease":0.4,"cost_efficiency":0.3,"time_to_value":0.5}},"data":{"recommendations":[{"action":"Implement automated billing verification system","rationale":"Prevents billing errors before they reach customers","expected_impact":"30% reduction in billing-related complaints","priority":3},{"action":"Add chat support to reduce call wait times","rationale":"Provides alternative support channel","expected_impact":"25% reduction in call volume","priority":2},{"action":"Improve mobile app self-service features","rationale":"Allows customers to resolve issues without contacting support","expected_impact":"20% reduction in basic inquiry calls","priority":1}]}}'
```

##### Go Example
```go
// Prioritize recommendations based on custom criteria
priorityReq := StandardAnalysisRequest{
    AnalysisType: "recommendations",
    Parameters: map[string]interface{}{
        "focus_area": "customer retention",
        "criteria": map[string]float64{
            "impact":              0.6,
            "implementation_ease": 0.4,
            "cost_efficiency":     0.3,
            "time_to_value":       0.5,
        },
    },
    Data: map[string]interface{}{
        "recommendations": existingRecommendations,
    },
}
priorityResp, err := client.PerformAnalysis(priorityReq)

// Access prioritized recommendations
if priorityResults, ok := priorityResp.Results.(map[string]interface{}); ok {
    if actions, ok := priorityResults["immediate_actions"].([]interface{}); ok {
        for _, a := range actions {
            if action, ok := a.(map[string]interface{}); ok {
                fmt.Printf("Action: %s\n", action["action"])
                fmt.Printf("Priority: %d\n", int(action["priority"].(float64)))
            }
        }
    }
}
```

##### JSON Request
```json
{
  "workflow_id": "test-priority-123",
  "analysis_type": "recommendations",
  "parameters": {
    "focus_area": "customer retention",
    "criteria": {
      "impact": 0.6,
      "implementation_ease": 0.4,
      "cost_efficiency": 0.3,
      "time_to_value": 0.5
    }
  },
  "data": {
    "recommendations": [
      {
        "action": "Implement automated billing verification system",
        "rationale": "Prevents billing errors before they reach customers",
        "expected_impact": "30% reduction in billing-related complaints",
        "priority": 3
      },
      {
        "action": "Add chat support to reduce call wait times",
        "rationale": "Provides alternative support channel",
        "expected_impact": "25% reduction in call volume",
        "priority": 2
      },
      {
        "action": "Improve mobile app self-service features",
        "rationale": "Allows customers to resolve issues without contacting support",
        "expected_impact": "20% reduction in basic inquiry calls",
        "priority": 1
      }
    ]
  }
}
```

#### Action Plan Generation

> Creates a comprehensive implementation plan from a set of recommendations. This endpoint transforms recommendations into a structured plan with timelines and responsibilities.
>
> **Inputs:** A list of recommendations and constraints like budget, timeline, and available resources.
>
> **Outputs:** A detailed action plan including goals, immediate/short-term/long-term actions, responsible parties, timeline, success metrics, and risk mitigations.

##### curl Example
```bash
# Test action plan generation with curl
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-plan-123","analysis_type":"plan","parameters":{"constraints":{"budget":50000,"timeline":"6 months","resources":["customer_support","engineering","marketing"]}},"data":{"recommendations":[{"id":"rec1","title":"Implement automated billing verification system","description":"Create an automated system to verify billing accuracy before charges are processed","impact":0.8,"implementation_ease":0.5,"priority_score":0.68},{"id":"rec2","title":"Add chat support to reduce call wait times","description":"Implement chat support option to divert 30% of calls to faster text-based resolution","impact":0.6,"implementation_ease":0.7,"priority_score":0.64},{"id":"rec3","title":"Develop mobile app account management features","description":"Add self-service account management features to the mobile app","impact":0.7,"implementation_ease":0.4,"priority_score":0.58}]}}'
```

##### Go Example
```go
// Create an action plan from recommendations
planReq := StandardAnalysisRequest{
    AnalysisType: "plan",
    Parameters: map[string]interface{}{
        "constraints": map[string]interface{}{
            "budget": 50000,
            "timeline": "6 months",
            "resources": []string{"customer_support", "engineering", "marketing"},
        },
    },
    Data: map[string]interface{}{
        "recommendations": recommendations,
    },
}
planResp, err := client.PerformAnalysis(planReq)

// Access action plan
if planResults, ok := planResp.Results.(map[string]interface{}); ok {
    if goals, ok := planResults["goals"].([]interface{}); ok {
        fmt.Println("Action Plan Goals:")
        for _, g := range goals {
            fmt.Printf("- %s\n", g)
        }
    }
    
    if actions, ok := planResults["immediate_actions"].([]interface{}); ok {
        fmt.Println("Immediate Actions:")
        for _, a := range actions {
            if action, ok := a.(map[string]interface{}); ok {
                fmt.Printf("- %s\n", action["action"])
            }
        }
    }
}
```

##### JSON Request
```json
{
  "workflow_id": "test-plan-123",
  "analysis_type": "plan",
  "parameters": {
    "constraints": {
      "budget": 50000,
      "timeline": "6 months",
      "resources": ["customer_support", "engineering", "marketing"]
    }
  },
  "data": {
    "recommendations": [
      {
        "id": "rec1",
        "title": "Implement automated billing verification system",
        "description": "Create an automated system to verify billing accuracy before charges are processed",
        "impact": 0.8,
        "implementation_ease": 0.5,
        "priority_score": 0.68
      },
      {
        "id": "rec2", 
        "title": "Add chat support to reduce call wait times",
        "description": "Implement chat support option to divert 30% of calls to faster text-based resolution",
        "impact": 0.6,
        "implementation_ease": 0.7,
        "priority_score": 0.64
      },
      {
        "id": "rec3",
        "title": "Develop mobile app account management features",
        "description": "Add self-service account management features to the mobile app",
        "impact": 0.7,
        "implementation_ease": 0.4,
        "priority_score": 0.58
      }
    ]
  }
}
```

#### Retention Strategy Generation

> Generates customer retention strategies based on cancellation analysis. This endpoint helps organizations reduce customer churn.
>
> **Inputs:** Analysis of customer cancellation data including reasons, demographics, and exit survey information.
>
> **Outputs:** A targeted retention strategy with immediate actions, process changes, training needs, and success metrics tailored to specific customer segments.

##### curl Example
```bash
# Test retention strategy generation with curl
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-retention-123","analysis_type":"recommendations","parameters":{"generate_retention_strategy":true},"data":{"cancellation_analysis":{"primary_reasons":["billing_issues","competitive_offers","poor_service_experience"],"demographics":{"segment":"small_business","tenure":"1-3 years"},"exit_survey_data":{"avg_satisfaction":2.3,"would_return_percent":35}}}}'
```

##### Go Example
```go
// Generate retention strategies based on cancellation analysis
retentionReq := StandardAnalysisRequest{
    AnalysisType: "recommendations",
    Parameters: map[string]interface{}{
        "generate_retention_strategy": true,
    },
    Data: map[string]interface{}{
        "cancellation_analysis": cancellationData,
    },
}
retentionResp, err := client.PerformAnalysis(retentionReq)

// Access retention strategy
if retentionResults, ok := retentionResp.Results.(map[string]interface{}); ok {
    fmt.Printf("Target Segment: %s\n", retentionResults["target_segment"])
    
    if actions, ok := retentionResults["immediate_actions"].([]interface{}); ok {
        fmt.Println("Immediate Actions:")
        for _, a := range actions {
            if action, ok := a.(map[string]interface{}); ok {
                fmt.Printf("- %s\n", action["action"])
            }
        }
    }
    
    if changes, ok := retentionResults["process_changes"].([]interface{}); ok {
        fmt.Println("Process Changes:")
        for _, c := range changes {
            fmt.Printf("- %s\n", c)
        }
    }
}
```

##### JSON Request
```json
{
  "workflow_id": "test-retention-123",
  "analysis_type": "recommendations",
  "parameters": {
    "generate_retention_strategy": true
  },
  "data": {
    "cancellation_analysis": {
      "primary_reasons": [
        "billing_issues",
        "competitive_offers",
        "poor_service_experience"
      ],
      "demographics": {
        "segment": "small_business",
        "tenure": "1-3 years"
      },
      "exit_survey_data": {
        "avg_satisfaction": 2.3,
        "would_return_percent": 35
      }
    }
  }
}
```

#### Timeline Generation

> Creates a detailed implementation timeline for an action plan. This endpoint helps organizations plan the execution of their initiatives.
>
> **Inputs:** An action plan with initiatives and tasks, plus information about available resources.
>
> **Outputs:** A timeline with phases, descriptions, durations, and milestones that considers dependencies between actions and resource constraints.

##### curl Example
```bash
# Test timeline generation with curl
curl -X POST http://localhost:8080/api/analysis -H "Content-Type: application/json" -d '{"workflow_id":"test-timeline-123","analysis_type":"plan","parameters":{"generate_timeline":true},"data":{"action_plan":{"initiatives":[{"id":"init1","title":"Billing System Enhancement","description":"Improve billing accuracy and automation","tasks":[{"id":"task1","title":"Audit current billing process","duration":"2 weeks","dependencies":[]},{"id":"task2","title":"Implement verification checks","duration":"4 weeks","dependencies":["task1"]}]},{"id":"init2","title":"Support Channel Expansion","description":"Add chat support capabilities","tasks":[{"id":"task3","title":"Select chat platform vendor","duration":"3 weeks","dependencies":[]},{"id":"task4","title":"Train support staff on chat system","duration":"2 weeks","dependencies":["task3"]}]}]},"resources":{"staff":5,"start_date":"2023-07-01"}}}'
```

##### Go Example
```go
// Generate a timeline for the action plan
timelineReq := StandardAnalysisRequest{
    AnalysisType: "plan",
    Parameters: map[string]interface{}{
        "generate_timeline": true,
    },
    Data: map[string]interface{}{
        "action_plan": actionPlan,
        "resources": map[string]interface{}{
            "staff": 5,
            "start_date": "2023-07-01",
        },
    },
}
timelineResp, err := client.PerformAnalysis(timelineReq)

// Access timeline
if timelineResults, ok := timelineResp.Results.(map[string]interface{}); ok {
    if timeline, ok := timelineResults["timeline"].([]interface{}); ok {
        for _, t := range timeline {
            if event, ok := t.(map[string]interface{}); ok {
                fmt.Printf("Phase: %s\n", event["phase"])
                fmt.Printf("Duration: %s\n", event["duration"])
                
                if milestones, ok := event["milestones"].([]interface{}); ok {
                    fmt.Println("Milestones:")
                    for _, m := range milestones {
                        fmt.Printf("- %s\n", m)
                    }
                }
            }
        }
    }
}
```

##### JSON Request
```json
{
  "workflow_id": "test-timeline-123",
  "analysis_type": "plan",
  "parameters": {
    "generate_timeline": true
  },
  "data": {
    "action_plan": {
      "initiatives": [
        {
          "id": "init1",
          "title": "Billing System Enhancement",
          "description": "Improve billing accuracy and automation",
          "tasks": [
            {
              "id": "task1",
              "title": "Audit current billing process",
              "duration": "2 weeks",
              "dependencies": []
            },
            {
              "id": "task2",
              "title": "Implement verification checks",
              "duration": "4 weeks",
              "dependencies": ["task1"]
            }
          ]
        },
        {
          "id": "init2",
          "title": "Support Channel Expansion",
          "description": "Add chat support capabilities",
          "tasks": [
            {
              "id": "task3",
              "title": "Select chat platform vendor",
              "duration": "3 weeks",
              "dependencies": []
            },
            {
              "id": "task4",
              "title": "Train support staff on chat system",
              "duration": "2 weeks",
              "dependencies": ["task3"]
            }
          ]
        }
      ]
    },
    "resources": {
      "staff": 5,
      "start_date": "2023-07-01"
    }
  }
}
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