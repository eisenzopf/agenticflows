package analysis

import (
	"context"
	"encoding/json"
	"fmt"
)

// TextGenerator handles text generation and attribute extraction
type TextGenerator struct {
	llmClient *LLMClient
	debug     bool
}

// NewTextGenerator creates a new TextGenerator instance
func NewTextGenerator(apiKey string, debug bool) (*TextGenerator, error) {
	llmClient, err := NewLLMClient(apiKey, debug)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}
	return &TextGenerator{
		llmClient: llmClient,
		debug:     debug,
	}, nil
}

// GenerateRequiredAttributes generates attributes needed to answer specific questions
func (t *TextGenerator) GenerateRequiredAttributes(ctx context.Context, questions []string, existingAttributes []string) ([]AttributeDefinition, error) {
	// Validate request
	if len(questions) == 0 {
		return nil, fmt.Errorf("questions are required")
	}

	// Format questions for the prompt
	questionsText := ""
	for i, q := range questions {
		questionsText += fmt.Sprintf("%d. %s\n", i+1, q)
	}

	// Format existing attributes if provided
	existingText := ""
	if len(existingAttributes) > 0 {
		existingText = "\nExisting attributes:\n"
		for _, attr := range existingAttributes {
			existingText += fmt.Sprintf("- %s\n", attr)
		}
	}

	prompt := fmt.Sprintf(`We need to determine what data attributes are required to answer these questions:
%s
%s

Return a JSON object with this structure:
{
  "attributes": [
    {
      "field_name": str,  // Database field name in snake_case
      "title": str,       // Human readable title
      "description": str, // Detailed description of the attribute
      "rationale": str    // Why this attribute is needed for the questions
    }
  ]
}`, questionsText, existingText)

	expectedFormat := map[string]interface{}{
		"attributes": []interface{}{},
	}

	result, err := t.llmClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract attributes from the result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	attributesRaw, ok := resultMap["attributes"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("attributes field is not an array")
	}

	// Convert to strongly typed attributes
	attributes := make([]AttributeDefinition, 0, len(attributesRaw))
	for _, attrRaw := range attributesRaw {
		attrMap, ok := attrRaw.(map[string]interface{})
		if !ok {
			continue // Skip invalid entries
		}

		attr := AttributeDefinition{
			FieldName:   getString(attrMap, "field_name"),
			Title:       getString(attrMap, "title"),
			Description: getString(attrMap, "description"),
			Rationale:   getString(attrMap, "rationale"),
		}

		// Only add if field_name is valid
		if attr.FieldName != "" {
			attributes = append(attributes, attr)
		}
	}

	return attributes, nil
}

// GenerateAttribute generates a value for a specific attribute from text
func (t *TextGenerator) GenerateAttribute(ctx context.Context, text string, attribute AttributeDefinition) (*AttributeValue, error) {
	// Validate inputs
	if text == "" {
		return &AttributeValue{
			FieldName:   attribute.FieldName,
			Value:       "No content",
			Confidence:  0.0,
			Explanation: "No content was provided for analysis. The text input was empty.",
		}, nil
	}

	prompt := fmt.Sprintf(`Analyze this text to determine the value for the following attribute:

Attribute: %s
Description: %s

Text to analyze:
%s

Return a JSON object with this structure:
{
  "value": str,           // The extracted or determined value
  "confidence": float,    // Confidence score between 0 and 1
  "explanation": str      // Explanation of how the value was determined
}

Ensure the response is specific to the attribute definition and supported by the text content.`, 
	attribute.Title, attribute.Description, truncateText(text, 5000))

	expectedFormat := map[string]interface{}{
		"value":       "",
		"confidence":  0.0,
		"explanation": "",
	}

	result, err := t.llmClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract values from the result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	// Create attribute value
	attrValue := &AttributeValue{
		FieldName:   attribute.FieldName,
		Value:       getString(resultMap, "value"),
		Confidence:  getFloat(resultMap, "confidence"),
		Explanation: getString(resultMap, "explanation"),
	}

	return attrValue, nil
}

