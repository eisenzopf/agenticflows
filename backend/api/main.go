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
	id := strings.TrimPrefix(r.URL.Path, "/api/workflows/")
	if id == "" {
		http.Error(w, "Workflow ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		// Get a specific workflow
		workflow, err := db.GetWorkflow(id)
		if err != nil {
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
}
