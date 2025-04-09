package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"agenticflows/backend/analysis"
	"agenticflows/backend/db"
)

// Generator handles workflow generation
type Generator struct {
	llmClient *analysis.LLMClient
}

// NewGenerator creates a new workflow generator
func NewGenerator() *Generator {
	// Get the API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("Warning: GEMINI_API_KEY environment variable not set")
		return &Generator{}
	}

	// Create LLM client
	llmClient, err := analysis.NewLLMClient(apiKey, false)
	if err != nil {
		log.Printf("Warning: failed to create LLM client: %s", err)
		return &Generator{}
	}

	return &Generator{
		llmClient: llmClient,
	}
}

// GenerateFromDescription uses LLM to generate a workflow based on the description
func (g *Generator) GenerateFromDescription(name, description string) (db.Workflow, error) {
	if g.llmClient == nil {
		return db.Workflow{}, fmt.Errorf("LLM client not initialized")
	}

	// Create function metadata for the prompt
	functionMetadata, err := getFunctionMetadataForLLM()
	if err != nil {
		return db.Workflow{}, fmt.Errorf("failed to get function metadata: %s", err)
	}

	// Create the prompt
	prompt := createWorkflowGenerationPrompt(name, description, functionMetadata)

	// Call the LLM API
	result, err := g.llmClient.GenerateContent(context.Background(), prompt, map[string]interface{}{
		"nodes": []interface{}{},
		"edges": []interface{}{},
	})
	if err != nil {
		return db.Workflow{}, fmt.Errorf("failed to generate workflow from LLM: %s", err)
	}

	// Parse the result
	workflow, err := parseWorkflowFromLLMResponse(name, result)
	if err != nil {
		return db.Workflow{}, fmt.Errorf("failed to parse LLM response: %s", err)
	}

	// Default database path setting
	const defaultDBPath = "/Users/jonathan/Documents/Work/discourse_ai/Research/corpora/banking_2025/db/standard_charter_bank.db"

	// Set defaults and metadata
	workflow.ID = fmt.Sprintf("wf-%d", time.Now().UnixNano())
	workflow.Date = time.Now().Format("2006-01-02")

	// Add a description node to include information about the default database
	nodesStr := string(workflow.Nodes)
	if nodesStr != "" {
		var nodes []map[string]interface{}
		err = json.Unmarshal(workflow.Nodes, &nodes)
		if err == nil {
			// Look for a text or note node we can update
			noteNodeFound := false
			for i, node := range nodes {
				nodeType, _ := node["type"].(string)
				if nodeType == "note" || nodeType == "text" {
					// Update existing note node
					if data, ok := node["data"].(map[string]interface{}); ok {
						data["text"] = fmt.Sprintf("Database: %s\n\n%s",
							defaultDBPath,
							data["text"])
						node["data"] = data
						nodes[i] = node
						noteNodeFound = true
						break
					}
				}
			}

			// If no note node exists, add one
			if !noteNodeFound {
				noteNode := map[string]interface{}{
					"id":   fmt.Sprintf("note-%d", time.Now().UnixNano()),
					"type": "note",
					"position": map[string]interface{}{
						"x": 100,
						"y": 100,
					},
					"data": map[string]interface{}{
						"text": fmt.Sprintf("Database Path: %s\n\nThis workflow is configured to use the banking conversations database.", defaultDBPath),
					},
				}
				nodes = append(nodes, noteNode)
			}

			// Update the nodes in the workflow
			updatedNodes, _ := json.Marshal(nodes)
			workflow.Nodes = json.RawMessage(updatedNodes)
		}
	}

	return workflow, nil
}

// GenerateDynamic uses LLM to generate a dynamic workflow with custom functions
func (g *Generator) GenerateDynamic(name, description string) (db.Workflow, error) {
	if g.llmClient == nil {
		return db.Workflow{}, fmt.Errorf("LLM client not initialized")
	}

	// Create the prompt for dynamic workflow
	prompt := createDynamicWorkflowGenerationPrompt(name, description)

	// Call the LLM API
	log.Printf("Generating dynamic workflow with prompt: %s", prompt)

	// Provide a structured template to help the LLM
	defaultTemplate := map[string]interface{}{
		"nodes": []interface{}{
			map[string]interface{}{
				"id":   "node-1",
				"type": "function",
				"position": map[string]interface{}{
					"x": 250,
					"y": 100,
				},
				"data": map[string]interface{}{
					"nodeType":    "function",
					"functionId":  "custom-function-1",
					"label":       "Custom Function 1",
					"description": "This is a custom function",
				},
			},
		},
		"edges": []interface{}{},
	}

	result, err := g.llmClient.GenerateContent(context.Background(), prompt, defaultTemplate)
	if err != nil {
		return db.Workflow{}, fmt.Errorf("failed to generate dynamic workflow from LLM: %s", err)
	}

	// Parse the result - convert interface{} to string
	resultStr, ok := result.(string)
	if !ok {
		log.Printf("Failed to convert LLM response to string, got type: %T", result)
		// Try to marshal the result to string
		resultBytes, err := json.Marshal(result)
		if err != nil {
			return db.Workflow{}, fmt.Errorf("failed to convert LLM response to string: %v", err)
		}
		resultStr = string(resultBytes)
	}

	// Log the raw response for debugging
	log.Printf("Raw LLM response: %s", resultStr)

	// Parse the result
	workflow, err := parseDynamicWorkflowFromLLMResponse(name, resultStr)
	if err != nil {
		return db.Workflow{}, fmt.Errorf("failed to parse LLM response: %s", err)
	}

	// Set defaults and metadata
	workflow.ID = fmt.Sprintf("wf-dyn-%d", time.Now().UnixNano())
	workflow.Date = time.Now().Format("2006-01-02")

	return workflow, nil
}

