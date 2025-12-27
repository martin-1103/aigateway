# Integration Design

## Purpose

Integrate AuthManager with existing services.
Update request flow to use health-aware selection and result tracking.

## Updated Request Flow

```
1. Request arrives at RouterService
2. RouterService calls AuthManager.Select(provider, model)
3. AuthManager returns healthy account (or error if all blocked)
4. RouterService executes request with selected account
5. RouterService calls AuthManager.MarkResult(account, status, body)
6. If failed + retryable, go back to step 2 (different account selected)
```

## RouterService Changes

```go
func (s *RouterService) Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error) {
    // 1. Select account
    accState, err := s.authManager.Select(ctx, req.ProviderID, req.Model)
    if err != nil {
        if allBlocked, ok := err.(*AllBlockedError); ok {
            // All accounts blocked - wait or return error
            if allBlocked.WaitDuration < 30*time.Second {
                time.Sleep(allBlocked.WaitDuration)
                return s.Execute(ctx, req) // Retry
            }
        }
        return nil, err
    }

    // 2. Execute
    resp, execErr := s.executeWithProvider(ctx, accState.Account, req)

    // 3. Mark result
    if execErr != nil {
        statusCode, body := extractErrorDetails(execErr)
        s.authManager.MarkResult(accState.Account.ID, req.Model, statusCode, body)
    } else {
        s.authManager.MarkResult(accState.Account.ID, req.Model, resp.StatusCode, resp.Payload)
    }

    // 4. Retry if needed
    if execErr != nil && s.isRetryable(execErr) {
        return s.Execute(ctx, req)
    }

    return resp, execErr
}
```

## AllBlockedError

When all accounts are blocked:

```go
type AllBlockedError struct {
    WaitDuration time.Duration
    Message      string
}

func (e *AllBlockedError) Error() string {
    return e.Message
}
```

## Retry Logic

Conditions for retry:
- Error is retryable (rate limit, transient, overloaded)
- Have not exceeded max retry attempts
- At least one other account available

```go
func (s *RouterService) isRetryable(err error) bool {
    if statusErr, ok := err.(StatusError); ok {
        switch statusErr.StatusCode() {
        case 429, 500, 502, 503, 504, 529:
            return true
        }
    }
    return false
}
```

## Provider Changes

Providers should return errors with status code and body:

```go
type StatusError interface {
    error
    StatusCode() int
    Body() []byte
}

type ProviderError struct {
    Code    int
    RawBody []byte
    Msg     string
}

func (e *ProviderError) Error() string      { return e.Msg }
func (e *ProviderError) StatusCode() int    { return e.Code }
func (e *ProviderError) Body() []byte       { return e.RawBody }
```

## Initialization

```go
// cmd/main.go

func main() {
    // Create AuthManager
    authManager := manager.NewManager(accountRepo, redisClient)

    // Register error parsers
    authManager.RegisterParser("claude", &errors.ClaudeErrorParser{})
    authManager.RegisterParser("codex", &errors.CodexErrorParser{})
    authManager.RegisterParser("antigravity", &errors.AntigravityErrorParser{})

    // Register token refreshers
    authManager.RegisterRefresher("claude", claude.NewRefresher(httpClient))
    authManager.RegisterRefresher("codex", codex.NewRefresher(httpClient))
    authManager.RegisterRefresher("antigravity", antigravity.NewRefresher(httpClient))

    // Load accounts
    authManager.LoadAccounts(ctx)

    // Start background refresh
    authManager.StartAutoRefresh(ctx, 30*time.Second)

    // Create services with AuthManager
    routerService := services.NewRouterService(authManager, ...)
}
```

## Backward Compatibility

For gradual migration:

```go
type RouterService struct {
    authManager    *manager.Manager
    accountService *AccountService  // Keep existing
    useAuthManager bool             // Feature flag
}

func (s *RouterService) selectAccount(ctx context.Context, providerID, model string) (*models.Account, error) {
    if s.useAuthManager {
        accState, err := s.authManager.Select(ctx, providerID, model)
        if err != nil {
            return nil, err
        }
        return accState.Account, nil
    }
    // Fallback to existing
    return s.accountService.GetNextAccount(ctx, providerID, model)
}
```
