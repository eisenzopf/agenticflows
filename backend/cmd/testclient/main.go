package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	// Command line flags
	endpointFlag := flag.String("endpoint", "intent", "API endpoint to test (trends, patterns, findings, attributes, intent)")
	textFlag := flag.String("text", "", "Text to analyze")
	fileFlag := flag.String("file", "", "File containing text to analyze")
	workflowFlag := flag.String("workflow", "", "Workflow ID (optional)")
	flag.Parse()

	// Validate endpoint
	validEndpoints := map[string]string{
		"trends":     "/api/analysis/trends",
		"patterns":   "/api/analysis/patterns",
		"findings":   "/api/analysis/findings",
		"attributes": "/api/analysis/attributes",
		"intent":     "/api/analysis/intent",
		"results":    "/api/analysis/results",
	}

	endpoint, ok := validEndpoints[*endpointFlag]
	if !ok {
		fmt.Printf("Invalid endpoint: %s\n", *endpointFlag)
		fmt.Println("Valid endpoints: trends, patterns, findings, attributes, intent, results")
		os.Exit(1)
	}

	// Get text from file or command line
	text := *textFlag
	if *fileFlag != "" {
		data, err := os.ReadFile(*fileFlag)
		if err != nil {
			fmt.Printf("Error reading file: %v\n", err)
			os.Exit(1)
		}
		text = string(data)
	}

	// Prepare request
	var reqBody interface{}

	switch *endpointFlag {
	case "trends":
		reqBody = map[string]interface{}{
			"focus_areas": []string{"customer satisfaction", "service quality", "churn reasons"},
			"text":        text,
		}
	case "patterns":
		reqBody = map[string]interface{}{
			"pattern_types": []string{"complaints", "requests", "compliments"},
			"text":          text,
		}
	case "findings":
		reqBody = map[string]interface{}{
			"questions": []string{
				"What are the most common reasons for customer complaints?",
				"How effective are our customer service agents at resolving issues?",
			},
			"attribute_values": map[string]interface{}{
				"sample_data": "This is a placeholder. Normally you would include actual attribute values here.",
			},
		}
	case "attributes":
		reqBody = map[string]interface{}{
			"text": text,
			"attributes": []map[string]string{
				{
					"field_name":  "customer_sentiment",
					"title":       "Customer Sentiment",
					"description": "The overall sentiment of the customer in the conversation.",
				},
				{
					"field_name":  "issue_type",
					"title":       "Issue Type",
					"description": "The type of issue the customer is experiencing.",
				},
			},
			"workflow_id": *workflowFlag,
		}
	case "intent":
		reqBody = map[string]interface{}{
			"text":        text,
			"workflow_id": *workflowFlag,
		}
	case "results":
		// No request body needed, just use workflow ID
	}

	// Make API request
	url := "http://localhost:8080" + endpoint
	if *endpointFlag == "results" && *workflowFlag != "" {
		url += "?workflow_id=" + *workflowFlag
	}

	var req *http.Request
	var err error

	if *endpointFlag != "results" {
		// POST request with JSON body
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			fmt.Printf("Error marshaling request: %v\n", err)
			os.Exit(1)
		}

		req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			os.Exit(1)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		// GET request for results
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			os.Exit(1)
		}
	}

	// Set workflow ID header if provided
	if *workflowFlag != "" {
		req.Header.Set("X-Workflow-ID", *workflowFlag)
	}

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Pretty print response
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
		fmt.Printf("Error formatting JSON: %v\n", err)
		fmt.Println(string(body))
	} else {
		fmt.Println(prettyJSON.String())
	}
} 