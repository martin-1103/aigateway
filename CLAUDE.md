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

```
HTTP Request → ProxyHandler → ExecutorService → RouterService
    → AccountService (round-robin) → ProxyService (fill-first)
    → OAuthService (token caching) → Provider.Execute
    → StatsTrackerService (async) → Response
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
