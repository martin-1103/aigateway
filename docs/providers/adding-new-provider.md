# Adding a New Provider

This guide walks you through the process of adding support for a new AI provider to AIGateway.

## Prerequisites

- Access to the provider's API documentation
- API credentials (API key, OAuth client, etc.)
- Understanding of the provider's API format
- Go development environment

## Step 1: Define Provider Configuration

Insert provider metadata into the database.

### Database Provider Entry

```sql
INSERT INTO providers (id, name, base_url, auth_type, auth_strategy, supported_models, is_active, config)
VALUES (
  'new_provider',              -- Unique provider ID
  'New Provider',              -- Display name
  'https://api.newprovider.com', -- Base API URL
  'api_key',                   -- Auth type: oauth, api_key, bearer
  'api_key',                   -- Auth strategy name
  '["model-1", "model-2"]',    -- JSON array of supported models
  true,                        -- Active status
  '{}'                         -- Additional config (JSON)
);
```

### Provider Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `id` | VARCHAR(50) | Unique provider identifier | `new_provider` |
| `name` | VARCHAR(100) | Human-readable name | `New Provider` |
| `base_url` | VARCHAR(255) | Provider API base URL | `https://api.newprovider.com` |
| `auth_type` | ENUM | Authentication type | `oauth`, `api_key`, `bearer` |
| `auth_strategy` | VARCHAR(50) | Strategy implementation name | `api_key` |
| `supported_models` | JSON | Array of model names | `["model-1", "model-2"]` |
| `is_active` | BOOLEAN | Enable/disable provider | `true` |
| `config` | JSON | Additional configuration | `{}` |

## Step 2: Update Model Routing

Edit `providers/registry.go` to add routing logic for the new provider.

### Add Routing Case

```go
func (r *Registry) routeModel(model string) string {
    modelLower := strings.ToLower(model)

    switch {
    case strings.HasPrefix(modelLower, "newmodel-"):
        return "new_provider"
    // ... existing cases
    case strings.HasPrefix(modelLower, "gemini-"):
        return "antigravity"
    case strings.HasPrefix(modelLower, "gpt-"):
        return "openai"
    default:
        return ""
    }
}
```

### Routing Strategy

Choose routing strategy based on model naming:

**Prefix matching** (recommended):
```go
case strings.HasPrefix(modelLower, "prefix-"):
    return "provider_id"
```

**Exact matching**:
```go
case modelLower == "specific-model":
    return "provider_id"
```

**Pattern matching**:
```go
case strings.Contains(modelLower, "pattern"):
    return "provider_id"
```

## Step 3: Create Format Translator (if needed)

If the provider uses a non-standard format, add translation logic.

### Determine If Translation Needed

**No translation needed** if provider uses:
- OpenAI-compatible format
- Anthropic Messages format
- Standard REST JSON

**Translation needed** if provider uses:
- Custom JSON structure
- Protobuf/gRPC
- XML or other formats

### Implement Translation Functions

Edit `services/translator.service.go`:

```go
// Request translation: Claude format → Provider format
func (s *TranslatorService) ClaudeToNewProvider(payload []byte, model string) ([]byte, error) {
    // Parse Claude format
    var claudeReq struct {
        Model       string                   `json:"model"`
        Messages    []map[string]interface{} `json:"messages"`
        MaxTokens   int                      `json:"max_tokens"`
        Temperature float64                  `json:"temperature"`
        System      string                   `json:"system,omitempty"`
    }

    if err := json.Unmarshal(payload, &claudeReq); err != nil {
        return nil, fmt.Errorf("invalid claude format: %w", err)
    }

    // Transform to provider format
    providerReq := map[string]interface{}{
        "model": model,
        "messages": claudeReq.Messages, // Or transform as needed
        "config": map[string]interface{}{
            "max_tokens":  claudeReq.MaxTokens,
            "temperature": claudeReq.Temperature,
        },
    }

    if claudeReq.System != "" {
        providerReq["system"] = claudeReq.System
    }

    return json.Marshal(providerReq)
}

// Response translation: Provider format → Claude format
func (s *TranslatorService) NewProviderToClaude(payload []byte) ([]byte, error) {
    // Parse provider format
    var providerResp struct {
        Choices []struct {
            Message struct {
                Role    string `json:"role"`
                Content string `json:"content"`
            } `json:"message"`
        } `json:"choices"`
        Usage struct {
            PromptTokens     int `json:"prompt_tokens"`
            CompletionTokens int `json:"completion_tokens"`
        } `json:"usage"`
    }

    if err := json.Unmarshal(payload, &providerResp); err != nil {
        return nil, fmt.Errorf("invalid provider format: %w", err)
    }

    // Transform to Claude format
    claudeResp := map[string]interface{}{
        "role": "assistant",
        "content": []map[string]interface{}{
            {
                "type": "text",
                "text": providerResp.Choices[0].Message.Content,
            },
        },
        "stop_reason": "end_turn",
        "usage": map[string]interface{}{
            "input_tokens":  providerResp.Usage.PromptTokens,
            "output_tokens": providerResp.Usage.CompletionTokens,
        },
    }

    return json.Marshal(claudeResp)
}
```

### Update Executor Service

Add translation calls in `services/executor.service.go`:

