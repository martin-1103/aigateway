# AIGateway

Multi-provider AI API gateway with intelligent request routing, account management, and proxy pooling.

## Overview

AIGateway is a production-ready API gateway that unifies access to multiple AI providers (Anthropic, OpenAI, Google, etc.) through a single standardized interface. It provides automatic account rotation, proxy management, OAuth token caching, and comprehensive usage analytics.

## Features

- **Multi-provider architecture**: Extensible plugin-based system for AI providers
- **Intelligent request routing**: Automatic model-to-provider mapping
- **Account round-robin**: Load distribution across accounts per provider+model
- **Proxy pool management**: Fill-first strategy with persistent assignment
- **Authentication strategies**: OAuth 2.0, API Key, and Bearer token support
- **Token caching**: Redis-based OAuth token cache with auto-refresh
- **HTTP client pooling**: Connection reuse per proxy URL
- **Usage analytics**: MySQL daily aggregation + Redis real-time counters
- **Format translation**: Automatic request/response translation between formats

## Architecture

```
Client Request (OpenAI/Anthropic format)
  → Handler
  → Provider Registry (model-based routing)
  → Account Selector (round-robin per provider+model)
  → Proxy Assigner (fill-first with persistent assignment)
  → Auth Strategy (OAuth/API Key/Bearer)
  → Format Translator (input format → provider format)
  → HTTP Client (proxy-aware connection pool)
  → Provider API
  → Format Translator (provider format → output format)
  → Stats Tracker (async logging)
  → Client Response
```

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for detailed architecture documentation.

## Quick Start

### Prerequisites

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### Installation

```bash
cd aigateway

# Install dependencies
go mod download

# Configure the application
vim config/config.yaml

# Run the gateway
go run cmd/main.go
```

### Configuration

Edit `config/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "your-password"
  database: "aigateway"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

proxy:
  selection_strategy: "fill_first"
  health_check_interval: 60
  max_failures: 3
```

## API Endpoints

### AI Proxy Endpoints

Forward requests to AI providers with automatic routing:

```bash
# Anthropic Claude format
POST /v1/messages
Content-Type: application/json

{
  "model": "claude-sonnet-4-5",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "max_tokens": 1024
}

# OpenAI format
POST /v1/chat/completions
Content-Type: application/json

{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "Hello"}
  ],
  "max_tokens": 1024
}
```

### Management Endpoints

#### Accounts

```bash
# List all accounts
GET /api/v1/accounts

# Get account details
GET /api/v1/accounts/:id

# Create new account
POST /api/v1/accounts
{
  "provider_id": "anthropic",
  "label": "Production Account 1",
  "auth_data": {
    "api_key": "sk-ant-..."
  },
  "is_active": true
}

# Update account
PUT /api/v1/accounts/:id

# Delete account
DELETE /api/v1/accounts/:id
```

#### Proxies

```bash
# List proxies
GET /api/v1/proxies

# Get proxy details
GET /api/v1/proxies/:id

# Create proxy
POST /api/v1/proxies
{
  "label": "US Proxy 1",
  "proxy_url": "http://user:pass@proxy.example.com:8080",
  "is_active": true
}

# Update proxy
PUT /api/v1/proxies/:id

# Delete proxy
DELETE /api/v1/proxies/:id

# View proxy assignments
GET /api/v1/proxies/assignments

# Recalculate account distribution
POST /api/v1/proxies/recalculate
```

#### Statistics

```bash
# Get proxy usage statistics
GET /api/v1/stats/proxies/:id?days=7

# Get recent request logs
GET /api/v1/stats/logs?limit=100
```

## Provider Routing Logic

AIGateway automatically routes requests based on model name prefixes:

| Model Prefix | Provider | Example Models |
|--------------|----------|----------------|
| `gemini-*` | Antigravity | gemini-claude-sonnet-4-5, gemini-pro |
| `claude-sonnet-*` | Antigravity | claude-sonnet-4-5, claude-sonnet-3-5 |
| `gpt-*` | OpenAI | gpt-4, gpt-3.5-turbo |
| `glm-*` | GLM | glm-4, glm-3-turbo |

See [docs/PROVIDERS.md](docs/PROVIDERS.md) for complete provider documentation.

## Usage Examples

