# Antigravity Provider

**Provider ID**: `antigravity`

## Overview

Google's internal AI API gateway that provides access to Gemini models and Claude models via Google infrastructure.

**Base URL**: `https://cloudcode-pa.googleapis.com`

**Authentication**: OAuth 2.0

**Status**: Active

## Endpoints

- **Non-streaming**: `POST /v1internal:generateContent`
- **Streaming**: `POST /v1internal:streamGenerateContent?alt=sse`

## Supported Models

### Gemini 2.5 Series
- `gemini-2.5-flash` - Fast multimodal model with thinking support (0-24K budget)
- `gemini-2.5-flash-lite` - Lightweight version with thinking support (0-24K budget)
- `gemini-2.5-pro` - Most capable Gemini 2.5 model with thinking support (0-24K budget)
- `gemini-2.5-computer-use-preview-10-2025` - Computer use preview model

### Gemini 3 Series
- `gemini-3-pro-preview` - Next-gen flagship with thinking levels (128-32K budget, levels: low/high)
- `gemini-3-pro-image-preview` - Image generation with thinking levels (128-32K budget, levels: low/high)
- `gemini-3-flash-preview` - Fast next-gen with thinking levels (128-32K budget, levels: minimal/low/medium/high)

### Claude via Antigravity
- `gemini-claude-sonnet-4-5` - Claude Sonnet 4.5 without thinking
- `gemini-claude-sonnet-4-5-thinking` - Claude Sonnet 4.5 with extended thinking (1K-200K budget, max 64K tokens)
- `gemini-claude-opus-4-5` - Claude Opus 4.5 without thinking
- `gemini-claude-opus-4-5-thinking` - Claude Opus 4.5 with extended thinking (1K-200K budget, max 64K tokens)

## Model Routing

Models are automatically routed to Antigravity based on these prefixes:
- `gemini-2.5-*` → Antigravity (Gemini 2.5 series)
- `gemini-3-*` → Antigravity (Gemini 3 series)
- `gemini-claude-*` → Antigravity (Claude via Google infrastructure)

## Configuration

### Account Configuration Example

```json
{
  "provider_id": "antigravity",
  "label": "Google Antigravity Production",
  "auth_data": {
    "access_token": "ya29.a0AfB_byC...",
    "refresh_token": "1//0gHdBj...",
    "token_url": "https://oauth2.googleapis.com/token",
    "client_id": "123456789.apps.googleusercontent.com",
    "client_secret": "GOCSPX-...",
    "expires_at": "2024-12-26T10:00:00Z",
    "expires_in": 3600,
    "token_type": "Bearer"
  },
  "is_active": true
}
```

### OAuth Token Fields

| Field | Required | Description |
|-------|----------|-------------|
| `access_token` | Yes | Current OAuth access token |
| `refresh_token` | Yes | Long-lived refresh token for obtaining new access tokens |
| `token_url` | Yes | OAuth token endpoint for refresh |
| `client_id` | Yes | OAuth client ID |
| `client_secret` | Yes | OAuth client secret |
| `expires_at` | Yes | ISO 8601 timestamp when token expires |
| `expires_in` | Yes | Token validity in seconds (typically 3600) |
| `token_type` | No | Token type (default: "Bearer") |

## API Format

**Request Format**: Gemini API format

**Response Format**: Gemini API format

**Format Translation**: AIGateway automatically translates between Claude format and Gemini format.

See [Format Translation](format-translation.md) for detailed transformations.

## Token Management

### Automatic Token Refresh

AIGateway automatically refreshes OAuth tokens:
- Refresh triggered when <5 minutes remaining
- Tokens cached in Redis with TTL = expires_in
- Database persistence after refresh

### Token Lifecycle

```
1. Request arrives
   ↓
2. Check Redis cache: auth:oauth:antigravity:{account_id}
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
   → Update Redis cache
   → Update database auth_data
   → Return new token
```

## Setup Guide

### Step 1: Obtain OAuth Credentials

1. Visit [Google Cloud Console](https://console.cloud.google.com)
2. Create OAuth 2.0 Client ID
3. Configure consent screen
4. Generate credentials
5. Obtain initial access token and refresh token

### Step 2: Insert Provider

```sql
INSERT INTO providers (id, name, base_url, auth_type, auth_strategy, supported_models, is_active)
VALUES (
  'antigravity',
  'Google Antigravity',
  'https://cloudcode-pa.googleapis.com',
  'oauth',
  'oauth',
  '["gemini-2.5-flash", "gemini-2.5-pro", "gemini-3-pro-preview", "gemini-3-flash-preview", "gemini-claude-sonnet-4-5", "gemini-claude-opus-4-5"]',
  true
);
```

### Step 3: Create Account via API

```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "antigravity",
    "label": "Google Account 1",
    "auth_data": {
      "access_token": "ya29.xxx",
      "refresh_token": "1//xxx",
      "token_url": "https://oauth2.googleapis.com/token",
      "client_id": "xxx.apps.googleusercontent.com",
      "client_secret": "xxx",
      "expires_at": "2024-12-26T10:00:00Z",
      "expires_in": 3600,
      "token_type": "Bearer"
    },
    "is_active": true
  }'
```

### Step 4: Test the Integration

```bash
# Test with Gemini 2.5 Flash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [
      {"role": "user", "content": "test"}
    ],
    "max_tokens": 100
  }'

# Test with Claude Sonnet 4.5 via Antigravity
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-claude-sonnet-4-5",
    "messages": [
      {"role": "user", "content": "test"}
    ],
    "max_tokens": 100
  }'

# Test with Gemini 3 Pro with thinking
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-3-pro-preview",
    "messages": [
      {"role": "user", "content": "Explain quantum computing"}
    ],
    "max_tokens": 1000,
    "thinking": {
      "type": "enabled",
      "budget": 1000
    }
  }'
```

## Rate Limits

- Varies by project quota
- Use multiple accounts for higher throughput
- Typical limits: 60 req/min, 90k tokens/min per account

## Error Codes

| Code | Description | Resolution |
|------|-------------|------------|
| `401` | Token expired or invalid | Token will be auto-refreshed |
| `403` | Quota exceeded | Add more accounts or increase quota |
| `429` | Rate limit exceeded | Add more accounts for load distribution |

## Best Practices

1. **Credential Rotation**: Rotate OAuth credentials regularly
2. **Monitor Refresh**: Monitor token refresh failures
3. **Multiple Projects**: Use multiple Google Cloud projects for redundancy
4. **Quota Management**: Track quota usage via Google Cloud Console
5. **Token Security**: Keep refresh_token secure and encrypted

## Troubleshooting

### Token Refresh Failures

**Symptom**: "Failed to refresh OAuth token"

**Possible Causes**:
1. Invalid refresh_token
2. Invalid client_id or client_secret
3. OAuth credentials revoked

**Solutions**:
1. Obtain new OAuth credentials from Google Cloud Console
2. Update account auth_data with new credentials
3. Verify token_url is correct

### Quota Exceeded

**Symptom**: "403 Quota exceeded"

**Solutions**:
1. Increase quota in Google Cloud Console
2. Add more accounts for load distribution
3. Enable billing if using free tier

## Related Documentation

- [Provider Overview](README.md) - All providers
- [Authentication Strategies](authentication.md) - OAuth 2.0 details
- [Format Translation](format-translation.md) - Claude ↔ Gemini transformation
