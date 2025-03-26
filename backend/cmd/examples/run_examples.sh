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
WORKFLOW_ID="example-$(date +%Y%m%d%H%M%S)"
DEBUG=false
LIMIT=10
SAMPLE_SIZE=3
MIN_COUNT=5
TARGET_CLASS="fee dispute"
INTENTS="cancel,upgrade,billing"
THRESHOLD=0.7

# Function to display script usage
function show_usage {
    echo -e "${BLUE}Conversation Analysis Example Scripts Runner${NC}"
    echo ""
    echo "Usage: $0 [options] [script-name]"
    echo ""
    echo "Options:"
    echo "  -d, --database PATH     Path to SQLite database (required)"
    echo "  -o, --output DIR        Directory for output files (default: ./output)"
    echo "  -w, --workflow ID       Workflow ID (default: generated timestamp)"
    echo "  -l, --limit NUM         Limit number of items to process (default: 10)"
    echo "  -s, --sample NUM        Sample size for conversation analysis (default: 3)"
    echo "  -m, --min-count NUM     Minimum count threshold (default: 5)"
    echo "  -t, --target CLASS      Target class for analysis (default: 'fee dispute')"
    echo "  -i, --intents LIST      Comma-separated list of intents (default: cancel,upgrade,billing)"
    echo "  -c, --confidence NUM    Confidence threshold (default: 0.7)"
    echo "  -v, --verbose           Enable verbose/debug output"
    echo "  -h, --help              Show this help message"
    echo ""
    echo "Available scripts:"
    echo "  all                     Run all example scripts"
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
        -d|--database)
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
        -s|--sample)
            SAMPLE_SIZE="$2"
            shift 2
            ;;
        -m|--min-count)
            MIN_COUNT="$2"
            shift 2
            ;;
        -t|--target)
            TARGET_CLASS="$2"
            shift 2
            ;;
        -i|--intents)
            INTENTS="$2"
            shift 2
            ;;
        -c|--confidence)
            THRESHOLD="$2"
            shift 2
            ;;
        -v|--verbose)
            DEBUG=true
            shift
            ;;
        -h|--help)
            show_usage
            exit 0
            ;;
        -*)
            echo -e "${RED}Error: Unknown option $1${NC}"
            show_usage
            exit 1
            ;;
        *)
            SCRIPT_NAME="$1"
            shift
            ;;
    esac
done

# Validate required parameters
if [ -z "$DB_PATH" ]; then
    echo -e "${RED}Error: Database path is required (-d, --database)${NC}"
    show_usage
    exit 1
fi

if [ -z "$SCRIPT_NAME" ]; then
    echo -e "${RED}Error: Script name is required${NC}"
    show_usage
    exit 1
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Set debug flag
DEBUG_FLAG=""
if [ "$DEBUG" = true ]; then
    DEBUG_FLAG="--debug"
fi

# Run generate_intents script
function run_generate_intents {
    echo -e "${GREEN}Running generate_intents script...${NC}"
    OUTPUT_FILE="$OUTPUT_DIR/intents_$(date +%Y%m%d%H%M%S).json"
    
    go run utils.go generate_intents.go \
        --db "$DB_PATH" \
        --output "$OUTPUT_FILE" \
        --limit "$LIMIT" \
        --workflow "$WORKFLOW_ID" \
        $DEBUG_FLAG
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ generate_intents completed successfully${NC}"
        echo -e "  Output: $OUTPUT_FILE"
    else
        echo -e "${RED}✗ generate_intents failed${NC}"
    fi
    echo ""
}

# Run generate_attributes script
function run_generate_attributes {
    echo -e "${GREEN}Running generate_attributes script...${NC}"
    OUTPUT_FILE="$OUTPUT_DIR/attributes_$(date +%Y%m%d%H%M%S).json"
    
    go run utils.go generate_attributes.go \
        --db "$DB_PATH" \
        --output "$OUTPUT_FILE" \
        --min-count "$MIN_COUNT" \
        --sample-size "$SAMPLE_SIZE" \
        --target-class "$TARGET_CLASS" \
        --threshold "$THRESHOLD" \
        --workflow "$WORKFLOW_ID" \
        $DEBUG_FLAG
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ generate_attributes completed successfully${NC}"
        echo -e "  Output: $OUTPUT_FILE"
    else
        echo -e "${RED}✗ generate_attributes failed${NC}"
    fi
    echo ""
}

