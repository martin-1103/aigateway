package antigravity

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// TranslateAntigravityToClaude converts Antigravity API response format to Claude format
// Input: Antigravity format with candidates, parts, usageMetadata
// Output: Claude format with role, content, stop_reason, usage
func TranslateAntigravityToClaude(payload []byte) []byte {
	// Check if response is wrapped in "response" key
	responseNode := gjson.GetBytes(payload, "response")
	if !responseNode.Exists() {
		responseNode = gjson.ParseBytes(payload)
	}

	// Extract role and convert model to assistant
	// Antigravity: "candidates.0.content.role": "model"
	// Claude: "role": "assistant"
	role := responseNode.Get("candidates.0.content.role").String()
	if role == "model" {
		role = "assistant"
	}
	if role == "" {
		role = "assistant" // Default to assistant
	}

	contentJSON := `{"role":"","content":[]}`
	contentJSON, _ = sjson.Set(contentJSON, "role", role)

	// Convert parts to content blocks
	// Antigravity: "candidates.0.content.parts": [{"text": "..."}, {"functionCall": {...}}]
	// Claude: "content": [{"type": "text", "text": "..."}, {"type": "tool_use", ...}]
	parts := responseNode.Get("candidates.0.content.parts")
	if parts.IsArray() {
		for _, part := range parts.Array() {
			// Handle text parts
			if text := part.Get("text"); text.Exists() {
				textPart := `{"type":"text","text":""}`
				textPart, _ = sjson.Set(textPart, "text", text.String())
				contentJSON, _ = sjson.SetRaw(contentJSON, "content.-1", textPart)
			}

			// Handle function call (tool use)
			if functionCall := part.Get("functionCall"); functionCall.Exists() {
				toolUsePart := `{"type":"tool_use","id":"","name":"","input":{}}`
				name := functionCall.Get("name").String()
				args := functionCall.Get("args")

				// Generate a simple ID based on function name
				toolUsePart, _ = sjson.Set(toolUsePart, "id", "toolu_"+name)
				toolUsePart, _ = sjson.Set(toolUsePart, "name", name)
				if args.Exists() {
					toolUsePart, _ = sjson.SetRaw(toolUsePart, "input", args.Raw)
				}
				contentJSON, _ = sjson.SetRaw(contentJSON, "content.-1", toolUsePart)
			}

			// Handle function response (tool result)
			if functionResponse := part.Get("functionResponse"); functionResponse.Exists() {
				toolResultPart := `{"type":"tool_result","tool_use_id":"","content":""}`
				name := functionResponse.Get("name").String()
				response := functionResponse.Get("response")

				toolResultPart, _ = sjson.Set(toolResultPart, "tool_use_id", "toolu_"+name)

				// Extract result from response
				if result := response.Get("result"); result.Exists() {
					if result.Type == gjson.String {
						toolResultPart, _ = sjson.Set(toolResultPart, "content", result.String())
					} else {
						toolResultPart, _ = sjson.SetRaw(toolResultPart, "content", result.Raw)
					}
				} else {
					// If no result field, use entire response
					toolResultPart, _ = sjson.SetRaw(toolResultPart, "content", response.Raw)
				}
				contentJSON, _ = sjson.SetRaw(contentJSON, "content.-1", toolResultPart)
			}
		}
	}

	// Convert finish reason
	// Antigravity: "candidates.0.finishReason": "STOP", "MAX_TOKENS", "SAFETY", "OTHER"
	// Claude: "stop_reason": "end_turn", "max_tokens", "stop_sequence", "tool_use"
	finishReason := responseNode.Get("candidates.0.finishReason").String()
	stopReason := convertFinishReason(finishReason)
	contentJSON, _ = sjson.Set(contentJSON, "stop_reason", stopReason)

	// Add stop_sequence if applicable
	if stopReason == "stop_sequence" {
		contentJSON, _ = sjson.Set(contentJSON, "stop_sequence", "")
	}

	// Convert usage metadata
	// Antigravity: "usageMetadata": {"promptTokenCount": 10, "candidatesTokenCount": 20, "totalTokenCount": 30}
	// Claude: "usage": {"input_tokens": 10, "output_tokens": 20}
	usage := responseNode.Get("usageMetadata")
	if usage.Exists() {
		inputTokens := usage.Get("promptTokenCount").Int()
		outputTokens := usage.Get("candidatesTokenCount").Int()

		contentJSON, _ = sjson.Set(contentJSON, "usage.input_tokens", inputTokens)
		contentJSON, _ = sjson.Set(contentJSON, "usage.output_tokens", outputTokens)
	}

	// Add model ID if present
	if model := responseNode.Get("modelVersion"); model.Exists() {
		contentJSON, _ = sjson.Set(contentJSON, "model", model.String())
	}

	// Add response ID
	contentJSON, _ = sjson.Set(contentJSON, "id", "msg_antigravity")
	contentJSON, _ = sjson.Set(contentJSON, "type", "message")

	return []byte(contentJSON)
}

// convertFinishReason maps Antigravity finish reasons to Claude stop reasons
func convertFinishReason(finishReason string) string {
	switch finishReason {
	case "STOP":
		return "end_turn"
	case "MAX_TOKENS":
		return "max_tokens"
	case "SAFETY":
		return "end_turn"
	case "RECITATION":
		return "end_turn"
	case "OTHER":
		return "end_turn"
	case "":
		return "end_turn"
	default:
		return "end_turn"
	}
}

// TranslateAntigravityStreamToClaude converts streaming response chunks
func TranslateAntigravityStreamToClaude(chunk []byte) []byte {
	// For streaming, we need to handle SSE format
	// Antigravity sends: data: {"candidates": [...]}
	// Claude expects: data: {"type": "content_block_delta", ...}

	responseNode := gjson.ParseBytes(chunk)

	// Check if this is a complete message or delta
	if responseNode.Get("candidates.0.content.parts").Exists() {
		// This is a delta chunk
		eventJSON := `{"type":"content_block_delta","index":0,"delta":{}}`

		// Extract text delta
		text := responseNode.Get("candidates.0.content.parts.0.text").String()
		if text != "" {
			deltaJSON := `{"type":"text_delta","text":""}`
			deltaJSON, _ = sjson.Set(deltaJSON, "text", text)
			eventJSON, _ = sjson.SetRaw(eventJSON, "delta", deltaJSON)
			return []byte(eventJSON)
		}

		// Extract function call delta
		functionCall := responseNode.Get("candidates.0.content.parts.0.functionCall")
		if functionCall.Exists() {
			deltaJSON := `{"type":"input_json_delta","partial_json":""}`
			deltaJSON, _ = sjson.SetRaw(deltaJSON, "partial_json", functionCall.Raw)
			eventJSON, _ = sjson.SetRaw(eventJSON, "delta", deltaJSON)
			return []byte(eventJSON)
		}
	}

	// Handle finish message
	if finishReason := responseNode.Get("candidates.0.finishReason"); finishReason.Exists() {
		stopReason := convertFinishReason(finishReason.String())
		eventJSON := `{"type":"message_delta","delta":{"stop_reason":""}}`
		eventJSON, _ = sjson.Set(eventJSON, "delta.stop_reason", stopReason)

		// Add usage if present
		if usage := responseNode.Get("usageMetadata"); usage.Exists() {
			eventJSON, _ = sjson.Set(eventJSON, "usage.output_tokens", usage.Get("candidatesTokenCount").Int())
		}
		return []byte(eventJSON)
	}

	// Return original if we can't parse
	return chunk
}
