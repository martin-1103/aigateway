# Provider Documentation

This document describes all supported AI providers, their configuration, model mappings, and authentication requirements.

## Supported Providers

AIGateway currently supports the following providers:

| Provider ID | Provider Name | Base URL | Auth Type | Status |
|-------------|---------------|----------|-----------|--------|
| `antigravity` | Google Antigravity | `https://cloudcode-pa.googleapis.com` | OAuth 2.0 | Active |
| `openai` | OpenAI | `https://api.openai.com` | API Key | Planned |
| `anthropic` | Anthropic | `https://api.anthropic.com` | API Key | Planned |
| `glm` | Zhipu AI (GLM) | `https://open.bigmodel.cn` | API Key | Planned |

## Provider Details

### Antigravity (Google)

**Provider ID**: `antigravity`

**Description**: Google's internal AI API gateway that provides access to Gemini models and Claude models via Google infrastructure.

**Base URL**: `https://cloudcode-pa.googleapis.com`

**Authentication**: OAuth 2.0

**Endpoints**:
- Non-streaming: `POST /v1internal:generateContent`
- Streaming: `POST /v1internal:streamGenerateContent?alt=sse`

**Supported Models**:
- `gemini-claude-sonnet-4-5`
- `gemini-claude-sonnet-3-5`
- `gemini-pro`
- `gemini-pro-vision`
- `claude-sonnet-4-5` (routed to Antigravity)
- `claude-sonnet-3-5` (routed to Antigravity)

**Model Routing Prefixes**:
- `gemini-*` → Antigravity
- `claude-sonnet-*` → Antigravity

**Configuration Example**:

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

**OAuth Token Fields**:
- `access_token`: Current OAuth access token
- `refresh_token`: Long-lived refresh token for obtaining new access tokens
- `token_url`: OAuth token endpoint for refresh
- `client_id`: OAuth client ID
- `client_secret`: OAuth client secret
- `expires_at`: ISO 8601 timestamp when token expires
- `expires_in`: Token validity in seconds (typically 3600)
- `token_type`: Token type (usually "Bearer")

**Request Format**: Gemini API format

**Response Format**: Gemini API format

**Format Translation**:
- Input: Claude format → Gemini format
- Output: Gemini format → Claude format

**Token Management**:
- Automatic refresh when <5 minutes remaining
- Redis cache with TTL = expires_in
- Database persistence after refresh

---

### OpenAI

**Provider ID**: `openai`

**Description**: OpenAI's GPT model API.

**Base URL**: `https://api.openai.com`

**Authentication**: API Key

**Endpoints**:
- Chat: `POST /v1/chat/completions`
- Completions: `POST /v1/completions`

**Supported Models**:
- `gpt-4`
- `gpt-4-turbo`
- `gpt-4-turbo-preview`
- `gpt-3.5-turbo`
- `gpt-3.5-turbo-16k`

**Model Routing Prefixes**:
- `gpt-*` → OpenAI

**Configuration Example**:

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

**API Key Fields**:
- `api_key`: OpenAI API key (starts with `sk-proj-` or `sk-`)

**Request Format**: OpenAI Chat Completions format

**Response Format**: OpenAI Chat Completions format

**Format Translation**: None required (native OpenAI format)

**Token Management**: Static API key (no refresh needed)

---

### Anthropic

**Provider ID**: `anthropic`

**Description**: Anthropic's Claude model API.

**Base URL**: `https://api.anthropic.com`

**Authentication**: API Key

**Endpoints**:
- Messages: `POST /v1/messages`

**Supported Models**:
- `claude-opus-4-5`
- `claude-opus-4`
- `claude-sonnet-4`
- `claude-3-7-sonnet`
- `claude-3-5-sonnet`
- `claude-3-opus`
- `claude-3-haiku`

**Model Routing Prefixes**:
- `claude-opus-*` → Anthropic
- `claude-3-*` → Anthropic

**Configuration Example**:

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

**API Key Fields**:
- `api_key`: Anthropic API key (starts with `sk-ant-`)

**Request Format**: Anthropic Messages format

