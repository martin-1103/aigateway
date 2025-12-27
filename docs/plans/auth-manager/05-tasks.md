# Implementation Tasks

**Rule: Setiap file maksimal 300 lines. Lihat `06-file-structure.md` untuk breakdown.**

## Phase 1: Error Parser

**Files to create:** (total ~370 lines, 6 files)
- `auth/errors/types.go` (~50 lines)
- `auth/errors/parser.go` (~30 lines)
- `auth/errors/claude.go` (~80 lines)
- `auth/errors/codex.go` (~80 lines)
- `auth/errors/antigravity.go` (~80 lines)
- `auth/errors/helpers.go` (~50 lines)

**Tasks:**
- [ ] Define ErrorType constants
- [ ] Define ParsedError struct
- [ ] Define ErrorParser interface
- [ ] Implement ClaudeErrorParser
- [ ] Implement CodexErrorParser
- [ ] Implement AntigravityErrorParser
- [ ] Unit tests for each parser

**Dependencies:** None

---

## Phase 2: Auth Manager Core

**Files to create:** (total ~510 lines, 5 files)
- `auth/manager/types.go` (~60 lines)
- `auth/manager/account_state.go` (~80 lines)
- `auth/manager/manager.go` (~150 lines)
- `auth/manager/selector.go` (~100 lines)
- `auth/manager/refresh.go` (~120 lines)

**Tasks:**
- [ ] Define BlockReason constants
- [ ] Define ModelState struct
- [ ] Define QuotaState with backoff logic
- [ ] Define AccountState struct
- [ ] Implement Manager.Select()
- [ ] Implement Manager.MarkResult()
- [ ] Implement isBlockedForModel()
- [ ] Implement round-robin selection
- [ ] Unit tests for manager

**Dependencies:** Phase 1

---

## Phase 3: Token Refresh

**Files to create:** (total ~410 lines, 5 files)
- `auth/claude/types.go` (~40 lines)
- `auth/claude/auth.go` (~120 lines)
- `auth/codex/types.go` (~50 lines)
- `auth/codex/auth.go` (~120 lines)
- `auth/codex/jwt.go` (~80 lines)

**Tasks:**
- [ ] Define TokenRefresher interface
- [ ] Define TokenResult struct
- [ ] Implement ClaudeRefresher
- [ ] Implement CodexRefresher
- [ ] Implement JWT parsing for Codex
- [ ] Implement RefreshCoordinator
- [ ] Update TokenRefreshService for multi-provider
- [ ] Integration tests

**Dependencies:** Phase 2

---

## Phase 4: Integration

**Files to update:**
- `services/router.service.go`
- `cmd/main.go`

**Tasks:**
- [ ] Add AuthManager to RouterService
- [ ] Update Execute() to use Select()
- [ ] Add MarkResult() calls after execution
- [ ] Add retry logic with account switching
- [ ] Add StatusError interface to providers
- [ ] Update initialization in main.go
- [ ] End-to-end tests

**Dependencies:** Phase 3

---

## Phase 5: Observability

**Tasks:**
- [ ] Add metrics: account_health (per account per model)
- [ ] Add metrics: rotation_count (per provider)
- [ ] Add metrics: cooldown_events (per reason)
- [ ] Add logging for state changes
- [ ] Create dashboard for account status

**Dependencies:** Phase 4

---

## Testing Strategy

### Unit Tests
- Error parser parsing logic
- QuotaState backoff calculation
- Manager state updates
- Token refresh logic

### Integration Tests
- Select with blocked accounts
- MarkResult state transitions
- Retry with account switching
- Token refresh cycle

### E2E Tests
- Full request flow with rotation
- Recovery after rate limit
- Recovery after token refresh

---

## Configuration

Add to `config/config.yaml`:

```yaml
auth_manager:
  refresh_interval: 30s

  refresh_lead:
    antigravity: 5m
    claude: 5m
    codex: 5m

  cooldown:
    auth_failure: 30m
    rate_limit: 5s
    quota_exceeded: 1h
    transient: 1m

  quota_backoff:
    base: 1s
    max: 30m

  retry:
    max_attempts: 3
    max_wait: 30s
```

---

## Migration Plan

1. Deploy new code with feature flag OFF
2. Enable for one provider (e.g., antigravity)
3. Monitor error rates
4. Enable for remaining providers
5. Remove old AccountService after validation

---

## Success Metrics

- Reduced 429 error rate
- Faster recovery from errors
- Better token freshness
- Per-account health visibility