### Example 1: Using Anthropic Claude

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-5",
    "messages": [
      {"role": "user", "content": "Explain quantum computing"}
    ],
    "max_tokens": 1024
  }'
```

### Example 2: Using OpenAI GPT

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "user", "content": "Explain quantum computing"}
    ],
    "max_tokens": 1024
  }'
```

### Example 3: Streaming Response

```bash
curl -X POST "http://localhost:8080/v1/messages?stream=true" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-5",
    "messages": [
      {"role": "user", "content": "Write a story"}
    ],
    "max_tokens": 2048
  }'
```

## Authentication Strategies

AIGateway supports multiple authentication mechanisms:

### OAuth 2.0

For providers requiring OAuth (e.g., Google APIs):

```json
{
  "provider_id": "antigravity",
  "label": "OAuth Account",
  "auth_data": {
    "access_token": "ya29.xxx",
    "refresh_token": "1//xxx",
    "token_url": "https://oauth2.googleapis.com/token",
    "client_id": "xxx.apps.googleusercontent.com",
    "client_secret": "xxx",
    "expires_at": "2024-12-26T10:00:00Z"
  }
}
```

Features:
- Automatic token refresh before expiration
- Redis-based token caching (5-minute buffer)
- Database persistence of refreshed tokens

### API Key

For providers using static API keys (e.g., Anthropic, OpenAI):

```json
{
  "provider_id": "anthropic",
  "label": "Anthropic Account",
  "auth_data": {
    "api_key": "sk-ant-api03-xxx"
  }
}
```

### Bearer Token

For providers using bearer token authentication:

```json
{
  "provider_id": "custom",
  "label": "Custom Provider",
  "auth_data": {
    "token": "bearer-xxx"
  }
}
```

## Database Schema

AIGateway uses auto-migration via GORM. Tables:

| Table | Description |
|-------|-------------|
| `providers` | AI service provider configurations |
| `accounts` | Authentication credentials with proxy assignments |
| `proxy_pool` | Proxy server pool |
| `proxy_stats` | Daily aggregated usage statistics |
| `request_logs` | Request audit trail |

## Round-Robin and Proxy Logic

### Account Selection

Per-request round-robin based on provider + model:

- Redis counter: `account:rr:{provider}:{model}`
- Atomic increment with modulo for index calculation
- Persistent state across server restarts
- Fair distribution across all active accounts

### Proxy Assignment

Fill-first strategy with persistent assignment:

- Each account assigned to exactly one proxy
- Proxies filled sequentially to capacity
- Assignments stored in `accounts.proxy_id` and `accounts.proxy_url`
- Automatic recalculation when proxies disabled/removed
- Manual recalculation via `/api/v1/proxies/recalculate`

## Project Structure

```
aigateway/
├── cmd/main.go                      # Application entry point
├── config/
│   ├── config.go                   # Configuration loader
│   └── config.yaml                 # YAML configuration
├── database/
│   ├── mysql.go                    # MySQL connection
│   └── redis.go                    # Redis connection
├── models/
│   ├── provider.model.go           # Provider data model
│   ├── account.model.go            # Account data model
│   └── proxy.model.go              # Proxy + Stats data models
├── providers/
│   ├── provider.go                 # Provider interface
│   └── registry.go                 # Provider registry + routing
├── auth/
│   ├── strategy.go                 # Auth strategy interface
│   ├── oauth.strategy.go           # OAuth 2.0 implementation
│   ├── apikey.strategy.go          # API Key implementation
│   └── bearer.strategy.go          # Bearer token implementation
├── repositories/
│   ├── account.repository.go       # Account database operations
│   ├── proxy.repository.go         # Proxy database operations
│   └── stats.repository.go         # Statistics database operations
├── services/
│   ├── account.service.go          # Account round-robin logic
│   ├── proxy.service.go            # Proxy assignment logic
│   ├── oauth.service.go            # OAuth token management
│   ├── translator.service.go       # Format translation
│   ├── httpclient.service.go       # HTTP client pooling
│   ├── stats.service.go            # Statistics tracking
│   └── executor.service.go         # Request orchestration
├── handlers/
│   ├── proxy.handler.go            # Proxy endpoint handlers
│   ├── account.handler.go          # Account CRUD handlers
│   ├── proxy_mgmt.handler.go       # Proxy CRUD handlers
│   └── stats.handler.go            # Statistics handlers
└── routes/routes.go                 # Route definitions
```

