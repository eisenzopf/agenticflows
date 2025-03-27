package utils

import (
	"fmt"
	"time"
)

// Conversation represents a conversation record from the database
type Conversation struct {
	ID   string
	Text string
}

// GetString safely extracts a string value from a map[string]interface{}
func GetString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// GetFloat64 safely extracts a float64 value from a map[string]interface{}
func GetFloat64(m map[string]interface{}, key string) float64 {
	switch v := m[key].(type) {
	case float64:
		return v
	case int:
		return float64(v)
	default:
		return 0.0
	}
}

// PrintTimeTaken prints the time taken for an operation
func PrintTimeTaken(startTime time.Time, operation string) {
	duration := time.Since(startTime)
	fmt.Printf("\nTime taken for %s: %v\n", operation, duration)
}

// PrettyPrint formats and prints JSON data
func PrettyPrint(data interface{}) {
	fmt.Printf("%+v\n", data)
}

// GetInt safely extracts an integer value from a map
func GetInt(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return 0
	}
}

// GetStringArray safely extracts a string array from a map
func GetStringArray(m map[string]interface{}, key string) []string {
	if val, ok := m[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if s, ok := v.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// GetMapArray safely extracts an array of maps from a map
func GetMapArray(m map[string]interface{}, key string) []map[string]interface{} {
	if val, ok := m[key].([]interface{}); ok {
		result := make([]map[string]interface{}, 0, len(val))
		for _, v := range val {
			if m, ok := v.(map[string]interface{}); ok {
				result = append(result, m)
			}
		}
		return result
	}
	return nil
}
