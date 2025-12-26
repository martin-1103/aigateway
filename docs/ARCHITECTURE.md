# AIGateway Architecture

This document describes the system architecture, design patterns, and data flow of AIGateway.

## Overview

AIGateway is built on a multi-layered architecture that separates concerns into distinct components:

- **Handler Layer**: HTTP request/response handling
- **Provider Layer**: Provider abstraction and routing
- **Service Layer**: Business logic and orchestration
- **Repository Layer**: Data access and persistence
- **External Layer**: Provider APIs, Redis, MySQL

## Core Design Patterns

### 1. Provider Interface Pattern

The `Provider` interface defines a contract that all AI provider implementations must satisfy:

```go
type Provider interface {
    ID() string
    Name() string
    AuthStrategy() string
    SupportedModels() []string
    TranslateRequest(format string, payload []byte, model string) ([]byte, error)
    TranslateResponse(payload []byte) ([]byte, error)
    Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
}
```

**Benefits**:
- Uniform interface for all providers
- Easy to add new providers without modifying core logic
- Testable through mock implementations
- Decouples provider-specific logic from gateway logic

**Location**: `providers/provider.go`

### 2. Authentication Strategy Pattern

The `Strategy` interface enables pluggable authentication mechanisms:

```go
type Strategy interface {
    Name() string
    GetToken(ctx context.Context, authData map[string]interface{}) (string, error)
    RefreshToken(ctx context.Context, authData map[string]interface{}, oldToken string) (string, error)
    ValidateToken(ctx context.Context, token string) (bool, error)
}
```

**Implementations**:
- **OAuthStrategy**: OAuth 2.0 with token caching and refresh
- **APIKeyStrategy**: Static API key extraction
- **BearerStrategy**: Bearer token authentication

**Benefits**:
- Provider-agnostic authentication
- Centralized token management
- Easy to add new auth mechanisms
- Redis-based caching for OAuth tokens

**Location**: `auth/strategy.go`

### 3. Provider Registry Pattern

The `Registry` manages provider instances with thread-safe operations and intelligent routing:

```go
type Registry struct {
    mu        sync.RWMutex
    providers map[string]*models.Provider
}
```

**Key Methods**:
- `Register(id, provider)`: Add provider to registry
- `Get(id)`: Retrieve provider by ID
- `GetByModel(model)`: Route model to provider via prefix matching
- `routeModel(model)`: Model-to-provider mapping logic

**Routing Logic**:
```go
switch {
case strings.HasPrefix(modelLower, "gemini-"):
    return "antigravity"
case strings.HasPrefix(modelLower, "claude-sonnet-"):
    return "antigravity"
case strings.HasPrefix(modelLower, "gpt-"):
    return "openai"
case strings.HasPrefix(modelLower, "glm-"):
    return "glm"
}
```

**Location**: `providers/registry.go`

## Request Flow

### Complete Request Lifecycle

```
1. Client Request
   ↓
2. Handler (proxy.handler.go)
   - Parse request body
   - Extract model parameter
   - Validate input
   ↓
3. Provider Registry (registry.go)
   - Route model → provider ID
   - Validate provider is active
   ↓
4. Account Service (account.service.go)
   - Redis: INCR account:rr:{provider}:{model}
   - Calculate index: (counter - 1) % account_count
   - Select account via round-robin
   - Update last_used_at (async)
   ↓
5. Proxy Service (proxy.service.go)
   - Check if account has assigned proxy
   - If not: assign via fill-first algorithm
   - Update account.proxy_id and account.proxy_url
   ↓
6. OAuth Service (oauth.service.go)
   - Check Redis cache: auth:oauth:{provider}:{account}
   - If cached and valid (>5 min remaining): return token
   - If expired: extract from account.auth_data
   - If needs refresh: call OAuth token endpoint
   - Update Redis cache and database
   ↓
7. Translator Service (translator.service.go)
   - Translate request format (Claude → Provider)
   - Transform: messages, system, tools, parameters
   ↓
8. HTTP Client Service (httpclient.service.go)
   - Get or create HTTP client for proxy URL
   - Reuse connection pool per proxy
   ↓
9. Executor Service (executor.service.go)
   - Build HTTP request to provider API
   - Add Authorization header with token
   - Execute request via HTTP client
   - Measure latency
   ↓
10. Provider API
    - Process request
    - Return response
    ↓
11. Translator Service (translator.service.go)
    - Translate response format (Provider → Claude)
    - Transform: candidates, usage, finish_reason
    ↓
12. Stats Service (stats.service.go)
    - Record request log (async to MySQL)
    - Update proxy stats (async to Redis + MySQL)
    - Update failure counters if error
    ↓
13. Handler Response
    - Return translated response to client
```

