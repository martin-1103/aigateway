# Unified Account Creation Implementation Plan

## Overview
Transform fragmented OAuth + Account management into single unified flow with one provider endpoint and one account creation dialog.

---

## Phase 1: Database Schema Changes

### 1.1 Database Migration

**BACKUP FIRST** - Critical changes to providers table

```sql
-- Step 1: Add new column for supported auth types
ALTER TABLE providers
ADD COLUMN supported_auth_types JSON NOT NULL DEFAULT JSON_ARRAY('api_key')
AFTER name;

-- Step 2: Migrate existing data
UPDATE providers SET supported_auth_types = JSON_ARRAY('oauth') WHERE id='antigravity';
UPDATE providers SET supported_auth_types = JSON_ARRAY('oauth','api_key') WHERE id='openai';
UPDATE providers SET supported_auth_types = JSON_ARRAY('api_key') WHERE id='glm';

-- Step 3: Drop redundant enum column
ALTER TABLE providers DROP COLUMN auth_type;

-- Step 4: Make is_active non-nullable with default
ALTER TABLE providers MODIFY is_active TINYINT(1) NOT NULL DEFAULT 1;

-- Step 5: Add index for active filtering
ALTER TABLE providers ADD INDEX idx_is_active (is_active);
```

### 1.2 Data Verification

```sql
-- Verify migration
SELECT id, name, supported_auth_types FROM providers;

-- Should show:
-- antigravity | Antigravity | ["oauth"]
-- openai | OpenAI | ["oauth","api_key"]
-- glm | Zhipu AI (GLM) | ["api_key"]
```

---

## Phase 2: Backend Code Changes

### 2.1 Update Provider Model (`models/provider.model.go`)

**Changes:**
- Remove: `AuthType` field (enum)
- Add: `SupportedAuthTypes []string` field
- Keep: `AuthStrategy` (used by provider implementations)
- Update: JSON tags

```go
type Provider struct {
	ID                 string    `gorm:"primaryKey;size:50" json:"id"`
	Name               string    `gorm:"size:100;not null" json:"name"`
	BaseURL            string    `gorm:"size:255" json:"base_url"`
	SupportedAuthTypes []string  `gorm:"type:json;not null" json:"supported_auth_types"`
	AuthStrategy       string    `gorm:"size:50;not null" json:"auth_strategy"`
	SupportedModels    string    `gorm:"type:json" json:"supported_models"`
	IsActive           bool      `gorm:"not null;index:idx_is_active;default:1" json:"is_active"`
	Config             string    `gorm:"type:json" json:"config"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
```

**Files affected:** `models/provider.model.go`

---

### 2.2 Update Router Service (`services/router.service.go`)

**Changes:**
- Update `ProviderInfo` struct to use `SupportedAuthTypes` array instead of `AuthType` string
- Update `ListProviders()` method to fetch from database instead of registry

**Old code:**
```go
type ProviderInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
	AuthType string `json:"auth_type"`           // ← OLD: single value
	IsActive bool   `json:"is_active"`
}
```

**New code:**
```go
type ProviderInfo struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	BaseURL            string   `json:"base_url"`
	SupportedAuthTypes []string `json:"supported_auth_types"`  // ← NEW: array
	IsActive           bool     `json:"is_active"`
}
```

**Update ListProviders() method:**
```go
// OLD: From registry (code-based providers)
func (s *RouterService) ListProviders() []ProviderInfo {
	providerList := s.registry.ListActive()
	// ... builds from registry
}

// NEW: From database (actual configured providers)
func (s *RouterService) ListProviders() []ProviderInfo {
	var dbProviders []models.Provider
	if err := s.db.Where("is_active = ?", true).Find(&dbProviders).Error; err != nil {
		log.Printf("error fetching providers: %v", err)
		return []ProviderInfo{}
	}

	result := make([]ProviderInfo, 0, len(dbProviders))
	for _, p := range dbProviders {
		result = append(result, ProviderInfo{
			ID:                 p.ID,
			Name:               p.Name,
			BaseURL:            p.BaseURL,
			SupportedAuthTypes: p.SupportedAuthTypes,
			IsActive:           p.IsActive,
		})
	}
	return result
}
```

**Note:** RouterService needs DB connection - add to constructor:
```go
type RouterService struct {
	db                  *gorm.DB  // ← Add this
	registry            *providers.Registry
	// ... rest of fields
}
```

**Files affected:** `services/router.service.go`, `main.go` (constructor update)

---

### 2.3 Create Provider Repository (`repositories/provider.repository.go`)

**New file** - Simple provider queries

```go
package repositories

