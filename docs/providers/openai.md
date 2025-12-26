# OpenAI Provider

**Provider ID**: `openai`

## Overview

OpenAI's GPT model API for text generation and chat completions.

**Base URL**: `https://api.openai.com`

**Authentication**: API Key

**Status**: Planned

## Endpoints

- **Chat**: `POST /v1/chat/completions`
- **Completions**: `POST /v1/completions`

## Supported Models

- `gpt-4`
- `gpt-4-turbo`
- `gpt-4-turbo-preview`
- `gpt-3.5-turbo`
- `gpt-3.5-turbo-16k`

## Model Routing

Models are automatically routed to OpenAI based on this prefix:
- `gpt-*` → OpenAI

## Configuration

### Account Configuration Example

```json
{
  "provider_id": "openai",
  "label": "OpenAI Production",
  "auth_data": {
    "api_key": "sk-proj-..."
  },
  "is_active": true
}
```

### API Key Fields

| Field | Required | Description |
|-------|----------|-------------|
| `api_key` | Yes | OpenAI API key (starts with `sk-proj-` or `sk-`) |

## API Format

**Request Format**: OpenAI Chat Completions format

**Response Format**: OpenAI Chat Completions format

**Format Translation**: None required (native OpenAI format)

## Token Management

**Type**: Static API key (no refresh needed)

API keys do not expire and require no automatic refresh logic.

## Setup Guide

### Step 1: Obtain API Key

1. Visit [OpenAI Platform](https://platform.openai.com/api-keys)
2. Create new API key
3. Copy the key (starts with `sk-proj-` or `sk-`)

### Step 2: Insert Provider

```sql
INSERT INTO providers (id, name, base_url, auth_type, auth_strategy, supported_models, is_active)
VALUES (
  'openai',
  'OpenAI',
  'https://api.openai.com',
  'api_key',
  'api_key',
  '["gpt-4", "gpt-3.5-turbo"]',
  true
);
```

### Step 3: Create Account via API

```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "openai",
    "label": "OpenAI Account 1",
    "auth_data": {
      "api_key": "sk-proj-..."
    },
    "is_active": true
  }'
```

### Step 4: Update Model Routing

Ensure `providers/registry.go` includes:
```go
case strings.HasPrefix(modelLower, "gpt-"):
    return "openai"
```

### Step 5: Test the Integration

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "user", "content": "test"}
    ],
    "max_tokens": 100
  }'
```

## Rate Limits

**Tier-based Limits**:
- Free tier: Low RPM/TPM
- Pay-as-you-go: Medium RPM/TPM
- Scale: High RPM/TPM

**Metrics**:
- RPM (requests per minute)
- TPM (tokens per minute)

**Strategy**: Use multiple API keys to increase limits

## Error Codes

| Code | Description | Resolution |
|------|-------------|------------|
| `401` | Invalid API key | Verify API key is correct and active |
| `429` | Rate limit exceeded | Add more accounts or implement backoff |
| `500` | OpenAI server error | Retry with exponential backoff |

## Best Practices

1. **Usage Monitoring**: Monitor usage via OpenAI dashboard
2. **Backoff Strategy**: Implement exponential backoff for 429 errors
3. **Multiple Keys**: Use multiple API keys for production
4. **Cost Tracking**: Track token usage to manage costs
5. **Key Rotation**: Rotate API keys regularly for security

## Troubleshooting

### Invalid API Key

**Symptom**: "401 Invalid API key"

**Solutions**:
1. Verify API key format (starts with `sk-proj-` or `sk-`)
2. Check key is active in OpenAI dashboard
3. Ensure key has not been revoked

### Rate Limit Exceeded

**Symptom**: "429 Too Many Requests"

**Solutions**:
1. Add more API keys for load distribution
2. Implement exponential backoff
3. Upgrade to higher tier for increased limits

## Related Documentation

- [Provider Overview](README.md) - All providers
- [Authentication Strategies](authentication.md) - API Key details
- [Adding a New Provider](adding-new-provider.md) - Setup guide
