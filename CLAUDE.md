# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Structure

```
aigateway/
├── backend/          # Go API Gateway (module: aigateway-backend)
├── frontend/         # React Dashboard (Vite + TypeScript)
├── mcp-server/       # Python MCP Server
├── docs/             # Shared documentation
│   ├── plans/        # Design documents
│   └── api/          # OpenAPI specs
└── scripts/          # Shared scripts
```

## Backend (Go)

### Configuration

**IMPORTANT:** Backend uses **ONLY ONE** config file: `backend/config/config.yaml`

Do NOT create duplicate config files in other locations. All configuration is loaded from this single file.

**Configuration priority:**
1. `config/config.yaml` - Primary config (required)
2. Environment variables - Override specific values (optional)
3. `.env` file - Override specific values (optional)

**Example override:**
```bash
# config.yaml has: auth_manager.enabled: true
# Override with env var:
export USE_AUTH_MANAGER=false

# Or in .env file:
USE_AUTH_MANAGER=false
```

Backend uses `config/config.yaml` for all settings. Environment variables can override config values.

**Optional:** Create `.env` file to override config:
```bash
cd backend
cp .env.example .env
```

**Key Settings:**
- `auth_manager.enabled: true` - Enables health-aware account selection with auto-retry (recommended)
- `USE_AUTH_MANAGER=true` - Env var override for auth_manager.enabled

### Build & Run

```bash
cd backend

# Build
go build -o aigateway.exe .

# Run (default port 8088)
go run .

# Test
go test ./...
go test -cover ./...
go test ./services    # Single package
```

### Structure

```
backend/
├── main.go           # Entry point
├── go.mod            # module aigateway-backend
├── config/           # config.yaml location
├── internal/         # Private packages
│   ├── config/       # Config loading
│   ├── database/     # MySQL/Redis connections
│   └── utils/        # Utilities
├── handlers/         # HTTP handlers
├── services/         # Business logic
├── models/           # Data structures
├── repositories/     # Database operations
├── providers/        # AI provider implementations
│   ├── antigravity/
│   ├── openai/
│   └── glm/
├── auth/             # Authentication strategies
├── middleware/       # HTTP middleware
└── routes/           # Route definitions
```

### Request Pipeline

**Default (Legacy):**
```
HTTP Request → ProxyHandler → ExecutorService → RouterService
    → AccountService (round-robin) → ProxyService (fill-first)
    → OAuthService (token caching) → Provider.Execute
    → StatsTrackerService (async) → Response
```

**With AuthManager (Recommended):**
```
HTTP Request → ProxyHandler → ExecutorService → RouterService
    → AuthManager.Select (health-aware) → Auto-retry on failure
    → OAuthService (token caching) → Provider.Execute
    → StatsTrackerService (async) → Response
```

**Enable AuthManager in `config/config.yaml`:**
```yaml
auth_manager:
  enabled: true  # Enable health-aware selection with auto-retry
  periodic_reconcile_interval_min: 5
  auto_retry: true
  max_retries: 3
```

### Provider System

All providers implement `providers.Provider` interface:
- **Antigravity** - Google Cloud Code API, OAuth auth, Gemini + Claude models
- **OpenAI** - GPT models, API Key auth
- **GLM** - Chinese LLMs, Bearer token auth

Model routing in `providers/registry.go`:
- `gemini-*`, `claude-sonnet-*` → Antigravity
- `gpt-*` → OpenAI
- `glm-*` → GLM

### Redis Keys

- `account:rr:{provider}:{model}` - Round-robin counter
- `auth:{provider}:{account_id}` - Cached OAuth tokens
- `stats:proxy:{id}:requests:today` - Daily request count

## Frontend (React)

### Setup

**IMPORTANT:** Frontend requires `.env` file to run properly. If missing, create it:

```bash
cd frontend
cp .env.example .env
```

Default `.env` contents:
```
VITE_API_URL=http://localhost:8088
VITE_APP_NAME=AIGateway
```

### Build & Run

```bash
cd frontend
npm install
npm run dev      # Development server (port 5173)
npm run build    # Production build
```

### Common Issues

**Login not working?**
- Check if `.env` file exists with correct `VITE_API_URL`
- Restart dev server after creating/modifying `.env`
- Backend must be running on port 8088

## MCP Server (Python)

```bash
cd mcp-server
pip install -r requirements.txt
python -m src.main
```

## API Reference

Full API spec: `docs/api/index.yaml`

## Default Ports

| Service | Port |
|---------|------|
| Backend | 8088 |
| Frontend | 5173 |
| MySQL | 3306 |
| Redis | 6380 |
