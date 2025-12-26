# API Reference

Complete API endpoint documentation for AIGateway.

## Base URL

```
http://localhost:8080
```

## Proxy Endpoints

These endpoints forward requests to AI providers with automatic routing.

### POST /v1/messages

**Anthropic Claude format** (Messages API)

**Description**: Send a message request in Anthropic format. Automatically routed based on model name.

**Request**:

```http
POST /v1/messages
Content-Type: application/json

{
  "model": "claude-sonnet-4-5",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "system": "You are a helpful assistant",
  "max_tokens": 1024,
  "temperature": 0.7
}
```

**Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `model` | string | Yes | Model name (e.g., "claude-sonnet-4-5") |
| `messages` | array | Yes | Array of message objects |
| `messages[].role` | string | Yes | "user" or "assistant" |
| `messages[].content` | string/array | Yes | Message content |
| `system` | string | No | System instruction |
| `max_tokens` | integer | No | Maximum tokens to generate |
| `temperature` | float | No | Sampling temperature (0.0-1.0) |

**Response**:

```json
{
  "role": "assistant",
  "content": [
    {"type": "text", "text": "Hello! How can I help you?"}
  ],
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 12,
    "output_tokens": 25
  }
}
```

**Status Codes**:
- `200 OK` - Success
- `400 Bad Request` - Invalid request format
- `401 Unauthorized` - Authentication failed
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

---

### POST /v1/chat/completions

**OpenAI format** (Chat Completions API)

**Description**: Send a chat completion request in OpenAI format.

**Request**:

```http
POST /v1/chat/completions
Content-Type: application/json

{
  "model": "gpt-4",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ],
  "max_tokens": 1024,
  "temperature": 0.7
}
```

**Parameters**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `model` | string | Yes | Model name (e.g., "gpt-4") |
| `messages` | array | Yes | Array of message objects |
| `messages[].role` | string | Yes | "system", "user", or "assistant" |
| `messages[].content` | string | Yes | Message content |
| `max_tokens` | integer | No | Maximum tokens to generate |
| `temperature` | float | No | Sampling temperature (0.0-2.0) |

**Response**:

```json
{
  "id": "chatcmpl-xxx",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "gpt-4",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! How can I help you?"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 12,
    "completion_tokens": 25,
    "total_tokens": 37
  }
}
```

**Status Codes**:
- `200 OK` - Success
- `400 Bad Request` - Invalid request format
- `401 Unauthorized` - Authentication failed
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

---

## Management Endpoints

### Accounts API

#### GET /api/v1/accounts

**Description**: List all accounts.

**Request**:

```http
GET /api/v1/accounts
```

**Response**:

```json
[
  {
    "id": "acc-123",
    "provider_id": "antigravity",
    "label": "Account 1",
    "is_active": true,
    "proxy_url": "http://proxy.example.com:8080",
    "proxy_id": 1,
    "last_used_at": "2024-12-26T10:00:00Z",
    "usage_count": 1234,
    "created_at": "2024-12-01T00:00:00Z",
    "updated_at": "2024-12-26T10:00:00Z"
  }
]
```

---

#### GET /api/v1/accounts/:id

**Description**: Get account details.

**Request**:

```http
GET /api/v1/accounts/acc-123
```

**Response**:

```json
{
  "id": "acc-123",
  "provider_id": "antigravity",
  "label": "Account 1",
  "auth_data": {
    "access_token": "ya29.xxx",
    "expires_at": "2024-12-26T10:00:00Z"
  },
  "metadata": {},
  "is_active": true,
  "proxy_url": "http://proxy.example.com:8080",
  "proxy_id": 1,
  "last_used_at": "2024-12-26T10:00:00Z",
  "usage_count": 1234,
  "created_at": "2024-12-01T00:00:00Z",
  "updated_at": "2024-12-26T10:00:00Z"
}
```

---

#### POST /api/v1/accounts

**Description**: Create a new account.

**Request**:

```http
POST /api/v1/accounts
Content-Type: application/json

{
  "provider_id": "antigravity",
  "label": "My Account",
  "auth_data": {
    "access_token": "ya29.xxx",
    "refresh_token": "1//xxx",
    "token_url": "https://oauth2.googleapis.com/token",
    "client_id": "xxx.apps.googleusercontent.com",
    "client_secret": "xxx",
    "expires_at": "2024-12-26T10:00:00Z",
    "expires_in": 3600,
    "token_type": "Bearer"
  },
  "is_active": true
}
```

**Response**:

```json
{
  "id": "acc-new",
  "provider_id": "antigravity",
  "label": "My Account",
  "is_active": true,
  "created_at": "2024-12-26T10:00:00Z"
}
```

**Status Codes**:
- `201 Created` - Account created
- `400 Bad Request` - Invalid request
- `500 Internal Server Error` - Server error

---

#### PUT /api/v1/accounts/:id

**Description**: Update account.

**Request**:

```http
PUT /api/v1/accounts/acc-123
Content-Type: application/json

{
  "label": "Updated Label",
  "is_active": false
}
```

**Response**:

```json
{
  "id": "acc-123",
  "provider_id": "antigravity",
  "label": "Updated Label",
  "is_active": false,
  "updated_at": "2024-12-26T10:00:00Z"
}
```

