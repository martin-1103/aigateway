# Auth Manager & Rotation System - Overview

**Date:** 2025-12-27
**Status:** Draft

## Problem

Current limitations:
- Round-robin tanpa health check
- Error handling hanya status code, bukan error body
- Token refresh hanya antigravity
- Tidak ada cooldown/quota tracking per account per model

## Solution

Adopt reference architecture:
- **Auth Manager** - Centralized state management
- **Error Parser** - Parse error body per provider
- **Multi-provider Token Refresh** - Claude, Codex, Antigravity
- **Smart Rotation** - Health-aware dengan cooldown & quota

## Architecture

```
Request → RouterService → AuthManager.Select() → Provider
                              │
                              ▼
                      ┌───────────────┐
                      │  AuthManager  │
                      └───────┬───────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
  ErrorParser          TokenRefresh            Selector
  (per provider)       (per provider)      (health-aware)
```

## New Directory Structure

```
auth/
├── manager/
│   ├── manager.go      # Main orchestrator
│   ├── selector.go     # Account selection
│   ├── state.go        # State structs
│   └── refresh.go      # Background refresh
├── errors/
│   ├── parser.go       # Interface
│   ├── claude.go       # Claude parser
│   ├── codex.go        # Codex parser
│   └── antigravity.go  # Antigravity parser
├── claude/
│   └── auth.go         # Claude OAuth
└── codex/
    ├── auth.go         # Codex OAuth
    └── jwt.go          # JWT parsing
```

## Code Quality Rules

- **Max 300 lines per file** - Break into focused files if larger
- **Clear naming** - `types.go` for structs, `{feature}.go` for main logic
- **Single responsibility** - Each file has one purpose

## Related Docs

1. [Error Parser](./01-error-parser.md)
2. [Auth Manager](./02-auth-manager.md)
3. [Token Refresh](./03-token-refresh.md)
4. [Integration](./04-integration.md)
5. [Tasks](./05-tasks.md)
6. [File Structure](./06-file-structure.md) - Detailed file breakdown
