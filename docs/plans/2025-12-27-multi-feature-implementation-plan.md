# Implementation Plan: Multi-Feature

**Date:** 2025-12-27
**Design Doc:** 2025-12-27-multi-feature-design.md

## Execution Strategy

4 features implemented in parallel by separate workers, then integrated.

---

## Phase 1: Foundation (Sequential)

### Task 1.1: PKCE Utility
**Files:** `auth/pkce/pkce.go`

```go
// Functions needed:
GeneratePKCECodes() (*PKCECodes, error)
  - Generate 96 random bytes → code_verifier (128 chars base64url)
  - SHA256(verifier) → code_challenge (base64url)

type PKCECodes struct {
    CodeVerifier  string
    CodeChallenge string
}
```

### Task 1.2: Provider Interface Update
**Files:** `providers/provider.go`

```go
// Add to Provider interface:
ExecuteStream(ctx context.Context, req *ExecuteRequest) (*StreamResponse, error)
SupportsStreaming() bool

// Add new type:
type StreamResponse struct {
    StatusCode int
    Headers    map[string]string
    DataCh     <-chan []byte
    ErrCh      <-chan error
    Done       <-chan struct{}
}
```

---

## Phase 2: Parallel Implementation

### Stream A: OAuth (3 tasks)

#### Task A.1: OAuth Handler
**Files:** `handlers/oauth.handler.go`

```go
type OAuthHandler struct {
    oauthFlowService *services.OAuthFlowService
    accountRepo      *repositories.AccountRepository
}

// Endpoints:
POST /api/v1/oauth/init      → InitOAuth()
GET  /api/v1/oauth/callback  → HandleCallback()
POST /api/v1/oauth/exchange  → ExchangeCode()
GET  /api/v1/oauth/providers → ListProviders()
POST /api/v1/oauth/refresh   → RefreshToken()
```

#### Task A.2: OAuth Flow Service
**Files:** `services/oauth.flow.service.go`

```go
type OAuthFlowService struct {
    redis       *redis.Client
    accountRepo *repositories.AccountRepository
    httpClient  *http.Client
}

// Methods:
InitFlow(provider, accountName, flowType) (*InitResponse, error)
  - Generate state + PKCE
  - Store in Redis with TTL 10min
  - Build provider-specific auth URL
  - Return auth_url, state, expires_at

ExchangeCode(callbackURL string) (*Account, error)
  - Parse URL → extract code, state
  - Lookup Redis session
  - POST to provider token endpoint
  - Parse response → access_token, refresh_token
  - Create/update Account
  - Delete Redis session

GetProviders() []ProviderInfo
```

#### Task A.3: Provider OAuth Configs
**Files:** `auth/codex/oauth.go`, `auth/claude/oauth.go`

```go
// Per provider:
type CodexOAuthConfig struct {
    AuthURL      string  // https://auth.openai.com/oauth/authorize
    TokenURL     string  // https://auth.openai.com/oauth/token
    ClientID     string  // app_EMoamEEZ73f0CkXaXp7hrann
    Scopes       []string
    DefaultRedirect string
}

func (c *CodexOAuthConfig) BuildAuthURL(state, codeChallenge, redirectURI string) string
func (c *CodexOAuthConfig) ExchangeCode(code, codeVerifier, redirectURI string) (*TokenResponse, error)
```

---

### Stream B: Streaming (3 tasks)

#### Task B.1: Stream Forwarder
**Files:** `providers/stream_forwarder.go`

```go
type StreamForwarder struct {
    writer     http.ResponseWriter
    flusher    http.Flusher
    translator StreamTranslator
}

type StreamTranslator interface {
    TranslateChunk(chunk []byte) []byte
}

func NewStreamForwarder(w http.ResponseWriter, t StreamTranslator) (*StreamForwarder, error)
func (sf *StreamForwarder) Forward(dataCh <-chan []byte, errCh <-chan error) error
  - Loop: read from dataCh
  - Translate chunk
  - Write "data: {chunk}\n\n"
  - Flush
  - Handle errors from errCh
```

#### Task B.2: OpenAI Streaming
**Files:** `providers/openai/executor.go`, `providers/openai/stream_translator.go`

