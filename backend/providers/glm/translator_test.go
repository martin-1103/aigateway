package glm

import (
	"encoding/json"
	"testing"
)

// Request Translation Tests

func TestTranslateClaudeToGLM_ToolResult(t *testing.T) {
	claudeReq := `{
		"messages": [
			{"role": "user", "content": "What's the weather?"},
			{"role": "user", "content": [{"type": "tool_result", "tool_use_id": "call_123", "content": "25°C"}]}
		]
	}`

	result := TranslateClaudeToGLM([]byte(claudeReq), "glm-4")

	var glmReq map[string]interface{}
	json.Unmarshal(result, &glmReq)

	messages := glmReq["messages"].([]interface{})
	toolMsg := messages[1].(map[string]interface{})

	if toolMsg["role"] != "tool" {
		t.Errorf("role = %v, want 'tool'", toolMsg["role"])
	}
	if toolMsg["tool_call_id"] != "call_123" {
		t.Errorf("tool_call_id = %v, want 'call_123'", toolMsg["tool_call_id"])
	}
	if toolMsg["content"] != "25°C" {
		t.Errorf("content = %v, want '25°C'", toolMsg["content"])
	}
}

func TestTranslateClaudeToGLM_ImageContent(t *testing.T) {
	claudeReq := `{
		"messages": [{
			"role": "user",
			"content": [
				{"type": "text", "text": "Describe this image"},
				{"type": "image", "source": {"type": "base64", "media_type": "image/jpeg", "data": "base64data"}}
			]
		}]
	}`

	result := TranslateClaudeToGLM([]byte(claudeReq), "glm-4v")

	var glmReq map[string]interface{}
	json.Unmarshal(result, &glmReq)

	messages := glmReq["messages"].([]interface{})
	msg := messages[0].(map[string]interface{})
	content := msg["content"].([]interface{})

	if len(content) != 2 {
		t.Fatalf("content length = %d, want 2", len(content))
	}

	imagePart := content[1].(map[string]interface{})
	if imagePart["type"] != "image_url" {
		t.Errorf("content[1].type = %v, want 'image_url'", imagePart["type"])
	}

	imageURL := imagePart["image_url"].(map[string]interface{})
	if imageURL["url"] != "data:image/jpeg;base64,base64data" {
		t.Errorf("image_url.url = %v", imageURL["url"])
	}
}

func TestTranslateClaudeToGLM_SystemMessage(t *testing.T) {
	claudeReq := `{
		"system": "You are helpful",
		"messages": [{"role": "user", "content": "Hello"}]
	}`

	result := TranslateClaudeToGLM([]byte(claudeReq), "glm-4")

	var glmReq map[string]interface{}
	json.Unmarshal(result, &glmReq)

	messages := glmReq["messages"].([]interface{})
	if len(messages) != 2 {
		t.Fatalf("messages length = %d, want 2", len(messages))
	}

	sysMsg := messages[0].(map[string]interface{})
	if sysMsg["role"] != "system" {
		t.Errorf("messages[0].role = %v, want 'system'", sysMsg["role"])
	}
	if sysMsg["content"] != "You are helpful" {
		t.Errorf("messages[0].content = %v", sysMsg["content"])
	}
}

func TestTranslateClaudeToGLM_Tools(t *testing.T) {
	claudeReq := `{
		"tools": [{
			"name": "search",
			"description": "Search the web",
			"input_schema": {"type": "object", "properties": {"query": {"type": "string"}}}
		}],
		"messages": [{"role": "user", "content": "Search for Go tutorials"}]
	}`

	result := TranslateClaudeToGLM([]byte(claudeReq), "glm-4")

	var glmReq map[string]interface{}
	json.Unmarshal(result, &glmReq)

	tools := glmReq["tools"].([]interface{})
	if len(tools) != 1 {
		t.Fatalf("tools length = %d, want 1", len(tools))
	}

	tool := tools[0].(map[string]interface{})
	if tool["type"] != "function" {
		t.Errorf("tools[0].type = %v, want 'function'", tool["type"])
	}

	fn := tool["function"].(map[string]interface{})
	if fn["name"] != "search" {
		t.Errorf("function.name = %v, want 'search'", fn["name"])
	}
}

