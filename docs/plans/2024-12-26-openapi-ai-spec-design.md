# OpenAPI Spec for AI Consumption

## Goal

Create API documentation optimized for AI developer tools (Claude Code, Cursor, Copilot) with modular structure to minimize context loading.

## Design Decisions

### Modular File Structure

AI loads only what's needed per task:
- Entry point → specific paths file → relevant schemas
- No need to load entire spec for single endpoint work

### Simplified YAML Format

Not strict OpenAPI 3.0 - uses human/AI readable shorthand:
- Types inline with comments
- Union types with `|`
- Arrays with `[]` suffix
- No verbose JSON Schema

### Entry Point Pattern

`openapi/index.yaml` contains:
- `x-api-map`: endpoint groups with file locations
- `x-schemas`: schema file locations
- AI reads this first, navigates to specific files

## File Structure

```
openapi/
├── index.yaml              # Entry point (~50 lines)
├── paths/
│   ├── proxy.yaml          # AI provider endpoints
│   ├── accounts.yaml       # Account CRUD
│   ├── proxies.yaml        # Proxy pool CRUD
│   ├── stats.yaml          # Statistics queries
│   ├── providers.yaml      # Provider list
│   └── health.yaml         # Health check
└── schemas/
    ├── account.yaml        # Account model
    ├── proxy.yaml          # Proxy, ProxyStats, RequestLog
    ├── provider.yaml       # Provider model + routing
    └── common.yaml         # Shared types
```

## Context Loading Examples

| Task | Files Loaded |
|------|--------------|
| Work on accounts | index.yaml + paths/accounts.yaml + schemas/account.yaml |
| Debug proxy endpoint | index.yaml + paths/proxy.yaml |
| Understand stats | index.yaml + paths/stats.yaml + schemas/proxy.yaml |

## Verification

All specs verified against actual handlers:
- `handlers/account.handler.go`
- `handlers/proxy_mgmt.handler.go`
- `handlers/stats.handler.go`
- `handlers/proxy.handler.go`
- `providers/*/translator*.go`
