# AIGateway Documentation

Welcome to the AIGateway documentation. This guide will help you understand, deploy, and operate AIGateway for production use.

## What is AIGateway?

AIGateway is a production-ready API gateway that unifies access to multiple AI providers (Anthropic, OpenAI, Google, etc.) through a single standardized interface. It provides:

- **Multi-provider architecture** - Extensible plugin-based system
- **Intelligent request routing** - Automatic model-to-provider mapping
- **Account round-robin** - Load distribution across accounts
- **Proxy pool management** - Fill-first strategy with persistent assignment
- **Authentication strategies** - OAuth 2.0, API Key, Bearer token
- **Token caching** - Redis-based OAuth token cache with auto-refresh
- **Usage analytics** - MySQL daily aggregation + Redis real-time counters
- **Format translation** - Automatic request/response transformation

## Quick Navigation

### For First-Time Users
- **[Getting Started](GETTING-STARTED.md)** - Install, configure, and run your first request
- **[API Reference](API.md)** - Complete API endpoint documentation

### For Developers
- **[Architecture](architecture/README.md)** - System design and patterns
- **[Providers](providers/README.md)** - Provider integration guide
- **[Adding a Provider](providers/adding-new-provider.md)** - Step-by-step integration

### For Operators
- **[Database Setup](operations/database.md)** - Installation and seeding
- **[Monitoring](operations/monitoring.md)** - Health checks and dashboards
- **[Troubleshooting](operations/troubleshooting.md)** - Common issues
- **[Security](operations/security.md)** - Best practices

### Complete Index
- **[Documentation Index](INDEX.md)** - Complete navigation guide

## How AIGateway Works

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

## Key Features

### Multi-Provider Support

AIGateway supports multiple AI providers with automatic model routing:

| Provider | Auth Type | Models |
|----------|-----------|--------|
| Antigravity | OAuth 2.0 | Gemini, Claude (via Google) |
| OpenAI | API Key | GPT-4, GPT-3.5 |
| Anthropic | API Key | Claude Opus, Sonnet, Haiku |
| GLM | API Key | GLM-4, GLM-3 |

See [Provider Overview](providers/README.md) for details.

### Intelligent Routing

Requests are automatically routed based on model name:

```
claude-sonnet-4-5  → Antigravity
gpt-4             → OpenAI
claude-opus-4-5   → Anthropic
glm-4             → GLM
```

See [Model Routing](providers/README.md#model-routing-configuration) for configuration.

### Load Distribution

**Round-Robin Account Selection**:
- Per-provider, per-model distribution
- Redis-based atomic counters
- Persistent state across restarts

**Fill-First Proxy Assignment**:
- Automatic proxy assignment to accounts
- Capacity-based distribution
- Persistent assignments

See [Architecture](architecture/README.md) for implementation details.

### Authentication Strategies

**OAuth 2.0** (Antigravity):
- Automatic token refresh
- Redis caching (5-min buffer)
- Database persistence

**API Key** (OpenAI, Anthropic, GLM):
- Static key extraction
- No expiry management

**Bearer Token** (Custom providers):
- Simple bearer token auth

See [Authentication](providers/authentication.md) for details.

### Format Translation

Automatic translation between API formats:

**Claude → Antigravity**:
```json
{"system": "...", "messages": [...]}
→
{"systemInstruction": {...}, "contents": [...]}
```

**Antigravity → Claude**:
```json
{"candidates": [...], "usageMetadata": {...}}
→
{"role": "assistant", "content": [...], "usage": {...}}
```

See [Format Translation](providers/format-translation.md) for details.

## Documentation Structure

### Architecture Documentation
- **[Overview](architecture/README.md)** - System design and patterns
- **[Components](architecture/components.md)** - Layer-by-layer breakdown
- **[Database](architecture/database.md)** - Schema and Redis structures
- **[Concurrency](architecture/concurrency.md)** - Thread safety and errors
- **[Performance](architecture/performance.md)** - Optimization strategies
- **[Monitoring](architecture/monitoring.md)** - Observability

### Provider Documentation
- **[Overview](providers/README.md)** - All providers and routing
- **[Antigravity](providers/antigravity.md)** - Google OAuth setup
- **[OpenAI](providers/openai.md)** - GPT models
- **[Anthropic](providers/anthropic.md)** - Claude models
- **[GLM](providers/glm.md)** - Chinese language models
- **[Authentication](providers/authentication.md)** - Auth strategies
- **[Format Translation](providers/format-translation.md)** - Transformations
- **[Adding Provider](providers/adding-new-provider.md)** - Integration guide

### Operations Documentation
- **[Database Setup](operations/database.md)** - Installation and seeding
- **[Monitoring](operations/monitoring.md)** - Dashboards and alerts
- **[Troubleshooting](operations/troubleshooting.md)** - Common issues
- **[Security](operations/security.md)** - Best practices

## Common Use Cases

### Use Case 1: Multi-Account Load Balancing

**Scenario**: You have 10 OpenAI API keys and want to distribute load evenly.

**Solution**:
1. Add all 10 API keys as separate accounts
2. AIGateway automatically round-robins requests across all accounts
3. Monitor usage via statistics API

See [Getting Started](GETTING-STARTED.md) for account creation.

### Use Case 2: Provider Failover

**Scenario**: Use Antigravity for Claude Sonnet, fallback to Anthropic if unavailable.

**Solution**:
1. Configure both Antigravity and Anthropic providers
2. Update routing logic to try Antigravity first
3. Implement fallback in error handling

See [Adding a Provider](providers/adding-new-provider.md).

### Use Case 3: Proxy Rotation

**Scenario**: Route requests through multiple proxy servers for IP rotation.

**Solution**:
1. Add proxies to proxy pool
2. AIGateway automatically assigns accounts to proxies
3. Persistent assignments ensure consistent IPs per account

See [Architecture: Proxy Service](architecture/components.md#proxy-service).

### Use Case 4: Cost Optimization

**Scenario**: Track usage by account and optimize costs.

**Solution**:
1. Tag accounts with metadata (team, project, environment)
2. Query request_logs table for usage statistics
3. Analyze costs per account/provider/model

See [Monitoring](architecture/monitoring.md).

## Technology Stack

- **Language**: Go 1.21+
- **Web Framework**: Gin
- **Database**: MySQL 8.0+
- **Cache**: Redis 6.0+
- **ORM**: GORM

## Project Status

**Current Version**: 1.0.0

**Status**: Production-ready

**Supported Providers**:
- ✅ Antigravity (Active)
- 🔜 OpenAI (Planned)
- 🔜 Anthropic (Planned)
- 🔜 GLM (Planned)

See [Provider Overview](providers/README.md) for roadmap.

## Getting Help

1. **Documentation**: Start with [INDEX.md](INDEX.md)
2. **Troubleshooting**: Check [operations/troubleshooting.md](operations/troubleshooting.md)
3. **GitHub Issues**: [github.com/yourorg/aigateway/issues](https://github.com/yourorg/aigateway/issues)

## Contributing

We welcome contributions! See:
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Contribution guidelines
- [Adding a Provider](providers/adding-new-provider.md) - Provider integration guide

## License

MIT License - See [LICENSE](../LICENSE) for details.

## Next Steps

- **New Users**: Start with [Getting Started](GETTING-STARTED.md)
- **Developers**: Read [Architecture Overview](architecture/README.md)
- **Operators**: Review [Operations Guide](operations/monitoring.md)
- **Browse All**: See [Documentation Index](INDEX.md)
