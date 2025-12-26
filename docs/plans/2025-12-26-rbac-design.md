# RBAC Design for AIGateway

## Overview

Add Role-Based Access Control (RBAC) to AIGateway with 3 roles: admin, user, provider. Authentication via JWT (management) + API Key (AI proxy + management).

## Decisions Made

| Topic | Decision |
|-------|----------|
| Auth mechanism | JWT + API Key |
| Permission model | Simple role-based (hardcoded per role) |
| API keys per user | Multiple, with labels |
| API key expiry | No expiry (revoke manually) |
| OAuth account visibility | Admin: all, Provider: own, User: none |
| Model mapping ownership | Global (admin) + User-owned |
| Initial setup | Auto-seed admin on first run |

---

## Database Schema

### New Tables

```sql
-- users table
CREATE TABLE users (
    id            VARCHAR(36) PRIMARY KEY,
    username      VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role          ENUM('admin','user','provider') NOT NULL,
    is_active     BOOL DEFAULT true,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- api_keys table
CREATE TABLE api_keys (
    id            VARCHAR(36) PRIMARY KEY,
    user_id       VARCHAR(36) NOT NULL,
    key_hash      VARCHAR(64) NOT NULL,      -- SHA-256
    key_prefix    VARCHAR(8) NOT NULL,       -- "ak_xxxx" for display
    label         VARCHAR(100),
    is_active     BOOL DEFAULT true,
    last_used_at  TIMESTAMP NULL,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE INDEX idx_key_hash (key_hash),
    INDEX idx_user_id (user_id)
);
```

### Modified Tables

```sql
-- accounts: add creator tracking
ALTER TABLE accounts ADD COLUMN created_by VARCHAR(36);
ALTER TABLE accounts ADD INDEX idx_created_by (created_by);

-- model_mappings: add ownership (NULL = global)
ALTER TABLE model_mappings ADD COLUMN owner_id VARCHAR(36);
ALTER TABLE model_mappings ADD INDEX idx_owner_id (owner_id);

-- request_logs: add API key tracking
ALTER TABLE request_logs ADD COLUMN api_key_id VARCHAR(36);
ALTER TABLE request_logs ADD INDEX idx_api_key_id (api_key_id);
```

### API Key Format

- Pattern: `ak_` + 32 random alphanumeric chars
- Example: `ak_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6`
- Storage: SHA-256 hash (key shown only once at creation)

---

## Authentication Flow

### JWT (Management UI/API)

```
POST /api/v1/auth/login
Request:  { "username": "admin", "password": "admin123" }
Response: { "token": "eyJhbG...", "expires_in": 86400 }

JWT Payload: { "user_id", "username", "role", "exp" }
Validity: 24 hours
```

### API Key (AI Proxy + Management)

```
Header options:
  X-API-Key: ak_a1b2c3d4...
  Authorization: Bearer ak_a1b2c3d4...

Validation:
  1. Hash incoming key (SHA-256)
  2. Lookup in api_keys table
  3. Check is_active = true
  4. Get user_id → get role
  5. Cache in Redis 5 min: apikey:{hash} → {user_id, role}
```

### Middleware Pipeline

```
Request → ExtractAuth → ValidateToken/Key → SetUserContext → Handler
               ↓
         No auth? → 401 Unauthorized
         Invalid? → 401 Unauthorized
         Valid    → c.Set("user", &User{...})
```

---

## Role Permissions

### Admin

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| Users | ✓ | ✓ all | ✓ | ✓ |
| API Keys | ✓ own | ✓ all | - | ✓ all |
| OAuth Accounts | ✓ | ✓ all | ✓ | ✓ |
| Model Mappings | ✓ global + own | ✓ all | ✓ all | ✓ all |
| Stats/Usage | - | ✓ global + per user | - | - |
| Proxies | ✓ | ✓ | ✓ | ✓ |
| AI Proxy | ✓ | - | - | - |

### User

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| Users | ✗ | ✓ self | ✓ self (password) | ✗ |
| API Keys | ✓ own | ✓ own | - | ✓ own |
| OAuth Accounts | ✗ | ✗ | ✗ | ✗ |
| Model Mappings | ✓ own | ✓ global + own | ✓ own | ✓ own |
| Stats/Usage | - | ✓ own only | - | - |
| AI Proxy | ✓ | - | - | - |

### Provider

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| Users | ✗ | ✓ self | ✓ self (password) | ✗ |
| API Keys | ✗ | ✗ | ✗ | ✗ |
| OAuth Accounts | ✓ | ✓ own | ✓ own | ✗ |
| Model Mappings | ✗ | ✗ | ✗ | ✗ |
| Stats/Usage | ✗ | ✗ | ✗ | ✗ |
| AI Proxy | ✗ | - | - | - |

---

## API Endpoints

### Auth (public)

```
POST /api/v1/auth/login          # Login → JWT
POST /api/v1/auth/logout         # Invalidate (client-side)
GET  /api/v1/auth/me             # Get current user info (auth required)
PUT  /api/v1/auth/password       # Change own password (auth required)
```

### Users (admin only)