**Response Format**: Anthropic Messages format

**Format Translation**: None required (native Anthropic format)

**Token Management**: Static API key (no refresh needed)

---

### GLM (Zhipu AI)

**Provider ID**: `glm`

**Description**: Zhipu AI's GLM (General Language Model) API.

**Base URL**: `https://open.bigmodel.cn`

**Authentication**: API Key

**Endpoints**:
- Chat: `POST /api/paas/v4/chat/completions`

**Supported Models**:
- `glm-4`
- `glm-4-air`
- `glm-4-flash`
- `glm-3-turbo`

**Model Routing Prefixes**:
- `glm-*` → GLM

**Configuration Example**:

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

**API Key Fields**:
- `api_key`: GLM API key

**Request Format**: OpenAI-compatible format

**Response Format**: OpenAI-compatible format

**Format Translation**: Minimal (OpenAI-compatible)

**Token Management**: Static API key (no refresh needed)

## Model Routing Configuration

The `providers/registry.go` file contains the model routing logic:

```go
func (r *Registry) routeModel(model string) string {
    modelLower := strings.ToLower(model)

    switch {
    case strings.HasPrefix(modelLower, "gemini-"):
        return "antigravity"
    case strings.HasPrefix(modelLower, "claude-sonnet-"):
        return "antigravity"
    case strings.HasPrefix(modelLower, "gpt-"):
        return "openai"
    case strings.HasPrefix(modelLower, "glm-"):
        return "glm"
    default:
        return ""
    }
}
```

**To add a new routing rule**:

1. Edit `providers/registry.go`
2. Add a new case in the switch statement
3. Return the provider ID for the model prefix

## Authentication Strategies

### OAuth 2.0 Strategy

**Used by**: Antigravity

**Implementation**: `auth/oauth.strategy.go`

**Features**:
- Automatic token refresh
- Redis-based caching
- Database persistence
- 5-minute expiry buffer

**Token Lifecycle**:

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
   → Update Redis cache
   → Update database auth_data
   → Return new token
