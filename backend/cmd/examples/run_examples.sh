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
USE_MOCK=false

# Function to display script usage
function show_usage {
    echo -e "${BLUE}Conversation Analysis Example Scripts Runner${NC}"
    echo ""
    echo "Usage: $0 [options] [script-name]"
    echo ""
    echo "Options:"
    echo "  -d, --db PATH         Path to SQLite database (optional if using mock data)"
    echo "  -o, --output DIR      Directory for output files (default: ./output)"
    echo "  -w, --workflow ID     Workflow ID (default: generated timestamp)"
    echo "  -l, --limit NUM       Limit number of items to process (default: 10)"
    echo "  -v, --verbose         Enable verbose/debug output"
    echo "  -m, --mock            Use mock data instead of database"
    echo "  all                   Run all example scripts"
    echo "  -h, --help            Show this help message"
    echo ""
    echo "Available scripts:"
    echo "  generate_intents        Generate conversation intents"
    echo "  create_action_plan      Create action plan (doesn't need database)"
    echo ""
    echo "When using -m (mock) flag, only scripts that support mock data will run."
    echo ""
    echo "Examples:"
    echo "  $0 -d ./data.db all"
    echo "  $0 -d ./data.db -w my-workflow generate_intents"
    echo "  $0 -m all                         # Run with mock data"
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
        -m|--mock)
            USE_MOCK=true
            shift
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

# Validate database path if not using mock data
if [ "$USE_MOCK" = false ] && [ -z "$DB_PATH" ]; then
    echo "Error: Database path is required when not using mock data. Use -d or --db to specify, or use -m for mock data."
    exit 1
fi

# Check if database exists when specified
if [ -n "$DB_PATH" ] && [ ! -f "$DB_PATH" ]; then
    echo "Warning: Database file not found: $DB_PATH"
    echo "Using mock data instead."
    USE_MOCK=true
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Print configuration
echo -e "Starting Conversation Analysis Examples"
echo -e "Database: ${DB_PATH:-"Using mock data"}"
echo -e "Workflow ID: $WORKFLOW_ID"
echo -e "Output Directory: $OUTPUT_DIR"
echo -e ""

# Set up debug flag
DEBUG_FLAG=""
if [ "$DEBUG" = true ]; then
    DEBUG_FLAG="--debug"
fi

# Set up mock flag
MOCK_FLAG=""
if [ "$USE_MOCK" = true ]; then
    MOCK_FLAG="--mock"
    echo -e "${YELLOW}Using mock data for all examples${NC}"
    echo -e "${YELLOW}Note: Only some scripts support mock data at this time${NC}"
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
        local db_flag=""
        
        # Only add DB flag if not using mock data
        if [ "$USE_MOCK" = false ]; then
            db_flag="--db \"$DB_PATH\""
        fi
        
        case "$script_dir" in
            "group_intents")
                extra_flags="--min-count 5 --max-groups 10"
                ;;
            "analyze_fee_disputes")
                extra_flags="--max $LIMIT --batch 10"
                ;;
            "generate_recommendations")
                extra_flags="--limit $LIMIT --focus customer_retention"
                ;;
            "create_action_plan")
                extra_flags="--budget 50000 --timespan \"6 months\""
                db_flag="" # No DB needed for action plan
                ;;
            *)
                extra_flags="--limit $LIMIT"
                ;;
        esac
        
        # Run command with appropriate flags
        run_cmd="go run main.go $db_flag --workflow \"$WORKFLOW_ID\" $DEBUG_FLAG $MOCK_FLAG $extra_flags"
        eval $run_cmd
        
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

# Run example scripts
echo -e "Running example scripts..."

# When using mock data, only run scripts that support it
if [ "$USE_MOCK" = true ]; then
    # Generate Intents (supports mock data)
    run_script "generate_intents" "Generate Intents"
    
    # Create Action Plan (already uses sample data)
    run_script "create_action_plan" "Create Action Plan"
else
    # Run all scripts when using a real database
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

    # Generate Recommendations
    run_script "generate_recommendations" "Generate Recommendations"

    # Create Action Plan
    run_script "create_action_plan" "Create Action Plan"

    # Analyze Fee Disputes
    run_script "analyze_fee_disputes" "Analyze Fee Disputes"
fi

echo -e "\nAll tasks completed." 