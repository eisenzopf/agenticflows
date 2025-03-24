package db

// GetAllAgents returns all agent components from the database
func GetAllAgents() ([]Agent, error) {
	rows, err := DB.Query("SELECT id, type, label FROM agents ORDER BY label")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []Agent
	for rows.Next() {
		var agent Agent
		if err := rows.Scan(&agent.ID, &agent.Type, &agent.Label); err != nil {
			return nil, err
		}
		agents = append(agents, agent)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return agents, nil
}

// GetAllTools returns all tool components from the database
func GetAllTools() ([]Tool, error) {
	rows, err := DB.Query("SELECT id, type, label FROM tools ORDER BY label")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tools []Tool
	for rows.Next() {
		var tool Tool
		if err := rows.Scan(&tool.ID, &tool.Type, &tool.Label); err != nil {
			return nil, err
		}
		tools = append(tools, tool)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tools, nil
}

// AddAgent inserts a new agent component
func AddAgent(agent Agent) error {
	_, err := DB.Exec(
		"INSERT INTO agents (id, type, label) VALUES (?, ?, ?)",
		agent.ID, agent.Type, agent.Label,
	)
	return err
}

// UpdateAgent updates an existing agent component
func UpdateAgent(id string, agent Agent) error {
	_, err := DB.Exec(
		"UPDATE agents SET type = ?, label = ? WHERE id = ?",
		agent.Type, agent.Label, id,
	)
	return err
}

// DeleteAgent removes an agent component
func DeleteAgent(id string) error {
	_, err := DB.Exec("DELETE FROM agents WHERE id = ?", id)
	return err
}

// AddTool inserts a new tool component
func AddTool(tool Tool) error {
	_, err := DB.Exec(
		"INSERT INTO tools (id, type, label) VALUES (?, ?, ?)",
		tool.ID, tool.Type, tool.Label,
	)
	return err
}

// UpdateTool updates an existing tool component
func UpdateTool(id string, tool Tool) error {
	_, err := DB.Exec(
		"UPDATE tools SET type = ?, label = ? WHERE id = ?",
		tool.Type, tool.Label, id,
	)
	return err
}

// DeleteTool removes a tool component
func DeleteTool(id string) error {
	_, err := DB.Exec("DELETE FROM tools WHERE id = ?", id)
	return err
} 