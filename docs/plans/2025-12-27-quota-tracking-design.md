# Quota Tracking Design: Antigravity Provider

**Date:** 2025-12-27
**Status:** Draft
**Scope:** Antigravity provider (Google AI Pro accounts)

---

## Overview

Sistem untuk tracking usage quota per account per model, dengan kemampuan belajar dari pattern penggunaan. Menggunakan pendekatan **hybrid**: optimistic start + learning from quota exhaustion events.

## Goals

1. Track usage (requests + tokens) per account per model
2. Detect quota exhaustion dari error response (429 QUOTA_EXCEEDED)
3. Learn quota limits dari historical exhaustion patterns
4. Prevent requests ke exhausted accounts
5. Provide visibility via monitoring API

## Non-Goals

- Exact quota prediction (Google tidak publish angka pasti)
- Token counting dari request (hanya dari response)
- Support provider lain (fase 1 = Antigravity only)

---

## Research Findings

### Antigravity Quota Behavior

| Aspect | Detail |
|--------|--------|
| Scope | Per Google Account (bukan per project/API key) |
| Reset Window | 5 jam untuk AI Pro subscriber |
| Models | Gemini 3 Pro, Claude 4.5 Sonnet, GPT-OSS-120b |
| Measurement | "Work done" - correlates dengan token usage |
| Exact Limits | Tidak dipublish, perlu learn dari behavior |

### Error Response Format

```json
{
  "error": {
    "status": "RESOURCE_EXHAUSTED",
    "message": "Resource exhausted, please try again later.",
    "details": [{"reason": "QUOTA_EXCEEDED"}]
  }
}
```

**Sources:**
- https://blog.google/feed/new-antigravity-rate-limits-pro-ultra-subsribers/
- https://discuss.ai.google.dev/t/antigravity-quota-refresh/112317
- https://ai.google.dev/gemini-api/docs/rate-limits

---

## Architecture

### Storage Strategy

**Redis** - Real-time counters (fast, auto-expire)
```
quota:{account_id}:{model}:requests   → INT (TTL 5 hours)
quota:{account_id}:{model}:tokens     → INT (TTL 5 hours)
quota:{account_id}:{model}:exhausted  → BOOL (TTL 5 hours)
```

**MySQL** - Learned patterns (persistent)
```sql
CREATE TABLE account_quota_pattern (
    id                  BIGINT PRIMARY KEY AUTO_INCREMENT,
    account_id          VARCHAR(36) NOT NULL,
    model               VARCHAR(100) NOT NULL,

    -- Learned thresholds
    est_request_limit   INT NULL,
    est_token_limit     BIGINT NULL,
    confidence          FLOAT DEFAULT 0,
    sample_count        INT DEFAULT 0,

    -- State tracking
    last_exhausted_at   DATETIME NULL,
    last_reset_at       DATETIME NULL,

    created_at          DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at          DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    UNIQUE KEY idx_account_model (account_id, model)
);
```

### Why Hybrid Storage?

| Aspect | MySQL Only | Redis + MySQL |
|--------|------------|---------------|
| Counter updates | Write per request | In-memory INCR |
| Window reset | Manual/cron | Auto TTL expiry |
| Failure mode | DB bottleneck | Graceful degradation |
| Learned data | Mixed concerns | Clean separation |

---

## Service Layer

### QuotaTrackerService

```go
type QuotaTrackerService struct {
    redis     *redis.Client
    repo      *QuotaPatternRepository
    windowTTL time.Duration  // 5 hours
}

// Core Methods
func (s *QuotaTrackerService) RecordUsage(accountID, model string, tokens int64)
func (s *QuotaTrackerService) MarkExhausted(accountID, model string)
func (s *QuotaTrackerService) GetRemainingQuota(accountID, model string) *QuotaStatus
func (s *QuotaTrackerService) IsAvailable(accountID, model string) bool
func (s *QuotaTrackerService) GetEarliestReset(providerID, model string) time.Time
```

### QuotaStatus Response