## Component Details

### Handler Layer

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

### Service Layer

#### Executor Service

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

#### Account Service

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

#### Proxy Service

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

#### OAuth Service

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

#### Translator Service

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

### Repository Layer

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

## Data Flow Diagram

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       │ POST /v1/messages
       │ {model, messages, ...}
       ↓
┌──────────────────┐
│  Gin Handler     │
│  proxy.handler   │
└────────┬─────────┘
         │
         │ model name
         ↓
┌──────────────────┐
│ Provider Registry│ ─── routeModel(model) ───→ provider_id
└────────┬─────────┘
         │
         │ provider_id
         ↓
┌──────────────────┐
│ Account Service  │ ─── Redis INCR ───→ account index
└────────┬─────────┘
         │
         │ account
         ↓
┌──────────────────┐
│  Proxy Service   │ ─── fill-first ───→ proxy assignment
└────────┬─────────┘
         │
         │ account + proxy
         ↓
┌──────────────────┐
│  OAuth Service   │ ─── Redis cache ───→ access token
└────────┬─────────┘
         │
         │ token
         ↓
┌──────────────────┐
│ Translator Svc   │ ─── format transform ───→ provider payload
└────────┬─────────┘
         │
         │ translated payload
         ↓
┌──────────────────┐
│ HTTP Client Svc  │ ─── connection pool ───→ HTTP client
└────────┬─────────┘
         │
         │ HTTP POST
         ↓
┌──────────────────┐
│  Provider API    │
└────────┬─────────┘
         │
         │ response
         ↓
┌──────────────────┐
│ Translator Svc   │ ─── format transform ───→ Claude payload
└────────┬─────────┘
         │
         │ translated response
         ↓
┌──────────────────┐
│  Stats Service   │ ─── async logging ───→ MySQL + Redis
└────────┬─────────┘
         │
         │ final response
         ↓
┌──────────────────┐
│     Client       │
└──────────────────┘
```

## Database Schema Design

### Providers Table

```sql
CREATE TABLE providers (
    id              VARCHAR(50) PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    base_url        VARCHAR(255),
    auth_type       ENUM('oauth', 'api_key', 'bearer') NOT NULL,
    auth_strategy   VARCHAR(50),
    supported_models JSON,
    is_active       BOOLEAN DEFAULT true,
    config          JSON,
    created_at      TIMESTAMP,
    updated_at      TIMESTAMP
);

INDEX idx_active ON providers(is_active);
```

### Accounts Table

```sql
CREATE TABLE accounts (
    id          VARCHAR(36) PRIMARY KEY,
    provider_id VARCHAR(50) NOT NULL,
    label       VARCHAR(100) NOT NULL,
    auth_data   JSON NOT NULL,
    metadata    JSON,
    is_active   BOOLEAN DEFAULT true,
    proxy_url   VARCHAR(255),
    proxy_id    INT,
    last_used_at TIMESTAMP,
    usage_count BIGINT DEFAULT 0,
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP,

    FOREIGN KEY (provider_id) REFERENCES providers(id),
    FOREIGN KEY (proxy_id) REFERENCES proxy_pool(id)
);

