# Anthropic Provider

**Provider ID**: `anthropic`

## Overview

Anthropic's Claude model API for advanced conversational AI.

**Base URL**: `https://api.anthropic.com`

**Authentication**: API Key

**Status**: Planned

## Endpoints

- **Messages**: `POST /v1/messages`

## Supported Models

- `claude-opus-4-5`
- `claude-opus-4`
- `claude-sonnet-4`
- `claude-3-7-sonnet`
- `claude-3-5-sonnet`
- `claude-3-opus`
- `claude-3-haiku`

## Model Routing

Models are automatically routed to Anthropic based on these prefixes:
- `claude-opus-*` → Anthropic
- `claude-3-*` → Anthropic

**Note**: `claude-sonnet-*` models are routed to Antigravity by default. Update routing if you want direct Anthropic access.

## Configuration

### Account Configuration Example

```json
{
  "provider_id": "anthropic",
  "label": "Anthropic Production",
  "auth_data": {
    "api_key": "sk-ant-api03-..."
  },
  "is_active": true
}
```

### API Key Fields

| Field | Required | Description |
|-------|----------|-------------|
| `api_key` | Yes | Anthropic API key (starts with `sk-ant-`) |

## API Format

**Request Format**: Anthropic Messages format

**Response Format**: Anthropic Messages format

**Format Translation**: None required (native Anthropic format)

## Token Management

**Type**: Static API key (no refresh needed)

API keys do not expire and require no automatic refresh logic.

## Setup Guide

### Step 1: Obtain API Key

1. Visit [Anthropic Console](https://console.anthropic.com/)
2. Generate new API key
3. Copy the key (starts with `sk-ant-`)

### Step 2: Insert Provider

```sql
INSERT INTO providers (id, name, base_url, auth_type, auth_strategy, supported_models, is_active)
VALUES (
  'anthropic',
  'Anthropic',
  'https://api.anthropic.com',
  'api_key',
  'api_key',
  '["claude-opus-4-5", "claude-3-5-sonnet"]',
  true
);
```

### Step 3: Create Account via API

```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "anthropic",
    "label": "Anthropic Account 1",
    "auth_data": {
      "api_key": "sk-ant-api03-..."
    },
    "is_active": true
  }'
```

### Step 4: Update Model Routing (Optional)

To route Claude Sonnet models to Anthropic instead of Antigravity, edit `providers/registry.go`:

```go
case strings.HasPrefix(modelLower, "claude-opus-"):
    return "anthropic"
case strings.HasPrefix(modelLower, "claude-3-"):
    return "anthropic"
case strings.HasPrefix(modelLower, "claude-sonnet-"):
    return "anthropic"  // Changed from "antigravity"
```

### Step 5: Test the Integration

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-opus-4-5",
    "messages": [
      {"role": "user", "content": "test"}
    ],
    "max_tokens": 100
  }'
```

## Rate Limits

**Tier-based Limits**:
- Varies by plan (Free, Pro, Team, Enterprise)
- Requests per day and tokens per day limits
- Contact Anthropic for higher limits

**Strategy**: Use multiple API keys for increased capacity

## Error Codes

| Code | Description | Resolution |
|------|-------------|------------|
| `401` | Invalid API key | Verify API key is correct and active |
| `429` | Rate limit exceeded | Add more accounts or upgrade plan |
| `529` | Service overloaded | Implement retry logic with backoff |

## Best Practices

1. **API Key Monitoring**: Monitor API key usage via Anthropic Console
2. **Retry Logic**: Implement retry logic for 529 errors (service overload)
3. **Multiple Keys**: Use multiple API keys for production workloads
4. **Cost Optimization**: Choose appropriate model tier (Haiku < Sonnet < Opus)
5. **Key Security**: Store API keys securely and rotate regularly

## Troubleshooting

### Invalid API Key

**Symptom**: "401 Invalid API key"

**Solutions**:
1. Verify API key format (starts with `sk-ant-`)
2. Check key is active in Anthropic Console
3. Ensure workspace has not been deactivated

### Service Overloaded

**Symptom**: "529 Service overloaded"

**Solutions**:
1. Implement exponential backoff
2. Retry after brief delay
3. Spread requests across time

## Related Documentation

- [Provider Overview](README.md) - All providers
- [Authentication Strategies](authentication.md) - API Key details
- [Adding a New Provider](adding-new-provider.md) - Setup guide
