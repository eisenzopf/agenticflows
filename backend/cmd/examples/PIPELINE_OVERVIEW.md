# Conversation Analysis Pipeline Overview

This document explains how the various example scripts work together to form a comprehensive conversation analysis pipeline.

## The Complete Analysis Pipeline

![Analysis Pipeline Flow](https://mermaid.ink/img/pako:eNp1kc1qwzAQhF9FaNBCCn6AWDm0JG1KenBoD6IXxa5sy8iSkRU3IeTdK8eGJoWeLM3Ozn47cFJJRYiJrm77N1hWvTY9Z0hllmqQFSvOWWYqwWutuElkXWQnQnQbGxN5F_RbzJAZ8wDwEhRUNJyJ1f0-HF_hLYy1hTl4t_ZefzXWbgcVHSCLnc1rqDBvXj8Xe4kKoLkTPf4TFOFDgbKbHkexvw4kcAQVR0QfcSf46dPVgaMqbOHoGl61MKK2TZs4KjXVpH89U3_FZGh8MvPcZFzyvS5y9xqBP6Uaag8n0TiJCLUeSIdQpX0x-nT8N0_RbVQkYSZGrSSFpHCsJh9LjcYkWo-ZCDMo-Y-HdwxZMIhPTBQ8ZFRpDrXzXfmYfQP_nI6D)

## Pipeline Steps

### 1. Intent Generation (generate_intents.go)
- **Input**: Raw conversation texts from database
- **Process**: Analyzes each conversation to determine primary intent
- **Output**: JSON file with conversations and their associated intents
- **Purpose**: Categorize conversations by primary customer intent
- **API Endpoint**: `/api/analysis/intent`

### 2. Intent Grouping (group_intents.go)
- **Input**: Generated intents from step 1
- **Process**: Groups similar intents to identify patterns
- **Output**: JSON file with intent groups and their related conversations
- **Purpose**: Identify common intent categories and reduce intent fragmentation
- **API Endpoint**: Uses several endpoints including `/api/analysis/patterns`

### 3. Attribute Identification (identify_attributes.go)
- **Input**: Conversations filtered by specific intents
- **Process**: Analyzes conversations to identify what attributes could be extracted
- **Output**: JSON file with suggested attribute definitions
- **Purpose**: Determine what structured data fields would be useful for further analysis
- **API Endpoint**: `/api/analysis/attributes` with `generate_required` flag

### 4. Attribute Generation (generate_attributes.go)
- **Input**: Conversations and attribute definitions
- **Process**: Extracts specific attribute values from conversations
- **Output**: JSON file with conversations and their extracted attributes
- **Purpose**: Create structured data from unstructured conversations
- **API Endpoint**: `/api/analysis/attributes`

### 5. Intent Matching (match_intents.go)
- **Input**: Generated intents and predefined intent categories
- **Process**: Compares generated intents against expected categories
- **Output**: JSON file with matching results and evaluation metrics
- **Purpose**: Evaluate intent classification accuracy and improve categorization
- **API Endpoint**: `/api/analysis/intent`

### 6. Fee Dispute Analysis (analyze_fee_disputes.go)
- **Input**: Conversations specifically about fee disputes with generated attributes
- **Process**: Deep analysis of fee dispute patterns, causes, and resolutions
- **Output**: JSON file with analysis results, trends, and insights
- **Purpose**: Gain specific insights into fee dispute conversations
- **API Endpoints**: `/api/analysis/attributes`, `/api/analysis/trends`, and `/api/analysis/findings`

## API Integration

All scripts communicate with the API server running at http://localhost:8080 and utilize the following endpoints:

- `/api/analysis/intent` - For intent classification
- `/api/analysis/attributes` - For attribute extraction and identification
- `/api/analysis/patterns` - For pattern identification in data
- `/api/analysis/trends` - For trend analysis in structured data
- `/api/analysis/findings` - For deep analysis of specific insights

The `ApiClient` class in `utils.go` handles all API communications with retries and error handling.

## Data Flow

The pipeline typically flows in the order above, but each script can also be run independently with appropriate inputs. The workflow ID parameter helps track related analyses across multiple script runs.

### Database → Scripts → JSON Files

1. **Data Source**: SQLite database containing conversation records
2. **Processing**: Each script reads relevant data and processes it through the API
3. **Results Storage**: Analysis results are saved as JSON files
4. **Optional**: Results can be loaded back into the database or other systems

## Use Cases

### Customer Intent Understanding
Use `generate_intents.go` followed by `group_intents.go` to understand what customers are trying to accomplish in their conversations.

### Structured Data Extraction
Use `identify_attributes.go` followed by `generate_attributes.go` to convert unstructured conversation text into structured data fields.

### Quality Evaluation
Use `match_intents.go` to evaluate how well the system is classifying intents compared to expected categories.

### Focused Analysis
Use `analyze_fee_disputes.go` to deeply analyze specific conversation types (in this case, fee disputes).

## Extending the Pipeline

The pipeline can be extended with additional scripts for:
- Sentiment analysis
- Topic modeling
- Customer satisfaction prediction
- Custom attribute generation
- Conversation clustering

Any new script should follow the pattern of reading input data, processing it through the API, and producing structured output. 