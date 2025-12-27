package openai

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ClaudeToOpenAI converts Claude format request to OpenAI format
// Handles system messages, tool calling, tool results, and multimodal content
func ClaudeToOpenAI(payload []byte, model string) ([]byte, error) {
	result := string(payload)

	// Convert messages first (includes tool_result and image translation)
	result = convertMessages(payload, result)

	// Then prepend system message if exists
	result = prependSystemMessage(payload, result)

	// Convert tools
	result = convertTools(payload, result)

	// Set model
	result, _ = sjson.Set(result, "model", model)

	return []byte(result), nil
}

// prependSystemMessage prepends Claude system to OpenAI messages array
// This should be called after convertMessages to work on the already-converted messages
func prependSystemMessage(payload []byte, result string) string {
	systemResult := gjson.GetBytes(payload, "system")
	if !systemResult.Exists() {
		return result
	}

	// Get the already-converted messages from result
	messagesResult := gjson.Get(result, "messages")
	if !messagesResult.IsArray() {
		return result
	}

	// Prepend system message to messages array
	systemMsg := fmt.Sprintf(`{"role":"system","content":"%s"}`, systemResult.String())
	messages := messagesResult.Array()
	newMessages := `[` + systemMsg
	for _, msg := range messages {
		newMessages += `,` + msg.Raw
	}
	newMessages += `]`
	result, _ = sjson.SetRaw(result, "messages", newMessages)
	result, _ = sjson.Delete(result, "system")

	return result
}

// convertMessages handles message array translation including tool_result and images
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

// convertMessage translates a single message with content array handling
func convertMessage(msg gjson.Result) string {
	role := msg.Get("role").String()
	content := msg.Get("content")

	newMsg := `{"role":"","content":""}`
	newMsg, _ = sjson.Set(newMsg, "role", role)

	// Simple string content
	if content.Type == gjson.String {
		newMsg, _ = sjson.Set(newMsg, "content", content.String())
		return newMsg
	}

	// Array content - check for tool_result or multimodal
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
			// Text only: extract and concatenate
			textContent := ""
			for _, block := range blocks {
				if block.Get("type").String() == "text" {
					if textContent != "" {
						textContent += "\n"
					}
					textContent += block.Get("text").String()
				}
			}
			newMsg, _ = sjson.Set(newMsg, "content", textContent)
		}
	}

	return newMsg
}

// convertToolResultMessage converts Claude tool_result to OpenAI tool message
// Claude: {"role":"user","content":[{"type":"tool_result","tool_use_id":"call_xxx","content":"25°C"}]}
// OpenAI: {"role":"tool","tool_call_id":"call_xxx","content":"25°C"}
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

// translateContentPart converts Claude content block to OpenAI format
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
		// OpenAI: {"type":"image_url","image_url":{"url":"data:image/png;base64,iVBORw0..."}}
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

// convertTools handles Claude tools to OpenAI format
func convertTools(payload []byte, result string) string {
	toolsResult := gjson.GetBytes(payload, "tools")
	if !toolsResult.IsArray() {
		return result
	}

	functionsJSON := `[]`
	for _, tool := range toolsResult.Array() {
		functionJSON := `{}`

		// Map tool name
		if name := tool.Get("name"); name.Exists() {
			functionJSON, _ = sjson.Set(functionJSON, "name", name.String())
		}

		// Map description
		if desc := tool.Get("description"); desc.Exists() {
			functionJSON, _ = sjson.Set(functionJSON, "description", desc.String())
		}

		// Map input_schema to parameters
		if inputSchema := tool.Get("input_schema"); inputSchema.Exists() {
			functionJSON, _ = sjson.SetRaw(functionJSON, "parameters", inputSchema.Raw)
		}

		functionsJSON, _ = sjson.SetRaw(functionsJSON, "-1", functionJSON)
	}

	// OpenAI uses "tools" with type "function"
	openaiTools := `[]`
	for _, fn := range gjson.Parse(functionsJSON).Array() {
		toolWrapper := `{"type":"function"}`
		toolWrapper, _ = sjson.SetRaw(toolWrapper, "function", fn.Raw)
		openaiTools, _ = sjson.SetRaw(openaiTools, "-1", toolWrapper)
	}

	result, _ = sjson.SetRaw(result, "tools", openaiTools)

	return result
}