```

**Required Fields**:
- `access_token`: Current access token
- `refresh_token`: Refresh token
- `token_url`: OAuth token endpoint
- `client_id`: OAuth client ID
- `client_secret`: OAuth client secret
- `expires_at`: ISO 8601 expiry timestamp
- `expires_in`: Seconds until expiry

**Optional Fields**:
- `token_type`: Token type (default: "Bearer")

---

### API Key Strategy

**Used by**: OpenAI, Anthropic, GLM

**Implementation**: `auth/apikey.strategy.go`

**Features**:
- Simple key extraction
- No expiry or refresh
- Multiple field name support

**Key Extraction**:

Tries the following fields in order:
1. `api_key`
2. `apiKey`
3. `key`
4. `token`
5. `access_token`

**Required Fields**:
- One of the above fields containing the API key

---

### Bearer Token Strategy

**Used by**: Custom providers

**Implementation**: `auth/bearer.strategy.go`

**Features**:
- Static bearer token
- No expiry or refresh

**Required Fields**:
- `token`: Bearer token value

## Format Translation

### Claude to Antigravity (Gemini)

**Implementation**: `services/translator.service.go::ClaudeToAntigravity()`

**Transformations**:

| Claude Format | Antigravity Format |
|---------------|-------------------|
| `system: string` | `request.systemInstruction.parts[0].text` |
| `messages[].role: "assistant"` | `request.contents[].role: "model"` |
| `messages[].role: "user"` | `request.contents[].role: "user"` |
| `messages[].content: string` | `request.contents[].parts[].text` |
| `tools[].input_schema` | `request.tools[0].functionDeclarations[].parametersJsonSchema` |
| `max_tokens` | `request.generationConfig.maxOutputTokens` |
| `temperature` | `request.generationConfig.temperature` |
| `model` | `model` (top-level) |

**Example**:

Input (Claude):
```json
{
  "model": "claude-sonnet-4-5",
  "system": "You are a helpful assistant",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "max_tokens": 1024,
  "temperature": 0.7
}
```

Output (Antigravity):
```json
{
  "model": "claude-sonnet-4-5",
  "request": {
    "systemInstruction": {
      "role": "user",
      "parts": [{"text": "You are a helpful assistant"}]
    },
    "contents": [
      {
        "role": "user",
        "parts": [{"text": "Hello"}]
      }
    ],
    "generationConfig": {
      "maxOutputTokens": 1024,
      "temperature": 0.7
    }
  }
}
```

---

### Antigravity (Gemini) to Claude

**Implementation**: `services/translator.service.go::AntigravityToClaude()`

**Transformations**:

| Antigravity Format | Claude Format |
|-------------------|---------------|
| `response.candidates[0].content.role: "model"` | `role: "assistant"` |
| `response.candidates[0].content.parts[].text` | `content[].text` |
| `response.candidates[0].finishReason: "MAX_TOKENS"` | `stop_reason: "max_tokens"` |
| `response.candidates[0].finishReason: "STOP"` | `stop_reason: "end_turn"` |
| `response.usageMetadata.promptTokenCount` | `usage.input_tokens` |
| `response.usageMetadata.candidatesTokenCount` | `usage.output_tokens` |

**Example**:

Input (Antigravity):
```json
{
  "response": {
    "candidates": [
      {
        "content": {
          "role": "model",
          "parts": [{"text": "Hi there!"}]
        },
        "finishReason": "STOP"
      }
    ],
    "usageMetadata": {
      "promptTokenCount": 10,
      "candidatesTokenCount": 5
    }
  }
}
```

Output (Claude):
```json
{
  "role": "assistant",
  "content": [
    {"type": "text", "text": "Hi there!"}
  ],
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 10,
    "output_tokens": 5
  }
}
```

## Provider Setup Guide

### Adding Antigravity Provider

**Step 1**: Obtain OAuth credentials from Google Cloud Console

1. Create OAuth 2.0 Client ID
2. Configure consent screen
3. Generate credentials
4. Obtain initial access token and refresh token

**Step 2**: Insert provider into database

```sql
INSERT INTO providers (id, name, base_url, auth_type, auth_strategy, supported_models, is_active)
VALUES (
  'antigravity',
  'Google Antigravity',
  'https://cloudcode-pa.googleapis.com',
  'oauth',
  'oauth',
  '["gemini-claude-sonnet-4-5", "gemini-pro"]',
  true
);
```

**Step 3**: Create account via API

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

**Step 4**: Test the integration

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-5",
    "messages": [
      {"role": "user", "content": "test"}
    ],
    "max_tokens": 100
  }'
```

---

### Adding OpenAI Provider

**Step 1**: Obtain API key from OpenAI

1. Visit https://platform.openai.com/api-keys
2. Create new API key
3. Copy the key (starts with `sk-proj-` or `sk-`)

**Step 2**: Insert provider into database

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

**Step 3**: Create account via API

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

**Step 4**: Update model routing in `providers/registry.go`

Ensure routing logic includes:
```go
case strings.HasPrefix(modelLower, "gpt-"):
    return "openai"
```

**Step 5**: Test the integration

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

## Provider-Specific Considerations

### Antigravity

**Rate Limits**:
- Varies by project quota
- Use multiple accounts for higher throughput

**Token Refresh**:
- Tokens expire after 1 hour
- Automatic refresh when <5 minutes remaining
- Keep refresh_token secure

**Error Codes**:
- `401`: Token expired or invalid
- `403`: Quota exceeded
- `429`: Rate limit exceeded

**Best Practices**:
- Rotate OAuth credentials regularly
- Monitor token refresh failures
- Use multiple Google Cloud projects for redundancy

---

### OpenAI

**Rate Limits**:
- Tier-based (Free, Pay-as-you-go, Scale)
- RPM (requests per minute) and TPM (tokens per minute)
- Use multiple API keys to increase limits

**Error Codes**:
- `401`: Invalid API key
- `429`: Rate limit exceeded
- `500`: OpenAI server error

