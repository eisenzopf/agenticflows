# AgenticFlows API Server - Refactored Structure

This directory contains the refactored API server for AgenticFlows. The code has been organized into a modular structure to improve maintainability and separation of concerns.

## Directory Structure

- **api/** - API server code
  - **main.go** - Server entry point
  - **routes.go** - Route setup (moved to main.go)
  - **handlers/** - Request handlers
    - **analysis.go** - Analysis endpoint handlers
    - **components.go** - Agents and tools handlers
    - **questions.go** - Question answering handlers
    - **workflows.go** - Workflow CRUD handlers
  - **models/** - API data models
    - **api_models.go** - Request/response models

- **workflow/** - Workflow-related functionality
  - **generator.go** - Workflow generation logic
  - **execution.go** - Workflow execution logic

## API Endpoints

- **/api/agents** - Manage agents
- **/api/tools** - Manage tools
- **/api/workflows** - Manage workflows
- **/api/workflows/{id}** - Get, update, or delete a specific workflow
- **/api/workflows/{id}/execution-config** - Get execution configuration for a workflow
- **/api/workflows/{id}/execute** - Execute a workflow
- **/api/workflows/generate** - Generate a new workflow
- **/api/workflows/generate-dynamic** - Generate a dynamic workflow
- **/api/questions/answer** - Answer questions
- **/api/analysis** - Perform various types of analysis
- **/api/analysis/chain** - Perform chain analysis
- **/api/analysis/metadata** - Get analysis function metadata
- **/api/analysis/results** - Manage analysis results

## Code Organization

The refactoring aimed to improve the organization of the codebase by:

1. Separating handlers by functionality
2. Moving data models to their own package
3. Extracting workflow generation and execution logic
4. Keeping the main.go file focused on initialization and setup

## Development

To run the server:

```bash
go run main.go
```

The server will start on port 8080. 