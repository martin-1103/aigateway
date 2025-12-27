package openai

import (
	"encoding/json"
	"fmt"
)

// TranslateOpenAIStreamToClaude converts OpenAI SSE chunk to Claude SSE format
func TranslateOpenAIStreamToClaude(chunk []byte) []byte {
	var openaiChunk map[string]interface{}
	if err := json.Unmarshal(chunk, &openaiChunk); err != nil {
		return chunk
	}

	choices, ok := openaiChunk["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return buildClaudeChunk("message_start", map[string]interface{}{})
	}

	choice := choices[0].(map[string]interface{})
	delta, ok := choice["delta"].(map[string]interface{})
	if !ok {
		return chunk
	}

	// Check finish reason
	finishReason, _ := choice["finish_reason"].(string)
	if finishReason == "stop" || finishReason == "length" {
		return buildClaudeChunk("message_stop", map[string]interface{}{})
	}

	// Check for content delta
	if content, ok := delta["content"].(string); ok {
		return buildClaudeContentDelta(content)
	}

	// Check for role (message start)
	if role, ok := delta["role"].(string); ok && role == "assistant" {
		return buildClaudeChunk("message_start", map[string]interface{}{
			"message": map[string]interface{}{
				"role": "assistant",
			},
		})
	}

	return chunk
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
