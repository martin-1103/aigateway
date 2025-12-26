package glm

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// TranslateClaudeToGLM converts Claude format to GLM OpenAI-compatible format
func TranslateClaudeToGLM(payload []byte, model string) []byte {
	result := string(payload)

	result = convertSystem(payload, result)
	result = convertMessages(payload, result)
	result = convertTools(payload, result)

	if !gjson.GetBytes(payload, "stream").Exists() {
		result, _ = sjson.Set(result, "stream", false)
	}
	result, _ = sjson.Set(result, "model", model)
	return []byte(result)
}

func convertSystem(payload []byte, result string) string {
	systemResult := gjson.GetBytes(payload, "system")
	if !systemResult.Exists() {
		return result
	}

	systemContent := extractTextContent(systemResult)
	result, _ = sjson.Delete(result, "system")
	if systemContent == "" {
		return result
	}

	systemMsg := `{"role":"system","content":""}`
	systemMsg, _ = sjson.Set(systemMsg, "content", systemContent)

	messagesResult := gjson.GetBytes(payload, "messages")
	if messagesResult.IsArray() {
		newMessages := "[" + systemMsg + "]"
		for _, msg := range messagesResult.Array() {
			newMessages, _ = sjson.SetRaw(newMessages, "-1", msg.Raw)
		}
		result, _ = sjson.SetRaw(result, "messages", newMessages)
	}
	return result
}

func convertMessages(payload []byte, result string) string {
	messagesResult := gjson.GetBytes(payload, "messages")
	if !messagesResult.IsArray() {
		return result
	}

	newMessages := "[]"
	for _, msg := range messagesResult.Array() {
		newMessages, _ = sjson.SetRaw(newMessages, "-1", convertMessage(msg))
	}
	result, _ = sjson.SetRaw(result, "messages", newMessages)
	return result
}

func convertMessage(msg gjson.Result) string {
	role := msg.Get("role").String()
	content := msg.Get("content")

	newMsg := `{"role":"","content":""}`
	newMsg, _ = sjson.Set(newMsg, "role", role)

	if content.Type == gjson.String {
		newMsg, _ = sjson.Set(newMsg, "content", content.String())
		return newMsg
	}

	if content.IsArray() {
		textContent, toolCalls, toolCallID := processContentBlocks(content.Array())
		newMsg, _ = sjson.Set(newMsg, "content", textContent)
		if toolCalls != "[]" {
			newMsg, _ = sjson.SetRaw(newMsg, "tool_calls", toolCalls)
		}
		if toolCallID != "" {
			newMsg, _ = sjson.Set(newMsg, "tool_call_id", toolCallID)
			newMsg, _ = sjson.Set(newMsg, "role", "tool")
		}
	}
	return newMsg
}

func processContentBlocks(blocks []gjson.Result) (string, string, string) {
	textContent, toolCalls, toolCallID := "", "[]", ""

	for _, block := range blocks {
		switch block.Get("type").String() {
		case "text":
			if textContent != "" {
				textContent += "\n"
			}
			textContent += block.Get("text").String()
		case "tool_use":
			toolCall := `{"id":"","type":"function","function":{"name":"","arguments":""}}`
			toolCall, _ = sjson.Set(toolCall, "id", block.Get("id").String())
			toolCall, _ = sjson.Set(toolCall, "function.name", block.Get("name").String())
			toolCall, _ = sjson.Set(toolCall, "function.arguments", block.Get("input").Raw)
			toolCalls, _ = sjson.SetRaw(toolCalls, "-1", toolCall)
		case "tool_result":
			toolCallID = block.Get("tool_use_id").String()
			textContent = extractTextContent(block.Get("content"))
		}
	}
	return textContent, toolCalls, toolCallID
}

func convertTools(payload []byte, result string) string {
	toolsResult := gjson.GetBytes(payload, "tools")
	if !toolsResult.IsArray() {
		return result
	}

	newTools := "[]"
	for _, tool := range toolsResult.Array() {
		toolObj := `{"type":"function","function":{"name":"","description":"","parameters":{}}}`
		toolObj, _ = sjson.Set(toolObj, "function.name", tool.Get("name").String())
		toolObj, _ = sjson.Set(toolObj, "function.description", tool.Get("description").String())
		if inputSchema := tool.Get("input_schema"); inputSchema.Exists() {
			toolObj, _ = sjson.SetRaw(toolObj, "function.parameters", inputSchema.Raw)
		}
		newTools, _ = sjson.SetRaw(newTools, "-1", toolObj)
	}
	result, _ = sjson.SetRaw(result, "tools", newTools)
	return result
}

func extractTextContent(content gjson.Result) string {
	if content.Type == gjson.String {
		return content.String()
	}
	if content.IsArray() {
		for _, block := range content.Array() {
			if block.Get("type").String() == "text" {
				return block.Get("text").String()
			}
		}
	}
	return ""
}

// TranslateOpenAIToGLM converts OpenAI format to GLM format (minimal changes)
// GLM API is OpenAI-compatible, so mainly just ensure model is set
func TranslateOpenAIToGLM(payload []byte, model string) []byte {
	result := string(payload)
	result, _ = sjson.Set(result, "model", model)
	return []byte(result)
}

// TranslateGLMToClaude converts GLM response to Claude format
func TranslateGLMToClaude(payload []byte) []byte {
	result := `{"role":"assistant","content":[]}`
	choice := gjson.GetBytes(payload, "choices.0")
	if !choice.Exists() {
		return []byte(result)
	}

	message := choice.Get("message")
	role := message.Get("role").String()
	if role == "" {
		role = "assistant"
	}
	result, _ = sjson.Set(result, "role", role)

	// Add content and tool use blocks
	content := message.Get("content")
	contentIndex := 0
	if content.Exists() && content.String() != "" {
		textBlock := `{"type":"text","text":""}`
		textBlock, _ = sjson.Set(textBlock, "text", content.String())
		result, _ = sjson.SetRaw(result, "content.0", textBlock)
		contentIndex = 1
	}

	toolCalls := message.Get("tool_calls")
	if toolCalls.IsArray() {
		for _, toolCall := range toolCalls.Array() {
			toolUseBlock := `{"type":"tool_use","id":"","name":"","input":{}}`
			toolUseBlock, _ = sjson.Set(toolUseBlock, "id", toolCall.Get("id").String())
			toolUseBlock, _ = sjson.Set(toolUseBlock, "name", toolCall.Get("function.name").String())
			if args := toolCall.Get("function.arguments"); args.Exists() {
				toolUseBlock, _ = sjson.SetRaw(toolUseBlock, "input", args.Raw)
			}
			result, _ = sjson.SetRaw(result, "content."+string(rune(contentIndex)), toolUseBlock)
			contentIndex++
		}
	}

	// Add stop reason
	stopMap := map[string]string{"stop": "end_turn", "length": "max_tokens", "tool_calls": "tool_use", "function_call": "tool_use"}
	stopReason := stopMap[choice.Get("finish_reason").String()]
	if stopReason == "" {
		stopReason = "end_turn"
	}
	result, _ = sjson.Set(result, "stop_reason", stopReason)

	// Add usage and model
	usage := gjson.GetBytes(payload, "usage")
	if usage.Exists() {
		result, _ = sjson.Set(result, "usage.input_tokens", usage.Get("prompt_tokens").Int())
		result, _ = sjson.Set(result, "usage.output_tokens", usage.Get("completion_tokens").Int())
	}
	if model := gjson.GetBytes(payload, "model"); model.Exists() {
		result, _ = sjson.Set(result, "model", model.String())
	}

	return []byte(result)
}
