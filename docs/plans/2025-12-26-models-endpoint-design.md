# Design: GET /v1/models Endpoint + Model Naming Cleanup

## Overview

Add public endpoint for client discovery of available models, with data from database and Redis caching. Simplify model naming by removing unnecessary translation layers.

## Endpoint

**`GET /v1/models`**

Public endpoint (no auth required) for clients to discover available models.

### Response Format

```json
{
  "providers": [
    {
      "id": "antigravity",
      "name": "Google Cloud Code",
      "models": ["gemini-2.5-flash", "claude-sonnet-4-5", "claude-opus-4-5"]
    },
    {
      "id": "openai",
      "name": "OpenAI",
      "models": ["gpt-4o", "gpt-4o-mini", "gpt-4-turbo"]
    },
    {
      "id": "glm",
      "name": "Zhipu AI",
      "models": ["glm-4", "glm-4-flash", "glm-4-air", "glm-4-plus"]
    }
  ]
}
```

## Data Flow

1. Check Redis cache (`models:available`)
2. If cache miss, query `providers` table WHERE `is_active = 1`
3. Parse `supported_models` JSON column
4. Cache result in Redis (TTL: 5 minutes)
5. Return response

## Caching Strategy

- **Redis key:** `models:available`
- **TTL:** 5 minutes
- **Invalidation:** When provider is updated (update/create/delete)

## Model Naming Cleanup

### Problem

Current system has 3 different model name formats:
- Client request: `claude-sonnet-4-5-20250514`
- Database: `gemini-claude-sonnet-4-5`
- Upstream API: `claude-sonnet-4-5`

This requires translation functions that are hard to maintain.

### Solution

Standardize on upstream API names. No translation needed.

**Remove from `providers/antigravity/config.go`:**
- `NormalizeModelName()` function
- `Alias2ModelName()` function
- `SupportedModels` hardcoded list

**Update database:**
```sql
UPDATE providers SET supported_models =
  '["gemini-2.5-flash","gemini-2.5-flash-lite","gemini-2.5-pro","gemini-2.5-computer-use-preview-10-2025","gemini-3-pro-preview","gemini-3-pro-image-preview","gemini-3-flash-preview","claude-sonnet-4-5","claude-sonnet-4-5-thinking","claude-opus-4-5","claude-opus-4-5-thinking"]'
WHERE id = 'antigravity';
```

**Flow after cleanup:**
```
Client sends: "claude-sonnet-4-5"
Database has: "claude-sonnet-4-5"
Upstream API: "claude-sonnet-4-5"
```

No translation. Direct passthrough.

## Files to Create/Modify

### New Files
- `handlers/models.handler.go` - HTTP handler
- `services/models.service.go` - Business logic + caching

### Modified Files
- `routes/routes.go` - Add route
- `providers/antigravity/config.go` - Remove translation functions
- `providers/antigravity/provider.go` - Remove NormalizeModelName call
- `providers/antigravity/translator.request.go` - Remove Alias2ModelName call

## Implementation Order

1. Create models service with caching
2. Create models handler
3. Add route
4. Update database (rename gemini-claude-* to claude-*)
5. Remove translation functions from antigravity provider
6. Test endpoint