---

#### DELETE /api/v1/accounts/:id

**Description**: Delete account.

**Request**:

```http
DELETE /api/v1/accounts/acc-123
```

**Response**:

```json
{
  "message": "Account deleted successfully"
}
```

**Status Codes**:
- `200 OK` - Account deleted
- `404 Not Found` - Account not found
- `500 Internal Server Error` - Server error

---

### Proxies API

#### GET /api/v1/proxies

**Description**: List all proxies.

**Request**:

```http
GET /api/v1/proxies
```

**Response**:

```json
[
  {
    "id": 1,
    "label": "US Proxy",
    "proxy_url": "http://proxy.example.com:8080",
    "is_active": true,
    "max_failures": 3,
    "failure_count": 0,
    "created_at": "2024-12-01T00:00:00Z",
    "updated_at": "2024-12-26T10:00:00Z"
  }
]
```

---

#### GET /api/v1/proxies/:id

**Description**: Get proxy details.

**Request**:

```http
GET /api/v1/proxies/1
```

**Response**:

```json
{
  "id": 1,
  "label": "US Proxy",
  "proxy_url": "http://proxy.example.com:8080",
  "is_active": true,
  "max_failures": 3,
  "failure_count": 0,
  "created_at": "2024-12-01T00:00:00Z",
  "updated_at": "2024-12-26T10:00:00Z"
}
```

---

#### POST /api/v1/proxies

**Description**: Create a new proxy.

**Request**:

```http
POST /api/v1/proxies
Content-Type: application/json

{
  "label": "EU Proxy",
  "proxy_url": "http://eu-proxy.example.com:8080",
  "is_active": true
}
```

**Response**:

```json
{
  "id": 2,
  "label": "EU Proxy",
  "proxy_url": "http://eu-proxy.example.com:8080",
  "is_active": true,
  "created_at": "2024-12-26T10:00:00Z"
}
```

---

#### PUT /api/v1/proxies/:id

**Description**: Update proxy.

**Request**:

```http
PUT /api/v1/proxies/1
Content-Type: application/json

{
  "is_active": false
}
```

**Response**:

```json
{
  "id": 1,
  "label": "US Proxy",
  "is_active": false,
  "updated_at": "2024-12-26T10:00:00Z"
}
```

---

#### DELETE /api/v1/proxies/:id

**Description**: Delete proxy.

**Request**:

```http
DELETE /api/v1/proxies/1
```

**Response**:

```json
{
  "message": "Proxy deleted successfully"
}
```

---

#### GET /api/v1/proxies/assignments

**Description**: View proxy assignments.

**Request**:

```http
GET /api/v1/proxies/assignments
```

**Response**:

```json
{
  "proxies": [
    {
      "proxy_id": 1,
      "label": "US Proxy",
      "assigned_accounts": 5,
      "accounts": [
        {"id": "acc-1", "label": "Account 1"},
        {"id": "acc-2", "label": "Account 2"}
      ]
    }
  ]
}
```

---

#### POST /api/v1/proxies/recalculate

**Description**: Recalculate account-to-proxy assignments.

**Request**:

```http
POST /api/v1/proxies/recalculate
```

**Response**:

```json
{
  "message": "Proxy assignments recalculated",
  "assignments": {
    "proxy_1": 5,
    "proxy_2": 3
  }
}
```

---

### Statistics API

#### GET /api/v1/stats/proxies/:id

**Description**: Get proxy usage statistics.

**Request**:

```http
GET /api/v1/stats/proxies/1?days=7
```

**Query Parameters**:
- `days` (optional): Number of days to query (default: 7)

**Response**:

```json
{
  "proxy_id": 1,
  "stats": [
    {
      "date": "2024-12-26",
      "total_requests": 1000,
      "success_requests": 950,
      "failed_requests": 50,
      "avg_latency": 234,
      "total_tokens": 50000
    }
  ]
}
```

---

#### GET /api/v1/stats/logs

**Description**: Get recent request logs.

**Request**:

```http
GET /api/v1/stats/logs?limit=100
```

**Query Parameters**:
- `limit` (optional): Number of logs to return (default: 100, max: 1000)

**Response**:

```json
[
  {
    "id": 12345,
    "account_id": "acc-123",
    "provider_id": "antigravity",
    "proxy_id": 1,
    "model": "claude-sonnet-4-5",
    "status_code": 200,
    "latency_ms": 234,
    "error": null,
    "created_at": "2024-12-26T10:00:00Z"
  }
]
```

---

## Error Responses

All endpoints may return error responses in this format:

```json
{
  "error": "Error message description"
}
```

**Common Error Codes**:

| Code | Meaning |
|------|---------|
| `400` | Bad Request - Invalid input |
| `401` | Unauthorized - Authentication failed |
| `404` | Not Found - Resource not found |
| `429` | Too Many Requests - Rate limit exceeded |
| `500` | Internal Server Error - Server error |

## Related Documentation

- [Getting Started](GETTING-STARTED.md) - Installation and setup
- [Provider Documentation](providers/README.md) - Provider-specific details
- [Architecture](architecture/README.md) - System design
- [Troubleshooting](operations/troubleshooting.md) - Common issues
