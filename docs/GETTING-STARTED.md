# Getting Started with AIGateway

This guide will help you install, configure, and make your first request to AIGateway.

## Prerequisites

Before you begin, ensure you have:

- **Go 1.21+** installed
- **MySQL 8.0+** running
- **Redis 6.0+** running
- Basic understanding of REST APIs
- API credentials for at least one provider

## Installation

### Step 1: Clone the Repository

```bash
git clone https://github.com/yourorg/aigateway.git
cd aigateway
```

### Step 2: Install Dependencies

```bash
go mod download
```

### Step 3: Verify Installation

```bash
go version  # Should show 1.21 or higher
mysql --version  # Should show 8.0 or higher
redis-cli --version  # Should show 6.0 or higher
```

## Configuration

### Step 1: Create Configuration File

Copy the example configuration:

```bash
cp config/config.example.yaml config/config.yaml
```

### Step 2: Edit Configuration

Edit `config/config.yaml`:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  host: "localhost"
  port: 3306
  user: "root"
  password: "your-mysql-password"
  database: "aigateway"

redis:
  host: "localhost"
  port: 6379
  password: ""  # Leave empty if no password
  db: 0

proxy:
  selection_strategy: "fill_first"
  health_check_interval: 60
  max_failures: 3
```

### Step 3: Set Environment Variables (Optional)

Alternatively, use environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=your-password
export DB_NAME=aigateway

export REDIS_HOST=localhost
export REDIS_PORT=6379
```

## Database Setup

### Option 1: Quick Setup (Recommended)

Use the initialization script:

```bash
./scripts/init.sh
```

This will:
1. Create the database
2. Run auto-migrations
3. Load seed data

### Option 2: Manual Setup

```bash
# Create database
mysql -uroot -p -e "CREATE DATABASE aigateway CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# Run application (migrations run automatically)
go run cmd/main.go

# Load seed data (optional)
mysql -uroot -p aigateway < scripts/seed.sql
```

See [Database Setup](operations/database.md) for detailed instructions.

## Running the Application

### Development Mode

```bash
go run cmd/main.go
```

You should see:

```
[GIN] Listening on 0.0.0.0:8080
Database connected successfully
Redis connected successfully
```

### Production Build

```bash
# Build binary
go build -o aigateway cmd/main.go

# Run binary
./aigateway
```

### Verify Application is Running

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status": "ok"}
```

## Create Your First Provider

### Step 1: Add Provider to Database

```sql
INSERT INTO providers (id, name, base_url, auth_type, auth_strategy, supported_models, is_active)
VALUES (
  'antigravity',
  'Google Antigravity',
  'https://cloudcode-pa.googleapis.com',
  'oauth',
  'oauth',
  '["gemini-claude-sonnet-4-5", "claude-sonnet-4-5"]',
  true
);
```

### Step 2: Verify Provider

```bash
curl http://localhost:8080/api/v1/providers
```

## Create an Account

### Via API

```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "antigravity",
    "label": "My First Account",
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
  }'
```

### Via SQL

```sql
INSERT INTO accounts (id, provider_id, label, auth_data, is_active)
VALUES (
  UUID(),
  'antigravity',
  'My First Account',
  '{"access_token":"ya29.xxx","refresh_token":"1//xxx",...}',
  true
);
```

### Verify Account

```bash
curl http://localhost:8080/api/v1/accounts
```

## Making Your First Request

### Example 1: Claude Format (Anthropic Messages API)

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-5",
    "messages": [
      {"role": "user", "content": "Hello! How are you?"}
    ],
    "max_tokens": 1024
  }'
```

**Expected Response**:
```json
{
  "role": "assistant",
  "content": [
    {
      "type": "text",
      "text": "Hello! I'm doing well, thank you for asking..."
    }
  ],
  "stop_reason": "end_turn",
  "usage": {
    "input_tokens": 12,
    "output_tokens": 25
  }
}
```

### Example 2: OpenAI Format

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [
      {"role": "user", "content": "Hello! How are you?"}
    ],
    "max_tokens": 1024
  }'
