# Multi-Feature Implementation Design

**Date:** 2025-12-27
**Features:** OAuth Flow, Streaming, Function Calling, Multimodal

## Overview

Implement 4 features in parallel to enhance aigateway capabilities:
1. OAuth Flow - Full authorization code flow with PKCE
2. Streaming - SSE responses for all providers
3. Function Calling - Complete tool_use roundtrip
4. Multimodal - Image support for OpenAI/GLM

---

## 1. OAuth Flow

### Endpoints

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `POST` | `/api/v1/oauth/init` | Start OAuth, get auth URL |
| `GET` | `/api/v1/oauth/callback` | Auto flow - provider redirect |
| `POST` | `/api/v1/oauth/exchange` | Manual flow - paste URL |
| `GET` | `/api/v1/oauth/providers` | List OAuth providers |
| `POST` | `/api/v1/oauth/refresh` | Manual token refresh |

### Flow Types

**Automatic Flow:**
```
FE call /init → open popup → user consent → provider redirect to /callback
→ BE exchange token → HTML closes popup → FE refresh
```

**Manual Flow:**
```
FE call /init → open popup → user consent → user copy callback URL
→ paste in FE → FE call /exchange → BE parse & exchange token
```

### Init Endpoint

**Request:**
```json
{
  "provider": "antigravity",
  "account_name": "My Google Account",
  "flow_type": "auto"
}
```

**Response:**
```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?...",
  "state": "abc123",
  "flow_type": "auto",
  "expires_at": "2025-12-27T10:30:00Z"
}
```

### Callback Endpoint (Auto)

**Request:** `GET /api/v1/oauth/callback?code=AUTH_CODE&state=abc123`

**Response:** HTML that closes popup and notifies parent window

### Exchange Endpoint (Manual)

**Request:**
```json
{
  "callback_url": "http://localhost:1455/callback?code=AUTH_CODE&state=abc123"
}
```

**Response:**
```json
{
  "success": true,
  "account": {
    "id": 123,
    "name": "My Google Account",
    "provider": "antigravity",
    "email": "user@gmail.com"
  }
}
```

### Redis Storage

Key: `oauth:session:{state}`
Value: `{provider, redirect_uri, code_verifier, account_name}`
TTL: 10 minutes

### Provider OAuth Config

| Provider | Auth URL | Token URL | Client ID Source |
|----------|----------|-----------|------------------|
| Antigravity | accounts.google.com | oauth2.googleapis.com | Config |
| OpenAI Codex | auth.openai.com | auth.openai.com | Hardcoded |
| Claude | claude.ai | console.anthropic.com | Hardcoded |

### New Files

```
handlers/oauth.handler.go
services/oauth.flow.service.go
auth/pkce/pkce.go
auth/claude/oauth.go
auth/codex/oauth.go
```

---

## 2. Streaming

### Architecture

```
HTTP Request (stream: true)
    ↓
ProxyHandler.handleStreamingRequest()
    ↓
Set headers: text/event-stream
    ↓
Provider.ExecuteStream(ctx, req)
    ↓
StreamForwarder.Forward(dataCh, errCh)
    ↓
Translate chunks → Write → Flush
```

### Provider Interface Changes

```go
type Provider interface {
    Execute(ctx, req) (*ExecuteResponse, error)
    ExecuteStream(ctx, req) (*StreamResponse, error)  // NEW
    SupportsStreaming() bool                          // NEW
    GetName() string
}

type StreamResponse struct {
    StatusCode int
    Headers    map[string]string
    DataCh     <-chan []byte
    ErrCh      <-chan error
    Done       <-chan struct{}
}
```

### StreamForwarder Component

```go
type StreamForwarder struct {
    writer     http.ResponseWriter
    flusher    http.Flusher
    translator func([]byte) []byte
    done       chan struct{}
}

func (sf *StreamForwarder) Forward(dataCh <-chan []byte, errCh <-chan error) error
```

### SSE Output Format (Claude Compatible)

```
event: message_start
data: {"type":"message_start","message":{...}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}

event: message_stop
data: {"type":"message_stop"}
```

### New/Modified Files

```
providers/provider.go              (interface changes)
providers/stream_forwarder.go      (NEW)
handlers/proxy.handler.go          (add streaming handler)
providers/openai/executor.go       (add ExecuteStream)
providers/glm/executor.go          (add ExecuteStream)
providers/antigravity/executor.go  (adapt existing)
```

---

## 3. Function Calling

### Problem

- Response: `tool_calls` not properly translated to `tool_use`
- Request: `tool_result` not translated to provider format

### Solution: Enhance Translators

**Response Translation (tool_calls → tool_use):**

```go
// OpenAI response
{"choices":[{"message":{"tool_calls":[{"id":"call_xxx","function":{"name":"get_weather","arguments":"{\"city\":\"Jakarta\"}"}}]}}]}

// Translate to Claude format
{"content":[{"type":"tool_use","id":"call_xxx","name":"get_weather","input":{"city":"Jakarta"}}]}
```

**Request Translation (tool_result → provider format):**

```go
// Claude request
{"messages":[{"role":"user","content":[{"type":"tool_result","tool_use_id":"call_xxx","content":"25°C"}]}]}

// OpenAI format
{"messages":[{"role":"tool","tool_call_id":"call_xxx","content":"25°C"}]}

// Antigravity format
{"contents":[{"role":"function","parts":[{"functionResponse":{"name":"get_weather","response":{"result":"25°C"}}}]}]}
```

### Modified Files

```
providers/openai/translator.response.go
providers/openai/translator.request.go
providers/antigravity/translator.response.go
providers/antigravity/translator.request.go
providers/glm/translator.response.go
providers/glm/translator.request.go
```

---

## 4. Multimodal

### Problem

- OpenAI: No image translation
- GLM: No image translation
- Antigravity: Exists, verify complete

### Solution: translateContentPart()

**Claude → OpenAI:**

```go
// Claude format
{"type":"image","source":{"type":"base64","media_type":"image/png","data":"iVBORw0..."}}

// OpenAI format
{"type":"image_url","image_url":{"url":"data:image/png;base64,iVBORw0..."}}
```

**Claude → Antigravity:**

```go
// Claude format
{"type":"image","source":{"type":"base64","media_type":"image/png","data":"iVBORw0..."}}

// Antigravity format
{"inlineData":{"mimeType":"image/png","data":"iVBORw0..."}}
```

### Supported Formats

| Provider | Formats | Max Size |
|----------|---------|----------|
| OpenAI (GPT-4V) | PNG, JPEG, GIF, WebP | 20MB |
| Antigravity (Gemini) | PNG, JPEG, WebP | 20MB |
| GLM-4V | PNG, JPEG | 10MB |

### Modified Files

```
providers/openai/translator.request.go
providers/glm/translator.request.go
providers/antigravity/translator.request.go
```

---

## Implementation Summary

| Feature | New Files | Modified Files | Complexity |
|---------|-----------|----------------|------------|
| OAuth | 5 | 1 | Medium-High |
| Streaming | 1 | 5 | Medium |
| Function Calling | 0 | 6 | Low-Medium |
| Multimodal | 0 | 3 | Low |

### Dependency

All 4 features are independent and can be implemented in parallel.

```
OAuth ──────────┐
Streaming ──────┼──→ Integration Testing
Function Call ──┤
Multimodal ─────┘
```
