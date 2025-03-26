# Contact Center Analysis Package

This package provides Go implementations of contact center analysis functionality, ported from the Python library. It leverages Google's Gemini LLM API to analyze customer service conversations, extract attributes, identify intents, and generate insights.

## Features

- **Text Analysis**
  - Extract attributes from conversations
  - Identify customer intents
  - Generate required attributes for specific research questions

- **Data Analysis**
  - Analyze trends in conversation data
  - Identify conversation patterns
  - Analyze findings from attribute extraction

- **API Integration**
  - RESTful API for all analysis functions
  - Rate-limited Gemini API client
  - Database storage for analysis results

## API Endpoints

- `/api/analysis/trends` - Analyze trends in conversation data
- `/api/analysis/patterns` - Identify patterns in conversation data
- `/api/analysis/findings` - Analyze findings from attribute extraction
- `/api/analysis/attributes` - Extract or generate attributes from text
- `/api/analysis/intent` - Identify the primary intent in a conversation
- `/api/analysis/results` - Retrieve stored analysis results

## Requirements

- Go 1.21+
- Google Gemini API key (set as `GEMINI_API_KEY` environment variable)
- SQLite database

## Usage

### Command Line Testing

Use the included test client to test the API:

```bash
# Test intent classification
go run cmd/testclient/main.go -endpoint intent -text "I want to cancel my subscription"

# Process a file
go run cmd/testclient/main.go -endpoint attributes -file ./sample.txt

# View results for a workflow
go run cmd/testclient/main.go -endpoint results -workflow abc123
```

### Library Usage

```go
import "agenticflows/backend/analysis"

// Create a text generator
textGen, err := analysis.NewTextGenerator(apiKey, false)
if err != nil {
    log.Fatalf("Failed to create text generator: %v", err)
}

// Generate intent from text
intent, err := textGen.GenerateIntent(ctx, conversationText)
if err != nil {
    log.Fatalf("Failed to generate intent: %v", err)
}
fmt.Printf("Intent: %s (%s)\n", intent.LabelName, intent.Label)
fmt.Printf("Description: %s\n", intent.Description)
```

## Architecture

The package is organized into several components:

- `models.go` - Data structures for requests and responses
- `llm.go` - Gemini API integration
- `text.go` - Text analysis functions
- `analyzer.go` - Core analysis functions
- `ratelimiter.go` - API rate limiting

## Database Integration

Analysis results are stored in the SQLite database using the `analysis_results` table. Results can be retrieved by ID or by workflow ID. 