# Dashboard Lite dengan Access Key

## Overview

Fitur yang memungkinkan semua user mengakses dashboard "lite" tanpa login, menggunakan personal access key. Fokus utama: memudahkan provider menambah OAuth accounts tanpa harus login berulang kali.

## Problem Statement

- Provider perlu menambah banyak OAuth account (misal 10 Google accounts)
- Tiap Google account = beda browser/incognito session
- Login dashboard AIGateway dulu → baru OAuth ke Google = ribet
- Butuh cara akses dashboard tanpa login

## Solution

Personal access key (`ak_xxx`) yang bisa dipakai untuk akses dashboard lite tanpa login.

## Fitur per Role (Dashboard Lite)

| Fitur | Admin | User | Provider |
|-------|:-----:|:----:|:--------:|
| Lihat accounts sendiri | ✓ | ✓ | ✓ |
| Add OAuth account | ✓ | ✓ | ✓ |
| Lihat API Keys (read-only) | ✓ | ✓ | ✗ |

Fitur yang TIDAK tersedia di lite (butuh login penuh):
- Delete/edit account
- Create API key
- Manage users
- View stats
- Manage proxies

## Technical Design

### 1. Access Key System

**Database:**
```sql
ALTER TABLE users ADD COLUMN access_key VARCHAR(64) UNIQUE;
CREATE INDEX idx_users_access_key ON users(access_key);
```

**Key Format:** `ak_` + 64 char hex (256-bit entropy)

**Generation:**
- Auto-generate saat admin create user baru
- Auto-generate saat user login jika belum ada
- Regenerate via dashboard (key lama invalid)

### 2. URL Structure

```
Dashboard Lite: https://gateway.com/lite?key=ak_xxx
OAuth Callback: https://gateway.com/lite/oauth/callback?key=ak_xxx
```

Key diteruskan via `X-Access-Key` header untuk API calls.

### 3. Backend Routes

```go
lite := api.Group("/lite")
lite.Use(middleware.ValidateAccessKey())
lite.Use(middleware.LiteRateLimit())
{
    lite.GET("/accounts", liteHandler.ListAccounts)
    lite.GET("/api-keys", liteHandler.ListAPIKeys)
    lite.POST("/oauth/init", oauthHandler.InitFlow)
    lite.GET("/oauth/callback", oauthHandler.LiteCallback)
}
```

### 4. Frontend Structure

```
features/lite/
├── lite-layout.tsx
├── lite-accounts-page.tsx
├── lite-api-keys-page.tsx
├── lite-oauth-callback.tsx
├── components/
│   ├── lite-header.tsx
│   ├── lite-nav.tsx
│   └── lite-add-oauth.tsx
└── api/
    ├── lite-accounts.api.ts
    └── lite-api-keys.api.ts
```

## Security

| Concern | Mitigation |
|---------|------------|
| Key di URL terlihat di logs | `X-Access-Key` header untuk API, query param hanya initial load |
| Key bocor via referrer | `Referrer-Policy: no-referrer` |
| Brute force | Rate limit 10 req/menit per IP |
| Key tercuri | Regenerate dari dashboard |

## User Flow

```
1. User login ke dashboard biasa
2. Ambil access key dari Settings
3. Bookmark: https://gateway.com/lite?key=ak_xxx
4. Buka link → langsung lihat accounts tanpa login
5. Klik "Add OAuth" → redirect ke Google
6. Callback → account tersimpan → kembali ke list
```

## File Changes

**Backend:**
- `models/user.model.go` - +AccessKey field
- `handlers/auth.handler.go` - +GetMyKey, RegenerateKey
- `handlers/lite.handler.go` - NEW
- `middleware/lite.middleware.go` - NEW
- `routes/routes.go` - +lite routes
- `repositories/user.repo.go` - +GetByAccessKey

**Frontend:**
- `features/lite/` - NEW folder
- `routes/routes.tsx` - +lite routes
