package openai

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// OpenAIToClaude converts OpenAI response to Claude format
// Handles text content, tool_calls (tool_use), and usage statistics
func OpenAIToClaude(payload []byte) ([]byte, error) {
	result := gjson.ParseBytes(payload)

	// OpenAI response structure:
	// {
	//   "id": "chatcmpl-xxx",
	//   "object": "chat.completion",
	//   "created": 1234567890,
	//   "model": "gpt-4",
	//   "choices": [{
	//     "index": 0,
	//     "message": {"role": "assistant", "content": "...", "tool_calls": [...]},
	//     "finish_reason": "stop"
	//   }],
	//   "usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
	// }

	claudeResponse := `{}`

	// Extract message from first choice
	message := result.Get("choices.0.message")
	if !message.Exists() {
		return []byte(claudeResponse), nil
	}

	// Map role (should be "assistant")
	role := message.Get("role").String()
	if role == "" {
		role = "assistant"
	}
	claudeResponse, _ = sjson.Set(claudeResponse, "role", role)

	// Build content array
	claudeResponse = buildContentArray(message, claudeResponse)

	// Map finish_reason to stop_reason
	finishReason := result.Get("choices.0.finish_reason").String()
	stopReason := mapFinishReason(finishReason)
	claudeResponse, _ = sjson.Set(claudeResponse, "stop_reason", stopReason)

	// Map usage statistics
	usage := result.Get("usage")
	if usage.Exists() {
		claudeResponse, _ = sjson.Set(claudeResponse, "usage.input_tokens", usage.Get("prompt_tokens").Int())
		claudeResponse, _ = sjson.Set(claudeResponse, "usage.output_tokens", usage.Get("completion_tokens").Int())
	}

	// Map model
	model := result.Get("model")
	if model.Exists() {
		claudeResponse, _ = sjson.Set(claudeResponse, "model", model.String())
	}

	// Map ID
	id := result.Get("id")
	if id.Exists() {
		claudeResponse, _ = sjson.Set(claudeResponse, "id", id.String())
	}

	return []byte(claudeResponse), nil
}

// buildContentArray constructs Claude content array from OpenAI message
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

// buildToolUseBlock converts OpenAI tool_call to Claude tool_use block
// OpenAI: {"id":"call_xxx","type":"function","function":{"name":"get_weather","arguments":"{\"city\":\"Jakarta\"}"}}
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

// mapFinishReason maps OpenAI finish_reason to Claude stop_reason
func mapFinishReason(finishReason string) string {
	switch finishReason {
	case "stop":
		return "end_turn"
	case "length":
		return "max_tokens"
	case "tool_calls":
		return "tool_use"
	case "content_filter":
		return "stop_sequence"
	default:
		return "end_turn"
	}
}
