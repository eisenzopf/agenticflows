package workflow

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"agenticflows/backend/api/models"
	"agenticflows/backend/db"
)

// Executor handles workflow execution
type Executor struct {
	workflow db.Workflow
	nodes    []map[string]interface{}
	edges    []map[string]interface{}
}

// NewExecutor creates a workflow executor for a specific workflow
func NewExecutor(w db.Workflow) *Executor {
	// Parse nodes and edges
	var nodes []map[string]interface{}
	var edges []map[string]interface{}

	if err := json.Unmarshal(w.Nodes, &nodes); err != nil {
		log.Printf("Error parsing workflow nodes: %v", err)
		nodes = []map[string]interface{}{}
	}

	if err := json.Unmarshal(w.Edges, &edges); err != nil {
		log.Printf("Error parsing workflow edges: %v", err)
		edges = []map[string]interface{}{}
	}

	return &Executor{
		workflow: w,
		nodes:    nodes,
		edges:    edges,
	}
}

// Execute runs the workflow with the given inputs
func (e *Executor) Execute(text string, data map[string]interface{}, parameters map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("Executing workflow '%s' with %d nodes and %d edges", e.workflow.Name, len(e.nodes), len(e.edges))

	// Find all function nodes
	functionNodes := make([]map[string]interface{}, 0)
	for _, node := range e.nodes {
		data, ok := node["data"].(map[string]interface{})
		if !ok {
			continue
		}

		nodeType, _ := data["nodeType"].(string)
		if nodeType == "function" {
			functionNodes = append(functionNodes, node)
		}
	}

	// Sort nodes for execution based on dependencies
	sortedNodes, err := e.getExecutionOrder(functionNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to determine execution order: %s", err)
	}

	// Initialize results storage
	results := make(map[string]interface{})

	// Add initial data to results
	if data != nil {
		for k, v := range data {
			results[k] = v
		}
	}

	// If text was provided, add it to results
	if text != "" {
		results["text"] = text
	}

	// Add any additional parameters
	if parameters != nil {
		for k, v := range parameters {
			results[k] = v
		}
	}

	// Execute each node in order
	for _, node := range sortedNodes {
		nodeID, _ := node["id"].(string)
		data, _ := node["data"].(map[string]interface{})
		functionId, _ := data["functionId"].(string)

		// Skip if no function ID
		if functionId == "" {
			continue
		}

		// Parse the function type from the ID (e.g., "analysis-trends" -> "trends")
		parts := strings.Split(functionId, "-")
		if len(parts) < 2 {
			continue
		}

		// Get input data from connected nodes
		nodeInputs := make(map[string]interface{})

		// Find incoming edges to this node
		for _, edge := range e.edges {
			target, _ := edge["target"].(string)
			if target != nodeID {
				continue
			}

			source, _ := edge["source"].(string)
			edgeData, hasData := edge["data"].(map[string]interface{})

			// Apply data mappings if defined
			if hasData && edgeData != nil {
				mappings, hasMappings := edgeData["mappings"].([]interface{})
				if hasMappings && mappings != nil {
					// Get source node results
					sourceResults, exists := results[source].(map[string]interface{})
					if !exists {
						continue
					}

					// Apply each mapping
					for _, mappingObj := range mappings {
						mapping, isMap := mappingObj.(map[string]interface{})
						if !isMap {
							continue
						}

						sourceOutput, _ := mapping["sourceOutput"].(string)
						targetInput, _ := mapping["targetInput"].(string)

						if sourceOutput != "" && targetInput != "" {
							// Get the source value from results
							if sourceValue, exists := sourceResults[sourceOutput]; exists {
								nodeInputs[targetInput] = sourceValue
							}
						}
					}
				}
			}
		}

		// Merge with global data
		for k, v := range results {
			if _, exists := nodeInputs[k]; !exists {
				nodeInputs[k] = v
			}
		}

		// Create a placeholder for node results - in a real implementation,
		// we would delegate to specific function handlers
		nodeResult := map[string]interface{}{
			"status":         "executed",
			"function_id":    functionId,
			"execution_time": time.Now().Format(time.RFC3339),
			"inputs":         nodeInputs,
		}

		// Store results
		results[nodeID] = nodeResult
	}

	return results, nil
}