import (
	"aigateway-backend/models"
	"gorm.io/gorm"
)

type ProviderRepository struct {
	db *gorm.DB
}

func NewProviderRepository(db *gorm.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

func (r *ProviderRepository) GetByID(id string) (*models.Provider, error) {
	var provider models.Provider
	if err := r.db.First(&provider, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

func (r *ProviderRepository) GetActive(limit, offset int) ([]models.Provider, int64, error) {
	var providers []models.Provider
	var total int64

	if err := r.db.Where("is_active = ?", true).Count(&total).
		Limit(limit).Offset(offset).Find(&providers).Error; err != nil {
		return nil, 0, err
	}
	return providers, total, nil
}
```

**Files to create:** `repositories/provider.repository.go`

---

### 2.4 Update OAuthFlowService (Simplify)

**Key change:** When OAuth flow completes, create account via AccountService

**Changes:**
- OAuth service returns provider ID and token
- AccountService.CreateWithOAuth() handles the account creation
- Remove hardcoded logic, delegate to repo

**Files affected:** `services/oauth.flow.service.go`

---

### 2.5 Update Account Handler & Service

**Account creation should support both:**
1. OAuth callback → create account with token
2. Direct API key → create account with user-provided credentials

**New request structure:**
```go
type CreateAccountRequest struct {
	ProviderID string `json:"provider_id" binding:"required"`
	Label      string `json:"label" binding:"required"`
	AuthData   string `json:"auth_data" binding:"required"`  // JSON string
	AuthType   string `json:"auth_type" binding:"required"`  // "oauth" or "api_key"
	IsActive   bool   `json:"is_active"`
}
```

**Files affected:**
- `handlers/account.handler.go` - accept new request structure
- `services/account.service.go` - validate auth_data based on provider's supported_auth_types

---

### 2.6 Keep OAuth Package But Simplify Handler

**Keep:**
- `auth/oauth/` package (low-level token exchange)
- `services/oauth.service.go` (token refresh/caching)

**Remove/Merge:**
- `services/oauth.flow.service.go` - merge into account creation flow
- `handlers/oauth.handler.go` - endpoints merged into account handler

---

## Phase 3: Frontend Changes

### 3.1 Update Provider Types (`accounts.types.ts`)

```ts
interface Provider {
  id: string
  name: string
  base_url: string
  supported_auth_types: string[]  // ["oauth"], ["api_key"], or ["oauth", "api_key"]
  supported_models: string[]
  is_active: boolean
}

interface CreateAccountRequest {
  provider_id: string
  label: string
  auth_data: string  // JSON for API key or OAuth token
  auth_type: 'oauth' | 'api_key'
  is_active?: boolean
}
```

### 3.2 Unified Account Creation Dialog

**Remove:**
- `features/oauth/` folder (entire OAuth-specific page)
- OAuth form logic from account dialog

**Update:**
- `account-create-dialog.tsx` - single dialog supporting both auth types
- Fetch providers from `GET /api/v1/providers`
- Show auth type options based on `supported_auth_types`
- Handle OAuth popup within account creation flow

**Flow:**
```
User clicks "Add Account"
↓
Fetch providers from GET /api/v1/providers
↓
User selects provider (shows auth options based on supported_auth_types)
↓
If has "oauth" option: Show OAuth button → Open popup → Get token
If has "api_key" option: Show credential input field
↓
Submit POST /api/v1/accounts with auth_data and auth_type
↓
Account created ✓
```

---

## Phase 4: API Changes

### 4.1 Unified Endpoint

**GET /api/v1/providers**
- Returns all active providers with `supported_auth_types`
- Response includes metadata for UI to determine which auth methods to show

```json
{
  "providers": [
    {
      "id": "openai",
      "name": "OpenAI",
      "base_url": "https://api.openai.com/v1",
      "supported_auth_types": ["oauth", "api_key"],
      "supported_models": ["gpt-4o", "gpt-4o-mini"],
      "is_active": true
    },
    {
      "id": "antigravity",
      "name": "Google Cloud Code (Antigravity)",
      "supported_auth_types": ["oauth"],
      "supported_models": ["gemini-2.5-flash", "claude-opus-4-5"],
      "is_active": true
    },
    {
      "id": "glm",
      "name": "Zhipu AI (GLM)",
      "supported_auth_types": ["api_key"],
      "supported_models": ["glm-4", "glm-4-flash"],
      "is_active": true
    }
  ]
}
```

### 4.2 Routes Changes

**Keep:**
- `POST /api/v1/accounts` - create account (handles both OAuth + API key)
- `GET /api/v1/accounts` - list accounts
- `PUT /api/v1/accounts/:id` - update
- `DELETE /api/v1/accounts/:id` - delete

**Replace:**
- `GET /api/v1/providers` - now returns database providers with supported_auth_types

**Remove (cleanup only):**
- `GET /api/v1/oauth/providers`
- `POST /api/v1/oauth/init`
- `GET /api/v1/oauth/callback`
- `POST /api/v1/oauth/exchange`
- `POST /api/v1/oauth/refresh`

---

## Implementation Order (Critical)

1. **Database migration** (backup required)
2. **Provider model update** (GORM field changes)
3. **Router service update** (add DB query capability)
4. **Provider repository** (new file)
5. **Account creation** (support both auth types)
6. **Frontend types** (Provider interface)
7. **Unified dialog** (single account creation UI)
8. **Route consolidation** (keep /accounts, remove /oauth)
9. **OAuth cleanup** (remove handlers and routes)
10. **Frontend cleanup** (remove oauth folder)

---

## Backward Compatibility Notes

**Breaking changes:**
- `/api/v1/oauth/providers` endpoint removed
- Provider model JSON now has `supported_auth_types` instead of `auth_type`
- API responses for accounts with nested providers will now include `supported_auth_types`

**Migration path:**
- Old clients expecting `auth_type`: Will break, need update to use `supported_auth_types` array
- New clients: Use `supported_auth_types` to determine UI options

---

## Testing Checklist

- [ ] Database migration executes without errors
- [ ] Provider model loads correctly from DB with JSON array
- [ ] GET /api/v1/providers returns all 3 providers with correct supported_auth_types
- [ ] Account creation with API key creates account with auth_data
- [ ] Account creation with OAuth creates account with token
- [ ] Account list shows nested provider with supported_auth_types
- [ ] Frontend account dialog shows OAuth option for openai/antigravity
- [ ] Frontend account dialog shows API key option for openai/glm
- [ ] OAuth refresh still works with existing accounts
- [ ] Model routing still works correctly

---

## Files Summary

### Modified Files
- `models/provider.model.go` - Remove auth_type, add supported_auth_types
- `services/router.service.go` - Add DB, update ListProviders()
- `services/oauth.flow.service.go` - Simplify/merge logic
- `handlers/account.handler.go` - Accept new request format
- `services/account.service.go` - Validate auth types
- `frontend/src/features/accounts/accounts.types.ts` - Update Provider interface
- `routes/routes.go` - Keep /api/v1/providers, remove /oauth routes

### New Files
- `repositories/provider.repository.go` - Provider queries

### Deleted Files
- `handlers/oauth.handler.go` (functionality merged)
- `services/oauth.flow.service.go` (functionality merged)
- `frontend/src/features/oauth/` (entire folder)
- `frontend/src/features/oauth.service.ts` (if separate)

---

## Risk Assessment

**HIGH RISK:**
- Database schema change is irreversible (requires backup)
- Breaking API change (auth_type → supported_auth_types)

**MEDIUM RISK:**
- Provider model changes affect account/proxy preloads
- OAuth service refactoring could break token refresh

**LOW RISK:**
- Frontend changes (backward compatible if API returns both)
- New provider repository (new code, no existing dependencies)

---

## Rollback Plan

1. Restore database backup
2. Revert Provider model to use enum auth_type
3. Keep both AuthType and SupportedAuthTypes in responses (dual support)
4. Gradually migrate frontend to new field

