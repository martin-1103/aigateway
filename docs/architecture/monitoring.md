# Monitoring and Observability

This document describes metrics collection, logging, and monitoring strategies for AIGateway.

## Metrics Collection

### 1. Request Logs

**Table**: `request_logs`

**Collected Data**:
- Account ID
- Provider ID
- Proxy ID
- Model name
- Status code
- Latency (ms)
- Error message
- Timestamp

**Storage**: MySQL (async write)

**Usage**:
```sql
-- Recent errors
SELECT * FROM request_logs
WHERE status_code >= 400
ORDER BY created_at DESC
LIMIT 100;

-- Latency analysis
SELECT
    provider_id,
    AVG(latency_ms) as avg_latency,
    MAX(latency_ms) as max_latency,
    MIN(latency_ms) as min_latency
FROM request_logs
WHERE created_at >= NOW() - INTERVAL 1 HOUR
GROUP BY provider_id;
```

### 2. Proxy Stats

**Table**: `proxy_stats`

**Aggregated Daily**:
- Total requests
- Successful requests
- Failed requests
- Average latency
- Total tokens (if available)

**Usage**:
```sql
-- Proxy performance comparison
SELECT
    proxy_id,
    total_requests,
    (success_requests * 100.0 / total_requests) as success_rate,
    avg_latency
FROM proxy_stats
WHERE date = CURDATE()
ORDER BY success_rate DESC;
```

### 3. Real-Time Counters

**Redis Keys**:
- `stats:proxy:{id}:requests` - Request counter
- `stats:proxy:{id}:failures` - Failure counter
- `account:rr:{provider}:{model}` - Round-robin position

**Usage**:
```bash
# Get real-time stats
redis-cli GET stats:proxy:1:requests
redis-cli GET stats:proxy:1:failures

# Calculate error rate
requests=$(redis-cli GET stats:proxy:1:requests)
failures=$(redis-cli GET stats:proxy:1:failures)
error_rate=$(echo "scale=2; $failures * 100 / $requests" | bc)
```

## Key Metrics

### Request Metrics

**Latency** (per proxy, per provider):
- p50 (median)
- p95 (95th percentile)
- p99 (99th percentile)
- Max latency

**Success/Failure Rates**:
```sql
SELECT
    provider_id,
    COUNT(*) as total,
    SUM(CASE WHEN status_code < 400 THEN 1 ELSE 0 END) as success,
    SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as failure,
    (SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) * 100.0 / COUNT(*)) as error_rate
FROM request_logs
WHERE created_at >= NOW() - INTERVAL 1 DAY
GROUP BY provider_id;
```

### Account Metrics

**Usage Distribution**:
```sql
-- Account usage by provider
SELECT
    a.provider_id,
    a.label,
    COUNT(r.id) as request_count,
    AVG(r.latency_ms) as avg_latency
FROM accounts a
LEFT JOIN request_logs r ON r.account_id = a.id
WHERE r.created_at >= NOW() - INTERVAL 1 DAY
GROUP BY a.id, a.provider_id, a.label
ORDER BY request_count DESC;
```

**Last Used Tracking**:
```sql
-- Inactive accounts
SELECT id, label, last_used_at
FROM accounts
WHERE is_active = true
  AND last_used_at < NOW() - INTERVAL 7 DAY;
```

### Proxy Metrics

**Health Status**:
```sql
-- Proxy health overview
SELECT
    p.id,
    p.label,
    p.failure_count,
    ps.total_requests,
    ps.failed_requests,
    (ps.failed_requests * 100.0 / ps.total_requests) as error_rate,
    ps.avg_latency
FROM proxy_pool p
LEFT JOIN proxy_stats ps ON ps.proxy_id = p.id AND ps.date = CURDATE()
WHERE p.is_active = true;
```

### OAuth Token Metrics

**Refresh Rate**:
```sql
-- OAuth refresh frequency
SELECT
    provider_id,
    COUNT(*) as refresh_count
FROM request_logs
WHERE error LIKE '%token refresh%'
  AND created_at >= NOW() - INTERVAL 1 DAY
GROUP BY provider_id;
```

**Cache Hit Rate** (manual tracking in Redis):
```
cache_hits = redis.GET("stats:oauth:cache_hits")
cache_misses = redis.GET("stats:oauth:cache_misses")
hit_rate = cache_hits / (cache_hits + cache_misses)
```

## Log Aggregation

### Request Log Structure

```json
{
    "id": 12345,
    "account_id": "acc-123",
    "provider_id": "antigravity",
    "proxy_id": 1,
    "model": "claude-sonnet-4-5",
    "status_code": 200,
    "latency_ms": 234,
    "error": null,
    "created_at": "2024-12-26T10:00:00Z"
}
```

### Useful Queries

**Error Distribution**:
```sql
SELECT
    status_code,
    COUNT(*) as count,
    SUBSTRING(error, 1, 100) as error_sample
FROM request_logs
WHERE status_code >= 400
  AND created_at >= NOW() - INTERVAL 1 DAY
GROUP BY status_code, error_sample
ORDER BY count DESC
LIMIT 20;
```

