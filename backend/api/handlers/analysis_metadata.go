package handlers

// getFunctionMetadata returns metadata for all available analysis functions
func getFunctionMetadata() map[string]interface{} {
	return map[string]interface{}{
		"trends": map[string]interface{}{
			"name":        "Trend Analysis",
			"description": "Analyze trends in conversation data",
			"parameters": map[string]interface{}{
				"focus_areas": map[string]interface{}{
					"type":        "array",
					"description": "Areas to focus on in the analysis",
					"example":     []string{"Customer Satisfaction", "Response Time", "Issue Resolution"},
				},
			},
		},
		"patterns": map[string]interface{}{
			"name":        "Pattern Identification",
			"description": "Identify patterns in conversation data",
			"parameters": map[string]interface{}{
				"pattern_types": map[string]interface{}{
					"type":        "array",
					"description": "Types of patterns to look for",
					"example":     []string{"communication_patterns", "recurring_issues", "customer_behavior"},
				},
			},
		},
		"findings": map[string]interface{}{
			"name":        "Findings Analysis",
			"description": "Analyze findings from data",
			"parameters": map[string]interface{}{
				"questions": map[string]interface{}{
					"type":        "array",
					"description": "Questions to answer based on the data",
					"example":     []string{"What are the main customer pain points?", "How effective is the support team?"},
				},
			},
		},
		"attributes": map[string]interface{}{
			"name":        "Attribute Extraction",
			"description": "Extract attributes from conversation data",
			"parameters": map[string]interface{}{
				"attributes": map[string]interface{}{
					"type":        "array",
					"description": "Attributes to extract",
				},
				"generate_required": map[string]interface{}{
					"type":        "boolean",
					"description": "Whether to generate required attributes",
				},
			},
		},
		"intent": map[string]interface{}{
			"name":        "Intent Analysis",
			"description": "Analyze intents in conversation data",
		},
		"recommendations": map[string]interface{}{
			"name":        "Recommendations",
			"description": "Generate recommendations based on analysis",
			"parameters": map[string]interface{}{
				"objectives": map[string]interface{}{
					"type":        "array",
					"description": "Objectives for recommendations",
					"example":     []string{"Improve customer satisfaction", "Reduce response time"},
				},
			},
		},
		"plan": map[string]interface{}{
			"name":        "Action Plan Generation",
			"description": "Generate an action plan",
			"parameters": map[string]interface{}{
				"goals": map[string]interface{}{
					"type":        "array",
					"description": "Goals for the action plan",
					"example":     []string{"Increase customer retention", "Improve service quality"},
				},
			},
		},
	}
}

// Helper types for function metadata
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
