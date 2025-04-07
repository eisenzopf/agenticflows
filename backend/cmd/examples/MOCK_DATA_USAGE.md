# Using Mock Data with Example Scripts

## Overview

The example scripts in this directory can now be run with mock data instead of requiring a real database connection. This makes it easier to test and demonstrate the functionality without needing to set up and populate a database.

_Last Updated: [Current Date]_
_Version: 1.0.1_

## How to Use Mock Data

To use mock data, run the `run_examples.sh` script with the `-m` or `--mock` flag:

```bash
./run_examples.sh -m all
```

When using the mock flag:
1. The scripts will use predefined sample conversations and data instead of querying a database
2. You don't need to provide a database path with `-d` or `--db`
3. Only scripts that have been updated to support mock data will run

## Currently Supported Scripts with Mock Data

The following scripts currently support the mock data option:

- `generate_intents.go`: Demonstrates intent classification using mock customer service conversations
- `create_action_plan.go`: Creates an action plan based on sample recommendations
- `generate_attributes.go`: Extracts attributes from mock conversation data
- `identify_attributes.go`: Identifies potential attributes from mock conversations

**Note:** We're actively adding mock data support to additional scripts. This list will be updated as more scripts are updated.

## Adding Mock Data Support to Other Scripts

To add mock data support to additional scripts:

1. Add a `--mock` flag to the script's command-line parameters
2. Add a function like `createMockData()` to generate appropriate sample data
3. Modify the script to use the mock data when the flag is set instead of querying the database

Example implementation:

```go
// Add to command-line flags
mockFlag := flag.Bool("mock", false, "Use mock data instead of database")

// Validate required flags
if *dbPath == "" && !*mockFlag {
    fmt.Println("Error: --db flag is required unless --mock is used")
    flag.Usage()
    os.Exit(1)
}

// Use mock or database data
var data []SomeDataType
if *mockFlag {
    data = createMockData()
} else {
    data, err = fetchDataFromDatabase(*dbPath)
    if err != nil {
        // Handle error
    }
}

// Example mock data function
func createMockData() []SomeDataType {
    return []SomeDataType{
        {ID: "mock-1", Value: "Sample data 1"},
        {ID: "mock-2", Value: "Sample data 2"},
        // Add more realistic sample data
    }
}
```

## Mock Data Examples

Below are examples of mock data structures for currently supported scripts:

### generate_intents.go Mock Data

```go
func createMockConversations() []Conversation {
    return []Conversation{
        {
            ID: "conv-001",
            Text: "Customer: Hi, I was charged twice for my subscription last month.\nAgent: I'm sorry to hear that. Let me check your account.",
            Metadata: map[string]interface{}{"channel": "chat", "duration": 350},
        },
        {
            ID: "conv-002",
            Text: "Customer: How do I upgrade my account to the premium version?\nAgent: I'd be happy to help with that upgrade.",
            Metadata: map[string]interface{}{"channel": "email", "duration": 620},
        },
        // Additional mock conversations...
    }
}
```

### create_action_plan.go Mock Data

```go
func createMockRecommendations() []Recommendation {
    return []Recommendation{
        {
            ID: "rec-001",
            Description: "Implement proactive notification for subscription renewals",
            Impact: "High",
            Effort: "Medium",
        },
        {
            ID: "rec-002",
            Description: "Add troubleshooting guide for common login issues",
            Impact: "Medium",
            Effort: "Low",
        },
        // Additional mock recommendations...
    }
}
```

## Benefits of Mock Data

- Run examples and tests without database setup
- Consistent test data for debugging and development
- Faster execution without database queries
- Easier demonstration of functionality
- Supports development when disconnected from database resources

## Running the Full Suite

When all scripts support mock data, you will be able to run the full suite of examples with:

```bash
./run_examples.sh -m all
```

Until then, you can run individual scripts with mock data by specifying them directly:

```bash
./run_examples.sh -m generate_intents
```

## Troubleshooting

If you encounter issues with mock data:

1. Ensure you're using the latest version of the scripts
2. Check that the script explicitly supports mock data (see supported list above)
3. Verify that the `-m` flag is properly included in your command
4. Check the script's source code to confirm the mock data implementation is complete 