**Hourly Request Volume**:
```sql
SELECT
    DATE_FORMAT(created_at, '%Y-%m-%d %H:00:00') as hour,
    COUNT(*) as requests,
    AVG(latency_ms) as avg_latency
FROM request_logs
WHERE created_at >= NOW() - INTERVAL 24 HOUR
GROUP BY hour
ORDER BY hour;
```

**Provider Performance Comparison**:
```sql
SELECT
    provider_id,
    COUNT(*) as total_requests,
    AVG(latency_ms) as avg_latency,
    MAX(latency_ms) as max_latency,
    (SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) * 100.0 / COUNT(*)) as error_rate
FROM request_logs
WHERE created_at >= NOW() - INTERVAL 7 DAY
GROUP BY provider_id;
```

## Alerting Recommendations

### Critical Alerts

**High Error Rate**:
```sql
-- Alert if error rate > 5% in last 5 minutes
SELECT COUNT(*) as error_count
FROM request_logs
WHERE status_code >= 400
  AND created_at >= NOW() - INTERVAL 5 MINUTE;

-- Threshold: error_count > 5% of total requests
```

**High Latency**:
```sql
-- Alert if p95 latency > 1000ms
SELECT PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY latency_ms) as p95
FROM request_logs
WHERE created_at >= NOW() - INTERVAL 5 MINUTE;

-- Threshold: p95 > 1000
```

**OAuth Token Refresh Failures**:
```sql
-- Alert if refresh failures detected
SELECT COUNT(*) as refresh_failures
FROM request_logs
WHERE error LIKE '%refresh failed%'
  AND created_at >= NOW() - INTERVAL 5 MINUTE;

-- Threshold: refresh_failures > 0
```

**Proxy Failures**:
```sql
-- Alert if proxy failure rate > 10%
SELECT
    proxy_id,
    (failure_count * 100.0 / (failure_count + 1)) as failure_rate
FROM proxy_pool
WHERE failure_count > max_failures * 0.7;
```

### Warning Alerts

**Account Imbalance**:
```sql
-- Warn if account usage skewed (not round-robin)
SELECT
    provider_id,
    MAX(usage_count) - MIN(usage_count) as imbalance
FROM accounts
WHERE is_active = true
GROUP BY provider_id
HAVING imbalance > 1000;
```

**Cache Performance**:
```
-- Warn if OAuth cache hit rate < 90%
hit_rate = cache_hits / (cache_hits + cache_misses)
if hit_rate < 0.90:
    send_warning("OAuth cache hit rate low")
```

## Dashboard Recommendations

### Real-Time Dashboard

**Metrics to Display**:
- Current requests/second
- Active accounts count
- Active proxies count
- Current error rate (%)
- Average latency (last 5 min)

**Implementation** (Grafana example):
```sql
-- Requests per second (last 5 minutes)
SELECT
    UNIX_TIMESTAMP(created_at) DIV 60 * 60 as time_bucket,
    COUNT(*) / 60 as rps
FROM request_logs
WHERE created_at >= NOW() - INTERVAL 5 MINUTE
GROUP BY time_bucket;
```

### Historical Dashboard

**Metrics to Display**:
- Daily request volume (7 days)
- Error rate trend
- Latency percentiles (p50, p95, p99)
- Provider comparison
- Proxy performance

## Prometheus Integration (Future)

### Planned Metrics Endpoint

**Format**: OpenMetrics/Prometheus
**Path**: `/metrics`

**Example Metrics**:
```
# TYPE aigateway_requests_total counter
aigateway_requests_total{provider="antigravity",status="200"} 12345

# TYPE aigateway_request_duration_seconds histogram
aigateway_request_duration_seconds_bucket{provider="antigravity",le="0.1"} 1000
aigateway_request_duration_seconds_bucket{provider="antigravity",le="0.5"} 8000
aigateway_request_duration_seconds_sum{provider="antigravity"} 2500
aigateway_request_duration_seconds_count{provider="antigravity"} 10000

# TYPE aigateway_oauth_cache_hits_total counter
aigateway_oauth_cache_hits_total{provider="antigravity"} 9900

# TYPE aigateway_proxy_requests_total counter
aigateway_proxy_requests_total{proxy_id="1"} 5000
```

## Log Retention

### Retention Policy

**Request Logs**:
- Hot storage: Last 30 days (MySQL)
- Cold storage: 31-365 days (archive to S3/GCS)
- Deletion: After 1 year

**Proxy Stats**:
- Keep indefinitely (small size)
- Aggregate monthly for long-term trends

**Redis Counters**:
- Persistent (no expiry on stats keys)
- Backup daily to MySQL

### Archival Strategy

```sql
-- Archive old logs monthly
INSERT INTO request_logs_archive
SELECT * FROM request_logs
WHERE created_at < NOW() - INTERVAL 30 DAY;

DELETE FROM request_logs
WHERE created_at < NOW() - INTERVAL 30 DAY;
```

## Related Documentation

- [Architecture Overview](README.md) - System architecture
- [Database Schema](database.md) - Data structures
- [Performance](performance.md) - Optimization strategies
- [Operations Guide](../operations/troubleshooting.md) - Common issues
