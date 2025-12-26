# Database Scripts

## Files

### seed.sql
Seed data for AI Gateway database containing:
- **3 Providers**: antigravity (OAuth), openai (API Key), glm (Bearer)
- **3 Proxy Servers**: US (high-priority), EU (medium), Backup (SOCKS5)
- **8 Accounts**: Distributed across all providers with realistic configurations

### init.sh
Database initialization script that:
1. Creates the database (if not exists)
2. Runs Go auto-migrations
3. Loads seed data from seed.sql

## Usage

### Quick Start
```bash
# Using default MySQL root user
./scripts/init.sh

# With custom credentials
DB_USER=myuser DB_PASSWORD=mypass DB_NAME=aigateway ./scripts/init.sh
```

### Manual Steps
```bash
# 1. Create database
mysql -uroot -p -e "CREATE DATABASE aigateway CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 2. Run migrations (auto-runs on app start)
go run cmd/main.go

# 3. Load seed data
mysql -uroot -p aigateway < scripts/seed.sql
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| DB_HOST | localhost | MySQL host |
| DB_PORT | 3306 | MySQL port |
| DB_USER | root | MySQL user |
| DB_PASSWORD | (empty) | MySQL password |
| DB_NAME | aigateway | Database name |

## Seed Data Details

### Providers
- **antigravity**: OAuth-based, 60 req/min, 90k tokens/min
- **openai**: API key-based, 500 req/min, 150k tokens/min
- **glm**: Bearer token-based, 100 req/min, 100k tokens/min

### Accounts
All accounts use UUID primary keys and include:
- Realistic auth data (tokens, API keys)
- Metadata (owner, environment, quotas)
- Proxy assignments
- Active/inactive status examples

### Proxies
- **US Proxy**: HTTP, priority 10, weight 5, 120ms avg latency
- **EU Proxy**: HTTPS, priority 5, weight 3, 180ms avg latency
- **Backup**: SOCKS5, priority 1, weight 1, 250ms avg latency
