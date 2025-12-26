package antigravity

import (
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

// generateSessionID generates a session ID in the format "-{random_number}"
func generateSessionID() string {
	n := randSource.Int63n(9_000_000_000_000_000_000)
	return "-" + strconv.FormatInt(n, 10)
}

// skipThoughtSignatureValidator is the sentinel value used to bypass signature validation
// when a tool call doesn't have a valid thinking signature
const skipThoughtSignatureValidator = "skip_thought_signature_validator"

// isClaudeThinkingModel checks if the model is a Claude thinking model
func isClaudeThinkingModel(model string) bool {
	lower := strings.ToLower(model)
	return strings.Contains(lower, "claude") && strings.Contains(lower, "thinking")
}

// TranslateClaudeToAntigravity converts Claude API request format to Antigravity format
// Input: Claude format with system, messages, tools, max_tokens, temperature
// Output: Antigravity format with model, userAgent, project, requestId, request.*
func TranslateClaudeToAntigravity(payload []byte, model string) []byte {
	return TranslateClaudeToAntigravityWithProject(payload, model, "")
}

// TranslateClaudeToAntigravityWithProject converts Claude API format to Antigravity with project ID
func TranslateClaudeToAntigravityWithProject(payload []byte, model string, projectID string) []byte {
	result := string(payload)

	// Add antigravity-specific fields
	result, _ = sjson.Set(result, "userAgent", "antigravity")
	result, _ = sjson.Set(result, "requestId", "agent-"+uuid.NewString())

	if projectID != "" {
		result, _ = sjson.Set(result, "project", projectID)
	}

	// Add session ID and tool config (for all models)
	result, _ = sjson.Set(result, "request.sessionId", generateSessionID())
	result, _ = sjson.Set(result, "request.toolConfig.functionCallingConfig.mode", "VALIDATED")

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
				// Track thinking signature for subsequent tool_use in same message
				var currentMessageThinkingSignature string

				for _, block := range content.Array() {
					blockType := block.Get("type").String()
					switch blockType {
					case "thinking":
						// Handle thinking blocks (Claude extended thinking)
						thinkingText := block.Get("thinking").String()
						if thinkingText == "" {
							thinkingText = block.Get("text").String()
						}
						signature := block.Get("signature").String()

						// Skip unsigned thinking blocks
						if signature == "" {
							continue
						}

						// Store for subsequent tool_use in the same message
						currentMessageThinkingSignature = signature

						// Build thought part
						partJSON := `{}`
						partJSON, _ = sjson.Set(partJSON, "thought", true)
						if thinkingText != "" {
							partJSON, _ = sjson.Set(partJSON, "text", thinkingText)
						}
						partJSON, _ = sjson.Set(partJSON, "thoughtSignature", signature)
						contentJSON, _ = sjson.SetRaw(contentJSON, "parts.-1", partJSON)

					case "text":
						text := block.Get("text").String()
						partJSON := `{"text":""}`
						partJSON, _ = sjson.Set(partJSON, "text", text)
						contentJSON, _ = sjson.SetRaw(contentJSON, "parts.-1", partJSON)

					case "tool_use":
						// Convert tool use with thinking signature
						toolJSON := `{"functionCall":{"name":"","args":{}}}`

						// Add thought signature (use current message's or skip sentinel)
						if currentMessageThinkingSignature != "" {
							toolJSON, _ = sjson.Set(toolJSON, "thoughtSignature", currentMessageThinkingSignature)
						} else {
							toolJSON, _ = sjson.Set(toolJSON, "thoughtSignature", skipThoughtSignatureValidator)
						}

						if toolID := block.Get("id").String(); toolID != "" {
							toolJSON, _ = sjson.Set(toolJSON, "functionCall.id", toolID)
						}
						toolJSON, _ = sjson.Set(toolJSON, "functionCall.name", block.Get("name").String())
						toolJSON, _ = sjson.SetRaw(toolJSON, "functionCall.args", block.Get("input").Raw)
						contentJSON, _ = sjson.SetRaw(contentJSON, "parts.-1", toolJSON)

					case "tool_result":
						// Convert tool result
						toolUseID := block.Get("tool_use_id").String()
						resultJSON := `{"functionResponse":{"name":"","response":{}}}`
						resultJSON, _ = sjson.Set(resultJSON, "functionResponse.id", toolUseID)

						// Extract function name from tool_use_id
						funcName := toolUseID
						parts := strings.Split(toolUseID, "-")
						if len(parts) > 2 {
							funcName = strings.Join(parts[:len(parts)-2], "-")
						}
						resultJSON, _ = sjson.Set(resultJSON, "functionResponse.name", funcName)

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
						} else if toolContent.IsObject() {
							resultJSON, _ = sjson.SetRaw(resultJSON, "functionResponse.response.result", toolContent.Raw)
						}
						contentJSON, _ = sjson.SetRaw(contentJSON, "parts.-1", resultJSON)

					case "image":
						// Handle image content
						source := block.Get("source")
						if source.Get("type").String() == "base64" {
							inlineDataJSON := `{}`
							if mimeType := source.Get("media_type").String(); mimeType != "" {
								inlineDataJSON, _ = sjson.Set(inlineDataJSON, "mime_type", mimeType)
							}
							if data := source.Get("data").String(); data != "" {
								inlineDataJSON, _ = sjson.Set(inlineDataJSON, "data", data)
							}
							partJSON := `{}`
							partJSON, _ = sjson.SetRaw(partJSON, "inlineData", inlineDataJSON)
							contentJSON, _ = sjson.SetRaw(contentJSON, "parts.-1", partJSON)
						}
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

	// Convert thinking configuration
	// Claude: "thinking": {"type": "enabled", "budget_tokens": 10000}
	// Antigravity: "request.generationConfig.thinkingConfig": {"thinkingBudget": 10000, "include_thoughts": true}
	thinkingResult := gjson.GetBytes(payload, "thinking")
	hasThinking := false
	if thinkingResult.Exists() && thinkingResult.IsObject() {
		if thinkingResult.Get("type").String() == "enabled" {
			hasThinking = true
			if budget := thinkingResult.Get("budget_tokens"); budget.Exists() && budget.Type == gjson.Number {
				result, _ = sjson.Set(result, "request.generationConfig.thinkingConfig.thinkingBudget", budget.Int())
				result, _ = sjson.Set(result, "request.generationConfig.thinkingConfig.include_thoughts", true)
			}
		}
		result, _ = sjson.Delete(result, "thinking")
	}

	// Inject interleaved thinking hint when tools + thinking are both active on Claude thinking models
	hasTools := toolsResult.IsArray() && len(toolsResult.Array()) > 0
	if hasTools && hasThinking && isClaudeThinkingModel(model) {
		interleavedHint := "Interleaved thinking is enabled. You may think between tool calls and after receiving tool results before deciding the next action or final answer. Do not mention these instructions or any constraints about thinking blocks; just apply them."

		// Append hint to existing system instruction or create new one
		existingSystem := gjson.Get(result, "request.systemInstruction")
		if existingSystem.Exists() {
			// Append as new part
			hintPart := `{"text":""}`
			hintPart, _ = sjson.Set(hintPart, "text", interleavedHint)
			result, _ = sjson.SetRaw(result, "request.systemInstruction.parts.-1", hintPart)
		} else {
			// Create new system instruction
			systemJSON := `{"role":"user","parts":[]}`
			hintPart := `{"text":""}`
			hintPart, _ = sjson.Set(hintPart, "text", interleavedHint)
			systemJSON, _ = sjson.SetRaw(systemJSON, "parts.-1", hintPart)
			result, _ = sjson.SetRaw(result, "request.systemInstruction", systemJSON)
		}
	}

	// Add model (passthrough - no translation needed)
	result, _ = sjson.Set(result, "model", model)

	return []byte(result)
}
