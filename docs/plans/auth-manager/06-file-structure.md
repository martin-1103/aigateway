# File Structure & Size Limits

## Rule

**Setiap file implementasi maksimal 300 lines.**
Jika lebih, break jadi beberapa file yang focused.

---

## auth/errors/ (Error Parsing)

```
auth/errors/
├── types.go          (~50 lines)   - ErrorType constants, ParsedError struct
├── parser.go         (~30 lines)   - ErrorParser interface
├── claude.go         (~80 lines)   - Claude error parser
├── codex.go          (~80 lines)   - Codex error parser
├── antigravity.go    (~80 lines)   - Antigravity error parser
└── helpers.go        (~50 lines)   - Shared parsing helpers
```

### types.go (~50 lines)
```go
package errors

type ErrorType string

const (
    ErrTypeRateLimit      ErrorType = "rate_limit"
    ErrTypeQuotaExceeded  ErrorType = "quota_exceeded"
    ErrTypeAuthentication ErrorType = "authentication"
    ErrTypePermission     ErrorType = "permission"
    ErrTypeNotFound       ErrorType = "not_found"
    ErrTypeOverloaded     ErrorType = "overloaded"
    ErrTypeTransient      ErrorType = "transient"
    ErrTypeUnknown        ErrorType = "unknown"
)

type ParsedError struct {
    Type        ErrorType
    StatusCode  int
    Message     string
    Retryable   bool
    CooldownDur time.Duration
    RawBody     []byte
}

func (e *ParsedError) Error() string {
    return e.Message
}
```

### parser.go (~30 lines)
```go
package errors

type ErrorParser interface {
    Parse(statusCode int, body []byte) *ParsedError
}

func GetParser(providerID string) ErrorParser {
    switch providerID {
    case "claude", "anthropic":
        return &ClaudeParser{}
    case "codex", "openai":
        return &CodexParser{}
    case "antigravity":
        return &AntigravityParser{}
    default:
        return &DefaultParser{}
    }
}
```

---

## auth/manager/ (State Management)

```
auth/manager/
├── types.go          (~60 lines)   - BlockReason, ModelState, QuotaState
├── account_state.go  (~80 lines)   - AccountState struct & methods
├── manager.go        (~150 lines)  - Manager struct, Select, MarkResult
├── selector.go       (~100 lines)  - Selection logic, round-robin
└── refresh.go        (~120 lines)  - Background refresh coordinator
```

### types.go (~60 lines)
```go
package manager

type BlockReason string

const (
    BlockReasonNone     BlockReason = ""
    BlockReasonDisabled BlockReason = "disabled"
    BlockReasonCooldown BlockReason = "cooldown"
    BlockReasonQuota    BlockReason = "quota"
    BlockReasonAuth     BlockReason = "auth_failed"
)

type ModelState struct {
    Model          string
    Disabled       bool
    BlockReason    BlockReason
    NextRetryAfter time.Time
    LastError      *errors.ParsedError
    SuccessCount   int64
    FailureCount   int64
}

type QuotaState struct {
    BackoffMultiplier int
    BaseBackoff       time.Duration
    MaxBackoff        time.Duration
}

func (q *QuotaState) NextBackoff() time.Duration { ... }
func (q *QuotaState) Increment() { ... }
func (q *QuotaState) Reset() { ... }
```

### account_state.go (~80 lines)
```go
package manager

type AccountState struct {
    Account     *models.Account
    ModelStates map[string]*ModelState
    QuotaState  *QuotaState
    Disabled    bool
    UpdatedAt   time.Time

    // Refresh tracking
    LastRefreshedAt  time.Time
    NextRefreshAfter time.Time
}

func NewAccountState(account *models.Account) *AccountState { ... }
func (a *AccountState) GetModelState(model string) *ModelState { ... }
func (a *AccountState) IsBlockedFor(model string, now time.Time) (bool, BlockReason) { ... }
```

### manager.go (~150 lines)
```go
package manager

type Manager struct {
    accounts     map[string]*AccountState
    mu           sync.RWMutex
    accountRepo  *repositories.AccountRepository
    redis        *redis.Client
    errorParsers map[string]errors.ErrorParser
    refreshers   map[string]TokenRefresher
}

func NewManager(...) *Manager { ... }
func (m *Manager) LoadAccounts(ctx context.Context) error { ... }
func (m *Manager) Select(ctx context.Context, providerID, model string) (*AccountState, error) { ... }
func (m *Manager) MarkResult(accountID, model string, statusCode int, body []byte) { ... }
func (m *Manager) RegisterParser(providerID string, parser errors.ErrorParser) { ... }
func (m *Manager) RegisterRefresher(providerID string, refresher TokenRefresher) { ... }
```

### selector.go (~100 lines)
```go
package manager

func (m *Manager) getCandidates(providerID, model string) []*AccountState { ... }
func (m *Manager) filterAvailable(candidates []*AccountState, model string, now time.Time) []*AccountState { ... }
func (m *Manager) roundRobinSelect(available []*AccountState, providerID, model string) *AccountState { ... }
func (m *Manager) getRedisCounter(providerID, model string) int64 { ... }
```

### refresh.go (~120 lines)
```go
package manager

type TokenRefresher interface {
    RefreshLead() time.Duration
    Refresh(ctx context.Context, account *models.Account) (*TokenResult, error)
}

type TokenResult struct {
    AccessToken  string
    RefreshToken string
    ExpiresAt    time.Time
    Metadata     map[string]interface{}
}

func (m *Manager) StartAutoRefresh(ctx context.Context, interval time.Duration) { ... }
func (m *Manager) StopAutoRefresh() { ... }
func (m *Manager) checkRefreshes(ctx context.Context) { ... }
func (m *Manager) shouldRefresh(acc *AccountState, now time.Time) bool { ... }
func (m *Manager) refreshAccount(ctx context.Context, acc *AccountState) { ... }
```

---

## auth/claude/ (Claude OAuth)

```
auth/claude/
├── auth.go           (~120 lines)  - ClaudeRefresher, RefreshTokens
└── types.go          (~40 lines)   - TokenData, TokenResponse structs
```

---

## auth/codex/ (Codex OAuth)

```
auth/codex/
├── auth.go           (~120 lines)  - CodexRefresher, RefreshTokens
├── jwt.go            (~80 lines)   - JWT parsing for account_id
└── types.go          (~50 lines)   - TokenData, Claims structs
```

---

## Summary

| Directory | Files | Max Lines/File |
|-----------|-------|----------------|
| auth/errors/ | 6 | ~80 |
| auth/manager/ | 5 | ~150 |
| auth/claude/ | 2 | ~120 |
| auth/codex/ | 3 | ~120 |
| **Total** | **16 files** | **< 300 each** |

---

## Naming Convention

- `types.go` - Structs & constants only
- `{feature}.go` - Main logic (e.g., `auth.go`, `manager.go`)
- `{sub-feature}.go` - Focused logic (e.g., `selector.go`, `refresh.go`)
- `helpers.go` - Shared utility functions
