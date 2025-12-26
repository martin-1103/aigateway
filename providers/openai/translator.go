package openai

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ClaudeToOpenAI converts Claude format request to OpenAI format
// OpenAI and Claude formats are largely compatible, with minor differences
func ClaudeToOpenAI(payload []byte, model string) ([]byte, error) {
	result := string(payload)

	// Handle system message - OpenAI uses messages array with role "system"
	systemResult := gjson.GetBytes(payload, "system")
	if systemResult.Exists() {
		messagesResult := gjson.GetBytes(payload, "messages")
		if messagesResult.IsArray() {
			// Prepend system message to messages array
			systemMsg := fmt.Sprintf(`{"role":"system","content":"%s"}`, systemResult.String())
			messages := messagesResult.Array()
			newMessages := `[` + systemMsg
			for _, msg := range messages {
				newMessages += `,` + msg.Raw
			}
			newMessages += `]`
			result, _ = sjson.SetRaw(result, "messages", newMessages)
		}
		result, _ = sjson.Delete(result, "system")
	}

	// Handle tool calling format differences
	// Claude uses "tools" with "input_schema", OpenAI uses "functions" or "tools"
	toolsResult := gjson.GetBytes(payload, "tools")
	if toolsResult.IsArray() {
		functionsJSON := `[]`
		for _, tool := range toolsResult.Array() {
			functionJSON := `{}`

			// Map tool name
			if name := tool.Get("name"); name.Exists() {
				functionJSON, _ = sjson.Set(functionJSON, "name", name.String())
			}

			// Map description
			if desc := tool.Get("description"); desc.Exists() {
				functionJSON, _ = sjson.Set(functionJSON, "description", desc.String())
			}

			// Map input_schema to parameters
			if inputSchema := tool.Get("input_schema"); inputSchema.Exists() {
				functionJSON, _ = sjson.SetRaw(functionJSON, "parameters", inputSchema.Raw)
			}

			functionsJSON, _ = sjson.SetRaw(functionsJSON, "-1", functionJSON)
		}

		// OpenAI uses "tools" with type "function"
		openaiTools := `[]`
		for _, fn := range gjson.Parse(functionsJSON).Array() {
			toolWrapper := `{"type":"function"}`
			toolWrapper, _ = sjson.SetRaw(toolWrapper, "function", fn.Raw)
			openaiTools, _ = sjson.SetRaw(openaiTools, "-1", toolWrapper)
		}

		result, _ = sjson.SetRaw(result, "tools", openaiTools)
	}

	// Handle content format - ensure text content is properly formatted
	messagesResult := gjson.GetBytes(payload, "messages")
	if messagesResult.IsArray() {
		for i, msg := range messagesResult.Array() {
			content := msg.Get("content")

			// If content is array of objects (Claude format), convert to string or proper OpenAI format
			if content.IsArray() {
				hasMultipleParts := false
				textContent := ""

				for _, part := range content.Array() {
					if part.Get("type").String() == "text" {
						if textContent != "" {
							hasMultipleParts = true
						}
						textContent += part.Get("text").String()
					}
				}

				// If single text part, use string format
				if !hasMultipleParts && textContent != "" {
					result, _ = sjson.Set(result, fmt.Sprintf("messages.%d.content", i), textContent)
				}
			}
		}
	}

	// Set model
	result, _ = sjson.Set(result, "model", model)

	return []byte(result), nil
}

// OpenAIToClaude converts OpenAI response to Claude format
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
	//     "message": {"role": "assistant", "content": "..."},
	//     "finish_reason": "stop"
	//   }],
	//   "usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
	// }

	claudeResponse := `{}`

	// Extract message content from first choice
	message := result.Get("choices.0.message")
	if message.Exists() {
		// Map role (should be "assistant")
		role := message.Get("role").String()
		if role == "" {
			role = "assistant"
		}
		claudeResponse, _ = sjson.Set(claudeResponse, "role", role)

		// Map content - convert to array format for consistency
		content := message.Get("content")
		if content.Exists() {
			contentArray := `[]`
			textBlock := `{"type":"text"}`
			textBlock, _ = sjson.Set(textBlock, "text", content.String())
			contentArray, _ = sjson.SetRaw(contentArray, "0", textBlock)
			claudeResponse, _ = sjson.SetRaw(claudeResponse, "content", contentArray)
		}

		// Handle tool calls if present
		toolCalls := message.Get("tool_calls")
		if toolCalls.IsArray() {
			for _, toolCall := range toolCalls.Array() {
				toolUse := `{"type":"tool_use"}`
				toolUse, _ = sjson.Set(toolUse, "id", toolCall.Get("id").String())
				toolUse, _ = sjson.Set(toolUse, "name", toolCall.Get("function.name").String())

				// Parse arguments JSON string
				argsStr := toolCall.Get("function.arguments").String()
				if argsStr != "" {
					var args interface{}
					if err := json.Unmarshal([]byte(argsStr), &args); err == nil {
						argsBytes, _ := json.Marshal(args)
						toolUse, _ = sjson.SetRaw(toolUse, "input", string(argsBytes))
					}
				}

				claudeResponse, _ = sjson.SetRaw(claudeResponse, "content.-1", toolUse)
			}
		}
	}

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