// Create prompt for the LLM to generate a workflow
func createWorkflowGenerationPrompt(name, description string, functionMetadata []map[string]interface{}) string {
	prompt := fmt.Sprintf(`As an expert workflow designer, create a workflow called "%s" based on this description:

Description: %s

Available functions for the workflow:
`, name, description)

	// Add function descriptions
	for _, fn := range functionMetadata {
		prompt += fmt.Sprintf("\n- %s: %s\n", fn["id"], fn["description"])

		// Add inputs
		prompt += "  Inputs:\n"
		if inputs, ok := fn["inputs"].([]map[string]interface{}); ok {
			for _, input := range inputs {
				prompt += fmt.Sprintf("    - %s: %s\n", input["name"], input["description"])
			}
		}

		// Add outputs
		prompt += "  Outputs:\n"
		if outputs, ok := fn["outputs"].([]map[string]interface{}); ok {
			for _, output := range outputs {
				prompt += fmt.Sprintf("    - %s: %s\n", output["name"], output["description"])
			}
		}
	}

	prompt += `
The workflow will process data from a SQLite database with these tables:
- conversations (conversation_id, date_time, text, client_id, agent_id)
- attribute_definitions (id, field_name, title, description, type, enum_values, intent_type, workflow_id, created_at)

Default database path: /Users/jonathan/Documents/Work/discourse_ai/Research/corpora/banking_2025/db/standard_charter_bank.db

Please respond with a JSON object containing:
1. A list of nodes (workflow steps)
2. A list of edges (connections between steps)
3. For each node, specify the function type and any configuration
4. For edges, specify source node, target node, and data mappings

Format your response as valid JSON with this structure:
{
  "nodes": [
    {
      "id": "node-1",
      "type": "function",
      "position": { "x": 250, "y": 100 },
      "data": {
        "nodeType": "function",
        "functionId": "analysis-attributes",
        "label": "Extract Attributes"
      }
    },
    ...more nodes...
  ],
  "edges": [
    {
      "id": "edge-1-2",
      "source": "node-1",
      "target": "node-2",
      "data": {
        "mappings": [
          {
            "sourceOutput": "results.attributes",
            "targetInput": "parameters.attributes" 
          }
        ]
      }
    },
    ...more edges...
  ]
}

Create an effective workflow that processes banking conversations, extracts meaningful attributes, identifies patterns, and provides insights or recommendations based on the description.`

	return prompt
}

// getFunctionMetadataForLLM gets function metadata in a format usable by the LLM
func getFunctionMetadataForLLM() ([]map[string]interface{}, error) {
	// Create a list of function metadata
	metadata := []map[string]interface{}{
		{
			"id":          "analysis-trends",
			"label":       "Analyze Trends",
			"description": "Analyze trends in conversation data",
			"inputs": []map[string]interface{}{
				{
					"name":        "Focus Areas",
					"description": "Areas to focus trend analysis on",
					"required":    true,
				},
				{
					"name":        "Text",
					"description": "Text to analyze for trends",
					"required":    false,
				},
			},
			"outputs": []map[string]interface{}{
				{
					"name":        "Trends",
					"description": "Identified trends and patterns",
				},
				{
					"name":        "Metrics",
					"description": "Trend metrics and statistics",
				},
			},
		},
		{
			"id":          "analysis-patterns",
			"label":       "Identify Patterns",
			"description": "Identify patterns in conversation data",
			"inputs": []map[string]interface{}{
				{
					"name":        "Pattern Types",
					"description": "Types of patterns to identify",
					"required":    true,
				},
				{
					"name":        "Text",
					"description": "Text to analyze for patterns",
					"required":    false,
				},
			},
			"outputs": []map[string]interface{}{
				{
					"name":        "Patterns",
					"description": "Identified patterns",
				},
				{
					"name":        "Categories",
					"description": "Pattern categories",
				},
			},
		},
		// Add other function types as needed
	}

	return metadata, nil
}

