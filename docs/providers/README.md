# Provider Overview

This document provides an overview of all supported AI providers, their configuration, and model routing logic.

## Supported Providers

AIGateway currently supports the following providers:

| Provider ID | Provider Name | Base URL | Auth Type | Status |
|-------------|---------------|----------|-----------|--------|
| `antigravity` | Google Antigravity | `https://cloudcode-pa.googleapis.com` | OAuth 2.0 | Active |
| `openai` | OpenAI | `https://api.openai.com` | API Key | Planned |
| `anthropic` | Anthropic | `https://api.anthropic.com` | API Key | Planned |
| `glm` | Zhipu AI (GLM) | `https://open.bigmodel.cn` | API Key | Planned |

## Model Routing Configuration

The `providers/registry.go` file contains the model routing logic that automatically routes requests based on model name prefixes:

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

### Routing Rules

| Model Prefix | Routed To | Example Models |
|--------------|-----------|----------------|
| `gemini-*` | Antigravity | `gemini-claude-sonnet-4-5`, `gemini-pro` |
| `claude-sonnet-*` | Antigravity | `claude-sonnet-4-5`, `claude-sonnet-3-5` |
| `gpt-*` | OpenAI | `gpt-4`, `gpt-3.5-turbo` |
| `glm-*` | GLM | `glm-4`, `glm-3-turbo` |

### Adding a New Routing Rule

To add a new routing rule:

1. Edit `providers/registry.go`
2. Add a new case in the switch statement
3. Return the provider ID for the model prefix

**Example**:
```go
case strings.HasPrefix(modelLower, "claude-opus-"):
    return "anthropic"
```

## Provider Details

Each provider has specific configuration requirements and capabilities:

- **[Antigravity](antigravity.md)** - Google's internal AI API gateway with OAuth 2.0
- **[OpenAI](openai.md)** - GPT models with API key authentication
- **[Anthropic](anthropic.md)** - Claude models with API key authentication
- **[GLM](glm.md)** - Zhipu AI models with API key authentication

## Authentication Strategies

AIGateway supports multiple authentication mechanisms:

- **OAuth 2.0**: Automatic token refresh, Redis caching
- **API Key**: Static key extraction
- **Bearer Token**: Simple bearer token authentication

See [Authentication Strategies](authentication.md) for detailed documentation.

## Format Translation

AIGateway automatically translates between different API formats:

- Claude format ↔ Antigravity (Gemini) format
- OpenAI format (native)
- Anthropic format (native)

See [Format Translation](format-translation.md) for detailed transformations.

## Adding a New Provider

To add support for a new AI provider, follow the step-by-step guide in [Adding a New Provider](adding-new-provider.md).

## Provider Metrics

All providers track the following metrics:

**Request Logs** (`request_logs` table):
- Total requests
- Success rate
- Average latency
- Error distribution

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

## Related Documentation

- [Authentication Strategies](authentication.md) - Detailed auth mechanisms
- [Format Translation](format-translation.md) - Request/response transformations
- [Adding a New Provider](adding-new-provider.md) - Step-by-step guide
