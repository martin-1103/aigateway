# Format Translation

This document describes how AIGateway translates between different API formats for seamless provider integration.

## Overview

AIGateway supports automatic format translation between:
- **Claude format** (Anthropic Messages API)
- **Antigravity format** (Gemini API)
- **OpenAI format** (native, no translation)

## Claude to Antigravity Translation

**Implementation**: `services/translator.service.go::ClaudeToAntigravity()`

### Request Transformation

| Claude Format | Antigravity Format | Notes |
|---------------|-------------------|-------|
| `system: string` | `request.systemInstruction.parts[0].text` | System instruction |
| `messages[].role: "assistant"` | `request.contents[].role: "model"` | Role mapping |
| `messages[].role: "user"` | `request.contents[].role: "user"` | Unchanged |
| `messages[].content: string` | `request.contents[].parts[].text` | Content wrapping |
| `tools[].input_schema` | `request.tools[0].functionDeclarations[].parametersJsonSchema` | Tool schema |
| `max_tokens` | `request.generationConfig.maxOutputTokens` | Config mapping |
| `temperature` | `request.generationConfig.temperature` | Unchanged |
| `model` | `model` | Top-level field |

### Request Example

**Input (Claude format)**:
```json
{
  "model": "claude-sonnet-4-5",
  "system": "You are a helpful assistant",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "max_tokens": 1024,
  "temperature": 0.7
}
```

**Output (Antigravity format)**:
```json
{
  "model": "claude-sonnet-4-5",
  "request": {
    "systemInstruction": {
      "role": "user",
      "parts": [{"text": "You are a helpful assistant"}]
    },
    "contents": [
      {
        "role": "user",
        "parts": [{"text": "Hello"}]
      }
    ],
    "generationConfig": {
      "maxOutputTokens": 1024,
      "temperature": 0.7
    }
  }
}
```

### Message Content Transformation

**String content**:
```json
// Claude
{"role": "user", "content": "Hello"}

// Antigravity
{"role": "user", "parts": [{"text": "Hello"}]}
```

**Array content (multimodal)**:
```json
// Claude
{
  "role": "user",
  "content": [
    {"type": "text", "text": "Hello"},
    {"type": "image", "source": {...}}
  ]
}

// Antigravity
{
  "role": "user",
  "parts": [
    {"text": "Hello"},
    {"inlineData": {...}}
  ]
}
```

### Tool Definitions

**Claude format**:
```json
{
  "tools": [
    {
      "name": "get_weather",
      "description": "Get weather information",
      "input_schema": {
        "type": "object",
        "properties": {
          "location": {"type": "string"}
        },
        "required": ["location"]
      }
    }
  ]
}
```

**Antigravity format**:
```json
{
  "tools": [
    {
      "functionDeclarations": [
        {
          "name": "get_weather",
          "description": "Get weather information",
          "parametersJsonSchema": {
            "type": "object",
            "properties": {
              "location": {"type": "string"}
            },
            "required": ["location"]
          }
        }
      ]
    }
  ]
}
```

## Antigravity to Claude Translation

**Implementation**: `services/translator.service.go::AntigravityToClaude()`

### Response Transformation

| Antigravity Format | Claude Format | Notes |
|-------------------|---------------|-------|
| `response.candidates[0].content.role: "model"` | `role: "assistant"` | Role mapping |
| `response.candidates[0].content.parts[].text` | `content[].text` | Content extraction |
| `response.candidates[0].finishReason: "MAX_TOKENS"` | `stop_reason: "max_tokens"` | Reason mapping |
| `response.candidates[0].finishReason: "STOP"` | `stop_reason: "end_turn"` | Reason mapping |
| `response.usageMetadata.promptTokenCount` | `usage.input_tokens` | Token counting |
| `response.usageMetadata.candidatesTokenCount` | `usage.output_tokens` | Token counting |

### Response Example

**Input (Antigravity format)**:
```json
{
  "response": {
    "candidates": [
      {
        "content": {
          "role": "model",
          "parts": [{"text": "Hi there!"}]
        },
        "finishReason": "STOP"
      }
    ],
    "usageMetadata": {
      "promptTokenCount": 10,
      "candidatesTokenCount": 5
    }
  }
}
```

