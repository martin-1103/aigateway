package antigravity

import (
	"encoding/json"
	"fmt"
)

// TranslateAntigravityStreamToClaude converts Antigravity SSE to Claude format
func TranslateAntigravityStreamToClaude(data []byte, eventType string) []byte {
	var antigravityResp map[string]interface{}
	if err := json.Unmarshal(data, &antigravityResp); err != nil {
		return data
	}

	// Check for candidates (Gemini format)
	if candidates, ok := antigravityResp["candidates"].([]interface{}); ok && len(candidates) > 0 {
		candidate := candidates[0].(map[string]interface{})
		content, ok := candidate["content"].(map[string]interface{})
		if !ok {
			return data
		}

		parts, ok := content["parts"].([]interface{})
		if !ok || len(parts) == 0 {
			return data
		}

		part := parts[0].(map[string]interface{})

		// Check for text content
		if text, ok := part["text"].(string); ok {
			return buildClaudeContentDelta(text)
		}

		// Check for finish reason
		if finishReason, ok := candidate["finishReason"].(string); ok {
			if finishReason == "STOP" || finishReason == "MAX_TOKENS" {
				return buildClaudeChunk("message_stop", map[string]interface{}{})
			}
		}
	}

	// Default: message start
	return buildClaudeChunk("message_start", map[string]interface{}{
		"message": map[string]interface{}{
			"role": "assistant",
		},
	})
}

// buildClaudeContentDelta creates a Claude content_block_delta event
func buildClaudeContentDelta(text string) []byte {
	event := map[string]interface{}{
		"type":  "content_block_delta",
		"index": 0,
		"delta": map[string]interface{}{
			"type": "text_delta",
			"text": text,
		},
	}

	data, _ := json.Marshal(event)
	return []byte(fmt.Sprintf("event: content_block_delta\ndata: %s\n\n", data))
}

// buildClaudeChunk creates a Claude SSE event
func buildClaudeChunk(eventType string, data map[string]interface{}) []byte {
	data["type"] = eventType
	jsonData, _ := json.Marshal(data)
	return []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, jsonData))
}
