# Translator Patterns Reference

Quick reference for translation logic across providers.

## Function Calling Patterns

### Tool Response Translation (Provider → Claude)

**OpenAI/GLM Response:**
```json
{
  "choices": [{
    "message": {
      "tool_calls": [{
        "id": "call_123",
        "type": "function",
        "function": {
          "name": "get_weather",
          "arguments": "{\"city\":\"Jakarta\"}"  // JSON string!
        }
      }]
    }
  }]
}
```

**Claude Format (Output):**
```json
{
  "content": [{
    "type": "tool_use",
    "id": "call_123",
    "name": "get_weather",
    "input": {"city": "Jakarta"}  // Parsed object!
  }],
  "stop_reason": "tool_use"
}
```

**Key Code:**
```go
// Parse arguments JSON string to object
argsStr := toolCall.Get("function.arguments").String()
var argsObj interface{}
json.Unmarshal([]byte(argsStr), &argsObj)
argsBytes, _ := json.Marshal(argsObj)
toolUse, _ = sjson.SetRaw(toolUse, "input", string(argsBytes))
```

### Tool Result Translation (Claude → Provider)

**Claude Request:**
```json
{
  "messages": [{
    "role": "user",
    "content": [{
      "type": "tool_result",
      "tool_use_id": "call_123",
      "content": "28°C"
    }]
  }]
}
```

**OpenAI/GLM Format (Output):**
```json
{
  "messages": [{
    "role": "tool",
    "tool_call_id": "call_123",
    "content": "28°C"
  }]
}
```

**Antigravity Format (Output):**
```json
{
  "request": {
    "contents": [{
      "role": "user",
      "parts": [{
        "functionResponse": {
          "id": "call_123",
          "name": "get_weather",
          "response": {"result": "28°C"}
        }
      }]
    }]
  }
}
```

**Key Code (OpenAI/GLM):**
```go
// Detect tool_result in content array
if blocks[0].Get("type").String() == "tool_result" {
    toolMsg := `{"role":"tool","tool_call_id":"","content":""}`
    toolMsg, _ = sjson.Set(toolMsg, "tool_call_id", block.Get("tool_use_id").String())
    toolMsg, _ = sjson.Set(toolMsg, "content", block.Get("content").String())
    return toolMsg
}
```

## Multimodal Patterns

### Image Translation (Claude → Provider)

**Claude Format:**
```json
{
  "type": "image",
  "source": {
    "type": "base64",
    "media_type": "image/png",
    "data": "iVBORw0KGgoAAAA..."
  }
}
```

**OpenAI/GLM Format (Output):**
```json
{
  "type": "image_url",
  "image_url": {
    "url": "data:image/png;base64,iVBORw0KGgoAAAA..."
  }
}
```

**Antigravity Format (Output):**
```json
{
  "inlineData": {
    "mimeType": "image/png",
    "data": "iVBORw0KGgoAAAA..."
  }
}
```

**Key Code (OpenAI/GLM):**
```go
func translateContentPart(block gjson.Result) string {
    if block.Get("type").String() == "image" {
        source := block.Get("source")
        if source.Get("type").String() == "base64" {
            mediaType := source.Get("media_type").String()
            data := source.Get("data").String()

            // Build data URL
            dataURL := fmt.Sprintf("data:%s;base64,%s", mediaType, data)
            part := `{"type":"image_url","image_url":{"url":""}}`
            part, _ = sjson.Set(part, "image_url.url", dataURL)
            return part
        }
    }
}
```

## Message Processing Flow

### Request Translation (Claude → Provider)

```
ClaudeToOpenAI/GLM
  ├─ convertSystemMessage     → Prepend to messages array
  ├─ convertMessages
  │   └─ convertMessage
  │       ├─ Check for tool_result → convertToolResultMessage
  │       ├─ Check for images → translateContentPart (multimodal array)
  │       └─ Text only → concatenate text blocks
  └─ convertTools            → Map input_schema to parameters
```

### Response Translation (Provider → Claude)

```
OpenAIToClaude/GLMToClaude
  └─ buildContentArray
      ├─ Add text block if content exists
      └─ For each tool_call → buildToolUseBlock
          └─ Parse arguments JSON string to object
```

## Edge Cases Handled

### Empty/Missing Fields
```go
// Safe navigation with gjson
if !content.Exists() || content.String() == "" {
    // Handle missing content
}

// Safe JSON parsing
if err := json.Unmarshal([]byte(argsStr), &argsObj); err == nil {
    // Successfully parsed
} else {
    // Parse failed - set empty object
    toolUse, _ = sjson.Set(toolUse, "input", map[string]interface{}{})
}
```

### Content Arrays
```go
// Handle both string and array content
if content.Type == gjson.String {
    // Simple string
} else if content.IsArray() {
    // Array of blocks
}
```

### Tool Result Content Formats
```go
// Handle string, array, or object content
if toolContent.Type == gjson.String {
    result, _ = sjson.Set(result, "content", toolContent.String())
} else if toolContent.IsArray() {
    // Extract text from blocks
} else if toolContent.IsObject() {
    // Serialize as JSON
    result, _ = sjson.SetRaw(result, "content", toolContent.Raw)
}
```

## gjson/sjson Quick Reference

### Reading with gjson
```go
// Get nested field
value := gjson.GetBytes(payload, "choices.0.message.content")

// Check existence
if value.Exists() { ... }

// Type checking
if value.Type == gjson.String { ... }
if value.IsArray() { ... }
if value.IsObject() { ... }

// Array iteration
for _, item := range value.Array() { ... }
```

### Writing with sjson
```go
// Set value
result, _ = sjson.Set(result, "key", "value")

// Set raw JSON
result, _ = sjson.SetRaw(result, "key", jsonString)

// Append to array
result, _ = sjson.SetRaw(result, "-1", item)

// Delete field
result, _ = sjson.Delete(result, "key")
```

## Testing Template

```go
// 1. Create test payload
payload := `{...}`

// 2. Translate
result, err := provider.TranslateX([]byte(payload))

// 3. Parse and verify
var output map[string]interface{}
json.Unmarshal(result, &output)
prettyJSON, _ := json.MarshalIndent(output, "", "  ")
fmt.Println(string(prettyJSON))
```