// getExecutionOrder sorts nodes by dependencies to allow for proper execution order
func (e *Executor) getExecutionOrder(nodes []map[string]interface{}) ([]map[string]interface{}, error) {
	// Create a map of node dependencies
	dependencies := make(map[string][]string)
	nodeMap := make(map[string]map[string]interface{})

	// Initialize with empty dependencies
	for _, node := range nodes {
		id, _ := node["id"].(string)
		if id != "" {
			dependencies[id] = []string{}
			nodeMap[id] = node
		}
	}

	// Add dependencies based on edges
	for _, edge := range e.edges {
		target, hasTarget := edge["target"].(string)
		source, hasSource := edge["source"].(string)

		if hasTarget && hasSource && dependencies[target] != nil {
			dependencies[target] = append(dependencies[target], source)
		}
	}

	// Topological sort
	visited := make(map[string]bool)
	temp := make(map[string]bool) // For cycle detection
	result := make([]map[string]interface{}, 0)

	// DFS function for topological sort
	var dfs func(string) error
	dfs = func(nodeID string) error {
		// Skip if already visited
		if visited[nodeID] {
			return nil
		}

		// Check for cycles
		if temp[nodeID] {
			return fmt.Errorf("workflow contains cycles, which are not supported")
		}

		// Mark as temporarily visited
		temp[nodeID] = true

		// Visit all dependencies first
		for _, dep := range dependencies[nodeID] {
			if err := dfs(dep); err != nil {
				return err
			}
		}

		// Mark as visited
		visited[nodeID] = true
		temp[nodeID] = false

		// Add to result
		if node, exists := nodeMap[nodeID]; exists {
			result = append(result, node)
		}

		return nil
	}

	// Start DFS from all nodes
	for nodeID := range dependencies {
		if !visited[nodeID] {
			if err := dfs(nodeID); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

// GenerateExecutionConfig generates a configuration for executing a workflow
func GenerateExecutionConfig(w db.Workflow) (models.WorkflowExecutionConfig, error) {
	// Parse nodes and edges
	var nodes []map[string]interface{}
	var edges []map[string]interface{}

	if err := json.Unmarshal(w.Nodes, &nodes); err != nil {
		return models.WorkflowExecutionConfig{}, fmt.Errorf("error parsing workflow nodes: %v", err)
	}

	if err := json.Unmarshal(w.Edges, &edges); err != nil {
		// Non-critical error for configuration
		log.Printf("Warning: Failed to parse edges for workflow %s: %v", w.ID, err)
	}

	// Define basic data source configuration
	inputTabs := []map[string]interface{}{
		{
			"id":    "basicData",
			"label": "Data Sources",
			"dataSourceConfigs": []map[string]interface{}{
				{
					"id":          "manualInput",
					"name":        "Manual Input",
					"description": "Enter data manually for workflow execution",
					"fields": []map[string]interface{}{
						{
							"id":          "text",
							"label":       "Input Text",
							"type":        "textarea",
							"placeholder": "Enter text to analyze...",
							"required":    false,
						},
					},
				},
			},
		},
	}

	// Check for database connections
	hasDatabaseNode := false
	for _, node := range nodes {
		data, ok := node["data"].(map[string]interface{})
		if !ok {
			continue
		}

		nodeType, _ := data["nodeType"].(string)
		label, _ := data["label"].(string)

		if nodeType == "tool" &&
			(strings.Contains(strings.ToLower(label), "database") ||
				strings.Contains(strings.ToLower(label), "db")) {
			hasDatabaseNode = true
			break
		}
	}

	// Add database configuration if needed
	if hasDatabaseNode {
		inputTabs[0]["dataSourceConfigs"] = append(
			inputTabs[0]["dataSourceConfigs"].([]map[string]interface{}),
			map[string]interface{}{
				"id":          "databaseSource",
				"name":        "Database Connection",
				"description": "Configure database connection for data retrieval",
				"fields": []map[string]interface{}{
					{
						"id":          "dbPath",
						"label":       "Database Path",
						"type":        "text",
						"description": "Path to the SQLite database file",
						"required":    true,
					},
					{
						"id":           "maxItems",
						"label":        "Maximum Items",
						"type":         "number",
						"description":  "Maximum number of items to retrieve",
						"defaultValue": "100",
						"required":     false,
					},
				},
			},
		)
	}

	// Define basic execution parameters
	parameters := []map[string]interface{}{
		{
			"id":    "executionParams",
			"label": "Execution Parameters",
			"fields": []map[string]interface{}{
				{
					"id":           "batchSize",
					"label":        "Batch Size",
					"type":         "number",
					"description":  "Number of items to process in each batch",
					"defaultValue": "10",
					"required":     false,
				},
				{
					"id":           "debugMode",
					"label":        "Enable Debug Mode",
					"type":         "checkbox",
					"defaultValue": false,
					"required":     false,
				},
			},
		},
	}

	// Check for specific analysis nodes and add parameters accordingly
	hasNodeType := make(map[string]bool)

	for _, node := range nodes {
		data, ok := node["data"].(map[string]interface{})
		if !ok {
			continue
		}

		functionId, _ := data["functionId"].(string)
		if functionId == "" {
			continue
		}

		if strings.Contains(functionId, "analysis-trends") {
			hasNodeType["trends"] = true
		} else if strings.Contains(functionId, "analysis-patterns") {
			hasNodeType["patterns"] = true
		} else if strings.Contains(functionId, "analysis-findings") {
			hasNodeType["findings"] = true
		}
	}

	// Add parameters based on analysis types
	if hasNodeType["trends"] {
		parameters = append(parameters, map[string]interface{}{
			"id":    "trendsParams",
			"label": "Trends Analysis",
			"fields": []map[string]interface{}{
				{
					"id":           "focusAreas",
					"label":        "Focus Areas",
					"type":         "text",
					"description":  "Comma-separated list of focus areas for trend analysis",
					"defaultValue": "customer_impact,financial_impact",
					"required":     false,
				},
			},
		})
	}

	if hasNodeType["patterns"] {
		parameters = append(parameters, map[string]interface{}{
			"id":    "patternsParams",
			"label": "Patterns Analysis",
			"fields": []map[string]interface{}{
				{
					"id":           "patternTypes",
					"label":        "Pattern Types",
					"type":         "text",
					"description":  "Comma-separated list of pattern types to identify",
					"defaultValue": "behavior_patterns,resolution_patterns",
					"required":     false,
				},
			},
		})
	}

	if hasNodeType["findings"] {
		parameters = append(parameters, map[string]interface{}{
			"id":    "findingsParams",
			"label": "Findings Analysis",
			"fields": []map[string]interface{}{
				{
					"id":           "questions",
					"label":        "Analysis Questions",
					"type":         "textarea",
					"description":  "Enter questions for findings analysis (one per line)",
					"defaultValue": "What are the most common patterns?\nWhat are the key areas for improvement?",
					"required":     false,
				},
			},
		})
	}

	// Marshal to JSON
	inputTabsJson, _ := json.Marshal(inputTabs)
	parametersJson, _ := json.Marshal(parameters)

	// Prepare and return the configuration
	return models.WorkflowExecutionConfig{
		ID:          w.ID,
		Name:        w.Name,
		Description: "Execution configuration for " + w.Name,
		InputTabs:   inputTabsJson,
		Parameters:  parametersJson,
	}, nil
}