```go
func (s *ExecutorService) Execute(ctx context.Context, req Request) (Response, error) {
    // ... existing code ...

    // Translate request
    var translatedReq []byte
    var err error

    switch req.ProviderID {
    case "new_provider":
        translatedReq, err = s.translator.ClaudeToNewProvider(req.Payload, req.Model)
    // ... existing cases ...
    default:
        translatedReq = req.Payload
    }

    // ... execute request ...

    // Translate response
    var translatedResp []byte
    switch req.ProviderID {
    case "new_provider":
        translatedResp, err = s.translator.NewProviderToClaude(httpResp.Body)
    // ... existing cases ...
    default:
        translatedResp = httpResp.Body
    }

    return Response{Payload: translatedResp}, nil
}
```

## Step 4: Create Account Credentials

Add authentication credentials via API or SQL.

### Via API

```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "new_provider",
    "label": "New Provider Account 1",
    "auth_data": {
      "api_key": "npk-xxx"
    },
    "is_active": true
  }'
```

### Via SQL

```sql
INSERT INTO accounts (id, provider_id, label, auth_data, is_active)
VALUES (
  UUID(),
  'new_provider',
  'New Provider Account 1',
  '{"api_key": "npk-xxx"}',
  true
);
```

### Auth Data Examples

**API Key**:
```json
{"api_key": "key-value"}
```

**OAuth 2.0**:
```json
{
  "access_token": "token",
  "refresh_token": "refresh",
  "token_url": "https://provider.com/oauth/token",
  "client_id": "id",
  "client_secret": "secret",
  "expires_at": "2024-12-26T10:00:00Z",
  "expires_in": 3600
}
```

**Bearer Token**:
```json
{"token": "bearer-token"}
```

## Step 5: Test Integration

Test the new provider integration thoroughly.

### Basic Request Test

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "newmodel-latest",
    "messages": [
      {"role": "user", "content": "test"}
    ],
    "max_tokens": 100
  }'
```

### Expected Response

```json
{
  "role": "assistant",
  "content": [
    {"type": "text", "text": "Response from new provider"}
  ],
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 5,
    "output_tokens": 7
  }
}
```

### Test Checklist

- [ ] Basic request/response works
- [ ] Model routing is correct
- [ ] Authentication succeeds
- [ ] Format translation is accurate
- [ ] Error handling works
- [ ] Rate limiting is respected
- [ ] Token counting is accurate
- [ ] Streaming works (if supported)

## Step 6: Add Unit Tests

Create tests for the new provider.

### Translation Tests

```go
// File: services/translator_test.go

func TestClaudeToNewProvider(t *testing.T) {
    translator := NewTranslatorService()

    input := `{
        "model": "newmodel-1",
        "messages": [{"role": "user", "content": "test"}],
        "max_tokens": 100
    }`

    result, err := translator.ClaudeToNewProvider([]byte(input), "newmodel-1")
    assert.NoError(t, err)

    // Verify output format
    var output map[string]interface{}
    json.Unmarshal(result, &output)
    assert.Equal(t, "newmodel-1", output["model"])
}

func TestNewProviderToClaude(t *testing.T) {
    translator := NewTranslatorService()

    input := `{
        "choices": [{
            "message": {"role": "assistant", "content": "test response"}
        }],
        "usage": {"prompt_tokens": 5, "completion_tokens": 7}
    }`

    result, err := translator.NewProviderToClaude([]byte(input))
    assert.NoError(t, err)

    // Verify Claude format
    var output map[string]interface{}
    json.Unmarshal(result, &output)
    assert.Equal(t, "assistant", output["role"])
}
```

### Integration Tests

```go
// File: integration_test.go

func TestNewProviderIntegration(t *testing.T) {
    // Setup
    app := setupTestApp()

    // Create request
    req := httptest.NewRequest("POST", "/v1/messages", strings.NewReader(`{
        "model": "newmodel-1",
        "messages": [{"role": "user", "content": "test"}],
        "max_tokens": 100
    }`))

    // Execute
    resp := httptest.NewRecorder()
    app.ServeHTTP(resp, req)

    // Verify
    assert.Equal(t, 200, resp.Code)

    var result map[string]interface{}
    json.Unmarshal(resp.Body.Bytes(), &result)
    assert.Equal(t, "assistant", result["role"])
}
```

## Step 7: Document the Provider

Add provider-specific documentation.

### Create Provider Doc

Create `docs/providers/newprovider.md`:

```markdown
# New Provider

**Provider ID**: `new_provider`

## Overview
Brief description of the provider.

## Configuration
Auth requirements and setup steps.

## Supported Models
List of available models.

## Rate Limits
Provider-specific limits and quotas.

## Error Codes
Common errors and solutions.

## Best Practices
Recommendations for using this provider.
```

### Update Provider README

Add entry to `docs/providers/README.md`:

```markdown
## Supported Providers

| Provider ID | Provider Name | Auth Type | Status |
|-------------|---------------|-----------|--------|
| `new_provider` | New Provider | API Key | Active |
```

## Troubleshooting

### Provider Not Found

**Symptom**: "provider not found: new_provider"

**Solution**: Verify provider exists in database:
```sql
SELECT * FROM providers WHERE id = 'new_provider';
```

### No Available Accounts

**Symptom**: "no available accounts for provider new_provider"

**Solution**: Create at least one active account

### Authentication Failures

**Symptom**: "401 Unauthorized"

**Solution**: Verify auth_data format matches auth_strategy requirements

### Format Translation Errors

**Symptom**: "invalid JSON" or malformed responses

**Solution**: Add logging to translation functions to debug format issues

## Related Documentation

- [Provider Overview](README.md) - All providers
- [Authentication Strategies](authentication.md) - Auth details
- [Format Translation](format-translation.md) - Translation guide
- [Architecture](../architecture/components.md) - Service layer details
