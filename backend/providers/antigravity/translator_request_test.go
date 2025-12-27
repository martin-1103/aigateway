package antigravity

import (
	"encoding/json"
	"testing"

	"github.com/tidwall/gjson"
)

func TestTranslateClaudeToAntigravity_ImageContent(t *testing.T) {
	claudeReq := `{
		"messages": [{
			"role": "user",
			"content": [
				{"type": "text", "text": "Describe this image"},
				{"type": "image", "source": {"type": "base64", "media_type": "image/png", "data": "iVBORw0KGgo="}}
			]
		}]
	}`

	result := TranslateClaudeToAntigravity([]byte(claudeReq), "gemini-pro-vision")

	// Check request.contents structure
	contents := gjson.GetBytes(result, "request.contents")
	if !contents.IsArray() {
		t.Fatal("request.contents should be array")
	}

	content := contents.Array()[0]
	parts := content.Get("parts")
	if !parts.IsArray() || len(parts.Array()) != 2 {
		t.Fatalf("parts length = %d, want 2", len(parts.Array()))
	}

	// First part should be text
	textPart := parts.Array()[0]
	if textPart.Get("text").String() != "Describe this image" {
		t.Errorf("parts[0].text = %v", textPart.Get("text").String())
	}

	// Second part should be inlineData
	imagePart := parts.Array()[1]
	inlineData := imagePart.Get("inlineData")
	if !inlineData.Exists() {
		t.Fatal("parts[1].inlineData should exist")
	}
	if inlineData.Get("mime_type").String() != "image/png" {
		t.Errorf("inlineData.mime_type = %v, want 'image/png'", inlineData.Get("mime_type").String())
	}
	if inlineData.Get("data").String() != "iVBORw0KGgo=" {
		t.Errorf("inlineData.data = %v", inlineData.Get("data").String())
	}
}

func TestTranslateClaudeToAntigravity_ToolResult(t *testing.T) {
	claudeReq := `{
		"messages": [{
			"role": "user",
			"content": [{"type": "tool_result", "tool_use_id": "call_123", "content": "Weather is 25°C"}]
		}]
	}`

	result := TranslateClaudeToAntigravity([]byte(claudeReq), "gemini-pro")

	contents := gjson.GetBytes(result, "request.contents")
	content := contents.Array()[0]
	parts := content.Get("parts")

	funcResp := parts.Array()[0].Get("functionResponse")
	if !funcResp.Exists() {
		t.Fatal("functionResponse should exist")
	}

	if funcResp.Get("id").String() != "call_123" {
		t.Errorf("functionResponse.id = %v, want 'call_123'", funcResp.Get("id").String())
	}

	response := funcResp.Get("response.result").String()
	if response != "Weather is 25°C" {
		t.Errorf("functionResponse.response.result = %v", response)
	}
}

func TestTranslateClaudeToAntigravity_ToolUse(t *testing.T) {
	claudeReq := `{
		"messages": [{
			"role": "assistant",
			"content": [{"type": "tool_use", "id": "call_abc", "name": "get_weather", "input": {"city": "Jakarta"}}]
		}]
	}`

	result := TranslateClaudeToAntigravity([]byte(claudeReq), "gemini-pro")

	contents := gjson.GetBytes(result, "request.contents")
	content := contents.Array()[0]

	// Role should be mapped to "model"
	if content.Get("role").String() != "model" {
		t.Errorf("role = %v, want 'model'", content.Get("role").String())
	}

	parts := content.Get("parts")
	funcCall := parts.Array()[0].Get("functionCall")
	if !funcCall.Exists() {
		t.Fatal("functionCall should exist")
	}

	if funcCall.Get("id").String() != "call_abc" {
		t.Errorf("functionCall.id = %v", funcCall.Get("id").String())
	}
	if funcCall.Get("name").String() != "get_weather" {
		t.Errorf("functionCall.name = %v", funcCall.Get("name").String())
	}

	args := funcCall.Get("args")
	if args.Get("city").String() != "Jakarta" {
		t.Errorf("args.city = %v", args.Get("city").String())
	}
}

func TestTranslateClaudeToAntigravity_SystemInstruction(t *testing.T) {
	claudeReq := `{
		"system": "You are a helpful assistant.",
		"messages": [{"role": "user", "content": "Hello"}]
	}`

	result := TranslateClaudeToAntigravity([]byte(claudeReq), "gemini-pro")

	sysInstruction := gjson.GetBytes(result, "request.systemInstruction")
	if !sysInstruction.Exists() {
		t.Fatal("request.systemInstruction should exist")
	}

	if sysInstruction.Get("role").String() != "user" {
		t.Errorf("systemInstruction.role = %v, want 'user'", sysInstruction.Get("role").String())
	}

	parts := sysInstruction.Get("parts")
	if parts.Array()[0].Get("text").String() != "You are a helpful assistant." {
		t.Errorf("systemInstruction.parts[0].text = %v", parts.Array()[0].Get("text").String())
	}

	// Original system field should be removed
	if gjson.GetBytes(result, "system").Exists() {
		t.Error("system field should be removed")
	}
}

