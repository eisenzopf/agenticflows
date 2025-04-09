package main

import (
	"context"
	"log"
	"net/http"

	"agenticflows/backend/api/handlers"
	"agenticflows/backend/db"
)

// Main entry point for the API server
func main() {
	// Initialize database
	if err := db.Initialize(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize analysis handler
	analysisHandler, err := handlers.NewAnalysisHandler()
	if err != nil {
		log.Printf("Warning: Failed to initialize analysis handler: %v", err)
		log.Println("Analysis endpoints will not be available")
	}

	// Set up API routes
	setupRoutes(analysisHandler)

	// CORS middleware for development
	handler := corsMiddleware(http.DefaultServeMux)

	// Start server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

// CORS middleware for development
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

// setupRoutes configures all API routes
func setupRoutes(analysisHandler *handlers.AnalysisHandler) {
	// Basic API routes
	http.HandleFunc("/api/agents", handlers.HandleAgents)
	http.HandleFunc("/api/tools", handlers.HandleTools)
	http.HandleFunc("/api/workflows", handlers.HandleWorkflows)
	http.HandleFunc("/api/workflows/", handlers.HandleWorkflow)

	// Workflow generation endpoints
	http.HandleFunc("/api/workflows/generate", handlers.HandleGenerateWorkflow)
	http.HandleFunc("/api/workflows/generate-dynamic", handlers.HandleGenerateDynamicWorkflow)

	// Question answering endpoint
	// We need to pass the analysis handler to the questions handler
	http.HandleFunc("/api/questions/answer", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx = context.WithValue(ctx, "analysisHandler", analysisHandler)
		handlers.HandleAnswerQuestions(w, r.WithContext(ctx))
	})

	// Analysis routes (if initialized)
	if analysisHandler != nil {
		// New unified endpoint
		http.HandleFunc("/api/analysis", analysisHandler.HandleAnalysis)

		// Chain analysis endpoint for workflows
		http.HandleFunc("/api/analysis/chain", analysisHandler.HandleChainAnalysis)

		// Function metadata endpoint
		http.HandleFunc("/api/analysis/metadata", analysisHandler.HandleGetFunctionMetadata)

		// Enable debugging for analysis requests
		http.HandleFunc("/api/analysis/results", analysisHandler.HandleAnalysisResults)
	}
} 