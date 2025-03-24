package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	// Database file path - relative to the current directory
	dbName = "data/agenticflows.db"
)

var (
	// Single database connection instance
	DB *sql.DB
)

// Agent represents an agent component
type Agent struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

// Tool represents a tool component
type Tool struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
}

// Workflow represents a workflow with its ReactFlow configuration
type Workflow struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Date  string          `json:"date"`
	Nodes json.RawMessage `json:"nodes"`
	Edges json.RawMessage `json:"edges"`
}

// Initialize sets up the database connection and creates tables if they don't exist
func Initialize() error {
	// Ensure the database file exists
	dbPath := dbName
	
	// Get the absolute path for logging
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		absPath = dbPath
	}
	log.Printf("Using database at path: %s", absPath)
	
	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		dir := filepath.Dir(dbPath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create database directory: %w", err)
			}
		}
		log.Printf("Creating new database file at: %s", absPath)
		file, err := os.Create(dbPath)
		if err != nil {
			return fmt.Errorf("failed to create database file: %w", err)
		}
		file.Close()
	}

	// Open database connection
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Create tables if they don't exist
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}

	// Insert initial data for agents and tools
	if err := initializeComponentData(); err != nil {
		return fmt.Errorf("failed to initialize component data: %w", err)
	}

	log.Println("Database initialized successfully")
	return nil
}

// createTables creates the necessary tables if they don't exist
func createTables() error {
	// Create agents table
	_, err := DB.Exec(`
		CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			label TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create tools table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS tools (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			label TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	// Create workflows table
	_, err = DB.Exec(`
		CREATE TABLE IF NOT EXISTS workflows (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			date TEXT NOT NULL,
			nodes TEXT NOT NULL,
			edges TEXT NOT NULL
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// initializeComponentData inserts the initial agents and tools data
func initializeComponentData() error {
	// Get count of agents to check if we need to insert initial data
	var agentCount int
	err := DB.QueryRow("SELECT COUNT(*) FROM agents").Scan(&agentCount)
	if err != nil {
		return err
	}

	// Only insert initial data if the tables are empty
	if agentCount == 0 {
		// Agent components data
		agents := []Agent{
			{ID: "agent-assist", Type: "agent", Label: "Agent Assist"},
			{ID: "analyzer", Type: "agent", Label: "Analyzer"},
			{ID: "coach", Type: "agent", Label: "Coach"},
			{ID: "evaluator", Type: "agent", Label: "Evaluator"},
			{ID: "knowledge", Type: "agent", Label: "Knowledge"},
			{ID: "observer", Type: "agent", Label: "Observer"},
			{ID: "orchestrator", Type: "agent", Label: "Orchestrator"},
			{ID: "planner", Type: "agent", Label: "Planner"},
			{ID: "researcher", Type: "agent", Label: "Researcher"},
			{ID: "superviser-assist", Type: "agent", Label: "Superviser Assist"},
			{ID: "trainer", Type: "agent", Label: "Trainer"},
		}

		// Insert agents data
		for _, agent := range agents {
			_, err := DB.Exec(
				"INSERT INTO agents (id, type, label) VALUES (?, ?, ?)",
				agent.ID, agent.Type, agent.Label,
			)
			if err != nil {
				return err
			}
		}

		// Tool components data
		tools := []Tool{
			{ID: "agent-guide", Type: "tool", Label: "Agent Guide"},
			{ID: "agent-scorecard", Type: "tool", Label: "Agent Scorecard"},
			{ID: "agent-training", Type: "tool", Label: "Agent Training"},
			{ID: "call-observation", Type: "tool", Label: "Call Observation"},
			{ID: "competitive-analysis", Type: "tool", Label: "Competitive Analysis"},
			{ID: "goal-tracking", Type: "tool", Label: "Goal Tracking"},
			{ID: "talent-builder", Type: "tool", Label: "Talent Builder"},
		}

		// Insert tools data
		for _, tool := range tools {
			_, err := DB.Exec(
				"INSERT INTO tools (id, type, label) VALUES (?, ?, ?)",
				tool.ID, tool.Type, tool.Label,
			)
			if err != nil {
				return err
			}
		}

		log.Println("Initial component data inserted successfully")
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
} 