```go
// executor.go
func (p *OpenAIProvider) ExecuteStream(ctx, req) (*StreamResponse, error)
  - Make HTTP request with stream: true
  - Return channels that read SSE from OpenAI
  - Parse "data: {...}" lines

func (p *OpenAIProvider) SupportsStreaming() bool { return true }

// stream_translator.go
type OpenAIStreamTranslator struct{}
func (t *OpenAIStreamTranslator) TranslateChunk(chunk []byte) []byte
  - OpenAI SSE format → Claude SSE format
```

#### Task B.3: GLM Streaming
**Files:** `providers/glm/executor.go`, `providers/glm/stream_translator.go`

```go
// Same pattern as OpenAI
// GLM uses OpenAI-compatible SSE format
```

#### Task B.4: Handler Streaming Support
**Files:** `handlers/proxy.handler.go`

```go
func (h *ProxyHandler) Handle(c *gin.Context) {
    // Check stream flag
    if gjson.GetBytes(body, "stream").Bool() {
        h.handleStreaming(c, body)
        return
    }
    h.handleNonStreaming(c, body)
}

func (h *ProxyHandler) handleStreaming(c *gin.Context, body []byte) {
    // Set headers
    c.Header("Content-Type", "text/event-stream")
    c.Header("Cache-Control", "no-cache")
    c.Header("Connection", "keep-alive")

    // Get provider, execute stream
    streamResp, err := provider.ExecuteStream(ctx, req)

    // Forward with translator
    forwarder := NewStreamForwarder(c.Writer, translator)
    forwarder.Forward(streamResp.DataCh, streamResp.ErrCh)
}
```

---

### Stream C: Function Calling (2 tasks)

#### Task C.1: Response Translation Fix
**Files:**
- `providers/openai/translator.response.go`
- `providers/glm/translator.response.go`
- `providers/antigravity/translator.response.go`

```go
// OpenAI: Fix tool_calls → tool_use
func translateToolCalls(response []byte) []byte {
    toolCalls := gjson.GetBytes(response, "choices.0.message.tool_calls")
    if !toolCalls.Exists() {
        return response
    }

    var content []interface{}
    for _, tc := range toolCalls.Array() {
        // IMPORTANT: Parse arguments from string to object
        argsStr := tc.Get("function.arguments").String()
        var args map[string]interface{}
        json.Unmarshal([]byte(argsStr), &args)

        content = append(content, map[string]interface{}{
            "type":  "tool_use",
            "id":    tc.Get("id").String(),
            "name":  tc.Get("function.name").String(),
            "input": args,
        })
    }

    // Build Claude response
    return buildClaudeResponse(content, "tool_use")
}
```

#### Task C.2: Tool Result Translation
**Files:**
- `providers/openai/translator.request.go`
- `providers/glm/translator.request.go`
- `providers/antigravity/translator.request.go`

```go
// In message translation, detect tool_result
func translateMessage(msg gjson.Result) map[string]interface{} {
    content := msg.Get("content")

    // Check if content has tool_result
    if content.IsArray() {
        for _, part := range content.Array() {
            if part.Get("type").String() == "tool_result" {
                return translateToolResult(part)
            }
        }
    }

    // ... existing translation
}

// OpenAI format
func translateToolResult(part gjson.Result) map[string]interface{} {
    return map[string]interface{}{
        "role":         "tool",
        "tool_call_id": part.Get("tool_use_id").String(),
        "content":      part.Get("content").String(),
    }
}

// Antigravity format
func translateToolResultAntigravity(part gjson.Result) map[string]interface{} {
    return map[string]interface{}{
        "role": "function",
        "parts": []interface{}{
            map[string]interface{}{
                "functionResponse": map[string]interface{}{
                    "name":     part.Get("name").String(),
                    "response": map[string]interface{}{"result": part.Get("content").Value()},
                },
            },
        },
    }
}
```

---

### Stream D: Multimodal (1 task)

#### Task D.1: Image Content Translation
**Files:**
- `providers/openai/translator.request.go`
- `providers/glm/translator.request.go`
- `providers/antigravity/translator.request.go`