**Best Practices**:
- Monitor usage via OpenAI dashboard
- Implement exponential backoff for 429 errors
- Use multiple API keys for production

---

### Anthropic

**Rate Limits**:
- Tier-based (varies by plan)
- Requests per day and tokens per day
- Contact Anthropic for higher limits

**Error Codes**:
- `401`: Invalid API key
- `429`: Rate limit exceeded
- `529`: Service overloaded

**Best Practices**:
- Monitor API key usage
- Implement retry logic for 529 errors
- Use multiple API keys for production

## Proxy Configuration

All providers support proxy configuration through the account's assigned proxy.

**Proxy Assignment**:
- Automatic via fill-first algorithm
- Persistent per account
- Stored in `accounts.proxy_url` and `accounts.proxy_id`

**Proxy Format**:
```
http://username:password@proxy.example.com:8080
https://proxy.example.com:8080
socks5://proxy.example.com:1080
```

**Create Proxy**:
```bash
curl -X POST http://localhost:8080/api/v1/proxies \
  -H "Content-Type: application/json" \
  -d '{
    "label": "US Proxy 1",
    "proxy_url": "http://user:pass@proxy.example.com:8080",
    "is_active": true
  }'
```

**View Assignments**:
```bash
curl http://localhost:8080/api/v1/proxies/assignments
```

## Troubleshooting

### Provider Connection Errors

**Symptom**: "Failed to connect to provider"

**Possible Causes**:
1. Invalid base URL
2. Network connectivity issues
3. Proxy configuration errors

**Solutions**:
1. Verify provider base_url in database
2. Test network connectivity to provider
3. Verify proxy URL format and credentials

---

### Authentication Failures

**Symptom**: "401 Unauthorized" from provider

**Possible Causes**:
1. Invalid or expired API key
2. OAuth token expired
3. OAuth refresh token invalid

**Solutions**:
1. Verify API key is correct and active
2. Check OAuth token expiry in auth_data
3. Update refresh_token if invalid
4. Check Redis cache for stale tokens

---

### Rate Limit Errors

**Symptom**: "429 Too Many Requests"

**Possible Causes**:
1. Exceeded provider rate limits
2. Too few accounts for request volume
3. Account not rotating properly

**Solutions**:
1. Add more accounts for the provider
2. Verify round-robin is working (check Redis counters)
3. Contact provider for higher limits

---

### Token Refresh Failures

**Symptom**: "Failed to refresh OAuth token"

**Possible Causes**:
1. Invalid refresh_token
2. Invalid client_id or client_secret
3. OAuth credentials revoked

**Solutions**:
1. Obtain new OAuth credentials from provider
2. Update account auth_data with new credentials
3. Check token_url is correct

## Provider Metrics

Each provider tracks the following metrics:

**Request Logs** (`request_logs` table):
- Total requests
- Success rate
- Average latency
- Error distribution

**Proxy Stats** (`proxy_stats` table):
- Requests per proxy per day
- Success/failure counts
- Average latency per proxy

**Query Examples**:

```sql
-- Provider request volume (last 7 days)
SELECT provider_id, COUNT(*) as total_requests
FROM request_logs
WHERE created_at >= NOW() - INTERVAL 7 DAY
GROUP BY provider_id;

-- Provider success rate
SELECT provider_id,
       SUM(CASE WHEN status_code < 400 THEN 1 ELSE 0 END) / COUNT(*) as success_rate
FROM request_logs
GROUP BY provider_id;

-- Provider average latency
SELECT provider_id, AVG(latency_ms) as avg_latency
FROM request_logs
GROUP BY provider_id;
```

## Future Provider Support

Planned providers for future releases:

- **Cohere**: Text generation and embeddings
- **AI21 Labs**: Jurassic models
- **Hugging Face**: Inference API
- **Azure OpenAI**: Microsoft-hosted OpenAI models
- **AWS Bedrock**: Multi-model access via AWS
- **Google Vertex AI**: Google Cloud AI models

To request a new provider, please open a GitHub issue with:
- Provider name and API documentation
- Authentication requirements
- Desired models
- Use case description
