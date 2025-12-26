# Architecture Overview

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

## Related Documentation

- [Component Details](components.md) - Handler, Service, and Repository layers
- [Database Schema](database.md) - MySQL tables and Redis structures
- [Concurrency & Error Handling](concurrency.md) - Thread safety patterns
- [Performance & Scalability](performance.md) - Optimization strategies
- [Monitoring & Observability](monitoring.md) - Metrics and logging