```go
// Add to content part translation
func translateContentPart(part gjson.Result) interface{} {
    partType := part.Get("type").String()

    switch partType {
    case "text":
        return map[string]interface{}{
            "type": "text",
            "text": part.Get("text").String(),
        }

    case "image":
        return translateImage(part)

    case "tool_use":
        return translateToolUse(part)

    case "tool_result":
        return translateToolResult(part)
    }

    return nil
}

// OpenAI image translation
func translateImageOpenAI(part gjson.Result) map[string]interface{} {
    source := part.Get("source")
    mediaType := source.Get("media_type").String()
    data := source.Get("data").String()

    return map[string]interface{}{
        "type": "image_url",
        "image_url": map[string]interface{}{
            "url": fmt.Sprintf("data:%s;base64,%s", mediaType, data),
        },
    }
}

// Antigravity image translation
func translateImageAntigravity(part gjson.Result) map[string]interface{} {
    source := part.Get("source")
    return map[string]interface{}{
        "inlineData": map[string]interface{}{
            "mimeType": source.Get("media_type").String(),
            "data":     source.Get("data").String(),
        },
    }
}
```

---

## Phase 3: Integration & Routes

### Task 3.1: Register OAuth Routes
**Files:** `routes/routes.go`

```go
// OAuth routes
oauth := api.Group("/oauth")
{
    oauth.POST("/init", oauthHandler.InitOAuth)
    oauth.GET("/callback", oauthHandler.HandleCallback)
    oauth.POST("/exchange", oauthHandler.ExchangeCode)
    oauth.GET("/providers", oauthHandler.ListProviders)
    oauth.POST("/refresh", oauthHandler.RefreshToken)
}
```

### Task 3.2: Wire Dependencies
**Files:** `cmd/main.go`

```go
// Initialize OAuth
oauthFlowService := services.NewOAuthFlowService(redisClient, accountRepo)
oauthHandler := handlers.NewOAuthHandler(oauthFlowService, accountRepo)
```

---

## Phase 4: Testing

### Task 4.1: OAuth Tests
```
- Test init flow (auto + manual)
- Test callback parsing
- Test code exchange
- Test token refresh
- Test expired session handling
```

### Task 4.2: Streaming Tests
```
- Test stream detection in handler
- Test SSE format output
- Test provider stream translation
- Test error handling mid-stream
```

### Task 4.3: Function Calling Tests
```
- Test tool_calls → tool_use response
- Test tool_result → provider format request
- Test multi-turn tool conversation
```

### Task 4.4: Multimodal Tests
```
- Test image translation per provider
- Test mixed content (text + image)
- Test unsupported format handling
```

---

## Task Summary

| Phase | Task | Files | Est. Lines |
|-------|------|-------|------------|
| 1 | PKCE Utility | 1 | 60 |
| 1 | Provider Interface | 1 | 30 |
| 2-A | OAuth Handler | 1 | 200 |
| 2-A | OAuth Flow Service | 1 | 150 |
| 2-A | Provider OAuth Configs | 2 | 200 |
| 2-B | Stream Forwarder | 1 | 100 |
| 2-B | OpenAI Streaming | 2 | 150 |
| 2-B | GLM Streaming | 2 | 100 |
| 2-B | Handler Streaming | 1 | 80 |
| 2-C | Response Translation | 3 | 150 |
| 2-C | Tool Result Translation | 3 | 120 |
| 2-D | Image Translation | 3 | 90 |
| 3 | Routes + Wiring | 2 | 50 |
| 4 | Tests | 4 | 400 |

**Total:** ~1,880 lines across ~27 files

---

## Execution Order

```
Week 1:
├── Phase 1 (Foundation) - 1 person
└── Phase 2 (Parallel)
    ├── Stream A: OAuth - Person 1
    ├── Stream B: Streaming - Person 2
    ├── Stream C: Function Calling - Person 3
    └── Stream D: Multimodal - Person 4

Week 2:
├── Phase 3 (Integration)
└── Phase 4 (Testing)
```

Or single developer:
```
Day 1: Phase 1 + Stream D (Multimodal - smallest)
Day 2: Stream C (Function Calling)
Day 3-4: Stream B (Streaming)
Day 5-7: Stream A (OAuth)
Day 8: Phase 3 + 4 (Integration + Testing)
```
