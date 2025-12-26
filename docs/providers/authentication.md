# Authentication Strategies

This document describes the authentication mechanisms supported by AIGateway and their implementation details.

## Supported Strategies

AIGateway supports three authentication strategies:

1. **OAuth 2.0**: Token-based authentication with automatic refresh
2. **API Key**: Static key extraction
3. **Bearer Token**: Simple bearer token authentication

## OAuth 2.0 Strategy

**Used by**: Antigravity

**Implementation**: `auth/oauth.strategy.go`

### Features

- Automatic token refresh before expiration
- Redis-based caching for performance
- Database persistence of tokens
- 5-minute expiry buffer to prevent edge cases

### Token Lifecycle

```
1. Request arrives
   ↓
2. Check Redis cache: auth:oauth:{provider}:{account}
   ↓
3. If cached and valid (>5 min remaining):
   → Return cached token
   ↓
4. If expired or missing:
   → Extract access_token from auth_data
   → Check expiry from expires_at
   ↓
5. If <5 min remaining:
   → Call OAuth refresh endpoint
   → Update Redis cache (TTL = expires_in)
   → Update database auth_data
   → Return new token
```

### Required Fields

```json
{
  "access_token": "ya29.xxx",
  "refresh_token": "1//xxx",
  "token_url": "https://oauth2.googleapis.com/token",
  "client_id": "xxx.apps.googleusercontent.com",
  "client_secret": "xxx",
  "expires_at": "2024-12-26T10:00:00Z",
  "expires_in": 3600
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `access_token` | string | Yes | Current OAuth access token |
| `refresh_token` | string | Yes | Long-lived refresh token |
| `token_url` | string | Yes | OAuth token endpoint URL |
| `client_id` | string | Yes | OAuth client identifier |
| `client_secret` | string | Yes | OAuth client secret |
| `expires_at` | string | Yes | ISO 8601 expiry timestamp |
| `expires_in` | number | Yes | Seconds until token expiry |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `token_type` | string | "Bearer" | Token type |

### Token Refresh Request

When token needs refresh, AIGateway makes this request:

```http
POST {token_url}
Content-Type: application/x-www-form-urlencoded

grant_type=refresh_token&
refresh_token={refresh_token}&
client_id={client_id}&
client_secret={client_secret}
```

**Response**:
```json
{
  "access_token": "ya29.new_token",
  "expires_in": 3600,
  "token_type": "Bearer"
}
```

### Redis Caching

**Key Format**: `auth:oauth:{provider_id}:{account_id}`

**Value**: JSON string with token data
```json
{
  "access_token": "ya29.xxx",
  "refresh_token": "1//xxx",
  "expires_at": "2024-12-26T10:00:00Z",
  "token_type": "Bearer"
}
```

**TTL**: Set to `expires_in` value from OAuth response

### Performance Impact

```
Without cache: ~200ms OAuth endpoint call per request
With cache:    <1ms Redis GET per request
Improvement:   200x faster for cached tokens
```

**Cache Hit Rate**: ~99.9% for typical 1-hour token lifetime

## API Key Strategy

**Used by**: OpenAI, Anthropic, GLM

**Implementation**: `auth/apikey.strategy.go`

### Features

- Simple key extraction from auth_data
- No expiry or refresh logic
- Multiple field name support for compatibility

### Key Extraction Order

The strategy tries these fields in order:
1. `api_key`
2. `apiKey`
3. `key`
4. `token`
5. `access_token`

Returns the first non-empty value found.

### Required Fields

```json
{
  "api_key": "sk-proj-xxx"
}
```

**Alternative formats** (all supported):
```json
// Option 1
{"apiKey": "sk-proj-xxx"}

// Option 2
{"key": "sk-proj-xxx"}

