package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"agenticflows/backend/analysis/models"      // Analysis models
	apimodels "agenticflows/backend/api/models" // API models with alias
)

// HandleAnswerQuestions processes questions about the banking conversations
func HandleAnswerQuestions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req apimodels.QuestionRequest
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

	// Get analysis handler to use for processing questions
	analysisHandler, ok := r.Context().Value("analysisHandler").(*AnalysisHandler)
	if !ok || analysisHandler == nil {
		http.Error(w, "Analysis handler not available", http.StatusInternalServerError)
		return
	}

	// Process each question
	answers := make([]apimodels.QuestionAnswer, 0)
	for _, question := range req.Questions {
		// Create analysis request for this question
		analysisReq := models.StandardAnalysisRequest{
			AnalysisType: "findings",
			Text:         contextData,
			Parameters: map[string]interface{}{
				"questions": []string{question},
			},
		}

		// Execute the analysis
		response, err := analysisHandler.handleFindingsAnalysis(ctx, analysisReq)
		if err != nil {
			answers = append(answers, apimodels.QuestionAnswer{
				Question: question,
				Answer:   fmt.Sprintf("Error analyzing question: %s", err),
				Error:    true,
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
		answers = append(answers, apimodels.QuestionAnswer{
			Question: question,
			Answer:   answer,
		})
	}

	// Return the answers
	json.NewEncoder(w).Encode(apimodels.QuestionResponse{
		Answers: answers,
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
