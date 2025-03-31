package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"agenticflows/backend/analysis"
	"agenticflows/backend/db"
)

// FlowData represents a flow configuration
type FlowData struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Nodes json.RawMessage `json:"nodes"`
	Edges json.RawMessage `json:"edges"`
}

// WorkflowExecutionConfig represents a configuration for executing a workflow
type WorkflowExecutionConfig struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputTabs   json.RawMessage `json:"inputTabs"`
	Parameters  json.RawMessage `json:"parameters"`
}

var flows = make(map[string]FlowData)
var analysisHandler *AnalysisHandler

func main() {
	// Initialize database
	if err := db.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize analysis handler
	var err error
	analysisHandler, err = NewAnalysisHandler()
	if err != nil {
		log.Printf("Warning: Failed to initialize analysis handler: %v", err)
		log.Println("Analysis endpoints will not be available")
	}

	// API routes
	http.HandleFunc("/api/agents", handleAgents)
	http.HandleFunc("/api/tools", handleTools)
	http.HandleFunc("/api/workflows", handleWorkflows)
	http.HandleFunc("/api/workflows/", handleWorkflow)

	// Add new workflow generation endpoint
	http.HandleFunc("/api/workflows/generate", handleGenerateWorkflow)

	// Add new question answering endpoint
	http.HandleFunc("/api/questions/answer", handleAnswerQuestions)

	// Analysis routes (if initialized)
	if analysisHandler != nil {
		// New unified endpoint
		http.HandleFunc("/api/analysis", analysisHandler.handleAnalysis)

		// Chain analysis endpoint for workflows
		http.HandleFunc("/api/analysis/chain", analysisHandler.handleChainAnalysis)

		// Function metadata endpoint
		http.HandleFunc("/api/analysis/metadata", analysisHandler.handleGetFunctionMetadata)

		// Enable debugging for analysis requests
		log.Println("Analysis endpoints initialized with types: trends, patterns, findings, attributes, intent, recommendations, plan")

		// Legacy endpoints (kept for backward compatibility)
		/*http.HandleFunc("/api/analysis/trends", analysisHandler.handleAnalysisTrends)
		http.HandleFunc("/api/analysis/patterns", analysisHandler.handleAnalysisPatterns)
		http.HandleFunc("/api/analysis/findings", analysisHandler.handleAnalysisFindings)
		http.HandleFunc("/api/analysis/attributes", analysisHandler.handleTextAttributes)
		http.HandleFunc("/api/analysis/intent", analysisHandler.handleTextIntent)
		http.HandleFunc("/api/analysis/results", analysisHandler.handleAnalysisResults)
		http.HandleFunc("/api/analysis/results/", analysisHandler.handleAnalysisResults)*/
	}

	// CORS middleware for development
	handler := corsMiddleware(http.DefaultServeMux)

	// Start server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Handle /api/agents endpoint
func handleAgents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// Return all agents
		agents, err := db.GetAllAgents()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(agents)

	case "POST":
		// Create a new agent
		var agent db.Agent
		if err := json.NewDecoder(r.Body).Decode(&agent); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Simple validation
		if agent.ID == "" || agent.Label == "" {
			http.Error(w, "ID and Label are required", http.StatusBadRequest)
			return
		}
		agent.Type = "agent"

		if err := db.AddAgent(agent); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(agent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handle /api/tools endpoint
func handleTools(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// Return all tools
		tools, err := db.GetAllTools()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(tools)

	case "POST":
		// Create a new tool
		var tool db.Tool
		if err := json.NewDecoder(r.Body).Decode(&tool); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Simple validation
		if tool.ID == "" || tool.Label == "" {
			http.Error(w, "ID and Label are required", http.StatusBadRequest)
			return
		}
		tool.Type = "tool"

		if err := db.AddTool(tool); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(tool)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handle /api/workflows endpoint
func handleWorkflows(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// Return all workflows
		workflows, err := db.GetAllWorkflows()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(workflows)

	case "POST":
		// Create a new workflow
		var workflow db.Workflow
		if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Simple validation
		if workflow.ID == "" || workflow.Name == "" {
			http.Error(w, "ID and Name are required", http.StatusBadRequest)
			return
		}

		// Set date if not provided
		if workflow.Date == "" {
			workflow.Date = time.Now().Format("2006-01-02")
		}

		if err := db.CreateWorkflow(workflow); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(workflow)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Handle /api/workflows/{id} endpoint
func handleWorkflow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract workflow ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/workflows/")
	pathParts := strings.Split(path, "/")
	log.Printf("DEBUG: Adjusted path parts: %v", pathParts)

	if path == "" {
		http.Error(w, "Workflow ID is required", http.StatusBadRequest)
		return
	}

	// Check if it's a request for execution config
	if len(pathParts) >= 1 && pathParts[0] != "" {
		id := pathParts[0]

		// Check if it's a request for execution config
		if len(pathParts) > 1 && pathParts[1] == "execution-config" {
			log.Printf("DEBUG: Handling execution config request for workflow: %s", id)
			handleWorkflowExecutionConfig(w, r, id)
			return
		}

		// Check if it's a request to execute the workflow
		if len(pathParts) > 1 && pathParts[1] == "execute" {
			log.Printf("DEBUG: Handling execute request for workflow: %s", id)
			handleWorkflowExecute(w, r, id)
			return
		}

		log.Printf("DEBUG: Handling workflow request for ID: %s", id)

		switch r.Method {
		case "GET":
			// Get a specific workflow
			workflow, err := db.GetWorkflow(id)
			if err != nil {
				log.Printf("DEBUG: Error in GetWorkflow: %v (type: %T)", err, err)

				// Check if the workflow exists with direct database query
				exists, checkErr := db.WorkflowExists(id)
				if checkErr != nil {
					log.Printf("DEBUG: Error in WorkflowExists check: %v", checkErr)
				} else {
					log.Printf("DEBUG: WorkflowExists result: %v", exists)
				}

				// List all workflows in the database for debugging
				allWorkflows, listErr := db.GetAllWorkflows()
				if listErr != nil {
					log.Printf("DEBUG: Error listing all workflows: %v", listErr)
				} else {
					log.Printf("DEBUG: All workflows in DB:")
					for _, w := range allWorkflows {
						log.Printf("  - ID: %s, Name: %s", w.ID, w.Name)
					}
				}

				http.Error(w, "Workflow not found", http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(workflow)

		case "PUT":
			// Update a workflow
			var updatedWorkflow db.Workflow
			if err := json.NewDecoder(r.Body).Decode(&updatedWorkflow); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// Check if workflow exists
			exists, err := db.WorkflowExists(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !exists {
				http.Error(w, "Workflow not found", http.StatusNotFound)
				return
			}

			// Update the date
			if updatedWorkflow.Date == "" {
				updatedWorkflow.Date = time.Now().Format("2006-01-02")
			}

			// Ensure ID consistency
			updatedWorkflow.ID = id

			if err := db.UpdateWorkflow(id, updatedWorkflow); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(updatedWorkflow)

		case "DELETE":
			// Delete a workflow
			exists, err := db.WorkflowExists(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !exists {
				http.Error(w, "Workflow not found", http.StatusNotFound)
				return
			}

			if err := db.DeleteWorkflow(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	} else {
		http.Error(w, "Workflow ID is required", http.StatusBadRequest)
	}
}

// Handle /api/workflows/{id}/execution-config endpoint
func handleWorkflowExecutionConfig(w http.ResponseWriter, r *http.Request, workflowId string) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("DEBUG: Handling execution config for workflow ID: %s", workflowId)

	// Get the workflow to analyze its nodes and edges
	workflow, err := db.GetWorkflow(workflowId)
	if err != nil {
		log.Printf("DEBUG: Error fetching workflow for execution config: %v", err)
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}

	// Parse workflow nodes to detect types of analysis
	var nodes []map[string]interface{}
	var edges []map[string]interface{}

	err = json.Unmarshal([]byte(workflow.Nodes), &nodes)
	if err != nil {
		log.Printf("DEBUG: Error parsing workflow nodes: %v", err)
		http.Error(w, "Invalid workflow nodes structure", http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal([]byte(workflow.Edges), &edges)
	if err != nil {
		// Non-critical error for configuration
		log.Printf("Warning: Failed to parse edges for workflow %s: %v", workflowId, err)
	}

	// Build execution configuration based on workflow components
	config := generateWorkflowExecutionConfig(workflowId, workflow.Name, nodes, edges)

	// Return the configuration
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// Generate workflow execution configuration based on workflow components
func generateWorkflowExecutionConfig(workflowId string, workflowName string, nodes []map[string]interface{}, edges []map[string]interface{}) WorkflowExecutionConfig {
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
	return WorkflowExecutionConfig{
		ID:          workflowId,
		Name:        workflowName,
		Description: "Execution configuration for " + workflowName,
		InputTabs:   inputTabsJson,
		Parameters:  parametersJson,
	}
}

// Handle /api/workflows/{id}/execute endpoint
func handleWorkflowExecute(w http.ResponseWriter, r *http.Request, workflowId string) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Parameters map[string]interface{} `json:"parameters"`
		Data       map[string]interface{} `json:"data"`
		Text       string                 `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
		return
	}

	// Get the workflow from the database
	workflow, err := db.GetWorkflow(workflowId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get workflow: %s", err), http.StatusNotFound)
		return
	}

	// Parse the workflow nodes and edges
	var nodes []map[string]interface{}
	var edges []map[string]interface{}

	if err := json.Unmarshal([]byte(workflow.Nodes), &nodes); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse workflow nodes: %s", err), http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal([]byte(workflow.Edges), &edges); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse workflow edges: %s", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Executing workflow '%s' with %d nodes and %d edges", workflow.Name, len(nodes), len(edges))

	// Find all function nodes
	functionNodes := make([]map[string]interface{}, 0)
	for _, node := range nodes {
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
	sortedNodes, err := getExecutionOrder(functionNodes, edges)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to determine execution order: %s", err), http.StatusInternalServerError)
		return
	}

	// Initialize results storage
	results := make(map[string]interface{})

	// Add initial data to results
	if req.Data != nil {
		for k, v := range req.Data {
			results[k] = v
		}
	}

	// If text was provided, add it to results
	if req.Text != "" {
		results["text"] = req.Text
	}

	// Add any additional parameters
	if req.Parameters != nil {
		for k, v := range req.Parameters {
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
		functionType := parts[1]

		log.Printf("Executing node %s of type %s", nodeID, functionType)

		// Get input data from connected nodes
		nodeInputs := make(map[string]interface{})

		// Find incoming edges to this node
		for _, edge := range edges {
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

		// Execute the node based on its function type
		var nodeResult map[string]interface{}

		switch functionType {
		case "attributes":
			// Handle attributes analysis
			attrResult, err := executeAttributesAnalysis(nodeInputs, workflowId)
			if err != nil {
				log.Printf("Error executing attributes analysis: %v", err)
				results[nodeID] = map[string]interface{}{
					"error": fmt.Sprintf("Failed to execute attributes analysis: %v", err),
				}
				continue
			}
			nodeResult = attrResult

		case "intent":
			// Handle intent analysis
			intentResult, err := executeIntentAnalysis(nodeInputs, workflowId)
			if err != nil {
				log.Printf("Error executing intent analysis: %v", err)
				results[nodeID] = map[string]interface{}{
					"error": fmt.Sprintf("Failed to execute intent analysis: %v", err),
				}
				continue
			}
			nodeResult = intentResult

		case "trends":
			// Handle trends analysis
			trendsResult, err := executeTrendsAnalysis(nodeInputs, workflowId)
			if err != nil {
				log.Printf("Error executing trends analysis: %v", err)
				results[nodeID] = map[string]interface{}{
					"error": fmt.Sprintf("Failed to execute trends analysis: %v", err),
				}
				continue
			}
			nodeResult = trendsResult

		case "patterns":
			// Handle patterns analysis
			patternsResult, err := executePatternsAnalysis(nodeInputs, workflowId)
			if err != nil {
				log.Printf("Error executing patterns analysis: %v", err)
				results[nodeID] = map[string]interface{}{
					"error": fmt.Sprintf("Failed to execute patterns analysis: %v", err),
				}
				continue
			}
			nodeResult = patternsResult

		case "findings":
			// Handle findings analysis
			findingsResult, err := executeFindingsAnalysis(nodeInputs, workflowId)
			if err != nil {
				log.Printf("Error executing findings analysis: %v", err)
				results[nodeID] = map[string]interface{}{
					"error": fmt.Sprintf("Failed to execute findings analysis: %v", err),
				}
				continue
			}
			nodeResult = findingsResult

		case "recommendations":
			// Handle recommendations
			recommendationsResult, err := executeRecommendationsAnalysis(nodeInputs, workflowId)
			if err != nil {
				log.Printf("Error executing recommendations analysis: %v", err)
				results[nodeID] = map[string]interface{}{
					"error": fmt.Sprintf("Failed to execute recommendations analysis: %v", err),
				}
				continue
			}
			nodeResult = recommendationsResult

		case "plan":
			// Handle action plan
			planResult, err := executePlanAnalysis(nodeInputs, workflowId)
			if err != nil {
				log.Printf("Error executing plan analysis: %v", err)
				results[nodeID] = map[string]interface{}{
					"error": fmt.Sprintf("Failed to execute plan analysis: %v", err),
				}
				continue
			}
			nodeResult = planResult

		default:
			log.Printf("Unknown function type: %s", functionType)
			results[nodeID] = map[string]interface{}{
				"error": fmt.Sprintf("Unknown function type: %s", functionType),
			}
			continue
		}

		// Store results
		results[nodeID] = nodeResult
	}

	// Return the complete workflow results
	response := map[string]interface{}{
		"workflow_id":   workflowId,
		"workflow_name": workflow.Name,
		"timestamp":     time.Now(),
		"results":       results,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// getExecutionOrder sorts nodes by dependencies to allow for proper execution order
func getExecutionOrder(nodes []map[string]interface{}, edges []map[string]interface{}) ([]map[string]interface{}, error) {
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
	for _, edge := range edges {
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

// executeAttributesAnalysis performs attributes analysis using the analysis handler
func executeAttributesAnalysis(inputs map[string]interface{}, workflowID string) (map[string]interface{}, error) {
	if analysisHandler == nil {
		return nil, fmt.Errorf("analysis handler not initialized")
	}

	// Prepare the request
	req := analysis.StandardAnalysisRequest{
		AnalysisType: "attributes",
		WorkflowID:   workflowID,
		Text:         getStringValue(inputs, "text"),
		Parameters:   make(map[string]interface{}),
	}

	// Handle attributes properly - check both direct and nested locations
	if attributes, ok := inputs["attributes"]; ok {
		req.Parameters["attributes"] = attributes
	}

	// Add other parameters
	if params, ok := inputs["parameters"].(map[string]interface{}); ok {
		for k, v := range params {
			req.Parameters[k] = v
		}
	}

	// Execute the analysis
	resp, err := analysisHandler.handleAttributesAnalysis(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Convert to map for consistent return type
	resultMap := make(map[string]interface{})

	if resp != nil && resp.Results != nil {
		resultMap = convertToMap(resp.Results)
	}

	return resultMap, nil
}

// executeIntentAnalysis performs intent analysis using the analysis handler
func executeIntentAnalysis(inputs map[string]interface{}, workflowID string) (map[string]interface{}, error) {
	if analysisHandler == nil {
		return nil, fmt.Errorf("analysis handler not initialized")
	}

	// Prepare the request
	req := analysis.StandardAnalysisRequest{
		AnalysisType: "intent",
		WorkflowID:   workflowID,
		Text:         getStringValue(inputs, "text"),
		Parameters:   make(map[string]interface{}),
	}

	// Execute the analysis
	resp, err := analysisHandler.handleIntentAnalysis(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Convert to map for consistent return type
	resultMap := make(map[string]interface{})

	if resp != nil && resp.Results != nil {
		resultMap = convertToMap(resp.Results)
	}

	return resultMap, nil
}

// executeTrendsAnalysis performs trends analysis using the analysis handler
func executeTrendsAnalysis(inputs map[string]interface{}, workflowID string) (map[string]interface{}, error) {
	if analysisHandler == nil {
		return nil, fmt.Errorf("analysis handler not initialized")
	}

	// Prepare the request
	req := analysis.StandardAnalysisRequest{
		AnalysisType: "trends",
		WorkflowID:   workflowID,
		Text:         getStringValue(inputs, "text"),
		Data:         inputs,
		Parameters:   make(map[string]interface{}),
	}

	// Add focus areas if present
	if focusAreas, ok := inputs["focusAreas"]; ok {
		req.Parameters["focus_areas"] = focusAreas
	}

	// Execute the analysis
	resp, err := analysisHandler.handleTrendsAnalysis(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Convert to map for consistent return type
	resultMap := make(map[string]interface{})

	if resp != nil && resp.Results != nil {
		resultMap = convertToMap(resp.Results)
	}

	return resultMap, nil
}

// executePatternsAnalysis performs patterns analysis using the analysis handler
func executePatternsAnalysis(inputs map[string]interface{}, workflowID string) (map[string]interface{}, error) {
	if analysisHandler == nil {
		return nil, fmt.Errorf("analysis handler not initialized")
	}

	// Prepare the request
	req := analysis.StandardAnalysisRequest{
		AnalysisType: "patterns",
		WorkflowID:   workflowID,
		Text:         getStringValue(inputs, "text"),
		Data:         inputs,
		Parameters:   make(map[string]interface{}),
	}

	// Add pattern types if present
	if patternTypes, ok := inputs["patternTypes"]; ok {
		req.Parameters["pattern_types"] = patternTypes
	}

	// Execute the analysis
	resp, err := analysisHandler.handlePatternsAnalysis(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Convert to map for consistent return type
	resultMap := make(map[string]interface{})

	if resp != nil && resp.Results != nil {
		resultMap = convertToMap(resp.Results)
	}

	return resultMap, nil
}

// executeFindingsAnalysis performs findings analysis using the analysis handler
func executeFindingsAnalysis(inputs map[string]interface{}, workflowID string) (map[string]interface{}, error) {
	if analysisHandler == nil {
		return nil, fmt.Errorf("analysis handler not initialized")
	}

	// Prepare the request
	req := analysis.StandardAnalysisRequest{
		AnalysisType: "findings",
		WorkflowID:   workflowID,
		Text:         getStringValue(inputs, "text"),
		Data:         inputs,
		Parameters:   make(map[string]interface{}),
	}

	// Extract questions from the inputs
	if questionsRaw, ok := inputs["questions"]; ok {
		req.Parameters["questions"] = questionsRaw
	}

	// Execute the analysis
	resp, err := analysisHandler.handleFindingsAnalysis(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Convert to map for consistent return type
	resultMap := make(map[string]interface{})

	if resp != nil && resp.Results != nil {
		resultMap = convertToMap(resp.Results)
	}

	return resultMap, nil
}

// executeRecommendationsAnalysis performs recommendations analysis using the analysis handler
func executeRecommendationsAnalysis(inputs map[string]interface{}, workflowID string) (map[string]interface{}, error) {
	if analysisHandler == nil {
		return nil, fmt.Errorf("analysis handler not initialized")
	}

	// Prepare the request
	req := analysis.StandardAnalysisRequest{
		AnalysisType: "recommendations",
		WorkflowID:   workflowID,
		Text:         getStringValue(inputs, "text"),
		Data:         inputs,
		Parameters:   make(map[string]interface{}),
	}

	// Add focus area if present
	if focusArea, ok := inputs["focusArea"]; ok {
		req.Parameters["focus_area"] = focusArea
	}

	// Add criteria if present
	if criteria, ok := inputs["criteria"]; ok {
		req.Parameters["criteria"] = criteria
	}

	// Execute the analysis
	resp, err := analysisHandler.handleRecommendationsAnalysis(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Convert to map for consistent return type
	resultMap := make(map[string]interface{})

	if resp != nil && resp.Results != nil {
		resultMap = convertToMap(resp.Results)
	}

	return resultMap, nil
}

// executePlanAnalysis performs action plan analysis using the analysis handler
func executePlanAnalysis(inputs map[string]interface{}, workflowID string) (map[string]interface{}, error) {
	if analysisHandler == nil {
		return nil, fmt.Errorf("analysis handler not initialized")
	}

	// Prepare the request
	req := analysis.StandardAnalysisRequest{
		AnalysisType: "plan",
		WorkflowID:   workflowID,
		Text:         getStringValue(inputs, "text"),
		Data:         inputs,
		Parameters:   make(map[string]interface{}),
	}

	// Add recommendations if present
	if recommendations, ok := inputs["recommendations"]; ok {
		req.Parameters["recommendations"] = recommendations
	}

	// Add timeline requirements if present
	if timeline, ok := inputs["timeline"]; ok {
		req.Parameters["timeline"] = timeline
	}

	// Execute the analysis
	resp, err := analysisHandler.handlePlanAnalysis(context.Background(), req)
	if err != nil {
		return nil, err
	}

	// Convert to map for consistent return type
	resultMap := make(map[string]interface{})

	if resp != nil && resp.Results != nil {
		resultMap = convertToMap(resp.Results)
	}

	return resultMap, nil
}

// Helper function to get a string value from a map
func getStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

// Helper function to convert an interface{} to a map
func convertToMap(data interface{}) map[string]interface{} {
	if data == nil {
		return make(map[string]interface{})
	}

	// If it's already a map, return it
	if m, ok := data.(map[string]interface{}); ok {
		return m
	}

	// Otherwise, try to marshal and unmarshal to convert
	jsonData, err := json.Marshal(data)
	if err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to marshal data: %v", err),
		}
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(jsonData, &result); err != nil {
		return map[string]interface{}{
			"error": fmt.Sprintf("Failed to unmarshal data: %v", err),
		}
	}

	return result
}

// Create a new handler for workflow generation
func handleGenerateWorkflow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" || req.Description == "" {
		http.Error(w, "Name and description are required", http.StatusBadRequest)
		return
	}

	// Generate workflow
	workflow, err := generateWorkflowFromDescription(req.Name, req.Description)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate workflow: %s", err), http.StatusInternalServerError)
		return
	}

	// Save the generated workflow to the database
	if err := db.CreateWorkflow(workflow); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save workflow: %s", err), http.StatusInternalServerError)
		return
	}

	// Return the generated workflow
	json.NewEncoder(w).Encode(workflow)
}

// generateWorkflowFromDescription uses LLM to generate a workflow based on the description
func generateWorkflowFromDescription(name, description string) (db.Workflow, error) {
	// Get the API key from environment
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return db.Workflow{}, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	// Create LLM client
	llmClient, err := analysis.NewLLMClient(apiKey, false)
	if err != nil {
		return db.Workflow{}, fmt.Errorf("failed to create LLM client: %s", err)
	}

	// Create function metadata for the prompt
	functionMetadata, err := getFunctionMetadataForLLM()
	if err != nil {
		return db.Workflow{}, fmt.Errorf("failed to get function metadata: %s", err)
	}

	// Create the prompt
	prompt := createWorkflowGenerationPrompt(name, description, functionMetadata)

	// Call the LLM API
	result, err := llmClient.GenerateContent(context.Background(), prompt, map[string]interface{}{
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
		{
			"id":          "analysis-findings",
			"label":       "Analyze Findings",
			"description": "Generate findings from analysis results",
			"inputs": []map[string]interface{}{
				{
					"name":        "Questions",
					"description": "Questions to answer based on the analysis",
					"required":    true,
				},
				{
					"name":        "Text",
					"description": "Text to analyze for findings",
					"required":    false,
				},
			},
			"outputs": []map[string]interface{}{
				{
					"name":        "Findings",
					"description": "Key findings and insights",
				},
				{
					"name":        "Evidence",
					"description": "Supporting evidence for findings",
				},
			},
		},
		{
			"id":          "analysis-attributes",
			"label":       "Extract Attributes",
			"description": "Extract attributes from text",
			"inputs": []map[string]interface{}{
				{
					"name":        "Attributes",
					"description": "Attributes to extract",
					"required":    true,
				},
				{
					"name":        "Text",
					"description": "Text to extract attributes from",
					"required":    true,
				},
			},
			"outputs": []map[string]interface{}{
				{
					"name":        "Extracted Attributes",
					"description": "Extracted attribute values",
				},
				{
					"name":        "Confidence",
					"description": "Confidence scores for extracted attributes",
				},
			},
		},
		{
			"id":          "analysis-intent",
			"label":       "Generate Intent",
			"description": "Generate intent from text",
			"inputs": []map[string]interface{}{
				{
					"name":        "Text",
					"description": "Text to analyze for intent",
					"required":    true,
				},
			},
			"outputs": []map[string]interface{}{
				{
					"name":        "Intent",
					"description": "Identified intent",
				},
				{
					"name":        "Confidence",
					"description": "Confidence score for intent",
				},
			},
		},
		{
			"id":          "analysis-recommendations",
			"label":       "Generate Recommendations",
			"description": "Generate recommendations based on analysis results",
			"inputs": []map[string]interface{}{
				{
					"name":        "Focus Area",
					"description": "Area to focus recommendations on",
					"required":    true,
				},
				{
					"name":        "Findings",
					"description": "Findings to base recommendations on",
					"required":    true,
				},
			},
			"outputs": []map[string]interface{}{
				{
					"name":        "Recommendations",
					"description": "List of recommended actions",
				},
				{
					"name":        "Priority",
					"description": "Priority levels for recommendations",
				},
			},
		},
		{
			"id":          "analysis-plan",
			"label":       "Create Action Plan",
			"description": "Create action plan from recommendations",
			"inputs": []map[string]interface{}{
				{
					"name":        "Recommendations",
					"description": "Recommendations to include in the plan",
					"required":    true,
				},
				{
					"name":        "Timeline",
					"description": "Timeline requirements",
					"required":    false,
				},
			},
			"outputs": []map[string]interface{}{
				{
					"name":        "Plan",
					"description": "Detailed action plan",
				},
				{
					"name":        "Timeline",
					"description": "Implementation timeline",
				},
			},
		},
		{
			"id":          "analysis-chain",
			"label":       "Chain Analysis",
			"description": "Perform a complete chain of analysis steps",
			"inputs": []map[string]interface{}{
				{
					"name":        "Text",
					"description": "Text to analyze",
					"required":    true,
				},
				{
					"name":        "Config",
					"description": "Analysis configuration",
					"required":    true,
				},
			},
			"outputs": []map[string]interface{}{
				{
					"name":        "Results",
					"description": "Combined analysis results",
				},
			},
		},
	}

	return metadata, nil
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

// handleAnswerQuestions processes questions about the banking conversations
func handleAnswerQuestions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Questions    []string               `json:"questions"`
		Context      string                 `json:"context,omitempty"`
		DatabasePath string                 `json:"databasePath,omitempty"`
		Parameters   map[string]interface{} `json:"parameters,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %s", err), http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.Questions) == 0 {
		http.Error(w, "At least one question is required", http.StatusBadRequest)
		return
	}

	// Set default database path if not provided
	dbPath := req.DatabasePath
	if dbPath == "" {
		dbPath = "/Users/jonathan/Documents/Work/discourse_ai/Research/corpora/banking_2025/db/standard_charter_bank.db"
	}

	// Create context for the analysis
	ctx := context.Background()

	// Initialize context data if not provided
	contextData := req.Context
	if contextData == "" {
		// Fetch sample conversations from the database
		var err error
		contextData, err = getSampleConversationsFromDB(dbPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get sample conversations: %s", err), http.StatusInternalServerError)
			return
		}
	}

	// Process each question
	answers := make([]map[string]interface{}, 0)
	for _, question := range req.Questions {
		// Create analysis request for this question
		analysisReq := analysis.StandardAnalysisRequest{
			AnalysisType: "findings",
			Text:         contextData,
			Parameters: map[string]interface{}{
				"questions": []string{question},
			},
		}

		// Execute the analysis
		response, err := analysisHandler.handleFindingsAnalysis(ctx, analysisReq)
		if err != nil {
			answers = append(answers, map[string]interface{}{
				"question": question,
				"answer":   fmt.Sprintf("Error analyzing question: %s", err),
				"error":    true,
			})
			continue
		}

		// Get the answer from the results
		var answer string
		if response.Results != nil {
			// Extract the first finding as the answer
			findings, ok := extractFindingsFromResponse(response.Results)
			if ok && len(findings) > 0 {
				answer = findings[0]
			} else {
				answer = "No findings were generated for this question."
			}
		} else {
			answer = "No response was generated for this question."
		}

		// Add to answers
		answers = append(answers, map[string]interface{}{
			"question": question,
			"answer":   answer,
		})
	}

	// Return the answers
	json.NewEncoder(w).Encode(map[string]interface{}{
		"answers": answers,
	})
}

// Helper function to get sample conversations from database
func getSampleConversationsFromDB(dbPath string) (string, error) {
	// Open the database
	sqliteDB, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open database: %s", err)
	}
	defer sqliteDB.Close()

	// Query for sample conversations
	query := `SELECT text FROM conversations LIMIT 10`
	rows, err := sqliteDB.Query(query)
	if err != nil {
		return "", fmt.Errorf("failed to query conversations: %s", err)
	}
	defer rows.Close()

	// Build the context from conversations
	var conversations []string
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			return "", fmt.Errorf("failed to scan row: %s", err)
		}
		conversations = append(conversations, text)
	}

	if err := rows.Err(); err != nil {
		return "", fmt.Errorf("error iterating rows: %s", err)
	}

	if len(conversations) == 0 {
		return "No conversations found in the database.", nil
	}

	return strings.Join(conversations, "\n\n---\n\n"), nil
}

// Helper to extract findings from a response
func extractFindingsFromResponse(results interface{}) ([]string, bool) {
	// Try to cast directly to map
	resultsMap, ok := results.(map[string]interface{})
	if !ok {
		// Try to unmarshal from JSON if it's a string
		if strResults, ok := results.(string); ok {
			var parsedResults map[string]interface{}
			if err := json.Unmarshal([]byte(strResults), &parsedResults); err == nil {
				resultsMap = parsedResults
				ok = true
			}
		}
	}

	if !ok {
		return nil, false
	}

	// Look for findings in various possible formats
	if findings, ok := resultsMap["findings"].([]interface{}); ok {
		// Convert findings to strings
		stringFindings := make([]string, 0, len(findings))
		for _, finding := range findings {
			if str, ok := finding.(string); ok {
				stringFindings = append(stringFindings, str)
			} else {
				// Try to marshal to string
				if bytes, err := json.Marshal(finding); err == nil {
					stringFindings = append(stringFindings, string(bytes))
				}
			}
		}
		return stringFindings, true
	}

	return nil, false
}
