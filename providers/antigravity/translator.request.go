package antigravity

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// TranslateClaudeToAntigravity converts Claude API request format to Antigravity format
// Input: Claude format with system, messages, tools, max_tokens, temperature
// Output: Antigravity format with request.systemInstruction, request.contents, request.tools, request.generationConfig
func TranslateClaudeToAntigravity(payload []byte, model string) []byte {
	result := string(payload)

	// Convert system instruction
	// Claude: "system": "text" or {"type": "text", "text": "..."}
	// Antigravity: "request.systemInstruction": {"role": "user", "parts": [{"text": "..."}]}
	systemResult := gjson.GetBytes(payload, "system")
	if systemResult.Exists() {
		systemJSON := `{"role":"user","parts":[]}`
		if systemResult.Type == gjson.String {
			systemJSON, _ = sjson.Set(systemJSON, "parts.0.text", systemResult.String())
		} else if systemResult.IsArray() {
			// Handle array of content blocks
			for _, block := range systemResult.Array() {
				if block.Get("type").String() == "text" {
					text := block.Get("text").String()
					systemJSON, _ = sjson.Set(systemJSON, "parts.0.text", text)
					break
				}
			}
		}
		result, _ = sjson.SetRaw(result, "request.systemInstruction", systemJSON)
		result, _ = sjson.Delete(result, "system")
	}

	// Convert messages
	// Claude: "messages": [{"role": "user|assistant", "content": "text" or [blocks]}]
	// Antigravity: "request.contents": [{"role": "user|model", "parts": [{"text": "..."}]}]
	messagesResult := gjson.GetBytes(payload, "messages")
	if messagesResult.IsArray() {
		contentsJSON := "[]"
		for _, msg := range messagesResult.Array() {
			role := msg.Get("role").String()
			// Map assistant to model
			if role == "assistant" {
				role = "model"
			}

			contentJSON := `{"role":"","parts":[]}`
			contentJSON, _ = sjson.Set(contentJSON, "role", role)

			content := msg.Get("content")
			if content.Type == gjson.String {
				// Simple string content
				partJSON := `{"text":""}`
				partJSON, _ = sjson.Set(partJSON, "text", content.String())
				contentJSON, _ = sjson.SetRaw(contentJSON, "parts.-1", partJSON)
			} else if content.IsArray() {
				// Array of content blocks
				for _, block := range content.Array() {
					blockType := block.Get("type").String()
					switch blockType {
					case "text":
						text := block.Get("text").String()
						partJSON := `{"text":""}`
						partJSON, _ = sjson.Set(partJSON, "text", text)
						contentJSON, _ = sjson.SetRaw(contentJSON, "parts.-1", partJSON)
					case "tool_use":
						// Convert tool use
						toolJSON := `{"functionCall":{"name":"","args":{}}}`
						toolJSON, _ = sjson.Set(toolJSON, "functionCall.name", block.Get("name").String())
						toolJSON, _ = sjson.SetRaw(toolJSON, "functionCall.args", block.Get("input").Raw)
						contentJSON, _ = sjson.SetRaw(contentJSON, "parts.-1", toolJSON)
					case "tool_result":
						// Convert tool result
						resultJSON := `{"functionResponse":{"name":"","response":{}}}`
						resultJSON, _ = sjson.Set(resultJSON, "functionResponse.name", block.Get("tool_use_id").String())

						// Handle content in tool_result
						toolContent := block.Get("content")
						if toolContent.Type == gjson.String {
							resultJSON, _ = sjson.Set(resultJSON, "functionResponse.response.result", toolContent.String())
						} else if toolContent.IsArray() {
							// Extract text from content blocks
							for _, contentBlock := range toolContent.Array() {
								if contentBlock.Get("type").String() == "text" {
									resultJSON, _ = sjson.Set(resultJSON, "functionResponse.response.result", contentBlock.Get("text").String())
									break
								}
							}
						}
						contentJSON, _ = sjson.SetRaw(contentJSON, "parts.-1", resultJSON)
					}
				}
			}

			contentsJSON, _ = sjson.SetRaw(contentsJSON, "-1", contentJSON)
		}
		result, _ = sjson.SetRaw(result, "request.contents", contentsJSON)
		result, _ = sjson.Delete(result, "messages")
	}

	// Convert tools
	// Claude: "tools": [{"name": "...", "description": "...", "input_schema": {...}}]
	// Antigravity: "request.tools": [{"functionDeclarations": [{"name": "...", "description": "...", "parametersJsonSchema": {...}}]}]
	toolsResult := gjson.GetBytes(payload, "tools")
	if toolsResult.IsArray() {
		toolsJSON := `[{"functionDeclarations":[]}]`
		for _, tool := range toolsResult.Array() {
			inputSchema := tool.Get("input_schema")
			if inputSchema.Exists() {
				toolJSON := tool.Raw
				// Remove input_schema and add parametersJsonSchema
				toolJSON, _ = sjson.Delete(toolJSON, "input_schema")
				toolJSON, _ = sjson.SetRaw(toolJSON, "parametersJsonSchema", inputSchema.Raw)
				toolsJSON, _ = sjson.SetRaw(toolsJSON, "0.functionDeclarations.-1", toolJSON)
			}
		}
		result, _ = sjson.SetRaw(result, "request.tools", toolsJSON)
		result, _ = sjson.Delete(result, "tools")
	}

	// Convert max_tokens
	// Claude: "max_tokens": 1024
	// Antigravity: "request.generationConfig.maxOutputTokens": 1024
	if maxTokens := gjson.GetBytes(payload, "max_tokens"); maxTokens.Exists() {
		result, _ = sjson.Set(result, "request.generationConfig.maxOutputTokens", maxTokens.Int())
		result, _ = sjson.Delete(result, "max_tokens")
	}

	// Convert temperature
	// Claude: "temperature": 0.7
	// Antigravity: "request.generationConfig.temperature": 0.7
	if temp := gjson.GetBytes(payload, "temperature"); temp.Exists() {
		result, _ = sjson.Set(result, "request.generationConfig.temperature", temp.Float())
		result, _ = sjson.Delete(result, "temperature")
	}

	// Convert top_p
	if topP := gjson.GetBytes(payload, "top_p"); topP.Exists() {
		result, _ = sjson.Set(result, "request.generationConfig.topP", topP.Float())
		result, _ = sjson.Delete(result, "top_p")
	}

	// Convert top_k
	if topK := gjson.GetBytes(payload, "top_k"); topK.Exists() {
		result, _ = sjson.Set(result, "request.generationConfig.topK", topK.Int())
		result, _ = sjson.Delete(result, "top_k")
	}

	// Convert stop sequences
	if stopSeq := gjson.GetBytes(payload, "stop_sequences"); stopSeq.IsArray() {
		stopJSON := "[]"
		for i, seq := range stopSeq.Array() {
			stopJSON, _ = sjson.Set(stopJSON, string(rune(i)), seq.String())
		}
		result, _ = sjson.SetRaw(result, "request.generationConfig.stopSequences", stopJSON)
		result, _ = sjson.Delete(result, "stop_sequences")
	}

	// Add model
	result, _ = sjson.Set(result, "model", model)

	return []byte(result)
}
