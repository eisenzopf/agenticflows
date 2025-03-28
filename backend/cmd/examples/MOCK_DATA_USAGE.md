# Using Mock Data with Example Scripts

## Overview

The example scripts in this directory can now be run with mock data instead of requiring a real database connection. This makes it easier to test and demonstrate the functionality without needing to set up and populate a database.

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

- `generate_intents`: Demonstrates intent classification using mock customer service conversations
- `create_action_plan`: Creates an action plan based on sample recommendations (already used sample data)

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