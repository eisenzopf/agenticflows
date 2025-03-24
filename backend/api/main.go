package main

import (
	"encoding/json"
	"log"
	"net/http"
)

// FlowData represents a flow configuration
type FlowData struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Nodes json.RawMessage `json:"nodes"`
	Edges json.RawMessage `json:"edges"`
}

var flows = make(map[string]FlowData)

func main() {
	// API routes
	http.HandleFunc("/api/flows", handleFlows)
	http.HandleFunc("/api/flows/", handleFlow)

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

func handleFlows(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		// Return all flows
		flowsList := make([]FlowData, 0, len(flows))
		for _, flow := range flows {
			flowsList = append(flowsList, flow)
		}
		json.NewEncoder(w).Encode(flowsList)

	case "POST":
		// Create a new flow
		var flow FlowData
		if err := json.NewDecoder(r.Body).Decode(&flow); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Simple validation
		if flow.ID == "" || flow.Name == "" {
			http.Error(w, "ID and Name are required", http.StatusBadRequest)
			return
		}

		flows[flow.ID] = flow
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(flow)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleFlow(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract flow ID from URL
	id := r.URL.Path[len("/api/flows/"):]
	if id == "" {
		http.Error(w, "Flow ID is required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		// Get a specific flow
		flow, exists := flows[id]
		if !exists {
			http.Error(w, "Flow not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(flow)

	case "PUT":
		// Update a flow
		var updatedFlow FlowData
		if err := json.NewDecoder(r.Body).Decode(&updatedFlow); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, exists := flows[id]; !exists {
			http.Error(w, "Flow not found", http.StatusNotFound)
			return
		}

		updatedFlow.ID = id // Ensure ID consistency
		flows[id] = updatedFlow
		json.NewEncoder(w).Encode(updatedFlow)

	case "DELETE":
		// Delete a flow
		if _, exists := flows[id]; !exists {
			http.Error(w, "Flow not found", http.StatusNotFound)
			return
		}

		delete(flows, id)
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
} 