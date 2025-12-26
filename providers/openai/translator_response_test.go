package openai

import (
	"encoding/json"
	"testing"
)

func TestOpenAIToClaude_TextContent(t *testing.T) {
	openaiResp := `{
		"id": "chatcmpl-123",
		"object": "chat.completion",
		"model": "gpt-4",
		"choices": [{
			"index": 0,
			"message": {"role": "assistant", "content": "Hello, world!"},
			"finish_reason": "stop"
		}],
		"usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
	}`

	result, err := OpenAIToClaude([]byte(openaiResp))
	if err != nil {
		t.Fatalf("OpenAIToClaude() error = %v", err)
	}

	var claudeResp map[string]interface{}
	if err := json.Unmarshal(result, &claudeResp); err != nil {
		t.Fatalf("Failed to parse result: %v", err)
	}

	// Check role
	if claudeResp["role"] != "assistant" {
		t.Errorf("role = %v, want 'assistant'", claudeResp["role"])
	}

	// Check content array
	content := claudeResp["content"].([]interface{})
	if len(content) != 1 {
		t.Fatalf("content length = %d, want 1", len(content))
	}

	textBlock := content[0].(map[string]interface{})
	if textBlock["type"] != "text" {
		t.Errorf("content[0].type = %v, want 'text'", textBlock["type"])
	}
	if textBlock["text"] != "Hello, world!" {
		t.Errorf("content[0].text = %v, want 'Hello, world!'", textBlock["text"])
	}

	// Check stop_reason
	if claudeResp["stop_reason"] != "end_turn" {
		t.Errorf("stop_reason = %v, want 'end_turn'", claudeResp["stop_reason"])
	}

	// Check usage
	usage := claudeResp["usage"].(map[string]interface{})
	if usage["input_tokens"] != float64(10) {
		t.Errorf("usage.input_tokens = %v, want 10", usage["input_tokens"])
	}
	if usage["output_tokens"] != float64(20) {
		t.Errorf("usage.output_tokens = %v, want 20", usage["output_tokens"])
	}
}

func TestOpenAIToClaude_ToolCalls(t *testing.T) {
	openaiResp := `{
		"choices": [{
			"message": {
				"role": "assistant",
				"content": null,
				"tool_calls": [{
					"id": "call_abc123",
					"type": "function",
					"function": {
						"name": "get_weather",
						"arguments": "{\"city\":\"Jakarta\",\"unit\":\"celsius\"}"
					}
				}]
			},
			"finish_reason": "tool_calls"
		}]
	}`

	result, err := OpenAIToClaude([]byte(openaiResp))
	if err != nil {
		t.Fatalf("OpenAIToClaude() error = %v", err)
	}

	var claudeResp map[string]interface{}
	json.Unmarshal(result, &claudeResp)

	// Check content has tool_use block
	content := claudeResp["content"].([]interface{})
	if len(content) != 1 {
		t.Fatalf("content length = %d, want 1", len(content))
	}

	toolUse := content[0].(map[string]interface{})
	if toolUse["type"] != "tool_use" {
		t.Errorf("content[0].type = %v, want 'tool_use'", toolUse["type"])
	}
	if toolUse["id"] != "call_abc123" {
		t.Errorf("content[0].id = %v, want 'call_abc123'", toolUse["id"])
	}
	if toolUse["name"] != "get_weather" {
		t.Errorf("content[0].name = %v, want 'get_weather'", toolUse["name"])
	}

	// Check input is parsed as object (not string)
	input := toolUse["input"].(map[string]interface{})
	if input["city"] != "Jakarta" {
		t.Errorf("input.city = %v, want 'Jakarta'", input["city"])
	}
	if input["unit"] != "celsius" {
		t.Errorf("input.unit = %v, want 'celsius'", input["unit"])
	}

	// Check stop_reason
	if claudeResp["stop_reason"] != "tool_use" {
		t.Errorf("stop_reason = %v, want 'tool_use'", claudeResp["stop_reason"])
	}
}

func TestOpenAIToClaude_MultipleToolCalls(t *testing.T) {
	openaiResp := `{
		"choices": [{
			"message": {
				"role": "assistant",
				"content": "I'll check both cities.",
				"tool_calls": [
					{
						"id": "call_1",
						"type": "function",
						"function": {"name": "get_weather", "arguments": "{\"city\":\"Jakarta\"}"}
					},
					{
						"id": "call_2",
						"type": "function",
						"function": {"name": "get_weather", "arguments": "{\"city\":\"Tokyo\"}"}
					}
				]
			},
			"finish_reason": "tool_calls"
		}]
	}`

	result, _ := OpenAIToClaude([]byte(openaiResp))

	var claudeResp map[string]interface{}
	json.Unmarshal(result, &claudeResp)

	content := claudeResp["content"].([]interface{})
	// Should have text + 2 tool_use blocks
	if len(content) != 3 {
		t.Fatalf("content length = %d, want 3", len(content))
	}

	// First should be text
	if content[0].(map[string]interface{})["type"] != "text" {
		t.Error("First content block should be text")
	}

	// Second and third should be tool_use
	if content[1].(map[string]interface{})["type"] != "tool_use" {
		t.Error("Second content block should be tool_use")
	}
	if content[2].(map[string]interface{})["type"] != "tool_use" {
		t.Error("Third content block should be tool_use")
	}
}

func TestMapFinishReason(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"stop", "end_turn"},
		{"length", "max_tokens"},
		{"tool_calls", "tool_use"},
		{"content_filter", "stop_sequence"},
		{"unknown", "end_turn"},
		{"", "end_turn"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := mapFinishReason(tt.input)
			if got != tt.want {
				t.Errorf("mapFinishReason(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestOpenAIToClaude_EmptyResponse(t *testing.T) {
	openaiResp := `{"choices":[]}`

	result, err := OpenAIToClaude([]byte(openaiResp))
	if err != nil {
		t.Fatalf("OpenAIToClaude() error = %v", err)
	}

	// Should return valid JSON, even if empty
	var claudeResp map[string]interface{}
	if err := json.Unmarshal(result, &claudeResp); err != nil {
		t.Fatalf("Result is not valid JSON: %v", err)
	}
}