// GenerateAttributes generates values for multiple attributes from text in a single LLM call
func (t *TextGenerator) GenerateAttributes(ctx context.Context, text string, attributes []AttributeDefinition) ([]AttributeValue, error) {
	// Validate inputs
	if text == "" {
		emptyValues := make([]AttributeValue, 0, len(attributes))
		for _, attr := range attributes {
			emptyValues = append(emptyValues, AttributeValue{
				FieldName:   attr.FieldName,
				Value:       "No content",
				Confidence:  0.0,
				Explanation: "No content was provided for analysis. The text input was empty.",
			})
		}
		return emptyValues, nil
	}

	if len(attributes) == 0 {
		return []AttributeValue{}, nil
	}

	// Format attributes for the prompt
	attributesText := ""
	for _, attr := range attributes {
		attributesText += fmt.Sprintf("Attribute: %s\nField Name: %s\nDescription: %s\n\n", 
			attr.Title, attr.FieldName, attr.Description)
	}

	prompt := fmt.Sprintf(`Analyze this text to determine values for the following attributes:

%s
Text to analyze:
%s

Return a JSON object with this structure:
{
  "attribute_values": [
    {
      "field_name": str,     // Must match one of the field names provided above
      "value": str,          // The extracted or determined value
      "confidence": float,   // Confidence score between 0 and 1
      "explanation": str     // Explanation of how the value was determined
    }
  ]
}

Ensure each response is specific to the attribute definition and supported by the text content.
Include all requested attributes in your response, even if the confidence is low.`, 
	attributesText, truncateText(text, 8000))

	expectedFormat := map[string]interface{}{
		"attribute_values": []interface{}{},
	}

	result, err := t.llmClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract values from the result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	attrValuesRaw, ok := resultMap["attribute_values"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("attribute_values field is not an array")
	}

	// Convert to strongly typed attribute values
	attrValues := make([]AttributeValue, 0, len(attrValuesRaw))
	for _, valRaw := range attrValuesRaw {
		valMap, ok := valRaw.(map[string]interface{})
		if !ok {
			continue // Skip invalid entries
		}

		attrValue := AttributeValue{
			FieldName:   getString(valMap, "field_name"),
			Value:       getString(valMap, "value"),
			Confidence:  getFloat(valMap, "confidence"),
			Explanation: getString(valMap, "explanation"),
		}

		// Only add if field_name is valid
		if attrValue.FieldName != "" {
			attrValues = append(attrValues, attrValue)
		}
	}

	return attrValues, nil
}

// GenerateIntent generates the primary intent of a customer service conversation
func (t *TextGenerator) GenerateIntent(ctx context.Context, text string) (*IntentClassification, error) {
	// Validate input
	if text == "" {
		return &IntentClassification{
			LabelName:   "Unclear Intent",
			Label:       "unclear_intent",
			Description: "The conversation transcript is unclear or does not contain a discernible customer service request.",
		}, nil
	}

	prompt := fmt.Sprintf(`You are a helpful AI assistant specializing in classifying customer service conversations. Your task is to analyze a provided conversation transcript and determine the customer's *primary* intent for contacting customer service. Focus on the *main reason* the customer initiated the interaction, even if other topics are briefly mentioned.

**Input:** You will receive a conversation transcript as text.

**Output:** You will return a JSON object with the following *exact* keys and data types:

* **"label_name"**: (string) A natural language label describing the customer's primary intent. This label should be 2-3 words *maximum*. Use title case (e.g., "Update Address", "Cancel Order").
* **"label"**: (string) A lowercase version of "label_name", with underscores replacing spaces (e.g., "update_address", "cancel_order"). This should be machine-readable.
* **"description"**: (string) A concise description (1-2 sentences) of the customer's primary intent. Explain the *specific* problem or request the customer is making.

**Important Instructions and Constraints:**

1. **Primary Intent Focus:** Identify the *single, most important* reason the customer contacted support. Ignore minor side issues if they are not the core reason for the interaction.
2. **Conciseness:** Keep the "label_name" to 2-3 words and the "description" brief and to the point.
3. **JSON Format:** The output *must* be valid JSON. Do not include any extra text, explanations, or apologies outside of the JSON object. Only the JSON object should be returned.
4. **Specificity:** Be as specific as possible in the description. Don't just say "billing issue." Say "The customer is disputing a charge on their latest bill."
5. **Do not hallucinate information.** Base the classification solely on the provided transcript. Do not invent details.
6. **Do not respond in a conversational manner.** Your entire response should be only the requested json.

Conversation Transcript:
%s`, truncateText(text, 8000))

	expectedFormat := map[string]interface{}{
		"label_name":  "",
		"label":       "",
		"description": "",
	}

	result, err := t.llmClient.GenerateContent(ctx, prompt, expectedFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content: %w", err)
	}

	// Extract values from the result
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result format")
	}

	// Create intent classification
	intent := &IntentClassification{
		LabelName:   getString(resultMap, "label_name"),
		Label:       getString(resultMap, "label"),
		Description: getString(resultMap, "description"),
	}

	// Set default if missing
	if intent.LabelName == "" || intent.Label == "" {
		intent.LabelName = "Unclear Intent"
		intent.Label = "unclear_intent"
		if intent.Description == "" {
			intent.Description = "The conversation transcript is unclear or does not contain a discernible customer service request."
		}
	}

	return intent, nil
}

// Helper functions for type conversion

// getString safely extracts a string from a map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}

// getFloat safely extracts a float from a map
func getFloat(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case string:
			f, err := json.Number(v).Float64()
			if err == nil {
				return f
			}
		}
	}
	return 0.0
}

// truncateText safely truncates text to a maximum length
func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "... [text truncated]"
} 