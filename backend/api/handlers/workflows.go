package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"agenticflows/backend/api/models"
	"agenticflows/backend/db"
	"agenticflows/backend/workflow"
)

// HandleWorkflows handles /api/workflows endpoint
func HandleWorkflows(w http.ResponseWriter, r *http.Request) {
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

// HandleWorkflow handles /api/workflows/{id} endpoint
func HandleWorkflow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract workflow ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/workflows/")
	pathParts := strings.Split(path, "/")
	log.Printf("DEBUG: Adjusted path parts: %v", pathParts)

	if path == "" {
		http.Error(w, "Workflow ID is required", http.StatusBadRequest)
		return
	}

	// Check if it's a request for execution config or execution
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

// handleWorkflowExecutionConfig handles /api/workflows/{id}/execution-config endpoint
func handleWorkflowExecutionConfig(w http.ResponseWriter, r *http.Request, workflowId string) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("DEBUG: Handling execution config for workflow ID: %s", workflowId)

	// Get the workflow to analyze its nodes and edges
	workflowObj, err := db.GetWorkflow(workflowId)
	if err != nil {
		log.Printf("DEBUG: Error fetching workflow for execution config: %v", err)
		http.Error(w, "Workflow not found", http.StatusNotFound)
		return
	}

	// Generate execution configuration
	config, err := workflow.GenerateExecutionConfig(workflowObj)
	if err != nil {
		log.Printf("DEBUG: Error generating execution config: %v", err)
		http.Error(w, "Failed to generate execution configuration", http.StatusInternalServerError)
		return
	}

	// Return the configuration
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(config)
}

// handleWorkflowExecute handles /api/workflows/{id}/execute endpoint
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

	// Get the workflow
	workflowObj, err := db.GetWorkflow(workflowId)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get workflow: %s", err), http.StatusNotFound)
		return
	}

	// Execute the workflow
	executor := workflow.NewExecutor(workflowObj)
	results, err := executor.Execute(req.Text, req.Data, req.Parameters)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to execute workflow: %s", err), http.StatusInternalServerError)
		return
	}

	// Return the results
	response := models.WorkflowExecutionResponse{
		WorkflowID:   workflowId,
		WorkflowName: workflowObj.Name,
		Timestamp:    time.Now(),
		Results:      results,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// HandleGenerateWorkflow handles /api/workflows/generate endpoint
func HandleGenerateWorkflow(w http.ResponseWriter, r *http.Request) {
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
	generator := workflow.NewGenerator()
	newWorkflow, err := generator.GenerateFromDescription(req.Name, req.Description)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate workflow: %s", err), http.StatusInternalServerError)
		return
	}

	// Save the generated workflow to the database
	if err := db.CreateWorkflow(newWorkflow); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save workflow: %s", err), http.StatusInternalServerError)
		return
	}

	// Return the generated workflow
	json.NewEncoder(w).Encode(newWorkflow)
}

// HandleGenerateDynamicWorkflow handles /api/workflows/generate-dynamic endpoint
func HandleGenerateDynamicWorkflow(w http.ResponseWriter, r *http.Request) {
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

	// Generate dynamic workflow
	generator := workflow.NewGenerator()
	newWorkflow, err := generator.GenerateDynamic(req.Name, req.Description)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate dynamic workflow: %s", err), http.StatusInternalServerError)
		return
	}

	// Save the generated workflow to the database
	if err := db.CreateWorkflow(newWorkflow); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save workflow: %s", err), http.StatusInternalServerError)
		return
	}

	// Return the generated workflow
	json.NewEncoder(w).Encode(newWorkflow)
}
