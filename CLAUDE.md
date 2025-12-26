# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Build
go build -o aigateway cmd/main.go

# Run (default port 8088)
go run cmd/main.go

# Test
go test ./...
go test -cover ./...
go test ./services    # Single package

# Database setup
./scripts/init.sh     # Creates DB, runs migrations, seeds data
```

## Architecture Overview

AIGateway is a multi-provider AI API gateway that routes requests to different AI providers (Anthropic via Antigravity, OpenAI, GLM) through a unified interface.

### Request Pipeline

```
HTTP Request → ProxyHandler → ExecutorService → RouterService
    → AccountService (round-robin) → ProxyService (fill-first)
    → OAuthService (token caching) → Provider.Execute
    → StatsTrackerService (async) → Response
```

### Key Layers

| Layer | Purpose | Location |
|-------|---------|----------|
| Handlers | HTTP request/response | `handlers/` |
| Services | Business logic orchestration | `services/` |
| Providers | AI provider implementations | `providers/` |
| Repositories | Database CRUD operations | `repositories/` |
| Auth | Authentication strategies | `auth/` |
| Models | Data structures | `models/` |

### Provider System

All providers implement `providers.Provider` interface:
- **Antigravity** (`providers/antigravity/`) - Google Cloud Code API, OAuth auth, supports Gemini + Claude models
- **OpenAI** (`providers/openai/`) - GPT models, API Key auth
- **GLM** (`providers/glm/`) - Chinese LLMs, Bearer token auth

Model routing in `providers/registry.go`:
- `gemini-*`, `claude-sonnet-*` → Antigravity
- `gpt-*` → OpenAI
- `glm-*` → GLM

### Core Services

- **RouterService** (`services/router.service.go`) - Model-to-provider routing, full pipeline orchestration
- **AccountService** (`services/account.service.go`) - Round-robin account selection using Redis atomic counters
- **ProxyService** (`services/proxy.service.go`) - Fill-first proxy assignment with capacity tracking
- **OAuthService** (`services/oauth.service.go`) - Token caching in Redis, auto-refresh before expiry
- **StatsTrackerService** (`services/stats.tracker.service.go`) - Async request logging and daily aggregation

### Redis Keys

- `account:rr:{provider}:{model}` - Round-robin counter
- `auth:{provider}:{account_id}` - Cached OAuth tokens
- `stats:proxy:{id}:requests:today` - Daily request count
- `stats:proxy:{id}:errors:today` - Daily error count

### Database

MySQL with GORM auto-migration. 5 tables: `providers`, `accounts`, `proxy_pool`, `proxy_stats`, `request_logs`.

Config: `config/config.yaml`

### Default Ports

| Service | Port |
|---------|------|
| Server | 8088 |
| MySQL | 3306 |
| Redis | 6380 |

## API Reference

Full API spec: `openapi/index.yaml` (entry point with navigation to endpoint-specific files)

```
openapi/
├── index.yaml          # Entry point, endpoint map
├── paths/              # Endpoint definitions
│   ├── proxy.yaml      # /v1/messages, /v1/chat/completions
│   ├── accounts.yaml   # /api/v1/accounts/*
│   ├── proxies.yaml    # /api/v1/proxies/*
│   ├── stats.yaml      # /api/v1/stats/*
│   ├── providers.yaml  # /api/v1/providers
│   └── health.yaml     # /health
└── schemas/            # Data models
    ├── account.yaml
    ├── proxy.yaml
    ├── provider.yaml
    └── common.yaml
```

## Concurrency Patterns

- Redis atomic ops for counters (no locks)
- RWMutex for HTTP client pool in Antigravity provider
- Goroutines for async stats recording
- Token refresh is synchronous (blocks on HTTP)
