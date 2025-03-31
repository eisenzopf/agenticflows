# Intent Generation Workflow Example

This example demonstrates how to create and execute a workflow that generates intents from text input. It shows:

1. Creating a workflow with an intent generation node
2. Executing the workflow with sample data
3. Testing the workflow with multiple different intents

## Prerequisites

- The backend server must be running on `localhost:8080`
- The intent analysis function must be available in the backend

## Running the Example

1. Make sure the backend server is running
2. Run the example:

```bash
go run main.go
```

## What the Example Does

1. Creates a new workflow with a single intent generation node
2. Executes the workflow with a sample intent
3. Tests the workflow with multiple different intents to demonstrate its capabilities

## Sample Output

The example will output:
- The created workflow ID
- Results from processing the first intent
- Results from processing each additional intent in the sample list

## Understanding the Results

The results will include:
- The generated intent(s)
- Confidence scores
- Any additional metadata provided by the intent generation function

## Customization

You can modify the example by:
1. Changing the `sampleIntents` list to test different scenarios
2. Adjusting the `parameters` in the input data to configure the intent generation
3. Adding more nodes to the workflow to create more complex analysis chains 