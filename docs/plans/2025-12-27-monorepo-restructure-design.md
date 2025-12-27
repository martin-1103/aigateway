# Monorepo Restructure Design

**Date:** 2025-12-27
**Goal:** Clean monorepo structure for easier maintenance

## Target Structure

```
aigateway/
├── backend/                 # Go API Gateway (nested module)
│   ├── main.go
│   ├── go.mod               # module aigateway-backend
│   ├── go.sum
│   ├── internal/
│   │   ├── config/
│   │   ├── database/
│   │   └── utils/
│   ├── handlers/
│   ├── services/
│   ├── models/
│   ├── repositories/
│   ├── providers/
│   ├── auth/
│   ├── middleware/
│   ├── routes/
│   ├── aid/
│   └── referensi/
│
├── frontend/                # React Dashboard (existing)
│
├── mcp-server/              # Python MCP Server (existing)
│
├── docs/
│   ├── plans/
│   └── api/                 # OpenAPI specs (from openapi/)
│
├── scripts/                 # Shared scripts
│
├── README.md
├── CLAUDE.md
└── .gitignore
```

## Migration Steps

1. **Prepare backend folder** - Create go.mod, copy Go folders, update imports
2. **Verify backend works** - Build and test before cleanup
3. **Reorganize docs** - Move openapi/ to docs/api/
4. **Cleanup root** - Remove old Go folders and artifacts
5. **Update configs** - CLAUDE.md, README.md, .gitignore

## Decisions

- **Nested module**: backend/ has its own go.mod for independent builds
- **internal/ folder**: config, database, utils are private packages
- **main.go at backend root**: Single entry point, no cmd/ subfolder needed