```
GET    /api/v1/users             # List all users
POST   /api/v1/users             # Create user
GET    /api/v1/users/:id         # Get user detail
PUT    /api/v1/users/:id         # Update user (role, is_active)
DELETE /api/v1/users/:id         # Delete user
```

### API Keys (admin, user)

```
GET    /api/v1/api-keys          # List own keys (admin: ?user_id= filter)
POST   /api/v1/api-keys          # Create key → returns full key ONCE
DELETE /api/v1/api-keys/:id      # Revoke key
```

### Modified Existing

```
# Accounts - add auth + ownership filter
GET  /api/v1/accounts            # admin: all, provider: own only
POST /api/v1/accounts            # admin, provider (sets created_by)

# Model Mappings - add ownership
GET  /api/v1/model-mappings      # Returns global + own
POST /api/v1/model-mappings      # admin: can set global, user: own only

# Stats - add filtering
GET  /api/v1/stats/usage         # admin: all, user: own (by api_key)
GET  /api/v1/stats/usage/global  # admin only
```

### Protected Routes Summary

| Endpoint Pattern | Auth Required | Roles Allowed |
|-----------------|---------------|---------------|
| `GET /health` | No | - |
| `POST /v1/messages` | API Key | admin, user |
| `POST /v1/chat/completions` | API Key | admin, user |
| `POST /api/v1/auth/login` | No | - |
| `/api/v1/auth/*` (other) | JWT/API Key | any |
| `/api/v1/users/*` | JWT/API Key | admin |
| `/api/v1/api-keys/*` | JWT/API Key | admin, user |
| `/api/v1/accounts/*` | JWT/API Key | admin, provider |
| `/api/v1/model-mappings/*` | JWT/API Key | admin, user |
| `/api/v1/stats/*` | JWT/API Key | admin, user (filtered) |
| `/api/v1/proxies/*` | JWT/API Key | admin |

---

## File Structure

### New Files

```
models/
├── user.model.go              # User struct (~50 lines)
├── apikey.model.go            # APIKey struct (~40 lines)
└── role.enum.go               # Role constants + helpers (~30 lines)

repositories/
├── user.repository.go         # User CRUD (~80 lines)
├── user.repository.query.go   # Complex user queries (~60 lines)
├── apikey.repository.go       # APIKey CRUD (~80 lines)
└── apikey.repository.query.go # FindByHash, ListByUser (~60 lines)

services/
├── auth.service.go            # Login logic (~80 lines)
├── auth.jwt.service.go        # JWT create/validate (~70 lines)
├── auth.apikey.service.go     # API key validate/cache (~70 lines)
├── user.service.go            # User business logic (~80 lines)
└── password.service.go        # Hash/verify password (~40 lines)

handlers/
├── auth.handler.go            # Login, Logout (~80 lines)
├── auth.me.handler.go         # Me, ChangePassword (~60 lines)
├── user.handler.go            # List, Get user (~70 lines)
├── user.crud.handler.go       # Create, Update, Delete (~80 lines)
├── apikey.handler.go          # List, Create, Revoke (~80 lines)
└── apikey.response.go         # Response formatting (~30 lines)

middleware/
├── auth.middleware.go         # ExtractAuth from header (~60 lines)
├── auth.require.middleware.go # RequireAuth, RequireRole (~50 lines)
└── auth.context.go            # GetCurrentUser helper (~30 lines)

database/
└── seed.go                    # Auto-seed admin (~50 lines)
```

### Modified Files

```
routes/routes.go                      # Add new routes + middleware
database/mysql.go                     # Add User, APIKey to AutoMigrate
handlers/account.handler.go           # Add ownership filter
handlers/model.mapping.handler.go     # Add ownership logic
handlers/stats.handler.go             # Add user filtering
models/account.model.go               # Add CreatedBy field
models/model.mapping.go               # Add OwnerID field
repositories/account.repository.go    # Add GetByCreator method
services/stats.tracker.service.go     # Add api_key_id tracking
```

### New Dependencies

```bash
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5
```

---

## Auto-Seed Logic

```go
// database/seed.go
func SeedDefaultAdmin(db *gorm.DB) error {
    var count int64
    db.Model(&User{}).Count(&count)

    if count == 0 {
        hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
        admin := &User{
            ID:           uuid.New().String(),
            Username:     "admin",
            PasswordHash: string(hash),
            Role:         RoleAdmin,
            IsActive:     true,
        }
        if err := db.Create(admin).Error; err != nil {
            return err
        }
        log.Warn("Default admin created (admin/admin123) - CHANGE PASSWORD!")
    }
    return nil
}
```

Startup sequence in `cmd/main.go`:
1. Load config
2. Connect DB + Redis
3. AutoMigrate (existing + new tables)
4. **SeedDefaultAdmin()**
5. Start server

---

## Security Notes

- Passwords hashed with bcrypt (cost 10)
- API keys hashed with SHA-256 (fast lookup)
- JWT signed with HS256, secret from config
- API key shown only once at creation
- Log warning if default admin password unchanged
- Redis cache for API key validation (5 min TTL)
