# Conversation Analysis Example Scripts

_Last Updated: April 2024_
_Version: 1.0.2_

This directory contains a collection of Go scripts that demonstrate how to use the Discourse AI Analysis API for conversation analysis. These scripts replicate the functionality of the original Python scripts but utilize the new Go API endpoints.

## Scripts Overview

| Script | Purpose | API Endpoints Used |
|--------|---------|-------------------|
| `generate_intents.go` | Extracts the primary intent from conversation texts | `/api/analysis` with `analysis_type: "intent"` |
| `generate_attributes.go` | Generates structured attribute values from conversations | `/api/analysis` with `analysis_type: "attributes"` |
| `group_intents.go` | Groups similar intents together to identify patterns | `/api/analysis` with `analysis_type: "patterns"` |
| `identify_attributes.go` | Identifies potential attributes definitions from conversations | `/api/analysis` with `analysis_type: "attributes"` and appropriate parameters |
| `match_intents.go` | Matches and evaluates intent classifications | `/api/analysis` with `analysis_type: "intent"` |
| `analyze_fee_disputes.go` | Analyzes fee dispute conversations with detailed analytics | `/api/analysis` with various `analysis_type` values: `"attributes"`, `"trends"`, `"findings"` |
| `create_action_plan.go` | Generates actionable recommendations based on analysis | `/api/analysis` with `analysis_type: "recommendations"` and `analysis_type: "plan"` |

## Utility Files

| File | Purpose |
|------|---------|
| `utils.go` | Common utility functions and API client for all scripts |
| `run_examples.sh` | Shell script to run individual or all examples easily |
| `SCRIPT_USAGE.md` | Detailed instructions on using the shell script |
| `PIPELINE_OVERVIEW.md` | Overview of how the scripts work together in a pipeline |
| `MOCK_DATA_USAGE.md` | Instructions for using scripts with mock data instead of a database |

## Prerequisites

- Go 1.18 or later
- SQLite database with conversation data (or use mock data option)
- Access to a running Discourse AI Analysis API server (typically at http://localhost:8080)
- API key for the LLM service used by the backend
- jq (optional, for pretty-printing JSON output)
- curl (for API testing)

## Getting Started

1. Make sure the API server is running at http://localhost:8080
2. Ensure you have a valid API key configured in the API server for the LLM service
3. Ensure you have a SQLite database with conversation data (or use mock data - see below)
4. Make the shell script executable:
   ```bash
   chmod +x run_examples.sh
   ```
5. Run an example:
   ```bash
   ./run_examples.sh -d /path/to/database.db generate_intents
   ```

See `SCRIPT_USAGE.md` for detailed instructions and options.

## Using Mock Data

You can now run many of the example scripts without a database by using mock data:

```bash
./run_examples.sh -m all
```

This uses predefined sample data instead of querying a database. Currently, the following scripts support mock data:
- `generate_intents.go`
- `create_action_plan.go`
- `generate_attributes.go`
- `identify_attributes.go`

See `MOCK_DATA_USAGE.md` for more details on using and extending mock data support.

## Script Functionality

### generate_intents.go
Extracts the primary intent from conversation texts by sending each conversation to the API for intent classification. Uses the `/api/analysis` endpoint with `analysis_type: "intent"`.

### generate_attributes.go
Generates structured attribute values from conversations by extracting key information into a structured format. Uses the `/api/analysis` endpoint with `analysis_type: "attributes"`.

### group_intents.go
Groups similar intents together to identify patterns and common themes across conversations. Uses the `/api/analysis` endpoint with `analysis_type: "patterns"`.

### identify_attributes.go
Analyzes conversations to identify potential attribute definitions that could be extracted in future analysis. Uses the `/api/analysis` endpoint with `analysis_type: "attributes"` and parameters to indicate attribute definition generation.

### match_intents.go
Evaluates intent classification against known intents, calculating precision, recall, and F1 scores. Uses the `/api/analysis` endpoint with `analysis_type: "intent"`.

### analyze_fee_disputes.go
Performs detailed analysis on fee dispute conversations, extracting specific patterns and insights. Uses the `/api/analysis` endpoint with multiple analysis types including `"attributes"`, `"trends"`, and `"findings"`.

### create_action_plan.go
Generates actionable recommendations based on intent groups and attribute data, creating a prioritized action plan. Uses the `/api/analysis` endpoint with `analysis_type: "recommendations"` and `analysis_type: "plan"`.

## API Integration

All scripts use the common `ApiClient` defined in `utils.go` to interact with the Discourse AI Analysis API. The client handles:

- Making HTTP requests to the appropriate endpoints
- Formatting request payloads
- Processing and validating responses
- Error handling and retries
- Debug output for troubleshooting

The client connects to the API server at `http://localhost:8080` by default.

### API Endpoints and Request Format

The primary API endpoint used by all scripts is:

```
POST /api/analysis
```

With a standardized request format:

```json
{
  "workflow_id": "optional-workflow-id",
  "analysis_type": "intent|attributes|patterns|trends|findings|recommendations|plan",
  "text": "conversation text (when applicable)",
  "parameters": {
    // Analysis-specific parameters
  },
  "data": {
    // Input data for analysis (when applicable)
  }
}
```

Chain analysis is available through:

```
POST /api/analysis/chain
```

Function metadata can be retrieved from:

```
GET /api/analysis/metadata
```

### API Versioning

The current scripts work with API v1.0. When a new API version is released, the scripts will be updated to maintain compatibility.

## Database Integration

The scripts read conversation data from a SQLite database. The database schema is expected to have tables containing conversations with at least the following fields:
- Conversation ID
- Conversation text
- Metadata (where applicable)

Some scripts like `generate_intents.go` and `generate_attributes.go` can also save their results back to the database if a workflow ID is provided.

## Output Format

All scripts produce JSON output files containing the results of their analysis, saved to the specified output directory.

## Troubleshooting

Common issues and their solutions:

### API Connection Issues
- Ensure the API server is running at the expected URL
- Check network connectivity and firewall settings
- Verify API authorization if enabled

### Database Errors
- Confirm the database path is correct
- Ensure the database has the expected schema
- Check database permissions

### Script Execution Problems
- Ensure Go version 1.18+ is installed
- Make sure the run_examples.sh script is executable
- Check for missing dependencies

### Mock Data Issues
- Verify the script supports mock data (see list above)
- Ensure -m flag is used correctly
- Check the mock data implementation in the script

For more detailed troubleshooting, run scripts with the `-v` (verbose) flag. 