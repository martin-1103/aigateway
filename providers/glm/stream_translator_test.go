package glm

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTranslateGLMStreamToClaude_ContentDelta(t *testing.T) {
	glmChunk := `{"choices":[{"delta":{"content":"Hello from GLM"}}]}`

	result := TranslateGLMStreamToClaude([]byte(glmChunk))

	resultStr := string(result)
	if !strings.Contains(resultStr, "event: content_block_delta") {
		t.Error("Missing content_block_delta event")
	}
	if !strings.Contains(resultStr, "Hello from GLM") {
		t.Error("Missing content text")
	}
	if !strings.Contains(resultStr, "text_delta") {
		t.Error("Missing text_delta type")
	}
}

func TestTranslateGLMStreamToClaude_MessageStart(t *testing.T) {
	glmChunk := `{"choices":[{"delta":{"role":"assistant"}}]}`

	result := TranslateGLMStreamToClaude([]byte(glmChunk))

	resultStr := string(result)
	if !strings.Contains(resultStr, "event: message_start") {
		t.Error("Missing message_start event")
	}
}

func TestTranslateGLMStreamToClaude_MessageStop(t *testing.T) {
	tests := []struct {
		name         string
		finishReason string
	}{
		{"stop", "stop"},
		{"length", "length"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			glmChunk := `{"choices":[{"delta":{},"finish_reason":"` + tt.finishReason + `"}]}`

			result := TranslateGLMStreamToClaude([]byte(glmChunk))

			resultStr := string(result)
			if !strings.Contains(resultStr, "event: message_stop") {
				t.Errorf("Missing message_stop event for finish_reason=%s", tt.finishReason)
			}
		})
	}
}

func TestTranslateGLMStreamToClaude_EmptyChoices(t *testing.T) {
	glmChunk := `{"choices":[]}`

	result := TranslateGLMStreamToClaude([]byte(glmChunk))

	resultStr := string(result)
	if !strings.Contains(resultStr, "message_start") {
		t.Error("Should return message_start for empty choices")
	}
}

func TestTranslateGLMStreamToClaude_InvalidJSON(t *testing.T) {
	invalidJSON := `not valid json`

	result := TranslateGLMStreamToClaude([]byte(invalidJSON))

	// Should return original input for invalid JSON
	if string(result) != invalidJSON {
		t.Errorf("Should return original input, got: %s", string(result))
	}
}

func TestBuildClaudeContentDelta_GLM(t *testing.T) {
	result := buildClaudeContentDelta("GLM content")

	resultStr := string(result)

	// Check SSE format
	if !strings.HasPrefix(resultStr, "event: content_block_delta\n") {
		t.Error("Missing event prefix")
	}
	if !strings.Contains(resultStr, "data: ") {
		t.Error("Missing data prefix")
	}

	// Extract and verify JSON
	dataStart := strings.Index(resultStr, "data: ") + 6
	dataEnd := strings.LastIndex(resultStr, "\n\n")
	jsonData := resultStr[dataStart:dataEnd]

	var event map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &event); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	delta := event["delta"].(map[string]interface{})
	if delta["text"] != "GLM content" {
		t.Errorf("delta.text = %v, want 'GLM content'", delta["text"])
	}
}

func TestBuildClaudeChunk_GLM(t *testing.T) {
	data := map[string]interface{}{
		"message": map[string]interface{}{
			"role": "assistant",
		},
	}

	result := buildClaudeChunk("message_start", data)

	resultStr := string(result)

	if !strings.Contains(resultStr, "event: message_start") {
		t.Error("Missing message_start event")
	}

	dataStart := strings.Index(resultStr, "data: ") + 6
	dataEnd := strings.LastIndex(resultStr, "\n\n")
	jsonData := resultStr[dataStart:dataEnd]

	var parsed map[string]interface{}
	json.Unmarshal([]byte(jsonData), &parsed)

	if parsed["type"] != "message_start" {
		t.Errorf("type = %v, want message_start", parsed["type"])
	}
}
