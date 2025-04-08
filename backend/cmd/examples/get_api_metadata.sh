#!/bin/bash

# get_api_metadata.sh
# Script to retrieve and format the API function metadata from the analysis system
# Usage: ./get_api_metadata.sh [output_format]
# Where output_format is one of: json, pretty, table (default is pretty)

set -e

# Configuration
API_HOST="localhost:8080"
API_ENDPOINT="/api/analysis/metadata"
OUTPUT_DIR="./output"
FORMAT=${1:-"pretty"}  # Default to pretty if no format specified

# Create output directory if it doesn't exist
mkdir -p "${OUTPUT_DIR}"

echo "Fetching API metadata from http://${API_HOST}${API_ENDPOINT}..."

# Fetch the metadata
RESPONSE=$(curl -s "http://${API_HOST}${API_ENDPOINT}")

# Save raw JSON response
echo "${RESPONSE}" > "${OUTPUT_DIR}/api_metadata_raw.json"
echo "Raw JSON saved to ${OUTPUT_DIR}/api_metadata_raw.json"

# Format based on user preference
case "${FORMAT}" in
  "json")
    # Just save the raw JSON
    echo "Output format: JSON"
    cat "${OUTPUT_DIR}/api_metadata_raw.json"
    ;;
    
  "pretty")
    # Use jq to create a pretty-printed version
    echo "Output format: Pretty JSON"
    if ! command -v jq &> /dev/null; then
      echo "jq not found. Please install jq or use a different output format."
      exit 1
    fi
    
    # Save pretty JSON
    jq . "${OUTPUT_DIR}/api_metadata_raw.json" > "${OUTPUT_DIR}/api_metadata_pretty.json"
    echo "Pretty JSON saved to ${OUTPUT_DIR}/api_metadata_pretty.json"
    
    # Also display a summary
    echo ""
    echo "Available API Functions:"
    jq -r 'keys[] as $k | "  - \($k): \(.[$k].label) - \(.[$k].description)"' "${OUTPUT_DIR}/api_metadata_raw.json"
    ;;
    
  "table")
    # Create a markdown table of functions and their properties
    echo "Output format: Markdown Table"
    if ! command -v jq &> /dev/null; then
      echo "jq not found. Please install jq or use a different output format."
      exit 1
    fi
    
    # Create markdown table file
    OUTPUT_FILE="${OUTPUT_DIR}/api_metadata_table.md"
    
    # Write table header
    echo "# API Function Metadata" > "${OUTPUT_FILE}"
    echo "" >> "${OUTPUT_FILE}"
    echo "## Available Functions" >> "${OUTPUT_FILE}"
    echo "" >> "${OUTPUT_FILE}"
    echo "| ID | Label | Description |" >> "${OUTPUT_FILE}"
    echo "|---|---|---|" >> "${OUTPUT_FILE}"
    
    # Add function rows
    jq -r 'keys[] as $k | "| \(.[$k].id) | \(.[$k].label) | \(.[$k].description) |"' "${OUTPUT_DIR}/api_metadata_raw.json" >> "${OUTPUT_FILE}"
    
    # Add input/output sections for each function
    jq -r 'keys[] as $k | 
      "\n## \(.[$k].label) (\($k))\n\n### Inputs\n\n| Name | Path | Description | Required | Type |\n|---|---|---|---|---|\n" + 
      (.[$k].inputs | map("| \(.name) | \(.path) | \(.description) | \(.required) | \(.type) |") | join("\n")) +
      "\n\n### Outputs\n\n| Name | Path | Description | Type |\n|---|---|---|---|\n" + 
      (.[$k].outputs | map("| \(.name) | \(.path) | \(.description) | \(.type) |") | join("\n"))' "${OUTPUT_DIR}/api_metadata_raw.json" >> "${OUTPUT_FILE}"
    
    echo "Markdown table saved to ${OUTPUT_FILE}"
    
    # Show a preview of the table
    echo ""
    echo "Available Functions Preview:"
    jq -r 'keys[] as $k | "  - \(.[$k].id): \(.[$k].label)"' "${OUTPUT_DIR}/api_metadata_raw.json"
    ;;
    
  *)
    echo "Unknown format: ${FORMAT}"
    echo "Valid formats are: json, pretty, table"
    exit 1
    ;;
esac

echo ""
echo "Done!" 