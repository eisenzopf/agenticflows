package utils

import (
	"encoding/json"
	"fmt"
)

// ExtractValue extracts a value from a map with type checking
func ExtractValue[T any](m map[string]interface{}, key string) (T, error) {
	var empty T

	if m == nil {
		return empty, fmt.Errorf("map is nil")
	}

	value, exists := m[key]
	if !exists {
		return empty, fmt.Errorf("key %s not found in map", key)
	}

	result, ok := value.(T)
	if !ok {
		return empty, fmt.Errorf("value for key %s is not of expected type", key)
	}

	return result, nil
}

// ExtractStringFromMap extracts a string from a map
func ExtractStringFromMap(m map[string]interface{}, key string) (string, bool) {
	if m == nil {
		return "", false
	}

	value, exists := m[key]
	if !exists {
		return "", false
	}

	strValue, ok := value.(string)
	if !ok {
		return "", false
	}

	return strValue, true
}

// StringSliceFromInterface converts an interface{} to a []string
func StringSliceFromInterface(input interface{}) ([]string, error) {
	// If it's already a []string, return it
	if strSlice, ok := input.([]string); ok {
		return strSlice, nil
	}

	// If it's an []interface{}, convert it
	if slice, ok := input.([]interface{}); ok {
		result := make([]string, 0, len(slice))
		for _, item := range slice {
			if str, ok := item.(string); ok {
				result = append(result, str)
			} else {
				// Try to convert non-string items to string
				jsonBytes, err := json.Marshal(item)
				if err != nil {
					return nil, fmt.Errorf("failed to convert item to string: %w", err)
				}
				result = append(result, string(jsonBytes))
			}
		}
		return result, nil
	}

	// If it's a string, parse it as JSON
	if str, ok := input.(string); ok {
		var result []string
		if err := json.Unmarshal([]byte(str), &result); err != nil {
			// If it's not valid JSON, treat it as a single-item slice
			return []string{str}, nil
		}
		return result, nil
	}

	return nil, fmt.Errorf("cannot convert input to string slice")
}

// ExtractStringSlice extracts a string slice from a map
func ExtractStringSlice(params map[string]interface{}, key string) ([]string, error) {
	if params == nil {
		return nil, fmt.Errorf("params is nil")
	}

	value, exists := params[key]
	if !exists {
		return nil, fmt.Errorf("key %s not found in params", key)
	}

	return StringSliceFromInterface(value)
}

// ToJSONString converts an object to a JSON string
func ToJSONString(obj interface{}) string {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return fmt.Sprintf("Error marshaling to JSON: %v", err)
	}
	return string(jsonBytes)
}

// FromJSONString parses a JSON string into an object
func FromJSONString(jsonStr string, result interface{}) error {
	return json.Unmarshal([]byte(jsonStr), result)
}