```

### Example 3: Streaming Response

```bash
curl -X POST "http://localhost:8080/v1/messages?stream=true" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-sonnet-4-5",
    "messages": [
      {"role": "user", "content": "Write a short story"}
    ],
    "max_tokens": 2048
  }'
```

## Verify Request Logged

Check that the request was logged:

```bash
curl http://localhost:8080/api/v1/stats/logs?limit=10
```

Or via SQL:

```sql
SELECT * FROM request_logs ORDER BY created_at DESC LIMIT 10;
```

## Adding a Proxy (Optional)

### Create Proxy

```bash
curl -X POST http://localhost:8080/api/v1/proxies \
  -H "Content-Type: application/json" \
  -d '{
    "label": "US Proxy 1",
    "proxy_url": "http://user:pass@proxy.example.com:8080",
    "is_active": true
  }'
```

### View Proxy Assignments

```bash
curl http://localhost:8080/api/v1/proxies/assignments
```

### Recalculate Assignments

```bash
curl -X POST http://localhost:8080/api/v1/proxies/recalculate
```

See [Architecture: Proxy Service](architecture/components.md#proxy-service) for details.

## Next Steps

### Learn the Architecture

Understand how AIGateway works:
- [Architecture Overview](architecture/README.md)
- [Request Flow](architecture/README.md#request-flow)
- [Component Details](architecture/components.md)

### Add More Providers

Integrate additional AI providers:
- [Provider Overview](providers/README.md)
- [Adding a New Provider](providers/adding-new-provider.md)

### Set Up Monitoring

Monitor your deployment:
- [Monitoring Guide](operations/monitoring.md)
- [Setting Up Dashboards](operations/monitoring.md#dashboard-setup)
- [Configuring Alerts](operations/monitoring.md#alerting)

### Secure Your Deployment

Implement security best practices:
- [Security Guide](operations/security.md)
- [Access Control](operations/security.md#access-control)
- [Credential Storage](operations/security.md#credential-storage)

## Common Issues

### Issue: "Database connection failed"

**Solution**: Verify MySQL is running and credentials are correct:

```bash
mysql -u root -p -e "SELECT 1;"
```

### Issue: "Redis connection failed"

**Solution**: Verify Redis is running:

```bash
redis-cli PING
```

### Issue: "No available accounts"

**Solution**: Verify at least one active account exists:

```sql
SELECT * FROM accounts WHERE is_active = true;
```

### Issue: "401 Unauthorized from provider"

**Solution**: Check auth_data is correct and token hasn't expired:

```sql
SELECT JSON_EXTRACT(auth_data, '$.expires_at') FROM accounts WHERE provider_id = 'antigravity';
```

See [Troubleshooting Guide](operations/troubleshooting.md) for more issues.

## Development Tips

### Enable Debug Logging

```yaml
# config/config.yaml
logging:
  level: "debug"
  format: "json"
```

### Watch Logs

```bash
tail -f /var/log/aigateway/app.log
```

### Monitor Request Rate

```bash
watch -n 1 'mysql -uroot -p -e "SELECT COUNT(*) FROM aigateway.request_logs WHERE created_at >= NOW() - INTERVAL 1 MINUTE;"'
```

### Test with Different Models

```bash
# Claude Sonnet
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"test"}],"max_tokens":100}'

# GPT-4 (if configured)
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4","messages":[{"role":"user","content":"test"}],"max_tokens":100}'
```

## Additional Resources

- **[API Reference](API.md)** - Complete API documentation
- **[Provider Documentation](providers/README.md)** - Provider-specific guides
- **[Architecture](architecture/README.md)** - System design
- **[Operations](operations/monitoring.md)** - Production deployment

## Getting Help

1. Check the [Documentation Index](INDEX.md)
2. Review [Troubleshooting Guide](operations/troubleshooting.md)
3. Search [GitHub Issues](https://github.com/yourorg/aigateway/issues)
4. Open a new issue with detailed information

**Congratulations!** You've successfully set up AIGateway. Start making requests and explore the documentation to learn more.
