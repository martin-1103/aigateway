package glm

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// TranslateClaudeToGLM converts Claude format to GLM OpenAI-compatible format
// Handles system messages, tools, tool_result, and multimodal content
func TranslateClaudeToGLM(payload []byte, model string) []byte {
	result := string(payload)

	// Convert messages first (includes tool_result and image translation)
	result = convertMessages(payload, result)
	// Then prepend system message
	result = prependSystem(payload, result)
	result = convertTools(payload, result)

	if !gjson.GetBytes(payload, "stream").Exists() {
		result, _ = sjson.Set(result, "stream", false)
	}
	result, _ = sjson.Set(result, "model", model)
	return []byte(result)
}

// prependSystem prepends system message to already-converted messages array
func prependSystem(payload []byte, result string) string {
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

	// Get already-converted messages from result
	messagesResult := gjson.Get(result, "messages")
	if messagesResult.IsArray() {
		newMessages := "[" + systemMsg + "]"
		for _, msg := range messagesResult.Array() {
			newMessages, _ = sjson.SetRaw(newMessages, "-1", msg.Raw)
		}
		result, _ = sjson.SetRaw(result, "messages", newMessages)
	}
	return result
}

// convertMessages translates messages array including tool_result and images
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

// convertMessage translates a single message with content handling
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
		blocks := content.Array()

		// Check if this is a tool_result message
		if len(blocks) > 0 && blocks[0].Get("type").String() == "tool_result" {
			return convertToolResultMessage(blocks[0])
		}

		// Check for multimodal content (has images)
		hasImage := false
		for _, block := range blocks {
			if block.Get("type").String() == "image" {
				hasImage = true
				break
			}
		}

		if hasImage {
			// Multimodal: convert to OpenAI content array
			contentArray := "[]"
			for _, block := range blocks {
				contentArray, _ = sjson.SetRaw(contentArray, "-1", translateContentPart(block))
			}
			newMsg, _ = sjson.SetRaw(newMsg, "content", contentArray)
		} else {
			// Text only: process normally
			textContent, toolCalls, _ := processContentBlocks(blocks)
			newMsg, _ = sjson.Set(newMsg, "content", textContent)
			if toolCalls != "[]" {
				newMsg, _ = sjson.SetRaw(newMsg, "tool_calls", toolCalls)
			}
		}
	}
	return newMsg
}

// convertToolResultMessage converts Claude tool_result to GLM tool message
// Claude: {"role":"user","content":[{"type":"tool_result","tool_use_id":"call_xxx","content":"25°C"}]}
// GLM: {"role":"tool","tool_call_id":"call_xxx","content":"25°C"}
func convertToolResultMessage(block gjson.Result) string {
	toolMsg := `{"role":"tool","tool_call_id":"","content":""}`

	// Extract tool_use_id
	toolCallID := block.Get("tool_use_id").String()
	toolMsg, _ = sjson.Set(toolMsg, "tool_call_id", toolCallID)

	// Extract content
	toolContent := block.Get("content")
	if toolContent.Type == gjson.String {
		toolMsg, _ = sjson.Set(toolMsg, "content", toolContent.String())
	} else if toolContent.IsArray() {
		// Extract text from content array
		textContent := ""
		for _, contentBlock := range toolContent.Array() {
			if contentBlock.Get("type").String() == "text" {
				textContent += contentBlock.Get("text").String()
			}
		}
		toolMsg, _ = sjson.Set(toolMsg, "content", textContent)
	} else if toolContent.IsObject() {
		// Serialize object as JSON string
		toolMsg, _ = sjson.SetRaw(toolMsg, "content", toolContent.Raw)
	}

	return toolMsg
}

// translateContentPart converts Claude content block to GLM/OpenAI format
// Handles text and image types
func translateContentPart(block gjson.Result) string {
	blockType := block.Get("type").String()

	switch blockType {
	case "text":
		// Text block: {"type":"text","text":"..."}
		part := `{"type":"text","text":""}`
		part, _ = sjson.Set(part, "text", block.Get("text").String())
		return part

	case "image":
		// Image block: convert base64 to data URL
		// Claude: {"type":"image","source":{"type":"base64","media_type":"image/png","data":"iVBORw0..."}}
		// GLM: {"type":"image_url","image_url":{"url":"data:image/png;base64,iVBORw0..."}}
		source := block.Get("source")
		if source.Get("type").String() == "base64" {
			mediaType := source.Get("media_type").String()
			data := source.Get("data").String()

			dataURL := fmt.Sprintf("data:%s;base64,%s", mediaType, data)
			part := `{"type":"image_url","image_url":{"url":""}}`
			part, _ = sjson.Set(part, "image_url.url", dataURL)
			return part
		}

		// Fallback for URL-based images
		if source.Get("type").String() == "url" {
			url := source.Get("url").String()
			part := `{"type":"image_url","image_url":{"url":""}}`
			part, _ = sjson.Set(part, "image_url.url", url)
			return part
		}

	default:
		// Unknown type: return as text
		part := `{"type":"text","text":"[unsupported content]"}`
		return part
	}

	return `{"type":"text","text":""}`
}

// processContentBlocks extracts text and tool_use from content blocks
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

// convertTools handles Claude tools to GLM format
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

// extractTextContent extracts text from string or content blocks
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
