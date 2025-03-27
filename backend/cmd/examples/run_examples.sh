#!/bin/bash

# Define colors for better readability
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Default parameters
DB_PATH=""
OUTPUT_DIR="./output"
WORKFLOW_ID="example-workflow-$(date +%Y%m%d-%H%M%S)"
DEBUG=false
LIMIT=10
RUN_ALL=false

# Function to display script usage
function show_usage {
    echo -e "${BLUE}Conversation Analysis Example Scripts Runner${NC}"
    echo ""
    echo "Usage: $0 [options] [script-name]"
    echo ""
    echo "Options:"
    echo "  -d, --db PATH         Path to SQLite database (required)"
    echo "  -o, --output DIR      Directory for output files (default: ./output)"
    echo "  -w, --workflow ID     Workflow ID (default: generated timestamp)"
    echo "  -l, --limit NUM       Limit number of items to process (default: 10)"
    echo "  -v, --verbose         Enable verbose/debug output"
    echo "  all                   Run all example scripts"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Available scripts:"
    echo "  generate_intents        Generate conversation intents"
    echo "  generate_attributes     Generate attribute values for conversations"
    echo "  group_intents           Group similar intents together"
    echo "  identify_attributes     Identify attribute definitions for conversations"
    echo "  match_intents           Match and evaluate intent classifications"
    echo "  analyze_fee_disputes    Analyze fee dispute conversations"
    echo ""
    echo "Examples:"
    echo "  $0 -d ./data.db all"
    echo "  $0 -d ./data.db -w my-workflow generate_intents"
    echo "  $0 -d ./data.db -t 'subscription cancel' -v identify_attributes"
    echo ""
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -d|--db)
            DB_PATH="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -w|--workflow)
            WORKFLOW_ID="$2"
            shift 2
            ;;
        -l|--limit)
            LIMIT="$2"
            shift 2
            ;;
        --debug)
            DEBUG=true
            shift
            ;;
        all)
            RUN_ALL=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        *)
            echo "Unknown argument: $1"
            exit 1
            ;;
    esac
done

# Validate required parameters
if [ -z "$DB_PATH" ]; then
    echo "Error: Database path is required. Use -d or --db to specify."
    exit 1
fi

if [ ! -f "$DB_PATH" ]; then
    echo "Error: Database file not found: $DB_PATH"
    exit 1
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Print configuration
echo -e "Starting Conversation Analysis Examples"
echo -e "Database: $DB_PATH"
echo -e "Workflow ID: $WORKFLOW_ID"
echo -e "Output Directory: $OUTPUT_DIR"
echo -e ""

# Set up debug flag
DEBUG_FLAG=""
if [ "$DEBUG" = true ]; then
    DEBUG_FLAG="--debug"
fi

# Function to run a script and check its status
run_script() {
    local script_dir=$1
    local script_name=$2
    echo -e "\n=== Running $script_name ==="
    
    if [ -d "$script_dir" ]; then
        cd "$script_dir"
        
        # Set script-specific flags
        local extra_flags=""
        case "$script_dir" in
            "group_intents")
                extra_flags="--min-count 5 --max-groups 10"
                ;;
            *)
                extra_flags="--limit $LIMIT"
                ;;
        esac
        
        go run main.go --db "$DB_PATH" --workflow "$WORKFLOW_ID" $DEBUG_FLAG $extra_flags
        if [ $? -eq 0 ]; then
            echo -e "✓ $script_name completed successfully"
        else
            echo -e "✗ $script_name failed"
        fi
        cd ..
    else
        echo -e "✗ $script_name directory not found: $script_dir"
    fi
}

# Run all scripts
echo -e "Running example scripts..."

# Generate Intents
run_script "generate_intents" "Generate Intents"

# Generate Attributes
run_script "generate_attributes" "Generate Attributes"

# Group Intents
run_script "group_intents" "Group Intents"

# Identify Attributes
run_script "identify_attributes" "Identify Attributes"

# Match Intents
run_script "match_intents" "Match Intents"

# Analyze Fee Disputes
run_script "analyze_fee_disputes" "Analyze Fee Disputes"

echo -e "\nAll tasks completed." 