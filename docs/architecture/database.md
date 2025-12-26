# Database Schema Design

This document describes the MySQL database schema and Redis data structures used by AIGateway.

## MySQL Schema

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

**Purpose**: Maintain persistent round-robin state across gateway restarts.

**Usage**:
```go
key := fmt.Sprintf("account:rr:%s:%s", providerID, model)
counter, _ := redis.Incr(ctx, key).Result()
index := (counter - 1) % int64(len(accounts))
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

**Purpose**: Cache OAuth tokens to avoid repeated authentication calls.

**Expiration Strategy**:
- TTL set to token's `expires_in` value
- Automatic eviction when token expires
- 5-minute buffer before expiry triggers refresh

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

**Purpose**: Real-time proxy performance monitoring.

**Usage**:
```go
// Increment request counter
redis.Incr(ctx, fmt.Sprintf("stats:proxy:%d:requests", proxyID))

// Increment failure counter on error
if err != nil {
    redis.Incr(ctx, fmt.Sprintf("stats:proxy:%d:failures", proxyID))
}
```

## Database Indexing Strategy

**Optimized Queries**:
- `idx_provider_active` on `accounts(provider_id, is_active)` - Fast account lookup
- `idx_proxy` on `accounts(proxy_id)` - Efficient proxy assignment queries
- `idx_created` on `request_logs(created_at)` - Time-based log queries
- `idx_date` on `proxy_stats(date)` - Daily statistics aggregation

**Query Examples**:
```sql
-- Average latency per proxy
SELECT proxy_id, AVG(latency_ms)
FROM request_logs
GROUP BY proxy_id;

-- Error rate per provider
SELECT provider_id,
       SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) / COUNT(*) as error_rate
FROM request_logs
GROUP BY provider_id;

-- Daily request volume
SELECT DATE(created_at) as date, COUNT(*) as requests
FROM request_logs
WHERE created_at >= NOW() - INTERVAL 7 DAY
GROUP BY DATE(created_at);
```

## Data Integrity

**Foreign Key Constraints**:
- `accounts.provider_id` references `providers.id` - Ensures valid provider
- `accounts.proxy_id` references `proxy_pool.id` - Ensures valid proxy
- `request_logs.*_id` references - Maintains referential integrity

**Unique Constraints**:
- `proxy_pool.proxy_url` - Prevents duplicate proxy URLs
- `proxy_stats(proxy_id, date)` - One stats record per proxy per day

## Related Documentation

- [Architecture Overview](README.md) - System architecture
- [Components](components.md) - Service layer details
- [Database Operations](../operations/database.md) - Setup and seeding
