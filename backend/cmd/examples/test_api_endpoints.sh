#!/bin/bash

# Color definitions
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default settings
API_HOST="http://localhost:8080"
WORKFLOW_ID="test-api-$(date +%s)"
DEBUG=false
TEST_COUNT=0
PASS_COUNT=0
FAIL_COUNT=0

# Test conversation text
CONVERSATION="Customer: Hi, I've been charged twice for my monthly subscription. Can you help me fix this?\n\nAgent: I'm sorry to hear that. Let me check your account.\n\nCustomer: Thanks, I was billed on both the 1st and the 15th, but I should only be billed once per month.\n\nAgent: You're right. I can see the duplicate charge. I'll process a refund right away.\n\nCustomer: Great, how long will the refund take?\n\nAgent: The refund should appear in your account within 3-5 business days. Is there anything else I can help with?\n\nCustomer: No, that's all. Thank you!"

# Print header
print_header() {
    echo -e "\n${BLUE}===========================================================${NC}"
    echo -e "${BLUE}   Testing Discourse AI Analysis API Endpoints${NC}"
    echo -e "${BLUE}===========================================================${NC}"
    echo -e "${YELLOW}API Host:${NC} $API_HOST"
    echo -e "${YELLOW}Workflow ID:${NC} $WORKFLOW_ID"
    echo -e "${BLUE}===========================================================${NC}\n"
}

# Print usage
usage() {
    echo -e "Usage: $0 [options]"
    echo -e "Options:"
    echo -e "  -h, --host HOST    API host URL (default: $API_HOST)"
    echo -e "  -w, --workflow ID  Workflow ID (default: auto-generated)"
    echo -e "  -v, --verbose      Enable verbose output"
    echo -e "  --help             Display this help message and exit"
    exit 1
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -h|--host)
            API_HOST="$2"
            shift 2
            ;;
        -w|--workflow)
            WORKFLOW_ID="$2"
            shift 2
            ;;
        -v|--verbose)
            DEBUG=true
            shift
            ;;
        --help)
            usage
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            ;;
    esac
done

# Function to run a test
run_test() {
    local endpoint="$1"
    local payload="$2"
    local description="$3"
    
    TEST_COUNT=$((TEST_COUNT + 1))
    
    echo -e "\n${YELLOW}Test #$TEST_COUNT: $description${NC}"
    echo -e "${BLUE}Endpoint:${NC} $endpoint"
    
    if [ "$DEBUG" = true ]; then
        echo -e "${BLUE}Payload:${NC}"
        echo "$payload" | jq '.'
    fi
    
    # Make the API request
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "${API_HOST}${endpoint}")
    
    # Check if the response is valid JSON
    if echo "$response" | jq '.' &>/dev/null; then
        if [ "$DEBUG" = true ]; then
            echo -e "${BLUE}Response:${NC}"
            echo "$response" | jq '.'
        fi
        
        # Check if the response has an error field
        if [[ $(echo "$response" | jq -r 'has("error")') == "true" && $(echo "$response" | jq -r '.error') != "null" ]]; then
            echo -e "${RED}✘ FAILED:${NC} API returned an error"
            echo -e "${RED}Error:${NC} $(echo "$response" | jq -r '.error')"
            FAIL_COUNT=$((FAIL_COUNT + 1))
        else
            echo -e "${GREEN}✓ PASSED${NC}"
            PASS_COUNT=$((PASS_COUNT + 1))
        fi
    else
        echo -e "${RED}✘ FAILED:${NC} Invalid response"
        echo -e "${RED}Response:${NC} $response"
        FAIL_COUNT=$((FAIL_COUNT + 1))
    fi
}

# Run tests
run_tests() {
    # Test 1: Generate intent
    run_test "/api/analysis/intent" \
        "{\"text\":\"$CONVERSATION\",\"workflow_id\":\"$WORKFLOW_ID\"}" \
        "Generate intent from conversation"
    
    # Test 2: Generate attributes
    run_test "/api/analysis/attributes" \
        "{\"text\":\"$CONVERSATION\",\"workflow_id\":\"$WORKFLOW_ID\"}" \
        "Generate attributes from conversation"
    
    # Test 3: Generate required attributes
    run_test "/api/analysis/required_attributes" \
        "{\"text\":\"$CONVERSATION\",\"existing_attributes\":{\"intent\":\"billing issue\"},\"workflow_id\":\"$WORKFLOW_ID\"}" \
        "Generate required attributes based on existing attributes"
    
    # Test 4: Analyze trends
    run_test "/api/analysis/trends" \
        "{\"conversations\":[{\"id\":\"1\",\"text\":\"$CONVERSATION\",\"intent\":\"billing issue\"}],\"workflow_id\":\"$WORKFLOW_ID\"}" \
        "Analyze trends in conversations"
    
    # Test 5: Identify patterns
    run_test "/api/analysis/patterns" \
        "{\"data\":[{\"id\":\"1\",\"intent\":\"billing issue\",\"resolution\":\"refund\"}],\"pattern_types\":[\"intent_frequency\",\"resolution_time\"],\"workflow_id\":\"$WORKFLOW_ID\"}" \
        "Identify patterns in conversation data"
    
    # Test 6: Analyze findings
    run_test "/api/analysis/findings" \
        "{\"data\":{\"patterns\":[{\"name\":\"duplicate_charge\",\"frequency\":42}],\"attributes\":{\"avg_resolution_time\":\"3.5 days\"}},\"workflow_id\":\"$WORKFLOW_ID\"}" \
        "Analyze findings from patterns and attributes"
    
    # Print summary
    echo -e "\n${BLUE}===========================================================${NC}"
    echo -e "${YELLOW}Test Summary:${NC}"
    echo -e "${BLUE}Total Tests:${NC} $TEST_COUNT"
    echo -e "${GREEN}Passed:${NC} $PASS_COUNT"
    echo -e "${RED}Failed:${NC} $FAIL_COUNT"
    echo -e "${BLUE}===========================================================${NC}\n"
    
    # Return exit code based on test results
    if [ $FAIL_COUNT -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}Some tests failed!${NC}"
        return 1
    fi
}

# Main execution
print_header
run_tests
exit $? 