package glm

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// TranslateGLMToClaude converts GLM response to Claude format
// Handles text content, tool_calls (tool_use), and usage statistics
func TranslateGLMToClaude(payload []byte) []byte {
	result := `{"role":"assistant","content":[]}`
	choice := gjson.GetBytes(payload, "choices.0")
	if !choice.Exists() {
		return []byte(result)
	}

	message := choice.Get("message")
	role := message.Get("role").String()
	if role == "" {
		role = "assistant"
	}
	result, _ = sjson.Set(result, "role", role)

	// Build content array
	result = buildContentArray(message, result)

	// Add stop reason
	stopMap := map[string]string{
		"stop":          "end_turn",
		"length":        "max_tokens",
		"tool_calls":    "tool_use",
		"function_call": "tool_use",
	}
	stopReason := stopMap[choice.Get("finish_reason").String()]
	if stopReason == "" {
		stopReason = "end_turn"
	}
	result, _ = sjson.Set(result, "stop_reason", stopReason)

	// Add usage and model
	usage := gjson.GetBytes(payload, "usage")
	if usage.Exists() {
		result, _ = sjson.Set(result, "usage.input_tokens", usage.Get("prompt_tokens").Int())
		result, _ = sjson.Set(result, "usage.output_tokens", usage.Get("completion_tokens").Int())
	}
	if model := gjson.GetBytes(payload, "model"); model.Exists() {
		result, _ = sjson.Set(result, "model", model.String())
	}

	return []byte(result)
}

// buildContentArray constructs Claude content array from GLM message
// Handles text content and tool_calls
func buildContentArray(message gjson.Result, claudeResponse string) string {
	contentArray := "[]"
	contentIndex := 0

	// Add text content if present
	content := message.Get("content")
	if content.Exists() && content.String() != "" {
		textBlock := `{"type":"text","text":""}`
		textBlock, _ = sjson.Set(textBlock, "text", content.String())
		contentArray, _ = sjson.SetRaw(contentArray, "0", textBlock)
		contentIndex = 1
	}

	// Handle tool_calls - convert to tool_use blocks
	toolCalls := message.Get("tool_calls")
	if toolCalls.IsArray() && len(toolCalls.Array()) > 0 {
		for _, toolCall := range toolCalls.Array() {
			toolUseBlock := buildToolUseBlock(toolCall)
			contentArray, _ = sjson.SetRaw(contentArray, fmt.Sprintf("%d", contentIndex), toolUseBlock)
			contentIndex++
		}
	}

	claudeResponse, _ = sjson.SetRaw(claudeResponse, "content", contentArray)
	return claudeResponse
}

// buildToolUseBlock converts GLM tool_call to Claude tool_use block
// GLM: {"id":"call_xxx","type":"function","function":{"name":"get_weather","arguments":"{\"city\":\"Jakarta\"}"}}
// Claude: {"type":"tool_use","id":"call_xxx","name":"get_weather","input":{"city":"Jakarta"}}
func buildToolUseBlock(toolCall gjson.Result) string {
	toolUse := `{"type":"tool_use","id":"","name":"","input":{}}`

	// Extract tool_call ID
	id := toolCall.Get("id").String()
	toolUse, _ = sjson.Set(toolUse, "id", id)

	// Extract function name
	name := toolCall.Get("function.name").String()
	toolUse, _ = sjson.Set(toolUse, "name", name)

	// Parse arguments JSON string to object
	argsStr := toolCall.Get("function.arguments").String()
	if argsStr != "" {
		// Parse the JSON string arguments
		var argsObj interface{}
		if err := json.Unmarshal([]byte(argsStr), &argsObj); err == nil {
			// Successfully parsed - set as object
			argsBytes, _ := json.Marshal(argsObj)
			toolUse, _ = sjson.SetRaw(toolUse, "input", string(argsBytes))
		} else {
			// Parse failed - set empty object
			toolUse, _ = sjson.Set(toolUse, "input", map[string]interface{}{})
		}
	} else {
		// No arguments - set empty object
		toolUse, _ = sjson.Set(toolUse, "input", map[string]interface{}{})
	}

	return toolUse
}
