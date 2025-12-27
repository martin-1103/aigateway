# Auth Manager Design

## Purpose

Centralized state management for all accounts.
Tracks health, cooldown, and quota per account per model.

## Core Structs

### BlockReason

```go
type BlockReason string

const (
    BlockReasonNone     BlockReason = ""
    BlockReasonDisabled BlockReason = "disabled"
    BlockReasonCooldown BlockReason = "cooldown"
    BlockReasonQuota    BlockReason = "quota"
    BlockReasonAuth     BlockReason = "auth_failed"
)
```

### ModelState

Tracks state per account per model:

```go
type ModelState struct {
    Model          string
    Disabled       bool
    BlockReason    BlockReason
    NextRetryAfter time.Time
    LastError      *ParsedError
    SuccessCount   int64
    FailureCount   int64
}
```

### QuotaState

Exponential backoff for quota errors:

```go
type QuotaState struct {
    BackoffMultiplier int           // 1, 2, 4, 8...
    BaseBackoff       time.Duration // 1s
    MaxBackoff        time.Duration // 30m
}

func (q *QuotaState) NextBackoff() time.Duration {
    backoff := q.BaseBackoff * time.Duration(q.BackoffMultiplier)
    if backoff > q.MaxBackoff {
        return q.MaxBackoff
    }
    return backoff
}
```

### AccountState

Wraps account with state:

```go
type AccountState struct {
    Account     *models.Account
    ModelStates map[string]*ModelState
    QuotaState  *QuotaState
    Disabled    bool
}
```

## Manager Methods

### Select

Pick best available account:

```go
func (m *Manager) Select(ctx context.Context, providerID, model string) (*AccountState, error) {
    // 1. Get candidates for provider
    // 2. Filter out blocked accounts
    // 3. Round-robin among available
    // 4. If all blocked, return earliest retry time
}
```

### MarkResult

Update state after execution:

```go
func (m *Manager) MarkResult(accountID, model string, statusCode int, body []byte) {
    // 1. Get error parser for provider
    // 2. Parse error
    // 3. Update ModelState based on error type
}
```

## State Update Logic

### On Success
```go
ms.BlockReason = BlockReasonNone
ms.NextRetryAfter = time.Time{}
ms.SuccessCount++
quotaState.ResetBackoff()
```

### On Rate Limit (429 + rate_limit)
```go
ms.BlockReason = BlockReasonCooldown
ms.NextRetryAfter = now.Add(parsedError.CooldownDur)
```

### On Quota Exceeded (429 + quota)
```go
ms.BlockReason = BlockReasonQuota
quotaState.IncrementBackoff()
ms.NextRetryAfter = now.Add(quotaState.NextBackoff())
```

### On Auth Error (401/403)
```go
ms.BlockReason = BlockReasonAuth
account.Disabled = true
ms.NextRetryAfter = now.Add(30 * time.Minute)
```

## Blocking Check

```go
func (m *Manager) isBlockedForModel(acc *AccountState, model string, now time.Time) bool {
    if acc.Disabled {
        return true
    }
    ms := acc.ModelStates[model]
    if ms == nil {
        return false
    }
    if ms.Disabled {
        return true
    }
    if !ms.NextRetryAfter.IsZero() && now.Before(ms.NextRetryAfter) {
        return true
    }
    return false
}
```

## Implementation Files

- `auth/manager/state.go` - State structs
- `auth/manager/manager.go` - Manager with Select/MarkResult
- `auth/manager/selector.go` - Selection logic