INDEX idx_provider_active ON accounts(provider_id, is_active);
INDEX idx_label ON accounts(label);
INDEX idx_proxy ON accounts(proxy_id);
```

### Proxy Pool Table

```sql
CREATE TABLE proxy_pool (
    id          INT AUTO_INCREMENT PRIMARY KEY,
    label       VARCHAR(100) NOT NULL,
    proxy_url   VARCHAR(255) NOT NULL UNIQUE,
    is_active   BOOLEAN DEFAULT true,
    max_failures INT DEFAULT 3,
    failure_count INT DEFAULT 0,
    created_at  TIMESTAMP,
    updated_at  TIMESTAMP
);

INDEX idx_active ON proxy_pool(is_active);
```

### Request Logs Table

```sql
CREATE TABLE request_logs (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    account_id  VARCHAR(36),
    provider_id VARCHAR(50),
    proxy_id    INT,
    model       VARCHAR(100),
    status_code INT,
    latency_ms  INT,
    error       TEXT,
    created_at  TIMESTAMP,

    FOREIGN KEY (account_id) REFERENCES accounts(id),
    FOREIGN KEY (provider_id) REFERENCES providers(id),
    FOREIGN KEY (proxy_id) REFERENCES proxy_pool(id)
);

INDEX idx_created ON request_logs(created_at);
INDEX idx_provider ON request_logs(provider_id);
INDEX idx_proxy ON request_logs(proxy_id);
```

### Proxy Stats Table

```sql
CREATE TABLE proxy_stats (
    id              BIGINT AUTO_INCREMENT PRIMARY KEY,
    proxy_id        INT NOT NULL,
    date            DATE NOT NULL,
    total_requests  INT DEFAULT 0,
    success_requests INT DEFAULT 0,
    failed_requests INT DEFAULT 0,
    avg_latency     INT DEFAULT 0,
    total_tokens    BIGINT DEFAULT 0,

    FOREIGN KEY (proxy_id) REFERENCES proxy_pool(id),
    UNIQUE KEY unique_proxy_date (proxy_id, date)
);

INDEX idx_date ON proxy_stats(date);
```

## Redis Data Structures

### Round-Robin Counters

```
Key: account:rr:{provider_id}:{model}
Type: Integer
TTL: None (persistent)
Operations: INCR

Example:
account:rr:antigravity:claude-sonnet-4-5 → 127
```

### OAuth Token Cache

```
Key: auth:oauth:{provider_id}:{account_id}
Type: String (JSON)
TTL: expires_in seconds
Operations: GET, SET

Example:
auth:oauth:antigravity:acc-123 → {
  "access_token": "ya29.xxx",
  "refresh_token": "1//xxx",
  "expires_at": "2024-12-26T10:00:00Z",
  "token_type": "Bearer"
}
```

### Real-Time Stats

```
Key: stats:proxy:{proxy_id}:requests
Type: Integer
TTL: None
Operations: INCR

Key: stats:proxy:{proxy_id}:failures
Type: Integer
TTL: None
Operations: INCR
```

## Concurrency and Thread Safety

### Thread-Safe Components

1. **Provider Registry**:
   - Uses `sync.RWMutex` for read/write locks
   - Multiple readers, single writer pattern

2. **HTTP Client Service**:
   - Uses `sync.RWMutex` for client map access
   - Connection pooling per proxy URL

3. **Redis Operations**:
   - Atomic operations (INCR, GET, SET)
   - No race conditions on counters

### Async Operations

1. **Stats Logging**:
   ```go
   go s.statsService.RecordRequest(...)
   ```
   - Non-blocking request logging
   - Background goroutine

2. **Last Used Timestamp**:
   ```go
   go s.repo.UpdateLastUsed(selected.ID)
   ```
   - Non-blocking timestamp update
   - Background goroutine

## Error Handling Strategy

### Layered Error Handling

1. **Handler Layer**: HTTP status codes + JSON error responses
2. **Service Layer**: Structured errors with context
3. **Repository Layer**: Database errors with query context

### Error Types

```go
// Provider not found
return fmt.Errorf("provider not found: %s", id)

