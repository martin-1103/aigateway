# GLM Provider

**Provider ID**: `glm`

## Overview

Zhipu AI's GLM (General Language Model) API for Chinese and multilingual text generation.

**Base URL**: `https://open.bigmodel.cn`

**Authentication**: API Key

**Status**: Planned

## Endpoints

- **Chat**: `POST /api/paas/v4/chat/completions`

## Supported Models

- `glm-4`
- `glm-4-air`
- `glm-4-flash`
- `glm-3-turbo`

## Model Routing

Models are automatically routed to GLM based on this prefix:
- `glm-*` → GLM

## Configuration

### Account Configuration Example

```json
{
  "provider_id": "glm",
  "label": "GLM Production",
  "auth_data": {
    "api_key": "..."
  },
  "is_active": true
}
```

### API Key Fields

| Field | Required | Description |
|-------|----------|-------------|
| `api_key` | Yes | GLM API key |

## API Format

**Request Format**: OpenAI-compatible format

**Response Format**: OpenAI-compatible format

**Format Translation**: Minimal (OpenAI-compatible API)

## Token Management

**Type**: Static API key (no refresh needed)

API keys do not expire and require no automatic refresh logic.

## Setup Guide

### Step 1: Obtain API Key

1. Visit [Zhipu AI Open Platform](https://open.bigmodel.cn/)
2. Register or login to your account
3. Create new API key
4. Copy the API key

### Step 2: Insert Provider

```sql
INSERT INTO providers (id, name, base_url, auth_type, auth_strategy, supported_models, is_active)
VALUES (
  'glm',
  'Zhipu AI GLM',
  'https://open.bigmodel.cn',
  'api_key',
  'api_key',
  '["glm-4", "glm-4-air", "glm-3-turbo"]',
  true
);
```

### Step 3: Create Account via API

```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "glm",
    "label": "GLM Account 1",
    "auth_data": {
      "api_key": "your-api-key-here"
    },
    "is_active": true
  }'
```

### Step 4: Update Model Routing

Ensure `providers/registry.go` includes:
```go
case strings.HasPrefix(modelLower, "glm-"):
    return "glm"
```

### Step 5: Test the Integration

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "glm-4",
    "messages": [
      {"role": "user", "content": "你好"}
    ],
    "max_tokens": 100
  }'
```

## Rate Limits

**Varies by plan**:
- Free tier: Limited requests per day
- Paid tier: Higher limits based on subscription

**Metrics**:
- Requests per minute
- Tokens per minute

**Strategy**: Use multiple API keys for higher throughput

## Error Codes

| Code | Description | Resolution |
|------|-------------|------------|
| `401` | Invalid API key | Verify API key is correct and active |
| `429` | Rate limit exceeded | Add more accounts or upgrade plan |
| `500` | Server error | Retry with exponential backoff |

## Best Practices

1. **Chinese Language Support**: GLM excels at Chinese language tasks
2. **Multiple Keys**: Use multiple API keys for production
3. **Model Selection**: Choose appropriate model tier (Flash < Air < GLM-4)
4. **Usage Monitoring**: Track usage via Zhipu AI dashboard
5. **Key Security**: Store API keys securely

## Troubleshooting

### Invalid API Key

**Symptom**: "401 Invalid API key"

**Solutions**:
1. Verify API key is copied correctly
2. Check key is active in Zhipu AI dashboard
3. Ensure account has not been suspended

### Rate Limit Exceeded

**Symptom**: "429 Too Many Requests"

**Solutions**:
1. Add more API keys for load distribution
2. Implement request rate limiting
3. Upgrade to higher tier plan

## Related Documentation

- [Provider Overview](README.md) - All providers
- [Authentication Strategies](authentication.md) - API Key details
- [Adding a New Provider](adding-new-provider.md) - Setup guide