func createDynamicWorkflowGenerationPrompt(name, description string) string {
	prompt := fmt.Sprintf(`As an expert workflow designer, create a dynamic workflow called "%s" based on this description:

Description: %s

IMPORTANT: Unlike a standard workflow, this is a DYNAMIC workflow which means you need to:
1. Create custom functions that are specific to this workflow's needs
2. Define clear inputs and outputs for each function
3. Ensure the functions work together to achieve the workflow goal
4. Each function should have a description field explaining what it does

Your response MUST be a valid JSON object with the following structure:
{
  "nodes": [
    {
      "id": "node-1",
      "type": "function",
      "position": { "x": 250, "y": 100 },
      "data": {
        "nodeType": "function",
        "functionId": "custom-function-1",
        "label": "Extract Important Data",
        "description": "Extracts key data points from input text",
        "inputs": [
          {
            "name": "text",
            "description": "Input text to process",
            "required": true
          }
        ],
        "outputs": [
          {
            "name": "extractedData",
            "description": "Extracted data points"
          }
        ]
      }
    },
    ...more nodes...
  ],
  "edges": [
    ...
  ]
}

CREATE CUSTOM FUNCTIONS SPECIFIC TO THE WORKFLOW DESCRIPTION.
`, name, description)

	return prompt
}

// parseWorkflowFromLLMResponse parses the LLM response into a workflow
func parseWorkflowFromLLMResponse(name string, llmResponse interface{}) (db.Workflow, error) {
	respMap, ok := llmResponse.(map[string]interface{})
	if !ok {
		return db.Workflow{}, fmt.Errorf("unexpected response format, expected map")
	}

	// Extract nodes
	nodesRaw, ok := respMap["nodes"]
	if !ok {
		return db.Workflow{}, fmt.Errorf("nodes field missing from response")
	}

	// Extract edges
	edgesRaw, ok := respMap["edges"]
	if !ok {
		return db.Workflow{}, fmt.Errorf("edges field missing from response")
	}

	// Convert to JSON strings
	nodesBytes, err := json.Marshal(nodesRaw)
	if err != nil {
		return db.Workflow{}, fmt.Errorf("failed to marshal nodes: %s", err)
	}

	edgesBytes, err := json.Marshal(edgesRaw)
	if err != nil {
		return db.Workflow{}, fmt.Errorf("failed to marshal edges: %s", err)
	}

	// Create the workflow
	workflow := db.Workflow{
		ID:    fmt.Sprintf("wf-%d", time.Now().UnixNano()),
		Name:  name,
		Date:  time.Now().Format("2006-01-02"),
		Nodes: json.RawMessage(nodesBytes),
		Edges: json.RawMessage(edgesBytes),
	}

	return workflow, nil
}

func parseDynamicWorkflowFromLLMResponse(name string, llmResponse string) (db.Workflow, error) {
	// Parse the LLM response to extract the workflow structure
	var workflowData map[string]interface{}

	// First try to unmarshal the entire response
	err := json.Unmarshal([]byte(llmResponse), &workflowData)
	if err != nil {
		// Try to extract JSON from a text response
		jsonStart := strings.Index(llmResponse, "{")
		jsonEnd := strings.LastIndex(llmResponse, "}")

		if jsonStart >= 0 && jsonEnd > jsonStart {
			jsonStr := llmResponse[jsonStart : jsonEnd+1]
			if err := json.Unmarshal([]byte(jsonStr), &workflowData); err != nil {
				return db.Workflow{}, fmt.Errorf("failed to parse LLM response as JSON: %s", err)
			}
		} else {
			return db.Workflow{}, fmt.Errorf("failed to extract JSON from LLM response")
		}
	}

	// Debug logging
	jsonBytes, _ := json.MarshalIndent(workflowData, "", "  ")
	log.Printf("Parsed LLM response: %s", string(jsonBytes))

	// Check if nodes exists
	nodes, nodesExist := workflowData["nodes"]
	if !nodesExist {
		// Create default empty nodes array if missing
		nodes = []interface{}{}
		log.Printf("Nodes field missing from LLM response, using empty array")
	}

	// Check if edges exists
	edges, edgesExist := workflowData["edges"]
	if !edgesExist {
		// Create default empty edges array if missing
		edges = []interface{}{}
		log.Printf("Edges field missing from LLM response, using empty array")
	}

	// Convert to JSON
	nodesJSON, err := json.Marshal(nodes)
	if err != nil {
		return db.Workflow{}, fmt.Errorf("failed to marshal nodes: %s", err)
	}

	edgesJSON, err := json.Marshal(edges)
	if err != nil {
		return db.Workflow{}, fmt.Errorf("failed to marshal edges: %s", err)
	}

	// Create workflow
	workflow := db.Workflow{
		Name:  name,
		Nodes: nodesJSON,
		Edges: edgesJSON,
	}

	return workflow, nil
}
