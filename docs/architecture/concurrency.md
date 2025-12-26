# Concurrency and Error Handling

This document describes thread safety mechanisms and error handling strategies in AIGateway.

## Thread-Safe Components

### 1. Provider Registry

**Implementation**: `providers/registry.go`

```go
type Registry struct {
    mu        sync.RWMutex
    providers map[string]*models.Provider
}
```

**Pattern**: Multiple readers, single writer
- Read operations use `RLock()` for concurrent access
- Write operations use `Lock()` for exclusive access
- Prevents race conditions on provider map

**Usage**:
```go
// Read (concurrent)
func (r *Registry) Get(id string) (*models.Provider, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    // Safe concurrent reads
}

// Write (exclusive)
func (r *Registry) Register(id string, provider *models.Provider) {
    r.mu.Lock()
    defer r.mu.Unlock()
    // Exclusive write access
}
```

### 2. HTTP Client Service

**Implementation**: `services/httpclient.service.go`

```go
type HTTPClientService struct {
    mu      sync.RWMutex
    clients map[string]*http.Client
}
```

**Pattern**: Connection pooling with thread-safe map access
- Reuses HTTP clients per proxy URL
- Thread-safe client creation and retrieval
- Prevents duplicate client instances

### 3. Redis Operations

**Atomic Operations**:
- `INCR` - Atomic increment for round-robin counters
- `GET/SET` - Thread-safe token cache access
- No race conditions on shared state

**Example**:
```go
// Atomic round-robin increment
counter, _ := redis.Incr(ctx, key).Result()
// Guaranteed unique counter value per request
```

## Async Operations

### 1. Stats Logging

**Implementation**:
```go
go s.statsService.RecordRequest(
    account.ID,
    account.ProxyID,
    providerID,
    model,
    statusCode,
    latency,
)
```

**Benefits**:
- Non-blocking request logging
- Database write failures don't break requests
- Higher throughput

**Trade-offs**:
- Potential log loss on crash
- No immediate confirmation of log write

### 2. Last Used Timestamp

**Implementation**:
```go
go s.repo.UpdateLastUsed(selected.ID)
```

**Benefits**:
- Non-blocking account metadata update
- Minimal impact on request latency

## Error Handling Strategy

### Layered Error Handling

**1. Handler Layer**: HTTP status codes + JSON error responses
```go
if err != nil {
    c.JSON(http.StatusInternalServerError, gin.H{
        "error": err.Error(),
    })
    return
}
```

**2. Service Layer**: Structured errors with context
```go
return fmt.Errorf("no available accounts for provider %s", providerID)
```

**3. Repository Layer**: Database errors with query context
```go
return fmt.Errorf("failed to query accounts: %w", err)
```

### Error Types

**Provider Errors**:
```go
// Provider not found
return fmt.Errorf("provider not found: %s", id)

// No available accounts
return fmt.Errorf("no available accounts for provider %s", providerID)
```

**Authentication Errors**:
```go
// OAuth token refresh failed
return fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, body)

// Invalid API key
return fmt.Errorf("invalid api key format")
```

**Upstream Errors**:
```go
// Provider API error
return fmt.Errorf("upstream error: %d", httpResp.StatusCode)

// Proxy connection failed
return fmt.Errorf("proxy connection failed: %w", err)
```

### Failure Recording

**Strategy**: Log failures without breaking request flow
```go
if err != nil {
    s.statsService.RecordFailure(&account.ID, account.ProxyID, latency, err)
    return Response{}, err
}
```

**Tracked Metrics**:
- Failure count per proxy
- Error status codes
- Error messages
- Latency even on failures

## Race Condition Prevention

### Account Round-Robin

**Challenge**: Multiple goroutines selecting accounts concurrently

**Solution**: Redis atomic INCR
```go
// Atomic operation - no race condition
counter := redis.Incr(ctx, "account:rr:antigravity:claude-sonnet-4-5")
```

### HTTP Client Creation

**Challenge**: Multiple goroutines creating same client

**Solution**: Read-write mutex with double-check pattern
```go
// First check (read lock)
s.mu.RLock()
if client, exists := s.clients[proxyURL]; exists {
    s.mu.RUnlock()
    return client
}
s.mu.RUnlock()

// Create (write lock)
s.mu.Lock()
defer s.mu.Unlock()
// Double check after acquiring write lock
if client, exists := s.clients[proxyURL]; exists {
    return client
}
// Create new client
```

## Goroutine Management

### Best Practices

**1. Fire-and-forget for non-critical operations**:
```go
// Async logging - don't wait
go statsService.RecordRequest(...)
```

**2. Avoid goroutine leaks**:
- All spawned goroutines should complete
- No infinite loops without exit conditions
- Context cancellation for long-running ops

**3. Error handling in goroutines**:
```go
go func() {
    if err := operation(); err != nil {
        // Log error, don't panic
        log.Printf("async operation failed: %v", err)
    }
}()
```

## Related Documentation

- [Architecture Overview](README.md) - System architecture
- [Components](components.md) - Service implementations
- [Performance](performance.md) - Scalability considerations