func TestTranslateClaudeToGLM_ModelSet(t *testing.T) {
	claudeReq := `{"messages": [{"role": "user", "content": "Hi"}]}`

	result := TranslateClaudeToGLM([]byte(claudeReq), "glm-4-flash")

	var glmReq map[string]interface{}
	json.Unmarshal(result, &glmReq)

	if glmReq["model"] != "glm-4-flash" {
		t.Errorf("model = %v, want 'glm-4-flash'", glmReq["model"])
	}
}

// Response Translation Tests

func TestTranslateGLMToClaude_TextContent(t *testing.T) {
	glmResp := `{
		"choices": [{
			"message": {"role": "assistant", "content": "Hello!"},
			"finish_reason": "stop"
		}],
		"usage": {"prompt_tokens": 5, "completion_tokens": 10},
		"model": "glm-4"
	}`

	result := TranslateGLMToClaude([]byte(glmResp))

	var claudeResp map[string]interface{}
	json.Unmarshal(result, &claudeResp)

	if claudeResp["role"] != "assistant" {
		t.Errorf("role = %v, want 'assistant'", claudeResp["role"])
	}

	content := claudeResp["content"].([]interface{})
	if len(content) != 1 {
		t.Fatalf("content length = %d, want 1", len(content))
	}

	textBlock := content[0].(map[string]interface{})
	if textBlock["type"] != "text" {
		t.Errorf("content[0].type = %v, want 'text'", textBlock["type"])
	}
	if textBlock["text"] != "Hello!" {
		t.Errorf("content[0].text = %v, want 'Hello!'", textBlock["text"])
	}

	if claudeResp["stop_reason"] != "end_turn" {
		t.Errorf("stop_reason = %v, want 'end_turn'", claudeResp["stop_reason"])
	}

	usage := claudeResp["usage"].(map[string]interface{})
	if usage["input_tokens"] != float64(5) {
		t.Errorf("usage.input_tokens = %v, want 5", usage["input_tokens"])
	}
}

func TestTranslateGLMToClaude_ToolCalls(t *testing.T) {
	glmResp := `{
		"choices": [{
			"message": {
				"role": "assistant",
				"tool_calls": [{
					"id": "call_abc",
					"type": "function",
					"function": {"name": "get_weather", "arguments": "{\"city\":\"Beijing\"}"}
				}]
			},
			"finish_reason": "tool_calls"
		}]
	}`

	result := TranslateGLMToClaude([]byte(glmResp))

	var claudeResp map[string]interface{}
	json.Unmarshal(result, &claudeResp)

	content := claudeResp["content"].([]interface{})
	if len(content) != 1 {
		t.Fatalf("content length = %d, want 1", len(content))
	}

	toolUse := content[0].(map[string]interface{})
	if toolUse["type"] != "tool_use" {
		t.Errorf("content[0].type = %v, want 'tool_use'", toolUse["type"])
	}
	if toolUse["id"] != "call_abc" {
		t.Errorf("content[0].id = %v, want 'call_abc'", toolUse["id"])
	}
	if toolUse["name"] != "get_weather" {
		t.Errorf("content[0].name = %v, want 'get_weather'", toolUse["name"])
	}

	input := toolUse["input"].(map[string]interface{})
	if input["city"] != "Beijing" {
		t.Errorf("input.city = %v, want 'Beijing'", input["city"])
	}

	if claudeResp["stop_reason"] != "tool_use" {
		t.Errorf("stop_reason = %v, want 'tool_use'", claudeResp["stop_reason"])
	}
}

func TestTranslateOpenAIToGLM(t *testing.T) {
	openaiReq := `{"model": "gpt-4", "messages": [{"role": "user", "content": "Hi"}]}`

	result := TranslateOpenAIToGLM([]byte(openaiReq), "glm-4")

	var glmReq map[string]interface{}
	json.Unmarshal(result, &glmReq)

	// Should just update model, keep rest same
	if glmReq["model"] != "glm-4" {
		t.Errorf("model = %v, want 'glm-4'", glmReq["model"])
	}
}
