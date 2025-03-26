# How to Use the Example Scripts Runner

The `run_examples.sh` script provides an easy way to run all the example Go scripts for conversation analysis. It handles passing the appropriate parameters to each script and organizes the output files.

## Prerequisites

- Bash shell environment
- Go 1.18 or later
- A running instance of the API server (typically at http://localhost:8080)
- Access to a SQLite database with conversation data

## Basic Usage

```bash
./run_examples.sh -d /path/to/database.db all
```

This will run all the example scripts using the provided database, generating output files in the `./output` directory.

## Options

| Option | Description | Default |
|--------|-------------|---------|
| `-d, --database PATH` | Path to SQLite database (**required**) | - |
| `-o, --output DIR` | Directory for output files | `./output` |
| `-w, --workflow ID` | Workflow ID | Generated timestamp |
| `-l, --limit NUM` | Limit number of items to process | 10 |
| `-s, --sample NUM` | Sample size for conversation analysis | 3 |
| `-m, --min-count NUM` | Minimum count threshold | 5 |
| `-t, --target CLASS` | Target class for analysis | "fee dispute" |
| `-i, --intents LIST` | Comma-separated list of intents | "cancel,upgrade,billing" |
| `-c, --confidence NUM` | Confidence threshold | 0.7 |
| `-v, --verbose` | Enable verbose/debug output | false |
| `-h, --help` | Show usage information | - |

## Available Scripts

Instead of running all scripts, you can run individual ones:

```bash
./run_examples.sh -d /path/to/database.db generate_intents
```

Available script names:
- `all` - Run all scripts
- `generate_intents` - Generate conversation intents
- `generate_attributes` - Generate attribute values for conversations
- `group_intents` - Group similar intents together
- `identify_attributes` - Identify attribute definitions for conversations
- `match_intents` - Match and evaluate intent classifications
- `analyze_fee_disputes` - Analyze fee dispute conversations

## Examples

1. Run all examples with a specific workflow ID:
   ```bash
   ./run_examples.sh -d /path/to/database.db -w my-workflow-123 all
   ```

2. Run intent generation with custom limits:
   ```bash
   ./run_examples.sh -d /path/to/database.db -l 20 generate_intents
   ```

3. Analyze a specific type of conversation:
   ```bash
   ./run_examples.sh -d /path/to/database.db -t "subscription cancel" identify_attributes
   ```

4. Run with debug output to see API requests and responses:
   ```bash
   ./run_examples.sh -d /path/to/database.db -v generate_attributes
   ```

5. Custom output directory:
   ```bash
   ./run_examples.sh -d /path/to/database.db -o ./results match_intents
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