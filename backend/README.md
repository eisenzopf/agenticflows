# Discourse AI Backend

This is the backend for the Discourse AI application.

## Getting Started

### Prerequisites

- Go 1.21+
- Google Gemini API key (set as `GEMINI_API_KEY` environment variable)
- SQLite database (automatically initialized during startup)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/eisenzopf/agenticflows.git
   cd agenticflows
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the server:
   ```bash
   cd backend
   go build -o server ./api
   ```

### Running the Server

1. Set the required environment variables:
   ```bash
   export GEMINI_API_KEY="your-api-key-here"
   ```

2. Start the server:
   ```bash
   ./server
   ```

   The server will start on port 8080 by default. You should see:
   ```
   Starting server on :8080
   Analysis endpoints initialized with types: trends, patterns, findings, attributes, intent, recommendations, plan
   ```

3. For development with mock data (no API key required):
   
   When making API requests, add the `use_mock_data: true` parameter to avoid making actual LLM API calls (see example below).

## API Endpoints

### Analysis Endpoint

`POST /api/analysis`

The analysis endpoint provides various analytical capabilities for conversation data.

#### Parameters

- `analysis_type`: (Required) String. The type of analysis to perform. Supported values:
  - `intent`
  - `attributes`
  - `required_attributes`
  - `trends`
  - `patterns`
  - `findings`
  - `recommendations`
  - `action_plan`
  - `timeline`

- `use_mock_data`: (Optional) Boolean. When set to `true`, the API will return predefined mock data instead of making actual LLM API calls. This is useful for:
  - Testing environments
  - Demonstrations
  - Development when the LLM API is unavailable or rate-limited
  - CI/CD pipelines

Example request with mock data:
```json
{
  "analysis_type": "recommendations",
  "parameters": {
    "conversations": [...],
    "use_mock_data": true
  }
}
```

When `use_mock_data` is not specified or set to `false`, the API will use actual data processing and LLM calls to generate results.

## Running Examples

See the `cmd/examples` directory for example implementations and the `run_examples.sh` script to execute them.

## Testing

Run the API tests using:
```
cd cmd/examples
./test_api_endpoints.sh
```

The test script uses the `use_mock_data` parameter to avoid relying on external APIs during testing. 