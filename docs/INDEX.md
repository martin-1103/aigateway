# AIGateway Documentation Index

Complete navigation guide for AIGateway documentation.

## Quick Start

- **[Getting Started](GETTING-STARTED.md)** - Installation, configuration, and first steps
- **[API Reference](API.md)** - Complete API endpoint documentation
- **[Documentation Overview](README.md)** - Introduction to AIGateway docs

## Architecture

**[Architecture Overview](architecture/README.md)** - System design, patterns, and request flow

- **[Component Details](architecture/components.md)** - Handler, Service, and Repository layers
- **[Database Schema](architecture/database.md)** - MySQL tables and Redis structures
- **[Concurrency & Error Handling](architecture/concurrency.md)** - Thread safety and error strategies
- **[Performance & Scalability](architecture/performance.md)** - Optimization and scaling
- **[Monitoring & Observability](architecture/monitoring.md)** - Metrics, logging, and alerts

## Providers

**[Provider Overview](providers/README.md)** - Supported providers and routing configuration

### Provider Documentation

- **[Antigravity (Google)](providers/antigravity.md)** - OAuth 2.0, Gemini & Claude models
- **[OpenAI](providers/openai.md)** - API key, GPT models
- **[Anthropic](providers/anthropic.md)** - API key, Claude models
- **[GLM (Zhipu AI)](providers/glm.md)** - API key, GLM models

### Provider Guides

- **[Authentication Strategies](providers/authentication.md)** - OAuth, API Key, Bearer token
- **[Format Translation](providers/format-translation.md)** - Request/response transformations
- **[Adding a New Provider](providers/adding-new-provider.md)** - Step-by-step integration guide

## Operations

**Operations Guides** - Production deployment and maintenance

- **[Database Setup](operations/database.md)** - Installation, seeding, migrations, backup
- **[Monitoring](operations/monitoring.md)** - Health checks, dashboards, alerting
- **[Troubleshooting](operations/troubleshooting.md)** - Common issues and solutions
- **[Security](operations/security.md)** - Best practices, compliance, incident response

## By Topic

### Getting Started
1. [Installation & Setup](GETTING-STARTED.md#installation)
2. [Configuration](GETTING-STARTED.md#configuration)
3. [Database Initialization](operations/database.md)
4. [First Request](GETTING-STARTED.md#making-your-first-request)

### Providers & Authentication
1. [Provider Overview](providers/README.md)
2. [Model Routing](providers/README.md#model-routing-configuration)
3. [OAuth 2.0 Setup](providers/authentication.md#oauth-20-strategy)
4. [API Key Setup](providers/authentication.md#api-key-strategy)
5. [Adding Providers](providers/adding-new-provider.md)

### Architecture & Design
1. [System Overview](architecture/README.md#overview)
2. [Design Patterns](architecture/README.md#core-design-patterns)
3. [Request Lifecycle](architecture/README.md#request-flow)
4. [Service Layer](architecture/components.md#service-layer)
5. [Database Design](architecture/database.md)

### Performance & Scaling
1. [Performance Optimization](architecture/performance.md#performance-optimizations)
2. [Horizontal Scaling](architecture/performance.md#horizontal-scaling)
3. [Bottlenecks & Mitigations](architecture/performance.md#bottlenecks-and-mitigations)
4. [Connection Pooling](architecture/performance.md#connection-pooling)

### Operations & Monitoring
1. [Health Checks](operations/monitoring.md#quick-health-checks)
2. [Real-Time Monitoring](operations/monitoring.md#real-time-monitoring)
3. [Alerting](operations/monitoring.md#alerting)
4. [Log Management](operations/monitoring.md#log-management)
5. [Troubleshooting](operations/troubleshooting.md)

### Security
1. [Credential Storage](operations/security.md#credential-storage)
2. [Access Control](operations/security.md#access-control)
3. [Token Security](operations/security.md#token-security)
4. [Network Security](operations/security.md#network-security)
5. [Security Checklist](operations/security.md#security-checklist)

## API Reference

### Proxy Endpoints
- [POST /v1/messages](API.md#post-v1messages) - Anthropic Claude format
- [POST /v1/chat/completions](API.md#post-v1chatcompletions) - OpenAI format

### Management Endpoints
- [Accounts API](API.md#accounts-api) - Create, list, update, delete accounts
- [Proxies API](API.md#proxies-api) - Manage proxy pool
- [Statistics API](API.md#statistics-api) - Usage metrics and logs

## Common Tasks

### Initial Setup
1. [Install AIGateway](GETTING-STARTED.md#installation)
2. [Configure Database](operations/database.md#usage)
3. [Add Provider](providers/adding-new-provider.md)
4. [Create Account](GETTING-STARTED.md#create-an-account)
5. [Test Integration](GETTING-STARTED.md#making-your-first-request)

### Adding a Provider
1. [Define Provider Config](providers/adding-new-provider.md#step-1-define-provider-configuration)
2. [Update Model Routing](providers/adding-new-provider.md#step-2-update-model-routing)
3. [Create Format Translator](providers/adding-new-provider.md#step-3-create-format-translator-if-needed)
4. [Create Account](providers/adding-new-provider.md#step-4-create-account-credentials)
5. [Test Integration](providers/adding-new-provider.md#step-5-test-integration)

### Troubleshooting
1. [Connection Issues](operations/troubleshooting.md#connection-issues)
2. [Authentication Issues](operations/troubleshooting.md#authentication-issues)
3. [Rate Limiting](operations/troubleshooting.md#rate-limiting-issues)
4. [Performance Issues](operations/troubleshooting.md#performance-issues)

### Monitoring Production
1. [Set Up Dashboards](operations/monitoring.md#dashboard-setup)
2. [Configure Alerts](operations/monitoring.md#alerting)
3. [Monitor Health](operations/monitoring.md#quick-health-checks)
4. [Review Logs](operations/monitoring.md#log-management)

## External Resources

- **GitHub Repository**: [github.com/yourorg/aigateway](https://github.com/yourorg/aigateway)
- **Issue Tracker**: [github.com/yourorg/aigateway/issues](https://github.com/yourorg/aigateway/issues)
- **Changelog**: [CHANGELOG.md](../CHANGELOG.md)
- **Contributing**: [CONTRIBUTING.md](../CONTRIBUTING.md)

## Document Conventions

### Code Examples
- Go code snippets show implementation examples
- SQL queries show database operations
- Bash commands show CLI operations
- YAML/JSON show configuration examples

### File References
- Relative paths from project root: `cmd/main.go`
- Full paths in examples: `/var/log/aigateway/app.log`

### Placeholders
- `{variable}` - Replace with actual value
- `your-value-here` - Replace with your configuration
- `...` - Additional content or parameters

## Getting Help

1. **Check Documentation**: Search this index and relevant docs
2. **Review Issues**: Check [GitHub issues](https://github.com/yourorg/aigateway/issues)
3. **Troubleshooting Guide**: See [operations/troubleshooting.md](operations/troubleshooting.md)
4. **Open an Issue**: Include:
   - Clear problem description
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant logs and configurations

## Contributing to Docs

To improve documentation:

1. Fork the repository
2. Edit markdown files in `/docs`
3. Test links and code examples
4. Submit pull request
5. See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines

## Documentation Updates

This documentation is version-controlled with the codebase. Check git history for changes:

```bash
# See recent doc changes
git log --oneline -- docs/

# See changes to specific doc
git log -p -- docs/architecture/README.md
```

---

**Last Updated**: 2024-12-26
**Version**: 1.0.0
