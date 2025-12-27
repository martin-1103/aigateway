# Error Parser Design

## Purpose

Parse error response body to get accurate error classification.
Status code alone is not enough (e.g., 429 can be rate_limit or quota_exceeded).

## Error Types

```go
type ErrorType string

const (
    ErrTypeRateLimit      ErrorType = "rate_limit"       // Retry quickly
    ErrTypeQuotaExceeded  ErrorType = "quota_exceeded"   // Wait long
    ErrTypeAuthentication ErrorType = "authentication"   // Disable account
    ErrTypePermission     ErrorType = "permission"       // Disable account
    ErrTypeNotFound       ErrorType = "not_found"        // Disable for model
    ErrTypeOverloaded     ErrorType = "overloaded"       // Retry with backoff
    ErrTypeTransient      ErrorType = "transient"        // Retry quickly
)
```

## Interface

```go
// auth/errors/parser.go

type ParsedError struct {
    Type        ErrorType
    StatusCode  int
    Message     string
    Retryable   bool
    CooldownDur time.Duration
}

type ErrorParser interface {
    Parse(statusCode int, body []byte) *ParsedError
}
```

## Provider: Claude

**Error format:**
```json
{"error": {"type": "rate_limit_error", "message": "..."}}
```

**Mapping:**

| Status | error.type | ErrorType | Cooldown |
|--------|-----------|-----------|----------|
| 401 | authentication_error | Authentication | 30m |
| 403 | permission_error | Permission | 30m |
| 429 | rate_limit_error | RateLimit | parse Retry-After |
| 500 | api_error | Transient | 1m |
| 529 | overloaded_error | Overloaded | 30s |

## Provider: Codex/OpenAI

**Error format:**
```json
{"error": {"type": "...", "code": "insufficient_quota", "message": "..."}}
```

**Key distinction for 429:**

| error.code | ErrorType | Cooldown |
|-----------|-----------|----------|
| insufficient_quota | QuotaExceeded | 24h |
| rate_limit_exceeded | RateLimit | 5s |

## Provider: Antigravity/Google

**Error format:**
```json
{"error": {"status": "...", "message": "...", "details": [{"reason": "QUOTA_EXCEEDED"}]}}
```

**Key distinction for 429:**

| error.details.reason | ErrorType | Cooldown |
|---------------------|-----------|----------|
| RATE_LIMIT_EXCEEDED | RateLimit | 1s |
| QUOTA_EXCEEDED | QuotaExceeded | 1h |
| USER_RATE_LIMIT_EXCEEDED | RateLimit | 10s |

## Implementation Files

- `auth/errors/parser.go` - Interface & types
- `auth/errors/claude.go` - Claude parser
- `auth/errors/codex.go` - Codex parser
- `auth/errors/antigravity.go` - Antigravity parser
