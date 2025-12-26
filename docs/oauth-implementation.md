# OAuth Flow Implementation

## Overview

This document describes the OAuth 2.0 authorization flow implementation for aigateway. The implementation supports PKCE (Proof Key for Code Exchange) for secure authentication with multiple AI providers.

## Architecture

### Components

```
auth/pkce/pkce.go              - PKCE code generation (verifier + challenge)
auth/oauth/providers.go        - OAuth configs for Antigravity, Codex, Claude
services/oauth.flow.service.go - OAuth flow orchestration service
handlers/oauth.handler.go      - HTTP endpoints for OAuth flow
```

### Flow Types

**Automatic Flow:**
1. Frontend calls `/api/v1/oauth/init` to start flow
2. User opens auth URL in popup window
3. After consent, provider redirects to `/api/v1/oauth/callback`
4. Backend exchanges code for token, saves account
5. HTML response closes popup and notifies parent window
6. Frontend refreshes account list

**Manual Flow:**
1. Frontend calls `/api/v1/oauth/init` to start flow
2. User opens auth URL in popup window
3. After consent, user copies callback URL from browser
4. User pastes URL in frontend form
5. Frontend calls `/api/v1/oauth/exchange` with callback URL
6. Backend exchanges code for token, saves account
7. Frontend shows success message

## API Endpoints

### POST /api/v1/oauth/init

Start OAuth authorization flow.

**Request:**
```json
{
  "provider": "antigravity",
  "account_name": "My Google Account",
  "flow_type": "auto",
  "redirect_uri": "http://localhost:8088/api/v1/oauth/callback"
}
```

**Response:**
```json
{
  "auth_url": "https://accounts.google.com/o/oauth2/v2/auth?...",
  "state": "abc123-xyz",
  "flow_type": "auto",
  "expires_at": "2025-12-27T10:30:00Z"
}
```

**Flow Types:**
- `auto` - Provider redirects to callback endpoint (popup closes automatically)
- `manual` - User copies callback URL and pastes in frontend

**Redirect URI:**
- Optional field
- Defaults to `http://localhost:8088/api/v1/oauth/callback`
- Must match provider configuration

### GET /api/v1/oauth/callback

OAuth callback endpoint for automatic flow. Provider redirects here after user consent.

**Query Parameters:**
- `code` - Authorization code from provider
- `state` - State parameter for session validation

**Response:**
- HTML page that closes popup and notifies parent window
- Success: Green UI with checkmark, sends postMessage to opener
- Error: Red UI with error message, sends error postMessage

### POST /api/v1/oauth/exchange

Exchange callback URL for access token (manual flow).

**Request:**
```json
{
  "callback_url": "http://localhost:8088/api/v1/oauth/callback?code=AUTH_CODE&state=abc123"
}
```

**Response:**
```json
{
  "success": true,
  "account": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "provider_id": "antigravity",
    "label": "My Google Account",
    "is_active": true,
    "expires_at": "2025-12-28T10:00:00Z",
    "created_at": "2025-12-27T09:00:00Z"
  }
}
```

### GET /api/v1/oauth/providers

List available OAuth providers.

**Response:**
```json
{
  "providers": [
    {
      "id": "antigravity",
      "name": "Google Cloud Code (Antigravity)"
    },
    {
      "id": "codex",
      "name": "OpenAI Codex"
    },
    {
      "id": "claude",
      "name": "Anthropic Claude"
    }
  ]
}
```

### POST /api/v1/oauth/refresh

Manually refresh an account's OAuth token.

**Request:**
```json
{
  "account_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response:**
```json
{
  "message": "token refreshed successfully"
}
```

## Providers

### Antigravity (Google Cloud Code)

**OAuth Config:**
- Auth URL: `https://accounts.google.com/o/oauth2/v2/auth`
- Token URL: `https://oauth2.googleapis.com/token`
- Client ID: `1071006060591-tmhssin2h21lcre235vtolojh4g403ep.apps.googleusercontent.com`
- Scope: `https://www.googleapis.com/auth/cloudcode`
- Supports: Gemini models, Claude models via Antigravity

**Token Exchange:**
- Method: POST (form-urlencoded)
- Requires: client_secret

### OpenAI Codex

**OAuth Config:**
- Auth URL: `https://auth.openai.com/oauth/authorize`
- Token URL: `https://auth.openai.com/oauth/token`
- Client ID: `app_EMoamEEZ73f0CkXaXp7hrann`
- Scope: `openid email profile offline_access`
- Extra Params: `id_token_add_organizations`, `codex_cli_simplified_flow`

**Token Exchange:**
- Method: POST (form-urlencoded)
- No client_secret required (public client)

### Anthropic Claude

**OAuth Config:**
- Auth URL: `https://claude.ai/oauth/authorize`
- Token URL: `https://console.anthropic.com/v1/oauth/token`
- Client ID: `9d1c250a-e61b-44d9-88ed-5944d1962f5e`
- Scope: `org:create_api_key user:profile user:inference`
- Extra Params: `code=true`

**Token Exchange:**
- Method: POST (JSON body)
- No client_secret required (public client)

## Security

### PKCE Implementation

All providers use PKCE (RFC 7636) with S256 method:

1. **Code Verifier**: 96 random bytes → 128 chars base64url (no padding)
2. **Code Challenge**: SHA256(verifier) → base64url (no padding)
3. **Challenge Method**: S256

**Example:**
```go
codes, _ := pkce.GeneratePKCECodes()
// codes.CodeVerifier: "xkjPl3...128chars"
// codes.CodeChallenge: "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"
```