// Option 3
{"token": "sk-proj-xxx"}
```

### Usage Example

**OpenAI**:
```json
{
  "provider_id": "openai",
  "auth_data": {
    "api_key": "sk-proj-..."
  }
}
```

**Anthropic**:
```json
{
  "provider_id": "anthropic",
  "auth_data": {
    "api_key": "sk-ant-api03-..."
  }
}
```

### HTTP Header

API keys are added to requests as:
```http
Authorization: Bearer {api_key}
```

## Bearer Token Strategy

**Used by**: Custom providers

**Implementation**: `auth/bearer.strategy.go`

### Features

- Static bearer token authentication
- No expiry or refresh logic
- Simple extraction from auth_data

### Required Fields

```json
{
  "token": "bearer-token-value"
}
```

### HTTP Header

Bearer tokens are added to requests as:
```http
Authorization: Bearer {token}
```

## Strategy Interface

All authentication strategies implement this interface:

```go
type Strategy interface {
    Name() string
    GetToken(ctx context.Context, authData map[string]interface{}) (string, error)
    RefreshToken(ctx context.Context, authData map[string]interface{}, oldToken string) (string, error)
    ValidateToken(ctx context.Context, token string) (bool, error)
}
```

### Method Descriptions

**Name()**: Returns strategy name (e.g., "oauth", "api_key", "bearer")

**GetToken()**: Retrieves or generates authentication token
- OAuth: Checks cache, refreshes if needed
- API Key: Extracts from auth_data
- Bearer: Extracts from auth_data

**RefreshToken()**: Refreshes an expired token
- OAuth: Calls OAuth endpoint with refresh_token
- API Key: No-op (keys don't expire)
- Bearer: No-op (tokens don't expire)

**ValidateToken()**: Validates token is still valid
- OAuth: Checks expiry timestamp
- API Key: Always returns true
- Bearer: Always returns true

## Adding a New Strategy

To add a new authentication strategy:

### 1. Create Strategy Implementation

Create `auth/newstrategy.strategy.go`:

```go
package auth

type NewStrategy struct {
    redis *redis.Client
    db    *gorm.DB
}

func (s *NewStrategy) Name() string {
    return "new_strategy"
}

func (s *NewStrategy) GetToken(ctx context.Context, authData map[string]interface{}) (string, error) {
    // Implementation
}

func (s *NewStrategy) RefreshToken(ctx context.Context, authData map[string]interface{}, oldToken string) (string, error) {
    // Implementation
}

func (s *NewStrategy) ValidateToken(ctx context.Context, token string) (bool, error) {
    // Implementation
}
```

### 2. Register Strategy

Update strategy factory in service initialization:

```go
strategies := map[string]auth.Strategy{
    "oauth":        oauth.NewOAuthStrategy(redis, db),
    "api_key":      apikey.NewAPIKeyStrategy(),
    "bearer":       bearer.NewBearerStrategy(),
    "new_strategy": newstrategy.NewStrategy(), // Add this
}
```

### 3. Update Provider Configuration

Set provider to use new strategy:

```sql
UPDATE providers
SET auth_strategy = 'new_strategy'
WHERE id = 'provider_id';
```

## Security Considerations

### OAuth Tokens

- **Storage**: Encrypted in database auth_data column
- **Transit**: TLS for all OAuth token requests
- **Caching**: Redis with TTL, ensure Redis uses AUTH password
- **Rotation**: Automatic via refresh token mechanism

### API Keys

- **Storage**: Encrypted in database auth_data column
- **Transit**: TLS for all provider API requests
- **Rotation**: Manual rotation recommended every 90 days
- **Scope**: Use minimum required scopes

### Best Practices

1. **Encryption at Rest**: Encrypt `auth_data` column in MySQL
2. **TLS Everywhere**: Use TLS 1.2+ for all API communications
3. **Secret Management**: Use secret management service (Vault, AWS Secrets Manager)
4. **Access Control**: Restrict database access to application only
5. **Audit Logging**: Log all authentication failures
6. **Regular Rotation**: Rotate credentials regularly

## Troubleshooting

### OAuth Token Refresh Failures

**Check**:
1. Refresh token is still valid
2. client_id and client_secret are correct
3. token_url is reachable
4. Network connectivity to OAuth endpoint

**Debug**:
```bash
# Check Redis cache
redis-cli GET "auth:oauth:antigravity:account-id"

# Check database auth_data
SELECT auth_data FROM accounts WHERE id = 'account-id';
```

### API Key Failures

**Check**:
1. API key format is correct
2. Key is active in provider dashboard
3. Key has required permissions
4. No typos in key value

## Related Documentation

- [Provider Overview](README.md) - All providers
- [Antigravity Provider](antigravity.md) - OAuth example
- [OpenAI Provider](openai.md) - API Key example
