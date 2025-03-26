# Conversation Analysis Example Scripts

This directory contains a collection of Go scripts that demonstrate how to use the Discourse AI Analysis API for conversation analysis. These scripts replicate the functionality of the original Python scripts but utilize the new Go API endpoints.

## Scripts Overview

| Script | Purpose |
|--------|---------|
| `generate_intents.go` | Extracts the primary intent from conversation texts |
| `generate_attributes.go` | Generates structured attribute values from conversations |
| `group_intents.go` | Groups similar intents together to identify patterns |
| `identify_attributes.go` | Identifies potential attributes definitions from conversations |
| `match_intents.go` | Matches and evaluates intent classifications |
| `analyze_fee_disputes.go` | Analyzes fee dispute conversations with detailed analytics |

## Utility Files

| File | Purpose |
|------|---------|
| `utils.go` | Common utility functions and API client for all scripts |
| `run_examples.sh` | Shell script to run individual or all examples easily |
| `SCRIPT_USAGE.md` | Detailed instructions on using the shell script |

## Getting Started

1. Make sure the API server is running (typically at http://localhost:8080)
2. Ensure you have a SQLite database with conversation data
3. Make the shell script executable:
   ```bash
   chmod +x run_examples.sh
   ```
4. Run an example:
   ```bash
   ./run_examples.sh -d /path/to/database.db generate_intents
   ```

See `SCRIPT_USAGE.md` for detailed instructions and options.

## Script Functionality

### generate_intents.go
Extracts the primary intent from conversation texts by sending each conversation to the API for intent classification.

### generate_attributes.go
Generates structured attribute values from conversations by extracting key information into a structured format.

### group_intents.go
Groups similar intents together to identify patterns and common themes across conversations.

### identify_attributes.go
Analyzes conversations to identify potential attribute definitions that could be extracted in future analysis.

### match_intents.go
Evaluates intent classification against known intents, calculating precision, recall, and F1 scores.

### analyze_fee_disputes.go
Performs detailed analysis on fee dispute conversations, extracting specific patterns and insights.

## API Integration

All scripts use the common `ApiClient` defined in `utils.go` to interact with the Discourse AI Analysis API. The client handles:

- Making HTTP requests to the appropriate endpoints
- Formatting request payloads
- Processing and validating responses
- Error handling and retries
- Debug output for troubleshooting

## Database Integration

The scripts read conversation data from a SQLite database. The database schema is expected to have tables containing conversations with at least the following fields:
- Conversation ID
- Conversation text
- Metadata (where applicable)

## Output Format

All scripts produce JSON output files containing the results of their analysis, saved to the specified output directory. 