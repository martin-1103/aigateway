# OAuth Flow - Quick Reference

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| GET | `/api/v1/oauth/providers` | Public | List OAuth providers |
| GET | `/api/v1/oauth/callback` | Public | OAuth callback (auto flow) |
| POST | `/api/v1/oauth/init` | User/Admin | Start OAuth flow |
| POST | `/api/v1/oauth/exchange` | User/Admin | Exchange code (manual) |
| POST | `/api/v1/oauth/refresh` | User/Admin | Refresh token |

## Provider IDs

- `antigravity` - Google Cloud Code (Gemini + Claude)
- `codex` - OpenAI (GPT models)
- `claude` - Anthropic (Claude direct)

## Init Flow Request

```json
{
  "provider": "antigravity",
  "account_name": "My Google Account",
  "flow_type": "auto"
}
```

**Flow Types:**
- `auto` - Popup closes automatically after callback
- `manual` - User copies callback URL

## Exchange Request

```json
{
  "callback_url": "http://localhost:8088/api/v1/oauth/callback?code=AUTH_CODE&state=STATE"
}
```

## Refresh Request

```json
{
  "account_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Frontend Example

```typescript
// 1. Init flow
const { auth_url } = await fetch('/api/v1/oauth/init', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    provider: 'antigravity',
    account_name: 'My Account',
    flow_type: 'auto'
  })
}).then(r => r.json());

// 2. Open popup
const popup = window.open(auth_url, 'oauth', 'width=600,height=700');

// 3. Listen for success
window.addEventListener('message', (e) => {
  if (e.data.type === 'oauth_success') {
    console.log('Connected:', e.data.account);
  }
});
```

## Redis Keys

| Key | TTL | Purpose |
|-----|-----|---------|
| `oauth:session:{state}` | 10min | OAuth flow session |
| `auth:{provider}:{account_id}` | Dynamic | Token cache |

## File Structure

```
auth/
  pkce/
    pkce.go                   # PKCE code generation
  oauth/
    providers.go              # OAuth configs

services/
  oauth.flow.service.go       # Flow orchestration
  oauth.service.go            # Token caching (existing)

handlers/
  oauth.handler.go            # HTTP endpoints
```

## cURL Examples

**List providers:**
```bash
curl http://localhost:8088/api/v1/oauth/providers
```

**Init flow:**
```bash
curl -X POST http://localhost:8088/api/v1/oauth/init \
  -H "Content-Type: application/json" \
  -d '{"provider":"antigravity","account_name":"Test","flow_type":"manual"}'
```

**Exchange (manual):**
```bash
curl -X POST http://localhost:8088/api/v1/oauth/exchange \
  -H "Content-Type: application/json" \
  -d '{"callback_url":"http://localhost:8088/api/v1/oauth/callback?code=CODE&state=STATE"}'
```

**Refresh:**
```bash
curl -X POST http://localhost:8088/api/v1/oauth/refresh \
  -H "Content-Type: application/json" \
  -d '{"account_id":"ACCOUNT_ID"}'
```
