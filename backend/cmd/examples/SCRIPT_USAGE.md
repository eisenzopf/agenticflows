# How to Use the Example Scripts Runner

The `run_examples.sh` script provides an easy way to run all the example Go scripts for conversation analysis. It handles passing the appropriate parameters to each script and organizes the output files.

## Prerequisites

- Bash shell environment
- Go 1.18 or later
- A running instance of the API server (typically at http://localhost:8080)
- Access to a SQLite database with conversation data (or use mock data option)

## Basic Usage

```bash
./run_examples.sh -d /path/to/database.db all
```

This will run all the example scripts using the provided database, generating output files in the `./output` directory.

## Options

| Option | Description | Default | API Parameter |
|--------|-------------|---------|--------------|
| `-d, --database PATH` | Path to SQLite database (required unless using -m) | - | N/A |
| `-o, --output DIR` | Directory for output files | `./output` | N/A |
| `-w, --workflow ID` | Workflow ID | Generated timestamp | `workflow_id` |
| `-l, --limit NUM` | Limit number of items to process | 10 | N/A |
| `-s, --sample NUM` | Sample size for conversation analysis | 3 | N/A |
| `-m, --mock` | Use mock data instead of database | false | N/A |
| `-i, --intents LIST` | Comma-separated list of intents | "cancel,upgrade,billing" | Used in matching |
| `-c, --confidence NUM` | Confidence threshold | 0.7 | Used in intent filtering |
| `-v, --verbose` | Enable verbose/debug output | false | N/A |
| `-h, --help` | Show usage information | - | N/A |

## Available Scripts

Instead of running all scripts, you can run individual ones:

```bash
./run_examples.sh -d /path/to/database.db generate_intents
```

Available script names:
- `all` - Run all scripts in sequence
- `generate_intents` - Generate conversation intents (uses `/api/analysis/intent`)
- `generate_attributes` - Generate attribute values for conversations (uses `/api/analysis/attributes`)
- `group_intents` - Group similar intents together (uses `/api/analysis/patterns`)
- `identify_attributes` - Identify attribute definitions for conversations (uses `/api/analysis/attributes`)
- `match_intents` - Match and evaluate intent classifications (uses `/api/analysis/intent`)
- `analyze_fee_disputes` - Analyze fee dispute conversations (uses multiple endpoints)

## Using Mock Data

You can run the scripts without a database by using mock data:

```bash
./run_examples.sh -m all
```

When using the `-m` flag:
- No database connection is required
- Scripts use predefined sample data
- Only scripts that support mock data will run

Currently, the following scripts support mock data:
- `generate_intents` - Uses sample customer service conversations
- `create_action_plan` - Already uses sample data by default

For more details on the mock data implementation, see `MOCK_DATA_USAGE.md`.

## Examples

1. Run all examples with a specific workflow ID:
   ```bash
   ./run_examples.sh -d /path/to/database.db -w my-workflow-123 all
   ```

2. Run intent generation with custom limits:
   ```bash
   ./run_examples.sh -d /path/to/database.db -l 20 generate_intents
   ```

3. Run with mock data and debug output:
   ```bash
   ./run_examples.sh -m -v generate_intents
   ```

4. Custom output directory with mock data:
   ```bash
   ./run_examples.sh -m -o ./results -w test-workflow generate_intents
   ```

5. Run with debug output to see API requests and responses:
   ```bash
   ./run_examples.sh -d /path/to/database.db -v generate_attributes
   ```

## Output

The script creates timestamped JSON files in the output directory (default: `./output`). Each file contains the results of the corresponding analysis script.

Example output structure:
```
./output/
  ├── intents_20230515123045.json
  ├── attributes_20230515123112.json
  ├── intent_groups_20230515123157.json
  ├── attribute_definitions_20230515123245.json
  ├── intent_matching_20230515123318.json
  └── fee_dispute_analysis_20230515123402.json
```

## Script to API Mapping

The scripts interact with the API server through the following endpoints:

1. `generate_intents.go` → `/api/analysis/intent`
2. `generate_attributes.go` → `/api/analysis/attributes`
3. `group_intents.go` → `/api/analysis/patterns` and others
4. `identify_attributes.go` → `/api/analysis/attributes` with `generate_required` flag
5. `match_intents.go` → `/api/analysis/intent`
6. `analyze_fee_disputes.go` → Multiple endpoints (`/api/analysis/attributes`, `/api/analysis/trends`, `/api/analysis/findings`)

The command-line options from this script are passed to the individual scripts and ultimately mapped to API request parameters in each script's implementation. 