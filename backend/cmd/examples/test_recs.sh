#!/bin/bash

# Simple test for the recommendations API endpoint
API_HOST="http://localhost:8080"
WORKFLOW_ID="test-recs-$(date +%s)"

echo "Testing recommendations endpoint..."
echo "API Host: $API_HOST"
echo "Workflow ID: $WORKFLOW_ID"

# Create a sample payload
PAYLOAD="{
  \"workflow_id\": \"$WORKFLOW_ID\",
  \"analysis_type\": \"recommendations\",
  \"parameters\": {
    \"focus_area\": \"customer_retention\",
    \"criteria\": {
      \"impact\": 0.6,
      \"implementation_ease\": 0.4
    }
  },
  \"data\": {
    \"trends\": [
      {
        \"focus_area\": \"customer_satisfaction\",
        \"trend\": \"Declining satisfaction scores in Q3\"
      }
    ],
    \"patterns\": [
      {
        \"type\": \"user_behavior\",
        \"description\": \"Customers frequently check order status multiple times\"
      }
    ]
  }
}"

echo "Request payload:"
echo "$PAYLOAD" | jq '.'

# Make the API request
echo "Sending request to $API_HOST/api/analysis..."
RESPONSE=$(curl -s -X POST \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d "$PAYLOAD" \
  "${API_HOST}/api/analysis")

# Check if the response is valid JSON
if echo "$RESPONSE" | jq '.' &>/dev/null; then
  echo "Response (JSON):"
  echo "$RESPONSE" | jq '.'
  
  # Check if there's an error
  if [[ $(echo "$RESPONSE" | jq -r 'has("error")') == "true" && $(echo "$RESPONSE" | jq -r '.error') != "null" ]]; then
    echo "ERROR: API returned an error"
    ERROR_CODE=$(echo "$RESPONSE" | jq -r '.error.code // "unknown"')
    ERROR_MSG=$(echo "$RESPONSE" | jq -r '.error.message // "Unknown error"')
    echo "Error code: $ERROR_CODE"
    echo "Error message: $ERROR_MSG"
    exit 1
  else
    echo "SUCCESS: Recommendation API call succeeded!"
    exit 0
  fi
else
  echo "ERROR: Invalid JSON response"
  echo "Raw response: $RESPONSE"
  exit 1
fi 