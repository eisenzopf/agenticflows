package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// AddTableForAnalysis adds the analysis_results table if it doesn't exist
func AddTableForAnalysis() error {
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS analysis_results (
			id TEXT PRIMARY KEY,
			workflow_id TEXT NOT NULL,
			analysis_type TEXT NOT NULL,
			results TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (workflow_id) REFERENCES workflows(id)
		)
	`)
	return err
}

// SaveAnalysisResult saves an analysis result to the database
func SaveAnalysisResult(id, workflowID, analysisType string, results interface{}) error {
	// Convert results to JSON
	resultBytes, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	// Insert into database
	_, err = DB.Exec(
		"INSERT INTO analysis_results (id, workflow_id, analysis_type, results, created_at) VALUES (?, ?, ?, ?, ?)",
		id, workflowID, analysisType, string(resultBytes), time.Now(),
	)

	return err
}

// GetAnalysisResult retrieves an analysis result by ID
func GetAnalysisResult(id string) (map[string]interface{}, error) {
	var result AnalysisResult
	var resultsStr string

	err := DB.QueryRow(
		"SELECT id, workflow_id, analysis_type, results, created_at FROM analysis_results WHERE id = ?",
		id,
	).Scan(
		&result.ID,
		&result.WorkflowID,
		&result.AnalysisType,
		&resultsStr,
		&result.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("analysis result not found")
		}
		return nil, err
	}

	// Parse results JSON
	var resultsMap map[string]interface{}
	if err := json.Unmarshal([]byte(resultsStr), &resultsMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal results: %w", err)
	}

	// Create a map with all the result data
	response := map[string]interface{}{
		"id":            result.ID,
		"workflow_id":   result.WorkflowID,
		"analysis_type": result.AnalysisType,
		"results":       resultsMap,
		"created_at":    result.CreatedAt.Format(time.RFC3339),
	}

	return response, nil
}

// GetAnalysisResultsByWorkflow retrieves all analysis results for a workflow
func GetAnalysisResultsByWorkflow(workflowID string) ([]map[string]interface{}, error) {
	rows, err := DB.Query(
		"SELECT id, workflow_id, analysis_type, results, created_at FROM analysis_results WHERE workflow_id = ? ORDER BY created_at DESC",
		workflowID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]interface{}
	for rows.Next() {
		var result AnalysisResult
		var resultsStr string

		err := rows.Scan(
			&result.ID,
			&result.WorkflowID,
			&result.AnalysisType,
			&resultsStr,
			&result.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse results JSON
		var resultsMap map[string]interface{}
		if err := json.Unmarshal([]byte(resultsStr), &resultsMap); err != nil {
			return nil, fmt.Errorf("failed to unmarshal results: %w", err)
		}

		// Create a map with all the result data
		resultMap := map[string]interface{}{
			"id":            result.ID,
			"workflow_id":   result.WorkflowID,
			"analysis_type": result.AnalysisType,
			"results":       resultsMap,
			"created_at":    result.CreatedAt.Format(time.RFC3339),
		}

		results = append(results, resultMap)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// DeleteAnalysisResult deletes an analysis result
func DeleteAnalysisResult(id string) error {
	_, err := DB.Exec("DELETE FROM analysis_results WHERE id = ?", id)
	return err
}

// AnalysisResult represents a record in the analysis_results table
type AnalysisResult struct {
	ID           string    `json:"id"`
	WorkflowID   string    `json:"workflow_id"`
	AnalysisType string    `json:"analysis_type"`
	Results      string    `json:"-"` // Stored as JSON string
	CreatedAt    time.Time `json:"created_at"`
} 