## Adding a New Provider

Follow these steps to add support for a new AI provider:

### 1. Define Provider Configuration

Insert provider metadata into the database:

```sql
INSERT INTO providers (id, name, base_url, auth_type, auth_strategy, supported_models, is_active)
VALUES (
  'new_provider',
  'New Provider',
  'https://api.newprovider.com',
  'api_key',
  'api_key',
  '["model-1", "model-2"]',
  true
);
```

### 2. Update Model Routing

Edit `providers/registry.go` in the `routeModel()` function:

```go
func (r *Registry) routeModel(model string) string {
    modelLower := strings.ToLower(model)

    switch {
    case strings.HasPrefix(modelLower, "newmodel-"):
        return "new_provider"
    // ... existing cases
    }
}
```

### 3. Create Format Translator (if needed)

If the provider uses a non-standard format, add translation logic in `services/translator.service.go`:

```go
func (s *TranslatorService) ClaudeToNewProvider(payload []byte, model string) []byte {
    // Translate from Claude format to provider format
}

func (s *TranslatorService) NewProviderToClaude(payload []byte) []byte {
    // Translate from provider format to Claude format
}
```

### 4. Create Accounts

Add authentication credentials via API:

```bash
POST /api/v1/accounts
{
  "provider_id": "new_provider",
  "label": "New Provider Account 1",
  "auth_data": {
    "api_key": "npk-xxx"
  },
  "is_active": true
}
```

### 5. Test Integration

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "newmodel-latest",
    "messages": [{"role": "user", "content": "test"}],
    "max_tokens": 100
  }'
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./services

# Run with coverage
go test -cover ./...
```

### Building for Production

```bash
# Build binary
go build -o aigateway cmd/main.go

# Run production binary
./aigateway
```

### Docker Deployment

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o aigateway cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/aigateway .
COPY --from=builder /app/config/config.yaml ./config/
CMD ["./aigateway"]
```

## Monitoring and Observability

### Request Logs

All requests are logged asynchronously to `request_logs` table:

- Account ID
- Provider ID
- Model
- Status code
- Latency (ms)
- Error details
- Timestamp

### Statistics Aggregation

Daily statistics per proxy in `proxy_stats` table:

- Total requests
- Successful requests
- Failed requests
- Average latency
- Total tokens (if available)

### Real-time Monitoring

Redis counters for real-time metrics:

- `stats:proxy:{id}:requests` - Request counter
- `stats:proxy:{id}:failures` - Failure counter
- `account:rr:{provider}:{model}` - Round-robin position

## Troubleshooting

### Issue: "No available accounts for provider"

**Cause**: No active accounts configured for the requested provider/model.

**Solution**: Add accounts via `/api/v1/accounts` endpoint.

### Issue: OAuth token refresh failures

**Cause**: Invalid refresh token or expired credentials.

**Solution**: Update account with fresh OAuth credentials from provider.

### Issue: Proxy connection errors

**Cause**: Invalid proxy URL or proxy server unreachable.

**Solution**: Verify proxy URL format and connectivity, disable if needed.

### Issue: High latency

**Cause**: Proxy performance issues or provider rate limiting.

**Solution**: Review proxy statistics, redistribute accounts across proxies.

## Security Considerations

- Store sensitive credentials (API keys, OAuth tokens) encrypted at rest
- Use HTTPS for all provider API communications
- Rotate API keys and OAuth tokens regularly
- Implement rate limiting to prevent abuse
- Monitor for unusual usage patterns
- Use proxy authentication where available
- Keep Redis and MySQL access restricted to localhost or private network

## Performance Optimization

- **Connection pooling**: HTTP clients reused per proxy URL
- **Token caching**: OAuth tokens cached in Redis (reduces auth overhead)
- **Async logging**: Statistics recorded in background goroutines
- **Database indexing**: Optimized queries on provider_id, is_active
- **Redis counters**: Atomic operations for round-robin state

## License

MIT

## Contributing

1. Fork the repository
2. Create a feature branch
3. Implement changes with tests
4. Submit pull request with detailed description

## Support

For issues and feature requests, please open a GitHub issue with:
- Detailed description
- Steps to reproduce (for bugs)
- Expected vs actual behavior
- Relevant logs and error messages
