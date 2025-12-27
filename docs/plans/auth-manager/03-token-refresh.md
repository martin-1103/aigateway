# Token Refresh Design

## Purpose

Multi-provider token refresh with per-provider logic.
Background job checks expiring tokens and refreshes before expiry.

## Current State

| Provider | Status | Notes |
|----------|--------|-------|
| Antigravity | ✅ Implemented | Uses Google OAuth |
| Claude | ❌ Missing | Needs new implementation |
| Codex | ❌ Missing | Needs new implementation |
| OpenAI | N/A | API key, no refresh |
| GLM | N/A | Bearer token, no refresh |

## Interface

```go
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
```

## Provider: Claude

**Token URL:** `https://console.anthropic.com/v1/oauth/token`
**Client ID:** `9d1c250a-e61b-44d9-88ed-5944d1962f5e`
**RefreshLead:** 5 minutes
**Body format:** JSON

```json
{
    "client_id": "...",
    "grant_type": "refresh_token",
    "refresh_token": "..."
}
```

**Response:**
```json
{
    "access_token": "...",
    "refresh_token": "...",
    "expires_in": 3600,
    "account": {"email_address": "..."}
}
```

## Provider: Codex

**Token URL:** `https://auth.openai.com/oauth/token`
**Client ID:** `app_EMoamEEZ73f0CkXaXp7hrann`
**RefreshLead:** 5 minutes
**Body format:** form-urlencoded

```
client_id=...&grant_type=refresh_token&refresh_token=...&scope=openid profile email
```

**Response:**
```json
{
    "access_token": "...",
    "refresh_token": "...",
    "id_token": "...",
    "expires_in": 3600
}
```

**Note:** Parse `id_token` JWT to get `account_id` and `email`.

## Provider: Antigravity

**Token URL:** `https://oauth2.googleapis.com/token`
**Client ID/Secret:** Google Cloud OAuth credentials
**RefreshLead:** 5 minutes (existing: 10 min)
**Body format:** form-urlencoded

Already implemented in `services/token.refresh.service.go`.

## Background Refresh Logic

```go
func (s *TokenRefreshService) checkRefreshes() {
    for _, acc := range s.manager.GetAllAccounts() {
        if acc.Disabled {
            continue
        }

        refresher := s.getRefresher(acc.Account.ProviderID)
        if refresher == nil {
            continue // No refresh for this provider
        }

        expiresAt := s.getExpiryTime(acc.Account)
        lead := refresher.RefreshLead()

        if time.Until(expiresAt) <= lead {
            go s.refreshAccount(acc, refresher)
        }
    }
}
```

## Failure Handling

```go
func (s *TokenRefreshService) refreshAccount(acc *AccountState, refresher TokenRefresher) {
    result, err := refresher.Refresh(ctx, acc.Account)
    if err != nil {
        // Set NextRefreshAfter with backoff (30s)
        acc.NextRefreshAfter = time.Now().Add(30 * time.Second)
        acc.LastError = err
        return
    }

    // Update account auth data
    s.updateAccountAuth(acc.Account, result)
    acc.LastRefreshedAt = time.Now()
    acc.NextRefreshAfter = time.Time{}
}
```

## Implementation Files

- `auth/claude/auth.go` - Claude refresher
- `auth/codex/auth.go` - Codex refresher
- `auth/codex/jwt.go` - JWT parsing for account_id
- `auth/manager/refresh.go` - Refresh coordinator
- `services/token.refresh.service.go` - Update for multi-provider