func TestTranslateClaudeToAntigravity_Tools(t *testing.T) {
	claudeReq := `{
		"tools": [{
			"name": "get_weather",
			"description": "Get weather info",
			"input_schema": {"type": "object", "properties": {"city": {"type": "string"}}}
		}],
		"messages": [{"role": "user", "content": "Weather?"}]
	}`

	result := TranslateClaudeToAntigravity([]byte(claudeReq), "gemini-pro")

	tools := gjson.GetBytes(result, "request.tools")
	if !tools.IsArray() {
		t.Fatal("request.tools should be array")
	}

	funcDecls := tools.Array()[0].Get("functionDeclarations")
	if !funcDecls.IsArray() {
		t.Fatal("functionDeclarations should be array")
	}

	funcDecl := funcDecls.Array()[0]
	if funcDecl.Get("name").String() != "get_weather" {
		t.Errorf("functionDeclarations[0].name = %v", funcDecl.Get("name").String())
	}
	if funcDecl.Get("description").String() != "Get weather info" {
		t.Errorf("functionDeclarations[0].description = %v", funcDecl.Get("description").String())
	}

	params := funcDecl.Get("parametersJsonSchema")
	if !params.Exists() {
		t.Error("parametersJsonSchema should exist")
	}
}

func TestTranslateClaudeToAntigravity_GenerationConfig(t *testing.T) {
	claudeReq := `{
		"max_tokens": 1024,
		"temperature": 0.7,
		"top_p": 0.9,
		"top_k": 40,
		"stop_sequences": ["END", "STOP"],
		"messages": [{"role": "user", "content": "Hello"}]
	}`

	result := TranslateClaudeToAntigravity([]byte(claudeReq), "gemini-pro")

	genConfig := gjson.GetBytes(result, "request.generationConfig")

	if genConfig.Get("maxOutputTokens").Int() != 1024 {
		t.Errorf("maxOutputTokens = %v, want 1024", genConfig.Get("maxOutputTokens").Int())
	}
	if genConfig.Get("temperature").Float() != 0.7 {
		t.Errorf("temperature = %v, want 0.7", genConfig.Get("temperature").Float())
	}
	if genConfig.Get("topP").Float() != 0.9 {
		t.Errorf("topP = %v, want 0.9", genConfig.Get("topP").Float())
	}
	if genConfig.Get("topK").Int() != 40 {
		t.Errorf("topK = %v, want 40", genConfig.Get("topK").Int())
	}

	stopSeqs := genConfig.Get("stopSequences")
	if !stopSeqs.IsArray() || len(stopSeqs.Array()) != 2 {
		t.Errorf("stopSequences length = %v, want 2", len(stopSeqs.Array()))
	}
}

func TestTranslateClaudeToAntigravity_ModelPassthrough(t *testing.T) {
	claudeReq := `{"messages": [{"role": "user", "content": "Hi"}]}`

	result := TranslateClaudeToAntigravity([]byte(claudeReq), "claude-sonnet-4-20250514")

	model := gjson.GetBytes(result, "model").String()
	if model != "claude-sonnet-4-20250514" {
		t.Errorf("model = %v, want 'claude-sonnet-4-20250514'", model)
	}
}

func TestTranslateClaudeToAntigravityWithProject(t *testing.T) {
	claudeReq := `{"messages": [{"role": "user", "content": "Hi"}]}`

	result := TranslateClaudeToAntigravityWithProject([]byte(claudeReq), "gemini-pro", "my-project-123")

	project := gjson.GetBytes(result, "project").String()
	if project != "my-project-123" {
		t.Errorf("project = %v, want 'my-project-123'", project)
	}
}

func TestTranslateClaudeToAntigravity_ThinkingConfig(t *testing.T) {
	claudeReq := `{
		"thinking": {"type": "enabled", "budget_tokens": 10000},
		"messages": [{"role": "user", "content": "Think about this"}]
	}`

	result := TranslateClaudeToAntigravity([]byte(claudeReq), "claude-thinking")

	thinkingConfig := gjson.GetBytes(result, "request.generationConfig.thinkingConfig")
	if !thinkingConfig.Exists() {
		t.Fatal("thinkingConfig should exist")
	}

	if thinkingConfig.Get("thinkingBudget").Int() != 10000 {
		t.Errorf("thinkingBudget = %v, want 10000", thinkingConfig.Get("thinkingBudget").Int())
	}
	if thinkingConfig.Get("include_thoughts").Bool() != true {
		t.Error("include_thoughts should be true")
	}
}

func TestIsClaudeThinkingModel(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{"claude-thinking", true},
		{"claude-sonnet-4-thinking", true},
		{"CLAUDE-THINKING-MODEL", true},
		{"claude-sonnet-4", false},
		{"gpt-4", false},
		{"gemini-pro", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got := isClaudeThinkingModel(tt.model)
			if got != tt.want {
				t.Errorf("isClaudeThinkingModel(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}

func TestGenerateSessionID(t *testing.T) {
	id1 := generateSessionID()
	id2 := generateSessionID()

	// Should start with "-"
	if id1[0] != '-' {
		t.Error("Session ID should start with '-'")
	}

	// Should be different each time
	if id1 == id2 {
		t.Error("Session IDs should be unique")
	}
}

func TestTranslateClaudeToAntigravity_AntigravityFields(t *testing.T) {
	claudeReq := `{"messages": [{"role": "user", "content": "Hi"}]}`

	result := TranslateClaudeToAntigravity([]byte(claudeReq), "gemini-pro")

	var parsed map[string]interface{}
	json.Unmarshal(result, &parsed)

	// Check antigravity-specific fields
	if parsed["userAgent"] != "antigravity" {
		t.Errorf("userAgent = %v, want 'antigravity'", parsed["userAgent"])
	}

	requestID, ok := parsed["requestId"].(string)
	if !ok || requestID == "" {
		t.Error("requestId should be set")
	}
	if len(requestID) < 10 {
		t.Error("requestId should be a UUID-like string")
	}

	// Check session ID
	request := parsed["request"].(map[string]interface{})
	sessionID, ok := request["sessionId"].(string)
	if !ok || sessionID == "" {
		t.Error("request.sessionId should be set")
	}
	if sessionID[0] != '-' {
		t.Error("sessionId should start with '-'")
	}
}