### Session Storage

OAuth sessions stored in Redis with 10-minute TTL:

**Key:** `oauth:session:{state}`

**Value:**
```json
{
  "provider": "antigravity",
  "account_name": "My Account",
  "flow_type": "auto",
  "redirect_uri": "http://localhost:8088/api/v1/oauth/callback",
  "code_verifier": "xkjPl3...128chars",
  "created_at": "2025-12-27T09:00:00Z",
  "created_by": "user-id-here"
}
```

**TTL:** 10 minutes (OAuthSessionTTL constant)

### Token Storage

Access tokens stored in Account.AuthData (JSON field):

```json
{
  "access_token": "ya29.a0...",
  "refresh_token": "1//0...",
  "token_type": "Bearer",
  "expires_at": "2025-12-27T10:00:00Z",
  "expires_in": 3600
}
```

**Auto-refresh:**
- Handled by existing `OAuthService.GetAccessToken()` method
- Refreshes 3000 seconds before expiry (RefreshSkew)
- Updates Redis cache and database

## Frontend Integration

### Automatic Flow (Popup)

```typescript
// 1. Initialize OAuth flow
const response = await fetch('/api/v1/oauth/init', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    provider: 'antigravity',
    account_name: 'My Google Account',
    flow_type: 'auto'
  })
});

const { auth_url } = await response.json();

// 2. Open popup
const popup = window.open(auth_url, 'oauth', 'width=600,height=700');

// 3. Listen for callback
window.addEventListener('message', (event) => {
  if (event.data.type === 'oauth_success') {
    console.log('Account connected:', event.data.account);
    // Refresh account list
  } else if (event.data.type === 'oauth_error') {
    console.error('OAuth failed:', event.data.error);
  }
});
```

### Manual Flow (Copy-Paste)

```typescript
// 1. Initialize OAuth flow
const response = await fetch('/api/v1/oauth/init', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    provider: 'antigravity',
    account_name: 'My Google Account',
    flow_type: 'manual'
  })
});

const { auth_url } = await response.json();

// 2. Open popup
window.open(auth_url, 'oauth', 'width=600,height=700');

// 3. User pastes callback URL
const callbackURL = prompt('Paste the callback URL from browser:');

// 4. Exchange code
const exchangeResp = await fetch('/api/v1/oauth/exchange', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ callback_url: callbackURL })
});

const { account } = await exchangeResp.json();
console.log('Account created:', account);
```

## Testing

### Test OAuth Flow

1. **Get providers:**
```bash
curl http://localhost:8088/api/v1/oauth/providers
```

2. **Initialize flow:**
```bash
curl -X POST http://localhost:8088/api/v1/oauth/init \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "antigravity",
    "account_name": "Test Account",
    "flow_type": "manual"
  }'
```

3. **Open auth_url in browser, copy callback URL**

4. **Exchange code:**
```bash
curl -X POST http://localhost:8088/api/v1/oauth/exchange \
  -H "Content-Type: application/json" \
  -d '{
    "callback_url": "http://localhost:8088/api/v1/oauth/callback?code=4/0...&state=abc123"
  }'
```

5. **Verify account created:**
```bash
curl http://localhost:8088/api/v1/accounts
```

### Test Token Refresh

```bash
curl -X POST http://localhost:8088/api/v1/oauth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

## Files Created

| File | Lines | Purpose |
|------|-------|---------|
| `auth/pkce/pkce.go` | 48 | PKCE code generation (S256 method) |
| `auth/oauth/providers.go` | 234 | OAuth configs + token exchange for 3 providers |
| `services/oauth.flow.service.go` | 298 | OAuth flow orchestration, session management |
| `handlers/oauth.handler.go` | 206 | HTTP endpoints with HTML popup responses |

**Total:** 786 lines across 4 new files

## Files Modified

| File | Changes |
|------|---------|
| `cmd/main.go` | Added OAuthFlowService initialization and handler |
| `routes/routes.go` | Added 5 OAuth endpoints |

## Error Handling

### Common Errors

**Invalid flow_type:**
```json
{
  "error": "invalid flow_type: must be 'auto' or 'manual'"
}
```

**Session expired:**
```json
{
  "error": "session not found or expired"
}
```

**Token exchange failed:**
```json
{
  "error": "token exchange failed: invalid_grant"
}
```

**Missing refresh token:**
```json
{
  "error": "no refresh token available"
}
```

## Integration with Existing Services

The OAuth flow service integrates with:

1. **AccountService**: Creates accounts after successful OAuth
2. **AccountRepository**: Persists OAuth credentials to database
3. **Redis**: Session storage with TTL
4. **Middleware**: RBAC protection for OAuth endpoints
5. **OAuthService**: Token caching and auto-refresh (existing)

## Future Enhancements

1. **OAuth Scopes**: Allow custom scopes per provider
2. **Multi-account**: Support multiple accounts per provider per user
3. **Webhook**: Provider webhook for token revocation events
4. **Admin UI**: Visual OAuth flow management
5. **Audit Log**: Track OAuth authorization events

## References

- [RFC 7636 - PKCE](https://datatracker.ietf.org/doc/html/rfc7636)
- [Google OAuth 2.0](https://developers.google.com/identity/protocols/oauth2)
- [OpenAI OAuth](https://platform.openai.com/docs/api-reference/authentication)
- [Anthropic OAuth](https://docs.anthropic.com/en/api/oauth)
