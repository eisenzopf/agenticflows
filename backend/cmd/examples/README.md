# Conversation Analysis Example Scripts

This directory contains a collection of Go scripts that demonstrate how to use the Discourse AI Analysis API for conversation analysis. These scripts replicate the functionality of the original Python scripts but utilize the new Go API endpoints.

## Scripts Overview

| Script | Purpose | API Endpoints Used |
|--------|---------|-------------------|
| `generate_intents.go` | Extracts the primary intent from conversation texts | `/api/analysis/intent` |
| `generate_attributes.go` | Generates structured attribute values from conversations | `/api/analysis/attributes` |
| `group_intents.go` | Groups similar intents together to identify patterns | `/api/analysis/patterns` |
| `identify_attributes.go` | Identifies potential attributes definitions from conversations | `/api/analysis/attributes` |
| `match_intents.go` | Matches and evaluates intent classifications | `/api/analysis/intent` |
| `analyze_fee_disputes.go` | Analyzes fee dispute conversations with detailed analytics | `/api/analysis/attributes`, `/api/analysis/trends`, `/api/analysis/findings` |

## Utility Files

| File | Purpose |
|------|---------|
| `utils.go` | Common utility functions and API client for all scripts |
| `run_examples.sh` | Shell script to run individual or all examples easily |
| `SCRIPT_USAGE.md` | Detailed instructions on using the shell script |
| `PIPELINE_OVERVIEW.md` | Overview of how the scripts work together in a pipeline |
| `MOCK_DATA_USAGE.md` | Instructions for using scripts with mock data instead of a database |

## Getting Started

1. Make sure the API server is running (typically at http://localhost:8080)
2. Ensure you have a SQLite database with conversation data (or use mock data - see below)
3. Make the shell script executable:
   ```bash
   chmod +x run_examples.sh
   ```
4. Run an example:
   ```bash
   ./run_examples.sh -d /path/to/database.db generate_intents
   ```

See `SCRIPT_USAGE.md` for detailed instructions and options.

## Using Mock Data

You can now run the example scripts without a database by using mock data:

```bash
./run_examples.sh -m all
```

This uses predefined sample data instead of querying a database. Currently, only some scripts support mock data. See `MOCK_DATA_USAGE.md` for more details.

## Script Functionality

### generate_intents.go
Extracts the primary intent from conversation texts by sending each conversation to the API for intent classification. Uses the `/api/analysis/intent` endpoint.

### generate_attributes.go
Generates structured attribute values from conversations by extracting key information into a structured format. Uses the `/api/analysis/attributes` endpoint.

### group_intents.go
Groups similar intents together to identify patterns and common themes across conversations. Uses multiple API endpoints including `/api/analysis/patterns`.

### identify_attributes.go
Analyzes conversations to identify potential attribute definitions that could be extracted in future analysis. Uses the `/api/analysis/attributes` endpoint with the `generate_required` flag.

### match_intents.go
Evaluates intent classification against known intents, calculating precision, recall, and F1 scores. Uses the `/api/analysis/intent` endpoint.

### analyze_fee_disputes.go
Performs detailed analysis on fee dispute conversations, extracting specific patterns and insights. Uses multiple endpoints including `/api/analysis/attributes`, `/api/analysis/trends`, and `/api/analysis/findings`.

## API Integration

All scripts use the common `ApiClient` defined in `utils.go` to interact with the Discourse AI Analysis API. The client handles:

- Making HTTP requests to the appropriate endpoints
- Formatting request payloads
- Processing and validating responses
- Error handling and retries
- Debug output for troubleshooting

The client connects to the API server at `http://localhost:8080/api/analysis` by default.

## Database Integration

The scripts read conversation data from a SQLite database. The database schema is expected to have tables containing conversations with at least the following fields:
- Conversation ID
- Conversation text
- Metadata (where applicable)

Some scripts like `generate_intents.go` and `generate_attributes.go` can also save their results back to the database if a workflow ID is provided.

## Output Format

All scripts produce JSON output files containing the results of their analysis, saved to the specified output directory. 