**Output (Claude format)**:
```json
{
  "role": "assistant",
  "content": [
    {"type": "text", "text": "Hi there!"}
  ],
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 10,
    "output_tokens": 5
  }
}
```

### Finish Reason Mapping

| Antigravity | Claude | Meaning |
|-------------|--------|---------|
| `STOP` | `end_turn` | Natural completion |
| `MAX_TOKENS` | `max_tokens` | Token limit reached |
| `SAFETY` | `end_turn` | Content filter triggered |
| `RECITATION` | `end_turn` | Recitation detected |

### Tool Call Responses

**Antigravity format**:
```json
{
  "candidates": [
    {
      "content": {
        "parts": [
          {
            "functionCall": {
              "name": "get_weather",
              "args": {"location": "San Francisco"}
            }
          }
        ]
      }
    }
  ]
}
```

**Claude format**:
```json
{
  "content": [
    {
      "type": "tool_use",
      "id": "toolu_123",
      "name": "get_weather",
      "input": {"location": "San Francisco"}
    }
  ]
}
```

## OpenAI Format (No Translation)

OpenAI format is used natively without translation.

**Request format**:
```json
{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "max_tokens": 1024,
  "temperature": 0.7
}
```

**Response format**:
```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 9,
    "total_tokens": 19
  }
}
```

## Translation Performance

### Overhead Analysis

**Request Translation**:
- Claude → Antigravity: ~0.5ms (JSON parsing + transformation)
- No translation (OpenAI): 0ms

**Response Translation**:
- Antigravity → Claude: ~0.5ms (JSON parsing + transformation)
- No translation (OpenAI): 0ms

**Total Overhead**: ~1ms per request for format translation

### Optimization

Current implementation uses:
- Direct JSON manipulation (no intermediate objects)
- Lazy parsing (only parse required fields)
- Zero-copy where possible

## Error Handling

### Translation Errors

**Invalid input format**:
```go
if model == "" {
    return nil, fmt.Errorf("model field is required")
}
```

**Missing required fields**:
```go
if len(messages) == 0 {
    return nil, fmt.Errorf("messages array cannot be empty")
}
```

**Malformed JSON**:
```go
if err := json.Unmarshal(payload, &req); err != nil {
    return nil, fmt.Errorf("invalid JSON: %w", err)
}
```

## Adding Translation for New Provider

To add format translation for a new provider:

### 1. Identify Source and Target Formats

**Example**: Adding Azure OpenAI support
- Source: Claude format
- Target: Azure OpenAI format (same as OpenAI)
- Decision: No translation needed

### 2. Implement Translation Functions

If translation needed, create in `services/translator.service.go`:

```go
func (s *TranslatorService) ClaudeToNewProvider(payload []byte, model string) ([]byte, error) {
    // Parse Claude format
    var claudeReq ClaudeRequest
    if err := json.Unmarshal(payload, &claudeReq); err != nil {
        return nil, err
    }

    // Transform to provider format
    providerReq := ProviderRequest{
        // Map fields
    }

    // Return JSON
    return json.Marshal(providerReq)
}

func (s *TranslatorService) NewProviderToClaude(payload []byte) ([]byte, error) {
    // Parse provider format
    // Transform to Claude format
    // Return JSON
}
```

### 3. Update Executor Service

Call translation functions in `services/executor.service.go`:

```go
switch providerID {
case "new_provider":
    translatedReq, err = s.translator.ClaudeToNewProvider(req.Payload, req.Model)
    // ...
    translatedResp, err = s.translator.NewProviderToClaude(resp.Payload)
}
```

## Testing Translation

### Unit Tests

```go
func TestClaudeToAntigravity(t *testing.T) {
    input := `{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"test"}]}`
    expected := // Expected Antigravity format

    result, err := translator.ClaudeToAntigravity([]byte(input), "claude-sonnet-4-5")
    assert.NoError(t, err)
    assert.JSONEq(t, expected, string(result))
}
```

### Integration Tests

Test with actual provider APIs to ensure compatibility.

## Related Documentation

- [Provider Overview](README.md) - All providers
- [Antigravity Provider](antigravity.md) - Format example
- [Architecture](../architecture/components.md) - Translator service
