# Dashboard Lite Implementation Plan

Reference: `2025-12-29-dashboard-lite-design.md`

## Phase 1: Backend - User Model & Repository

### Task 1.1: Add AccessKey to User Model
**File:** `models/user.model.go`
```go
type User struct {
    // existing fields...
    AccessKey *string `json:"-" gorm:"size:64;uniqueIndex"`
}
```

### Task 1.2: Create Migration
**File:** `internal/database/migrations.go` or manual SQL
```sql
ALTER TABLE users ADD COLUMN access_key VARCHAR(64) UNIQUE;
CREATE INDEX idx_users_access_key ON users(access_key);
```

### Task 1.3: Add Repository Method
**File:** `repositories/user.repo.go`
```go
func (r *UserRepository) GetByAccessKey(key string) (*models.User, error)
func (r *UserRepository) UpdateAccessKey(userID, key string) error
```

---

## Phase 2: Backend - Lite Middleware

### Task 2.1: Create ValidateAccessKey Middleware
**File:** `middleware/lite.middleware.go`
- Extract key from query param or X-Access-Key header
- Validate against database
- Set current_user context

### Task 2.2: Add Rate Limiting
**File:** `middleware/lite.middleware.go`
- 10 requests per minute per IP
- Return 429 if exceeded

---

## Phase 3: Backend - Auth Endpoints

### Task 3.1: Add Key Management Endpoints
**File:** `handlers/auth.handler.go`
```go
GET  /api/v1/auth/my-key        → GetMyKey (masked)
POST /api/v1/auth/regenerate-key → RegenerateKey
```

### Task 3.2: Auto-generate Key on Login
**File:** `handlers/auth.handler.go`
- Check if user.AccessKey is nil after login
- Generate and save if nil

### Task 3.3: Auto-generate Key on User Create
**File:** `handlers/user.handler.go`
- Generate AccessKey when creating new user

---

## Phase 4: Backend - Lite Handler & Routes

### Task 4.1: Create Lite Handler
**File:** `handlers/lite.handler.go`
```go
GET  /api/v1/lite/me        → Get current user info
GET  /api/v1/lite/accounts  → List own accounts
GET  /api/v1/lite/api-keys  → List own API keys (admin/user only)
POST /api/v1/lite/oauth/init → Init OAuth flow
```

### Task 4.2: Add Lite Routes
**File:** `routes/routes.go`
```go
lite := api.Group("/lite")
lite.Use(middleware.ValidateAccessKey())
lite.Use(middleware.LiteRateLimit())
```

### Task 4.3: Lite OAuth Callback
**File:** `handlers/lite.handler.go`
- Handle OAuth callback with key in state
- Redirect back to lite dashboard

---

## Phase 5: Frontend - Lite Feature

### Task 5.1: Create Folder Structure
```
features/lite/
├── index.ts
├── lite-layout.tsx
├── lite-accounts-page.tsx
├── lite-api-keys-page.tsx
├── lite-oauth-callback.tsx
├── components/
│   ├── lite-header.tsx
│   ├── lite-nav.tsx
│   └── lite-add-oauth-button.tsx
├── api/
│   ├── lite.client.ts
│   ├── lite-me.api.ts
│   ├── lite-accounts.api.ts
│   └── lite-api-keys.api.ts
└── hooks/
    └── use-lite-auth.ts
```

### Task 5.2: Create Lite API Client
**File:** `features/lite/api/lite.client.ts`
- Axios instance with X-Access-Key header
- Read key from URL query param

### Task 5.3: Create Lite Layout
**File:** `features/lite/lite-layout.tsx`
- Validate key on mount
- Simple header with user info
- Navigation tabs

### Task 5.4: Create Lite Accounts Page
**File:** `features/lite/lite-accounts-page.tsx`
- List accounts (reuse existing table component)
- Add OAuth button

### Task 5.5: Create Lite OAuth Flow
**File:** `features/lite/lite-oauth-callback.tsx`
- Handle callback, show success/error
- Redirect back to accounts list

### Task 5.6: Add Routes
**File:** `routes/routes.tsx`
```tsx
{ path: '/lite', element: <LiteLayout /> }
```

---

## Phase 6: Frontend - Settings Page

### Task 6.1: Add Access Key Section to Settings
- Show masked key
- Copy button
- Regenerate button with confirmation

---

## Implementation Order

1. **Backend Model + Repo** (foundation)
2. **Backend Middleware** (auth for lite)
3. **Backend Auth Endpoints** (key management)
4. **Backend Lite Handler** (API endpoints)
5. **Frontend Lite Feature** (UI)
6. **Frontend Settings** (key management UI)
7. **Testing** (end-to-end)

## Testing Checklist

- [ ] Key auto-generated on user create
- [ ] Key auto-generated on login (existing users)
- [ ] Lite dashboard accessible with valid key
- [ ] Invalid key returns 401
- [ ] Rate limiting works
- [ ] OAuth flow works via lite
- [ ] Regenerate key invalidates old key
- [ ] Role-based access (provider can't see API keys)
