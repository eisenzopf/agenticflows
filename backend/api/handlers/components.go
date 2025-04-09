package handlers

import (
	"encoding/json"
	"net/http"

	"agenticflows/backend/db"
)

// HandleAgents handles the /api/agents endpoint
func HandleAgents(w http.ResponseWriter, r *http.Request) {
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

// HandleTools handles the /api/tools endpoint
func HandleTools(w http.ResponseWriter, r *http.Request) {
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
