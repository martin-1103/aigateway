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
	// Antigravity: "candidates.0.content.parts": [{"text": "..."}, {"functionCall": {...}}, {"thought": true, "text": "...", "thoughtSignature": "..."}]
	// Claude: "content": [{"type": "text", "text": "..."}, {"type": "tool_use", ...}, {"type": "thinking", "thinking": "...", "signature": "..."}]
	parts := responseNode.Get("candidates.0.content.parts")
	if parts.IsArray() {
		for _, part := range parts.Array() {
			// Handle thinking/thought blocks (must come before text check)
			if thought := part.Get("thought"); thought.Exists() && thought.Bool() {
				thinkingPart := `{"type":"thinking","thinking":""}`
				thinkingText := part.Get("text").String()
				thinkingPart, _ = sjson.Set(thinkingPart, "thinking", thinkingText)

				// Add signature if present
				signature := part.Get("thoughtSignature").String()
				if signature == "" {
					signature = part.Get("thought_signature").String()
				}
				if signature != "" {
					thinkingPart, _ = sjson.Set(thinkingPart, "signature", signature)
				}
				contentJSON, _ = sjson.SetRaw(contentJSON, "content.-1", thinkingPart)
				continue
			}

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

				// Use ID from response or generate one
				toolID := functionCall.Get("id").String()
				if toolID == "" {
					toolID = "toolu_" + name
				}
				toolUsePart, _ = sjson.Set(toolUsePart, "id", toolID)
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

				// Use ID from response or generate one
				toolID := functionResponse.Get("id").String()
				if toolID == "" {
					toolID = "toolu_" + name
				}
				toolResultPart, _ = sjson.Set(toolResultPart, "tool_use_id", toolID)

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

			// Handle inline data (images)
			if inlineData := part.Get("inlineData"); inlineData.Exists() {
				imagePart := `{"type":"image","source":{"type":"base64"}}`
				if mimeType := inlineData.Get("mime_type").String(); mimeType != "" {
					imagePart, _ = sjson.Set(imagePart, "source.media_type", mimeType)
				}
				if data := inlineData.Get("data").String(); data != "" {
					imagePart, _ = sjson.Set(imagePart, "source.data", data)
				}
				contentJSON, _ = sjson.SetRaw(contentJSON, "content.-1", imagePart)
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

	// Check for response wrapper
	if responseNode.Get("response").Exists() {
		responseNode = responseNode.Get("response")
	}

	// Check if this is a complete message or delta
	if responseNode.Get("candidates.0.content.parts").Exists() {
		parts := responseNode.Get("candidates.0.content.parts")
		if !parts.IsArray() || len(parts.Array()) == 0 {
			return chunk
		}

		part := parts.Array()[0]

		// Handle thinking/thought delta
		if thought := part.Get("thought"); thought.Exists() && thought.Bool() {
			eventJSON := `{"type":"content_block_delta","index":0,"delta":{}}`
			thinkingText := part.Get("text").String()

			deltaJSON := `{"type":"thinking_delta","thinking":""}`
			deltaJSON, _ = sjson.Set(deltaJSON, "thinking", thinkingText)
			eventJSON, _ = sjson.SetRaw(eventJSON, "delta", deltaJSON)

			// Include signature if present
			signature := part.Get("thoughtSignature").String()
			if signature == "" {
				signature = part.Get("thought_signature").String()
			}
			if signature != "" {
				eventJSON, _ = sjson.Set(eventJSON, "signature", signature)
			}

			return []byte(eventJSON)
		}

		// Extract text delta
		text := part.Get("text").String()
		if text != "" {
			eventJSON := `{"type":"content_block_delta","index":0,"delta":{}}`
			deltaJSON := `{"type":"text_delta","text":""}`
			deltaJSON, _ = sjson.Set(deltaJSON, "text", text)
			eventJSON, _ = sjson.SetRaw(eventJSON, "delta", deltaJSON)
			return []byte(eventJSON)
		}

		// Extract function call delta
		functionCall := part.Get("functionCall")
		if functionCall.Exists() {
			eventJSON := `{"type":"content_block_delta","index":0,"delta":{}}`
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