```go
type QuotaStatus struct {
    RequestsUsed    int
    TokensUsed      int64
    EstRequestLimit *int       // nil = unknown
    EstTokenLimit   *int64     // nil = unknown
    PercentUsed     *float64   // nil if limit unknown
    IsExhausted     bool
    ResetsAt        *time.Time
}
```

---

## Integration Points

### 1. AuthManager.Select() - Filter Exhausted Accounts

```go
func (m *AuthManager) Select(ctx context.Context, providerID, model string) (*AccountState, error) {
    accounts := m.getActiveAccounts(providerID)

    // Filter out exhausted accounts
    available := make([]*Account, 0)
    for _, acc := range accounts {
        if m.quotaTracker.IsAvailable(acc.ID, model) {
            available = append(available, acc)
        }
    }

    if len(available) == 0 {
        resetAt := m.quotaTracker.GetEarliestReset(providerID, model)
        return nil, &AllExhaustedError{ResetAt: resetAt}
    }

    return m.roundRobinSelect(available, model)
}
```

### 2. AuthManager.MarkResult() - Track Usage & Exhaustion

```go
func (m *AuthManager) MarkResult(accountID, model string, statusCode int, payload []byte) {
    parsed := m.errorParser.Parse(statusCode, payload)

    // Track successful usage
    if statusCode >= 200 && statusCode < 300 {
        tokens := extractTokenUsage(payload)
        m.quotaTracker.RecordUsage(accountID, model, tokens)
    }

    // Learn from quota exhaustion
    if parsed.Type == ErrTypeQuotaExceeded {
        m.quotaTracker.MarkExhausted(accountID, model)
    }
}
```

### 3. Token Extraction from Antigravity Response

```go
// Response format:
// {"usageMetadata": {"promptTokenCount": 10, "candidatesTokenCount": 50, "totalTokenCount": 60}}

func extractTokenUsage(payload []byte) int64 {
    total := gjson.GetBytes(payload, "usageMetadata.totalTokenCount").Int()
    if total > 0 {
        return total
    }
    prompt := gjson.GetBytes(payload, "usageMetadata.promptTokenCount").Int()
    candidates := gjson.GetBytes(payload, "usageMetadata.candidatesTokenCount").Int()
    if prompt+candidates > 0 {
        return prompt + candidates
    }
    // Fallback: estimate from payload size (~4 chars per token)
    return int64(len(payload) / 4)
}
```

---

## Learning Algorithm

### On Quota Exhaustion (MarkExhausted)

```go
func (s *QuotaTrackerService) MarkExhausted(accountID, model string) {
    // 1. Get current usage from Redis
    requests, _ := s.redis.Get(ctx, s.reqKey(accountID, model)).Int()
    tokens, _ := s.redis.Get(ctx, s.tokenKey(accountID, model)).Int64()

    // 2. Update learned pattern
    pattern := s.repo.GetOrCreate(accountID, model)

    if pattern.EstRequestLimit == nil {
        // First time - set directly
        pattern.EstRequestLimit = &requests
        pattern.EstTokenLimit = &tokens
    } else {
        // Weighted average with existing estimate
        weight := pattern.Confidence
        pattern.EstRequestLimit = weightedAvg(*pattern.EstRequestLimit, requests, weight)
        pattern.EstTokenLimit = weightedAvg(*pattern.EstTokenLimit, tokens, weight)
    }

    pattern.SampleCount++
    pattern.Confidence = min(1.0, float64(pattern.SampleCount) / 10.0)
    pattern.LastExhaustedAt = time.Now()

    s.repo.Save(pattern)

    // 3. Mark as exhausted in Redis
    s.redis.Set(ctx, s.exhaustedKey(accountID, model), true, s.windowTTL)
}

func weightedAvg(old, new int, weight float64) int {
    return int((float64(old)*weight + float64(new)) / (weight + 1))
}
```

### Confidence Decay (Stale Data)

```go
func (s *QuotaTrackerService) GetLearnedLimit(accountID, model string) *int {
    pattern := s.repo.Get(accountID, model)
    if pattern == nil {
        return nil
    }

    // Decay confidence if data is old
    daysSinceLastHit := time.Since(pattern.LastExhaustedAt).Hours() / 24
    if daysSinceLastHit > 7 {
        pattern.Confidence *= 0.5
    }

    if pattern.Confidence < 0.3 {
        return nil  // Too stale
    }
    return pattern.EstRequestLimit
}
```

