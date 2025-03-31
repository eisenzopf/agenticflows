package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

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
