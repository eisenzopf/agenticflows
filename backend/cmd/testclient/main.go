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
	analysisTypeFlag := flag.String("type", "intent", "Analysis type (trends, patterns, findings, attributes, intent, recommendations, plan)")
	textFlag := flag.String("text", "", "Text to analyze")
	fileFlag := flag.String("file", "", "File containing text to analyze")
	workflowFlag := flag.String("workflow", "", "Workflow ID (optional)")
	resultsFlag := flag.Bool("results", false, "Retrieve analysis results for workflow")
	flag.Parse()

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

	// Handle fetching results
	if *resultsFlag {
		if *workflowFlag == "" {
			fmt.Println("Workflow ID is required when fetching results")
			os.Exit(1)
		}
		fetchResults(*workflowFlag)
		return
	}

	// Prepare parameters based on analysis type
	parameters := buildParameters(*analysisTypeFlag, text)

	// Call standardized API
	callStandardAPI(*analysisTypeFlag, text, *workflowFlag, parameters)
}

// buildParameters creates appropriate parameters based on analysis type
func buildParameters(analysisType string, text string) map[string]interface{} {
	parameters := make(map[string]interface{})

	switch analysisType {
	case "trends":
		parameters["focus_areas"] = []string{"customer satisfaction", "service quality", "churn reasons"}
	case "patterns":
		parameters["pattern_types"] = []string{"complaints", "requests", "compliments"}
	case "findings":
		parameters["questions"] = []string{
			"What are the most common reasons for customer complaints?",
			"How effective are our customer service agents at resolving issues?",
		}
	case "attributes":
		parameters["attributes"] = []map[string]string{
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
		}
	case "recommendations":
		parameters["focus_area"] = "customer retention"
		parameters["criteria"] = map[string]float64{
			"impact":              0.6,
			"implementation_ease": 0.4,
		}
	case "plan":
		parameters["constraints"] = map[string]interface{}{
			"budget":    50000,
			"timeline":  "6 months",
			"resources": []string{"customer_support", "engineering", "marketing"},
		}
	}

	return parameters
}

// callStandardAPI calls the standardized /api/analysis endpoint
func callStandardAPI(analysisType string, text string, workflowID string, parameters map[string]interface{}) {
	// Prepare request for the standardized API
	reqBody := map[string]interface{}{
		"analysis_type": analysisType,
		"text":          text,
		"workflow_id":   workflowID,
		"parameters":    parameters,
	}

	// For findings analysis, we need a data field with attribute values
	if analysisType == "findings" {
		reqBody["data"] = map[string]interface{}{
			"attribute_values": map[string]interface{}{
				"sample_data": "This is a placeholder. Normally you would include actual attribute values here.",
			},
		}
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	// Create request
	url := "http://localhost:8080/api/analysis"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("Content-Type", "application/json")

	// Set workflow ID header if provided
	if workflowID != "" {
		req.Header.Set("X-Workflow-ID", workflowID)
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

// fetchResults fetches analysis results for a workflow
func fetchResults(workflowID string) {
	// Create request
	url := "http://localhost:8080/api/analysis/results?workflow_id=" + workflowID
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		os.Exit(1)
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
