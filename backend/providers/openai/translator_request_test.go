package openai

import (
	"encoding/json"
	"testing"
)

func TestClaudeToOpenAI_ToolResult(t *testing.T) {
	claudeReq := `{
		"messages": [
			{"role": "user", "content": "What's the weather in Jakarta?"},
			{"role": "assistant", "content": [{"type": "tool_use", "id": "call_123", "name": "get_weather", "input": {"city": "Jakarta"}}]},
			{"role": "user", "content": [{"type": "tool_result", "tool_use_id": "call_123", "content": "25°C, sunny"}]}
		]
	}`

	result, err := ClaudeToOpenAI([]byte(claudeReq), "gpt-4")
	if err != nil {
		t.Fatalf("ClaudeToOpenAI() error = %v", err)
	}

	var openaiReq map[string]interface{}
	json.Unmarshal(result, &openaiReq)

	messages := openaiReq["messages"].([]interface{})
	if len(messages) != 3 {
		t.Fatalf("messages length = %d, want 3", len(messages))
	}

	// Third message should be tool role
	toolMsg := messages[2].(map[string]interface{})
	if toolMsg["role"] != "tool" {
		t.Errorf("messages[2].role = %v, want 'tool'", toolMsg["role"])
	}
	if toolMsg["tool_call_id"] != "call_123" {
		t.Errorf("messages[2].tool_call_id = %v, want 'call_123'", toolMsg["tool_call_id"])
	}
	if toolMsg["content"] != "25°C, sunny" {
		t.Errorf("messages[2].content = %v, want '25°C, sunny'", toolMsg["content"])
	}
}

func TestClaudeToOpenAI_ImageContent(t *testing.T) {
	claudeReq := `{
		"messages": [{
			"role": "user",
			"content": [
				{"type": "text", "text": "What's in this image?"},
				{"type": "image", "source": {"type": "base64", "media_type": "image/png", "data": "iVBORw0KGgo="}}
			]
		}]
	}`

	result, err := ClaudeToOpenAI([]byte(claudeReq), "gpt-4-vision")
	if err != nil {
		t.Fatalf("ClaudeToOpenAI() error = %v", err)
	}

	var openaiReq map[string]interface{}
	json.Unmarshal(result, &openaiReq)

	messages := openaiReq["messages"].([]interface{})
	msg := messages[0].(map[string]interface{})
	content := msg["content"].([]interface{})

	if len(content) != 2 {
		t.Fatalf("content length = %d, want 2", len(content))
	}

	// First should be text
	textPart := content[0].(map[string]interface{})
	if textPart["type"] != "text" {
		t.Errorf("content[0].type = %v, want 'text'", textPart["type"])
	}

	// Second should be image_url
	imagePart := content[1].(map[string]interface{})
	if imagePart["type"] != "image_url" {
		t.Errorf("content[1].type = %v, want 'image_url'", imagePart["type"])
	}

	imageURL := imagePart["image_url"].(map[string]interface{})
	expectedURL := "data:image/png;base64,iVBORw0KGgo="
	if imageURL["url"] != expectedURL {
		t.Errorf("content[1].image_url.url = %v, want %v", imageURL["url"], expectedURL)
	}
}

func TestClaudeToOpenAI_SystemMessage(t *testing.T) {
	claudeReq := `{
		"system": "You are a helpful assistant.",
		"messages": [
			{"role": "user", "content": "Hello"}
		]
	}`

	result, _ := ClaudeToOpenAI([]byte(claudeReq), "gpt-4")

	var openaiReq map[string]interface{}
	json.Unmarshal(result, &openaiReq)

	messages := openaiReq["messages"].([]interface{})
	if len(messages) != 2 {
		t.Fatalf("messages length = %d, want 2 (system + user)", len(messages))
	}

	// First message should be system
	sysMsg := messages[0].(map[string]interface{})
	if sysMsg["role"] != "system" {
		t.Errorf("messages[0].role = %v, want 'system'", sysMsg["role"])
	}

	// System field should be removed
	if _, exists := openaiReq["system"]; exists {
		t.Error("system field should be removed from request")
	}
}

func TestClaudeToOpenAI_Tools(t *testing.T) {
	claudeReq := `{
		"tools": [{
			"name": "get_weather",
			"description": "Get current weather",
			"input_schema": {
				"type": "object",
				"properties": {"city": {"type": "string"}},
				"required": ["city"]
			}
		}],
		"messages": [{"role": "user", "content": "Weather in Jakarta?"}]
	}`

	result, _ := ClaudeToOpenAI([]byte(claudeReq), "gpt-4")

	var openaiReq map[string]interface{}
	json.Unmarshal(result, &openaiReq)

	tools := openaiReq["tools"].([]interface{})
	if len(tools) != 1 {
		t.Fatalf("tools length = %d, want 1", len(tools))
	}

	tool := tools[0].(map[string]interface{})
	if tool["type"] != "function" {
		t.Errorf("tools[0].type = %v, want 'function'", tool["type"])
	}

	fn := tool["function"].(map[string]interface{})
	if fn["name"] != "get_weather" {
		t.Errorf("tools[0].function.name = %v, want 'get_weather'", fn["name"])
	}
	if fn["description"] != "Get current weather" {
		t.Errorf("tools[0].function.description = %v", fn["description"])
	}

	params := fn["parameters"].(map[string]interface{})
	if params["type"] != "object" {
		t.Errorf("parameters.type = %v, want 'object'", params["type"])
	}
}

func TestClaudeToOpenAI_ModelSet(t *testing.T) {
	claudeReq := `{"messages": [{"role": "user", "content": "Hello"}]}`

	result, _ := ClaudeToOpenAI([]byte(claudeReq), "gpt-4-turbo")

	var openaiReq map[string]interface{}
	json.Unmarshal(result, &openaiReq)

	if openaiReq["model"] != "gpt-4-turbo" {
		t.Errorf("model = %v, want 'gpt-4-turbo'", openaiReq["model"])
	}
}

func TestTranslateContentPart_URLImage(t *testing.T) {
	// Test URL-based image (not base64)
	claudeReq := `{
		"messages": [{
			"role": "user",
			"content": [
				{"type": "image", "source": {"type": "url", "url": "https://example.com/image.png"}}
			]
		}]
	}`

	result, _ := ClaudeToOpenAI([]byte(claudeReq), "gpt-4-vision")

	var openaiReq map[string]interface{}
	json.Unmarshal(result, &openaiReq)

	messages := openaiReq["messages"].([]interface{})
	msg := messages[0].(map[string]interface{})
	content := msg["content"].([]interface{})

	imagePart := content[0].(map[string]interface{})
	imageURL := imagePart["image_url"].(map[string]interface{})

	if imageURL["url"] != "https://example.com/image.png" {
		t.Errorf("image_url.url = %v, want URL", imageURL["url"])
	}
}
