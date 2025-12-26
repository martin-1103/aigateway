package openai

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTranslateOpenAIStreamToClaude_ContentDelta(t *testing.T) {
	openaiChunk := `{"choices":[{"delta":{"content":"Hello world"}}]}`

	result := TranslateOpenAIStreamToClaude([]byte(openaiChunk))

	resultStr := string(result)
	if !strings.Contains(resultStr, "event: content_block_delta") {
		t.Error("Missing content_block_delta event")
	}
	if !strings.Contains(resultStr, "Hello world") {
		t.Error("Missing content text")
	}
	if !strings.Contains(resultStr, "text_delta") {
		t.Error("Missing text_delta type")
	}
}

func TestTranslateOpenAIStreamToClaude_MessageStart(t *testing.T) {
	openaiChunk := `{"choices":[{"delta":{"role":"assistant"}}]}`

	result := TranslateOpenAIStreamToClaude([]byte(openaiChunk))

	resultStr := string(result)
	if !strings.Contains(resultStr, "event: message_start") {
		t.Error("Missing message_start event")
	}
}

func TestTranslateOpenAIStreamToClaude_MessageStop(t *testing.T) {
	tests := []struct {
		name         string
		finishReason string
	}{
		{"stop", "stop"},
		{"length", "length"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openaiChunk := `{"choices":[{"delta":{},"finish_reason":"` + tt.finishReason + `"}]}`

			result := TranslateOpenAIStreamToClaude([]byte(openaiChunk))

			resultStr := string(result)
			if !strings.Contains(resultStr, "event: message_stop") {
				t.Errorf("Missing message_stop event for finish_reason=%s", tt.finishReason)
			}
		})
	}
}

func TestTranslateOpenAIStreamToClaude_EmptyChoices(t *testing.T) {
	openaiChunk := `{"choices":[]}`

	result := TranslateOpenAIStreamToClaude([]byte(openaiChunk))

	resultStr := string(result)
	if !strings.Contains(resultStr, "message_start") {
		t.Error("Should return message_start for empty choices")
	}
}

func TestTranslateOpenAIStreamToClaude_InvalidJSON(t *testing.T) {
	invalidJSON := `not valid json`

	result := TranslateOpenAIStreamToClaude([]byte(invalidJSON))

	// Should return original input for invalid JSON
	if string(result) != invalidJSON {
		t.Errorf("Should return original input, got: %s", string(result))
	}
}

func TestBuildClaudeContentDelta(t *testing.T) {
	result := buildClaudeContentDelta("test content")

	resultStr := string(result)

	// Check SSE format
	if !strings.HasPrefix(resultStr, "event: content_block_delta\n") {
		t.Error("Missing event prefix")
	}
	if !strings.Contains(resultStr, "data: ") {
		t.Error("Missing data prefix")
	}
	if !strings.HasSuffix(resultStr, "\n\n") {
		t.Error("Missing SSE double newline")
	}

	// Extract and verify JSON
	dataStart := strings.Index(resultStr, "data: ") + 6
	dataEnd := strings.LastIndex(resultStr, "\n\n")
	jsonData := resultStr[dataStart:dataEnd]

	var event map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("Invalid JSON in data: %v", err)
	}

	if event["type"] != "content_block_delta" {
		t.Errorf("type = %v, want content_block_delta", event["type"])
	}
	if event["index"] != float64(0) {
		t.Errorf("index = %v, want 0", event["index"])
	}

	delta := event["delta"].(map[string]interface{})
	if delta["type"] != "text_delta" {
		t.Errorf("delta.type = %v, want text_delta", delta["type"])
	}
	if delta["text"] != "test content" {
		t.Errorf("delta.text = %v, want 'test content'", delta["text"])
	}
}

func TestBuildClaudeChunk(t *testing.T) {
	data := map[string]interface{}{
		"custom": "field",
	}

	result := buildClaudeChunk("custom_event", data)

	resultStr := string(result)

	if !strings.Contains(resultStr, "event: custom_event") {
		t.Error("Missing custom event type")
	}

	// Verify type is added to data
	dataStart := strings.Index(resultStr, "data: ") + 6
	dataEnd := strings.LastIndex(resultStr, "\n\n")
	jsonData := resultStr[dataStart:dataEnd]

	var parsed map[string]interface{}
	json.Unmarshal([]byte(jsonData), &parsed)

	if parsed["type"] != "custom_event" {
		t.Errorf("type = %v, want custom_event", parsed["type"])
	}
	if parsed["custom"] != "field" {
		t.Errorf("custom = %v, want 'field'", parsed["custom"])
	}
}