// No available accounts
return fmt.Errorf("no available accounts for provider %s", providerID)

// OAuth token refresh failed
return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, body)

// Upstream API error
return fmt.Errorf("upstream error: %d", httpResp.StatusCode)
```

### Failure Recording

```go
if err != nil {
    s.statsService.RecordFailure(&account.ID, account.ProxyID, latency, err)
    return Response{}, err
}
```

## Performance Considerations

### Connection Pooling

**HTTP Client Reuse**:
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

### Token Caching

**OAuth Token Cache**:
- Redis cache with TTL = expires_in
- 5-minute buffer before expiry
- Reduces auth overhead by ~99%

### Async Logging

**Non-Blocking Operations**:
```go
go s.statsService.RecordRequest(...)
```

**Benefits**:
- Request latency not impacted by logging
- Database write failures don't break requests
- Higher throughput

### Database Indexing

**Optimized Queries**:
- `idx_provider_active` on `accounts(provider_id, is_active)`
- `idx_proxy` on `accounts(proxy_id)`
- `idx_created` on `request_logs(created_at)`

## Scalability Considerations

### Horizontal Scaling

**Stateless Design**:
- All state in Redis/MySQL
- Multiple gateway instances possible
- Load balancer in front

**Shared State**:
- Round-robin counters in Redis
- OAuth tokens in Redis
- Account data in MySQL

### Bottlenecks and Mitigations

1. **Redis INCR Contention**:
   - Mitigation: Redis is single-threaded, atomic ops are fast
   - Alternative: Shard by provider+model

2. **MySQL Write Load**:
   - Mitigation: Async logging
   - Alternative: Batch inserts every N seconds

3. **Provider API Rate Limits**:
   - Mitigation: Multiple accounts per provider
   - Round-robin distributes load

## Security Architecture

### Credential Storage

- **At Rest**: JSON in MySQL `auth_data` field
- **Recommendation**: Encrypt `auth_data` column
- **In Transit**: TLS for all provider API calls

### Access Control

- **No authentication on proxy endpoints**: Add API key middleware
- **Management API**: No auth currently (add JWT/OAuth)
- **Database**: Restrict to localhost or private network

### Token Security

- **OAuth Tokens**: Cached in Redis with TTL
- **Redis**: Should use AUTH password
- **MySQL**: Should use strong password

## Monitoring and Observability

### Metrics Collection

1. **Request Logs**: Every request logged to `request_logs`
2. **Proxy Stats**: Daily aggregation in `proxy_stats`
3. **Redis Counters**: Real-time request/failure counts

### Key Metrics

- Request latency (per proxy, per provider)
- Success/failure rates
- Account usage distribution
- Proxy health status
- OAuth token refresh rate

### Log Aggregation

**Structure**:
```
request_logs: account_id, provider_id, proxy_id, model, status_code, latency_ms, error, created_at
```

**Queries**:
```sql
-- Average latency per proxy
SELECT proxy_id, AVG(latency_ms) FROM request_logs GROUP BY proxy_id;

-- Error rate per provider
SELECT provider_id,
       SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) / COUNT(*)
FROM request_logs GROUP BY provider_id;
```

## Future Improvements

1. **Dynamic Provider Loading**: Load providers from database instead of hardcoding
2. **Rate Limiting**: Per-account, per-provider rate limits
3. **Circuit Breaker**: Disable failing accounts automatically
4. **Health Checks**: Periodic provider API health checks
5. **Metrics Export**: Prometheus metrics endpoint
6. **WebSocket Streaming**: Real-time streaming support
7. **Request Retries**: Automatic retry with exponential backoff
8. **Provider Failover**: Fallback to alternative providers
