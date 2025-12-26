# Component Details

This document describes the responsibilities and implementation details of each architectural layer.

## Handler Layer

**File**: `handlers/proxy.handler.go`

**Responsibilities**:
- HTTP request parsing
- Model extraction from request body
- Request validation
- Response formatting
- Error handling

**Key Code**:
```go
func (h *ProxyHandler) HandleProxy(c *gin.Context) {
    body, _ := io.ReadAll(c.Request.Body)
    model := gjson.GetBytes(body, "model").String()

    req := services.Request{
        ProviderID: "antigravity",  // TODO: Dynamic routing
        Model:      model,
        Payload:    body,
        Stream:     c.Query("stream") == "true",
    }

    resp, err := h.executor.Execute(req)
    c.Data(resp.StatusCode, "application/json", resp.Payload)
}
```

## Service Layer

### Executor Service

**File**: `services/executor.service.go`

**Responsibilities**:
- Orchestrate request execution
- Coordinate all services
- Build provider API request
- Handle streaming/non-streaming

**Dependencies**:
- AccountService
- ProxyService
- OAuthService
- TranslatorService
- HTTPClientService
- StatsService

### Account Service

**File**: `services/account.service.go`

**Responsibilities**:
- Round-robin account selection
- Redis-based counter management
- Account activation/deactivation

**Round-Robin Algorithm**:
```go
key := fmt.Sprintf("account:rr:%s:%s", providerID, model)
idx, _ := redis.Incr(ctx, key).Result()
selected := accounts[(idx-1) % int64(len(accounts))]
```

**Key Properties**:
- Atomic increment ensures thread-safety
- Persistent state across restarts (Redis)
- Per-provider, per-model distribution
- Automatic wrapping via modulo

### Proxy Service

**File**: `services/proxy.service.go`

**Responsibilities**:
- Fill-first proxy assignment
- Proxy health tracking
- Assignment recalculation

**Fill-First Algorithm**:
```
1. Query all active proxies
2. For each proxy in order:
   - Count currently assigned accounts
   - If count < capacity: assign to this proxy
   - Break
3. If all full: assign to first proxy (overflow)
4. Update account.proxy_id and account.proxy_url
```

### OAuth Service

**File**: `services/oauth.service.go`

**Responsibilities**:
- OAuth token retrieval
- Token refresh before expiration
- Redis-based caching
- Database persistence

**Token Refresh Logic**:
```
1. Check Redis: auth:oauth:{provider}:{account}
2. If cached:
   - Parse expiry
   - If >5 minutes remaining: return
3. If <5 minutes or missing:
   - Extract refresh_token from auth_data
   - Call OAuth token endpoint
   - Update Redis (TTL = expires_in)
   - Update database auth_data
   - Return new token
```

### Translator Service

**File**: `services/translator.service.go`

**Responsibilities**:
- Format conversion (Claude ↔ Antigravity)
- Message structure transformation
- Tool/function call translation

**Request Translation** (Claude → Antigravity):
```
messages[].role: "assistant" → "model"
messages[].content: string → parts[].text
system: string → systemInstruction.parts[].text
tools[].input_schema → tools[].functionDeclarations[].parametersJsonSchema
max_tokens → generationConfig.maxOutputTokens
temperature → generationConfig.temperature
```

**Response Translation** (Antigravity → Claude):
```
candidates[].content.role: "model" → "assistant"
candidates[].content.parts[].text → content[].text
candidates[].finishReason: "MAX_TOKENS" → "max_tokens"
usageMetadata.promptTokenCount → usage.input_tokens
usageMetadata.candidatesTokenCount → usage.output_tokens
```

### HTTP Client Service

**File**: `services/httpclient.service.go`

**Responsibilities**:
- HTTP client creation and pooling
- Proxy configuration
- Connection reuse per proxy URL

**Connection Pooling**:
```go
// One client per proxy URL
clients := make(map[string]*http.Client)

func GetClient(proxyURL string) *http.Client {
    if client, exists := clients[proxyURL]; exists {
        return client
    }
    // Create and cache
}
```

**Benefits**:
- Reduced connection overhead
- TCP connection reuse
- Lower latency for subsequent requests

### Stats Service

**File**: `services/stats.service.go`

**Responsibilities**:
- Request logging (async)
- Proxy statistics tracking (async)
- Failure counter updates

**Async Operations**:
```go
go s.statsService.RecordRequest(...)
```

**Benefits**:
- Request latency not impacted by logging
- Database write failures don't break requests
- Higher throughput

## Repository Layer

**Files**:
- `repositories/account.repository.go`
- `repositories/proxy.repository.go`
- `repositories/stats.repository.go`

**Responsibilities**:
- Database CRUD operations
- Query optimization
- Transaction management
- Data integrity

**Key Queries**:
```go
// Get active accounts for provider
GetActiveByProvider(providerID) → WHERE provider_id = ? AND is_active = true

// Get active proxies
GetActiveProxies() → WHERE is_active = true ORDER BY id

// Record stats (async)
RecordRequest(accountID, proxyID, providerID, model, status, latency)
```

## Related Documentation

- [Architecture Overview](README.md) - Design patterns and request flow
- [Database Schema](database.md) - Data models and structures
- [Performance](performance.md) - Optimization strategies