# Run group_intents script
function run_group_intents {
    echo -e "${GREEN}Running group_intents script...${NC}"
    OUTPUT_FILE="$OUTPUT_DIR/intent_groups_$(date +%Y%m%d%H%M%S).json"
    
    go run utils.go group_intents.go \
        --db "$DB_PATH" \
        --output "$OUTPUT_FILE" \
        --min-count "$MIN_COUNT" \
        --max-groups 20 \
        --workflow "$WORKFLOW_ID" \
        $DEBUG_FLAG
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ group_intents completed successfully${NC}"
        echo -e "  Output: $OUTPUT_FILE"
    else
        echo -e "${RED}✗ group_intents failed${NC}"
    fi
    echo ""
}

# Run identify_attributes script
function run_identify_attributes {
    echo -e "${GREEN}Running identify_attributes script...${NC}"
    
    # Run the script
    go run identify_attributes.go utils.go \
        --db "$DB_PATH" \
        --output "$OUTPUT_DIR/attribute_definitions.json" \
        --intent "$TARGET_CLASS" \
        --workflow "$WORKFLOW_ID" \
        --limit "$LIMIT" $DEBUG_FLAG
    
    # Check if the script ran successfully
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ identify_attributes completed successfully${NC}"
        return 0
    else
        echo -e "${RED}✗ identify_attributes failed${NC}"
        return 1
    fi
}

# Run match_intents script
function run_match_intents {
    echo -e "${GREEN}Running match_intents script...${NC}"
    OUTPUT_FILE="$OUTPUT_DIR/intent_matching_$(date +%Y%m%d%H%M%S).json"
    
    go run utils.go match_intents.go \
        --db "$DB_PATH" \
        --output "$OUTPUT_FILE" \
        --intents "$INTENTS" \
        --sample-size "$SAMPLE_SIZE" \
        --threshold "$THRESHOLD" \
        --workflow "$WORKFLOW_ID" \
        $DEBUG_FLAG
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ match_intents completed successfully${NC}"
        echo -e "  Output: $OUTPUT_FILE"
    else
        echo -e "${RED}✗ match_intents failed${NC}"
    fi
    echo ""
}

# Run analyze_fee_disputes script
function run_analyze_fee_disputes {
    echo -e "${GREEN}Running analyze_fee_disputes script...${NC}"
    OUTPUT_FILE="$OUTPUT_DIR/fee_dispute_analysis_$(date +%Y%m%d%H%M%S).json"
    
    go run utils.go analyze_fee_disputes.go \
        --db "$DB_PATH" \
        --output "$OUTPUT_FILE" \
        --min-count "$MIN_COUNT" \
        --workflow "$WORKFLOW_ID" \
        $DEBUG_FLAG
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ analyze_fee_disputes completed successfully${NC}"
        echo -e "  Output: $OUTPUT_FILE"
    else
        echo -e "${RED}✗ analyze_fee_disputes failed${NC}"
    fi
    echo ""
}

# Main script execution
echo -e "${BLUE}Starting Conversation Analysis Examples${NC}"
echo -e "Database: ${YELLOW}$DB_PATH${NC}"
echo -e "Workflow ID: ${YELLOW}$WORKFLOW_ID${NC}"
echo -e "Output Directory: ${YELLOW}$OUTPUT_DIR${NC}"
echo ""

case "$SCRIPT_NAME" in
    all)
        echo -e "${BLUE}Running all example scripts...${NC}"
        echo ""
        run_generate_intents
        run_generate_attributes
        run_group_intents
        run_identify_attributes
        run_match_intents
        run_analyze_fee_disputes
        ;;
    generate_intents)
        run_generate_intents
        ;;
    generate_attributes)
        run_generate_attributes
        ;;
    group_intents)
        run_group_intents
        ;;
    identify_attributes)
        run_identify_attributes
        ;;
    match_intents)
        run_match_intents
        ;;
    analyze_fee_disputes)
        run_analyze_fee_disputes
        ;;
    *)
        echo -e "${RED}Error: Unknown script name '$SCRIPT_NAME'${NC}"
        show_usage
        exit 1
        ;;
esac

echo -e "${GREEN}All tasks completed.${NC}" 