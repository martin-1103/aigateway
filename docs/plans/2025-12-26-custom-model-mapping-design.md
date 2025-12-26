# Custom Model Mapping Design

**Date:** 2025-12-26
**Status:** Approved

## Overview

Fitur untuk membuat alias model yang di-route ke provider dan upstream model tertentu.

**Contoh:** Client kirim `my-claude` → route ke provider `antigravity` dengan model `claude-sonnet-4-5`

## Decisions

| Aspect | Decision |
|--------|----------|
| Purpose | Alias → (provider, model) mapping |
| Storage | MySQL + Redis cache |
| Management | CRUD API `/api/v1/model-mappings` |
| Cache strategy | Write-through, lazy load, no preload |
| Fallback | Existing prefix matching tetap jalan |

## Data Model

### MySQL Table

```sql
CREATE TABLE model_mappings (
    id          BIGINT AUTO_INCREMENT PRIMARY KEY,
    alias       VARCHAR(100) NOT NULL UNIQUE,
    provider_id VARCHAR(50) NOT NULL,
    model_name  VARCHAR(100) NOT NULL,
    description VARCHAR(255),
    enabled     BOOLEAN DEFAULT TRUE,
    priority    INT DEFAULT 0,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    INDEX idx_alias (alias),
    INDEX idx_provider (provider_id)
);
```

### Go Struct

```go
type ModelMapping struct {
    ID          uint   `gorm:"primaryKey"`
    Alias       string `gorm:"uniqueIndex;size:100"`
    ProviderID  string `gorm:"size:50"`
    ModelName   string `gorm:"size:100"`
    Description string `gorm:"size:255"`
    Enabled     bool   `gorm:"default:true"`
    Priority    int    `gorm:"default:0"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
}
```

### Redis Key Pattern

- Key: `model:mapping:{alias}`
- Value: `{"provider_id": "antigravity", "model_name": "claude-sonnet-4-5"}`
- TTL: none (persistent sampai di-invalidate)

## Request Flow

```
Client Request (model: "my-claude")
    │
    ▼
┌─────────────────────────────┐
│  1. Check Redis cache       │
│     key: model:mapping:my-claude
└─────────────────────────────┘
    │
    ├── HIT ──► Dapat {provider_id, model_name}
    │
    ▼ MISS
┌─────────────────────────────┐
│  2. Query DB                │
│     SELECT * FROM model_mappings
│     WHERE alias = ? AND enabled = true
└─────────────────────────────┘
    │
    ├── FOUND ──► Cache ke Redis, lanjut
    │
    ▼ NOT FOUND
┌─────────────────────────────┐
│  3. Fallback: existing      │
│     prefix matching logic   │
│     (registry.routeModel)   │
└─────────────────────────────┘
```

## API Endpoints

**Base path:** `/api/v1/model-mappings`

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | List semua mappings (with pagination) |
| GET | `/:alias` | Get single mapping by alias |
| POST | `/` | Create new mapping |
| PUT | `/:alias` | Update existing mapping |
| DELETE | `/:alias` | Delete mapping |

### Request/Response Examples

**Create:**
```
POST /api/v1/model-mappings
{
    "alias": "my-claude",
    "provider_id": "antigravity",
    "model_name": "claude-sonnet-4-5",
    "description": "My default Claude model",
    "enabled": true
}

Response: 201 Created
{
    "alias": "my-claude",
    "provider_id": "antigravity",
    "model_name": "claude-sonnet-4-5",
    "description": "My default Claude model",
    "enabled": true,
    "created_at": "2025-12-26T10:00:00Z"
}
```

**List:**
```
GET /api/v1/model-mappings?page=1&limit=20

Response: 200 OK
{
    "data": [...],
    "total": 50,
    "page": 1,
    "limit": 20
}
```

### Validations

- `alias`: required, unique, alphanumeric + dash/underscore
- `provider_id`: required, must exist in registry
- `model_name`: required, must exist in provider's `supported_models` (e.g., `claude-sonnet-4-5`, `gpt-4o`, `glm-4-flash`)

## Cache Invalidation Logic

Write-through pattern:

```go
type ModelMappingService struct {
    repo  *ModelMappingRepository
    redis *redis.Client
}

func (s *ModelMappingService) Create(m *ModelMapping) error {
    // 1. Insert ke DB
    if err := s.repo.Create(m); err != nil {
        return err
    }
    // 2. Set ke Redis
    return s.cacheMapping(m)
}

func (s *ModelMappingService) Update(alias string, m *ModelMapping) error {
    // 1. Update DB
    if err := s.repo.Update(alias, m); err != nil {
        return err
    }
    // 2. Invalidate old key (kalau alias berubah)
    if alias != m.Alias {
        s.redis.Del(ctx, "model:mapping:"+alias)
    }
    // 3. Set new cache
    return s.cacheMapping(m)
}

func (s *ModelMappingService) Delete(alias string) error {
    // 1. Delete dari DB
    if err := s.repo.Delete(alias); err != nil {
        return err
    }
    // 2. Delete dari Redis
    return s.redis.Del(ctx, "model:mapping:"+alias).Err()
}

func (s *ModelMappingService) cacheMapping(m *ModelMapping) error {
    key := "model:mapping:" + m.Alias
    val, _ := json.Marshal(map[string]string{
        "provider_id": m.ProviderID,
        "model_name":  m.ModelName,
    })
    return s.redis.Set(ctx, key, val, 0).Err() // 0 = no expiry
}
```

**Edge cases:**
- DB success tapi Redis fail → log warning, next request akan cache dari DB
- Startup: tidak preload (lazy load saja)

## File Structure

### New Files

```
models/
    model.mapping.go              # ModelMapping struct

repositories/
    model.mapping.repository.go   # DB CRUD operations

services/
    model.mapping.service.go      # Business logic + cache

handlers/
    model.mapping.handler.go      # HTTP handlers

routes/
    model.mapping.routes.go       # Route registration
```

### Modified Files

```
providers/
    registry.go                   # Update GetByModel() untuk cek custom mapping dulu

services/
    router.service.go             # Inject ModelMappingService
```

### Integration Point

```go
// providers/registry.go

// Before (current)
func (r *Registry) GetByModel(model string) (Provider, error) {
    providerID := r.routeModel(model)  // hardcoded prefix matching
    ...
}

// After
func (r *Registry) GetByModel(model string) (Provider, error) {
    // 1. Check custom mapping first (via injected service)
    if mapping := r.mappingService.Resolve(model); mapping != nil {
        return r.Get(mapping.ProviderID)
    }
    // 2. Fallback to prefix matching
    providerID := r.routeModel(model)
    ...
}
```
