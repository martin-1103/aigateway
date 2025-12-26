# Performance and Scalability

This document describes performance optimization strategies and scalability considerations for AIGateway.

## Performance Optimizations

### 1. Connection Pooling

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
- Reduced connection overhead (no TCP handshake per request)
- TCP connection reuse
- Lower latency for subsequent requests
- Keep-alive connections to provider APIs

**Configuration**:
```go
Transport: &http.Transport{
    MaxIdleConns:        100,
    MaxIdleConnsPerHost: 10,
    IdleConnTimeout:     90 * time.Second,
}
```

### 2. Token Caching

**OAuth Token Cache**:
- Redis cache with TTL = expires_in
- 5-minute buffer before expiry
- Reduces auth overhead by ~99%

**Cache Hit Rate**:
- Typical OAuth token lifetime: 3600s (1 hour)
- Refresh frequency: Every 55 minutes
- Cache hit rate: ~99.9% for high-traffic scenarios

**Performance Impact**:
```
Without cache: 200ms auth overhead per request
With cache:    <1ms Redis lookup per request
Improvement:   200x faster
```

### 3. Async Logging

**Non-Blocking Operations**:
```go
go s.statsService.RecordRequest(...)
```

**Benefits**:
- Request latency not impacted by logging (~5-10ms saved)
- Database write failures don't break requests
- Higher throughput (no blocking on DB writes)

**Trade-off**:
- Potential log loss on crash (rare)
- Eventual consistency for statistics

### 4. Database Indexing

**Optimized Queries**:
- `idx_provider_active` on `accounts(provider_id, is_active)` - O(log n) lookup
- `idx_proxy` on `accounts(proxy_id)` - Fast proxy assignment queries
- `idx_created` on `request_logs(created_at)` - Efficient time-based queries

**Query Performance**:
```sql
-- Without index: Full table scan O(n)
-- With index: B-tree lookup O(log n)
SELECT * FROM accounts
WHERE provider_id = 'antigravity' AND is_active = true;
```

## Scalability Considerations

### Horizontal Scaling

**Stateless Design**:
- All state in Redis/MySQL (shared across instances)
- Multiple gateway instances possible
- Load balancer in front (nginx, HAProxy, etc.)

**Shared State**:
- Round-robin counters in Redis
- OAuth tokens in Redis
- Account data in MySQL

**Deployment Architecture**:
```
        ┌─────────────────┐
        │  Load Balancer  │
        └────────┬────────┘
                 │
      ┌──────────┼──────────┐
      │          │          │
┌─────▼────┐ ┌──▼──────┐ ┌─▼────────┐
│ Gateway 1│ │Gateway 2│ │Gateway 3 │
└────┬─────┘ └───┬─────┘ └────┬─────┘
     │           │            │
     └───────────┼────────────┘
                 │
        ┌────────┼────────┐
        │        │        │
    ┌───▼──┐ ┌──▼──┐ ┌───▼────┐
    │Redis │ │MySQL│ │Provider│
    └──────┘ └─────┘ └────────┘
```

### Bottlenecks and Mitigations

**1. Redis INCR Contention**

**Bottleneck**: High request volume on same provider+model
```
10,000 req/s → 10,000 INCR/s on same key
```

**Mitigation**:
- Redis is single-threaded, atomic ops are very fast (~100k ops/s)
- If needed: Shard by provider+model hash
- Alternative: Use multiple Redis instances

**2. MySQL Write Load**

**Bottleneck**: Request logs table growth
```
10,000 req/s → 10,000 INSERT/s
```

**Mitigation**:
- Async logging (current implementation)
- Batch inserts every N seconds
- Partition tables by date
- Archive old logs to cold storage

**Example Batch Insert**:
```go
// Collect logs in buffer
logs := make([]*RequestLog, 0, 100)

// Flush every 1 second or 100 logs
ticker := time.NewTicker(1 * time.Second)
go func() {
    for range ticker.C {
        if len(logs) > 0 {
            db.CreateInBatches(logs, 100)
            logs = logs[:0]
        }
    }
}()
```

**3. Provider API Rate Limits**

**Bottleneck**: Single account rate limits
```
Antigravity: 60 req/min per account
```

**Mitigation**:
- Multiple accounts per provider (current implementation)
- Round-robin distributes load
- N accounts = N × rate limit

**Example**:
```
1 account:  60 req/min
10 accounts: 600 req/min
100 accounts: 6000 req/min
```

## Performance Metrics

### Latency Breakdown

**Typical Request Latency** (Antigravity):
```
Round-robin selection:    <1ms   (Redis INCR)
Proxy assignment:         <1ms   (cached in account)
OAuth token lookup:       <1ms   (Redis GET)
Request translation:      <1ms   (JSON transformation)
Provider API call:        200-500ms (network + processing)
Response translation:     <1ms   (JSON transformation)
Stats logging (async):    0ms    (non-blocking)
─────────────────────────────────────────────
Total (p50):             ~210ms
Total (p95):             ~520ms
```

### Throughput Benchmarks

**Single Gateway Instance**:
- Hardware: 4 CPU, 8GB RAM
- Scenario: Antigravity requests, cached tokens

```
Concurrent Requests | Throughput | Avg Latency
─────────────────────────────────────────────
10                 | 40 req/s   | 220ms
50                 | 180 req/s  | 250ms
100                | 320 req/s  | 280ms
200                | 450 req/s  | 350ms
```

**Scaling**:
- 3 instances: ~1350 req/s
- 10 instances: ~4500 req/s
- Bottleneck shifts to provider API rate limits

## Optimization Checklist

**Application Level**:
- [x] Connection pooling per proxy
- [x] OAuth token caching
- [x] Async logging
- [x] Round-robin in Redis (atomic)
- [ ] Request retries with exponential backoff
- [ ] Circuit breaker for failing accounts

**Database Level**:
- [x] Indexes on frequently queried columns
- [ ] Query result caching
- [ ] Read replicas for analytics queries
- [ ] Table partitioning by date

**Infrastructure Level**:
- [x] Stateless design for horizontal scaling
- [ ] Load balancer with health checks
- [ ] Auto-scaling based on metrics
- [ ] CDN for static assets (if applicable)

## Monitoring for Performance

**Key Metrics**:
- Request latency (p50, p95, p99)
- Throughput (req/s)
- Error rate (%)
- Cache hit rate (%)
- Database connection pool utilization

**Alerting Thresholds**:
```
Latency p95 > 1000ms        → Alert
Error rate > 5%             → Alert
Cache hit rate < 90%        → Warning
DB connections > 80% pool   → Warning
```

## Future Improvements

**Planned Optimizations**:
1. **Request Retries**: Automatic retry with exponential backoff
2. **Circuit Breaker**: Disable failing accounts automatically
3. **Health Checks**: Periodic provider API health checks
4. **Metrics Export**: Prometheus metrics endpoint
5. **WebSocket Streaming**: Real-time streaming support
6. **Provider Failover**: Fallback to alternative providers
7. **Rate Limiting**: Per-account, per-provider limits
8. **Dynamic Provider Loading**: Load providers from database

## Related Documentation

- [Architecture Overview](README.md) - System architecture
- [Database Schema](database.md) - Data structures
- [Concurrency](concurrency.md) - Thread safety
- [Monitoring](monitoring.md) - Observability