---

## API Endpoints

### Monitoring Endpoints

```
GET /api/v1/quota/accounts                     # List all accounts with quota status
GET /api/v1/quota/accounts/{id}                # Single account detail
GET /api/v1/quota/accounts/{id}/history        # Exhaustion history
GET /api/v1/quota/providers/{provider}/summary # Provider-level summary
```

### Response: Account Quota Status

```json
{
  "accounts": [
    {
      "account_id": "abc-123",
      "label": "user1@gmail.com",
      "provider_id": "antigravity",
      "models": {
        "gemini-2.5-pro": {
          "requests_used": 45,
          "tokens_used": 128000,
          "est_request_limit": 100,
          "est_token_limit": 500000,
          "percent_used": 45.0,
          "is_exhausted": false,
          "resets_at": "2025-12-27T15:00:00Z",
          "confidence": 0.8
        }
      }
    }
  ]
}
```

### Response: Provider Summary

```json
{
  "provider_id": "antigravity",
  "total_accounts": 5,
  "available_accounts": 3,
  "exhausted_accounts": 2,
  "models": {
    "gemini-2.5-pro": {
      "total": 5,
      "available": 3,
      "exhausted": 2,
      "avg_percent_used": 67.5,
      "next_reset_at": "2025-12-27T14:30:00Z"
    }
  },
  "health": "degraded"
}
```

### Health Status Logic

```go
func calculateHealth(available, total int) string {
    ratio := float64(available) / float64(total)
    switch {
    case ratio >= 0.5:
        return "healthy"
    case ratio >= 0.2:
        return "degraded"
    default:
        return "critical"
    }
}
```

---

## Edge Cases & Mitigations

### 1. Redis Unavailable
**Strategy:** Fail open (optimistic)
```go
if err != nil {
    log.Warn("Redis unavailable, assuming quota available")
    return true
}
```

### 2. Token Extraction Fails
**Strategy:** Estimate from payload size
```go
if tokens == 0 {
    return int64(len(payload) / 4)  // ~4 chars per token
}
```

### 3. Streaming Response
**Strategy:** Token count dari final chunk, atau estimate untuk stream requests.

### 4. Window Boundary Race
**Strategy:** Atomic Redis operations with pipeline.

### 5. Retry Double-Counting
**Strategy:** Only record usage on final success, not intermediate attempts.

### 6. Stale Learned Limits
**Strategy:** Confidence decay over time (>7 days = 50% decay).

### 7. All Accounts Exhausted
**Strategy:** Return 429 with Retry-After header.
```go
c.Header("Retry-After", fmt.Sprintf("%d", waitSeconds))
c.JSON(429, gin.H{
    "error": "All accounts exhausted",
    "retry_after_seconds": waitSeconds,
    "next_available_at": resetAt,
})
```

---

## File Structure

```
services/
├── quota.tracker.service.go     # Core quota tracking logic
├── quota.pattern.service.go     # Learning algorithm

repositories/
├── quota.pattern.repository.go  # MySQL CRUD for patterns

handlers/
├── quota.handler.go             # API endpoints

models/
├── quota.model.go               # QuotaPattern, QuotaStatus structs
```

---

## Implementation Order

1. **Phase 1: Data Layer**
   - Create `account_quota_pattern` table
   - Add Redis key patterns
   - QuotaPatternRepository

2. **Phase 2: Service Layer**
   - QuotaTrackerService (RecordUsage, MarkExhausted, IsAvailable)
   - Token extraction helper

3. **Phase 3: Integration**
   - Modify AuthManager.Select() for filtering
   - Modify AuthManager.MarkResult() for tracking
   - Handle AllExhaustedError

4. **Phase 4: API**
   - QuotaHandler endpoints
   - Response models

5. **Phase 5: Testing**
   - Unit tests for learning algorithm
   - Integration tests for quota flow
