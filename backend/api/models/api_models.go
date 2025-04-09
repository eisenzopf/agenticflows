package models

import (
	"encoding/json"
	"time"
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

// WorkflowExecutionResponse represents the response from a workflow execution
type WorkflowExecutionResponse struct {
	WorkflowID   string                 `json:"workflow_id"`
	WorkflowName string                 `json:"workflow_name"`
	Timestamp    time.Time              `json:"timestamp"`
	Results      map[string]interface{} `json:"results"`
}

// QuestionRequest represents a request to answer questions
type QuestionRequest struct {
	Questions    []string               `json:"questions"`
	Context      string                 `json:"context,omitempty"`
	DatabasePath string                 `json:"databasePath,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

// QuestionResponse represents a response to a question
type QuestionResponse struct {
	Answers []QuestionAnswer `json:"answers"`
}

// QuestionAnswer represents an answer to a specific question
type QuestionAnswer struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
	Error    bool   `json:"error,omitempty"`
}
