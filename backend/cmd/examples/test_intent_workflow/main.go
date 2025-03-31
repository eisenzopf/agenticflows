package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Sample intents for testing
var sampleIntents = []string{
	"Customer wants to check order status",
	"Customer is reporting a technical issue",
	"Customer wants to cancel their subscription",
	"Customer is asking about pricing",
	"Customer wants to upgrade their plan",
}

func main() {
	// Create a workflow with an intent generation node
	workflow := createIntentWorkflow()

	// Create the workflow via API
	workflowID, err := createWorkflow(workflow)
	if err != nil {
		log.Fatalf("Failed to create workflow: %v", err)
	}
	log.Printf("Created workflow with ID: %s", workflowID)

	// Prepare input data for workflow execution
	inputData := map[string]interface{}{
		"text": sampleIntents[0], // Start with first intent
		"parameters": map[string]interface{}{
			"intent_types":         []string{"primary", "secondary"},
			"confidence_threshold": 0.7,
		},
	}

	// Execute the workflow
	results, err := executeWorkflow(workflowID, inputData)
	if err != nil {
		log.Fatalf("Failed to execute workflow: %v", err)
	}

	// Print results
	fmt.Printf("\nWorkflow Execution Results:\n")
	fmt.Printf("=========================\n")

	// Pretty print the results
	prettyResults, _ := json.MarshalIndent(results, "", "  ")
	fmt.Printf("%s\n", string(prettyResults))

	// Test with multiple intents
	fmt.Printf("\nTesting with multiple intents:\n")
	fmt.Printf("============================\n")

	for _, intent := range sampleIntents {
		fmt.Printf("\nProcessing intent: %s\n", intent)

		// Update input data with new intent
		inputData["text"] = intent

		// Execute workflow again
		results, err := executeWorkflow(workflowID, inputData)
		if err != nil {
			log.Printf("Failed to execute workflow for intent '%s': %v", intent, err)
			continue
		}

		// Print results for this intent
		fmt.Printf("Results:\n")
		prettyResults, _ := json.MarshalIndent(results, "", "  ")
		fmt.Printf("%s\n", string(prettyResults))
	}
}

func createIntentWorkflow() map[string]interface{} {
	// Create a workflow with a single intent generation node
	return map[string]interface{}{
		"id":   fmt.Sprintf("intent-workflow-%d", time.Now().Unix()),
		"name": "Intent Generation Workflow",
		"nodes": []map[string]interface{}{
			{
				"id":   "intent-node-1",
				"type": "default",
				"position": map[string]interface{}{
					"x": 250,
					"y": 100,
				},
				"data": map[string]interface{}{
					"label":      "Generate Intent",
					"nodeType":   "function",
					"functionId": "analysis-intent",
				},
				"style": map[string]interface{}{
					"background":   "rgba(16, 185, 129, 0.1)",
					"borderColor":  "#10B981",
					"borderWidth":  "2px",
					"padding":      "10px",
					"borderRadius": "8px",
					"color":        "#065F46",
					"fontWeight":   500,
				},
				"sourcePosition": "right",
				"targetPosition": "left",
			},
		},
		"edges": []map[string]interface{}{}, // No edges needed for single node
	}
}

func createWorkflow(workflow map[string]interface{}) (string, error) {
	// Convert workflow to JSON
	jsonData, err := json.Marshal(workflow)
	if err != nil {
		return "", fmt.Errorf("failed to marshal workflow: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "http://localhost:8080/api/workflows", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}

	return result.ID, nil
}

func executeWorkflow(workflowID string, inputData map[string]interface{}) (map[string]interface{}, error) {
	// Convert input data to JSON
	jsonData, err := json.Marshal(inputData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input data: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:8080/api/workflows/%s/execute", workflowID), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return result, nil
}
