package db

import (
	"encoding/json"
)

// GetAllWorkflows returns all workflows from the database
func GetAllWorkflows() ([]Workflow, error) {
	rows, err := DB.Query("SELECT id, name, date, nodes, edges FROM workflows")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workflows []Workflow
	for rows.Next() {
		var workflow Workflow
		var nodesStr, edgesStr string
		
		err := rows.Scan(
			&workflow.ID,
			&workflow.Name,
			&workflow.Date,
			&nodesStr,
			&edgesStr,
		)
		if err != nil {
			return nil, err
		}
		
		workflow.Nodes = json.RawMessage(nodesStr)
		workflow.Edges = json.RawMessage(edgesStr)
		
		workflows = append(workflows, workflow)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return workflows, nil
}

// GetWorkflow retrieves a workflow by ID
func GetWorkflow(id string) (Workflow, error) {
	var workflow Workflow
	var nodesStr, edgesStr string

	err := DB.QueryRow(
		"SELECT id, name, date, nodes, edges FROM workflows WHERE id = ?",
		id,
	).Scan(
		&workflow.ID,
		&workflow.Name,
		&workflow.Date,
		&nodesStr,
		&edgesStr,
	)
	
	if err != nil {
		return Workflow{}, err
	}
	
	workflow.Nodes = json.RawMessage(nodesStr)
	workflow.Edges = json.RawMessage(edgesStr)
	
	return workflow, nil
}

// CreateWorkflow inserts a new workflow into the database
func CreateWorkflow(workflow Workflow) error {
	_, err := DB.Exec(
		"INSERT INTO workflows (id, name, date, nodes, edges) VALUES (?, ?, ?, ?, ?)",
		workflow.ID,
		workflow.Name,
		workflow.Date,
		string(workflow.Nodes),
		string(workflow.Edges),
	)
	
	return err
}

// UpdateWorkflow updates an existing workflow
func UpdateWorkflow(id string, workflow Workflow) error {
	_, err := DB.Exec(
		"UPDATE workflows SET name = ?, date = ?, nodes = ?, edges = ? WHERE id = ?",
		workflow.Name,
		workflow.Date,
		string(workflow.Nodes),
		string(workflow.Edges),
		id,
	)
	
	return err
}

// DeleteWorkflow removes a workflow from the database
func DeleteWorkflow(id string) error {
	_, err := DB.Exec("DELETE FROM workflows WHERE id = ?", id)
	return err
}

// WorkflowExists checks if a workflow with the given ID exists
func WorkflowExists(id string) (bool, error) {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM workflows WHERE id = ?)", id).Scan(&exists)
	return exists, err
} 