# Database Setup

This document describes how to initialize and seed the AIGateway database.

## Files

### seed.sql

Located at: `scripts/seed.sql`

Seed data for AI Gateway database containing:
- **3 Providers**: antigravity (OAuth), openai (API Key), glm (Bearer)
- **3 Proxy Servers**: US (high-priority), EU (medium), Backup (SOCKS5)
- **8 Accounts**: Distributed across all providers with realistic configurations

### init.sh

Located at: `scripts/init.sh`

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

If you prefer to run steps manually:

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

## Database Migrations

AIGateway uses GORM's AutoMigrate feature to manage database schema.

### Auto-Migration

Migrations run automatically when the application starts:

```go
// In cmd/main.go
db.AutoMigrate(
    &models.Provider{},
    &models.Account{},
    &models.ProxyPool{},
    &models.ProxyStats{},
    &models.RequestLog{},
)
```

### Migration Process

1. Application starts
2. Connects to database
3. Checks current schema
4. Creates missing tables
5. Adds missing columns
6. Updates indexes

**Note**: AutoMigrate will NOT:
- Delete unused columns
- Delete unused tables
- Modify column types (must be done manually)

### Manual Migration

For complex schema changes, create manual migration scripts:

```sql
-- migrations/001_add_column.sql
ALTER TABLE accounts ADD COLUMN new_field VARCHAR(255);
```

Run manually:
```bash
mysql -uroot -p aigateway < migrations/001_add_column.sql
```

## Backup and Restore

### Backup Database

**Full backup**:
```bash
mysqldump -uroot -p aigateway > backup_$(date +%Y%m%d_%H%M%S).sql
```

**Schema only**:
```bash
mysqldump -uroot -p --no-data aigateway > schema_backup.sql
```

**Data only**:
```bash
mysqldump -uroot -p --no-create-info aigateway > data_backup.sql
```

### Restore Database

```bash
# Restore full backup
mysql -uroot -p aigateway < backup_20241226_100000.sql

# Restore from scratch
mysql -uroot -p -e "DROP DATABASE IF EXISTS aigateway;"
mysql -uroot -p -e "CREATE DATABASE aigateway CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
mysql -uroot -p aigateway < backup_20241226_100000.sql
```

## Database Maintenance

### Optimize Tables

```sql
-- Optimize all tables
OPTIMIZE TABLE accounts, providers, proxy_pool, request_logs, proxy_stats;

-- Analyze tables for query optimization
ANALYZE TABLE accounts, providers, proxy_pool, request_logs, proxy_stats;
```

### Clean Old Logs

```sql
-- Delete logs older than 30 days
DELETE FROM request_logs WHERE created_at < NOW() - INTERVAL 30 DAY;

-- Archive before deleting
INSERT INTO request_logs_archive SELECT * FROM request_logs WHERE created_at < NOW() - INTERVAL 30 DAY;
DELETE FROM request_logs WHERE created_at < NOW() - INTERVAL 30 DAY;
```

### Rebuild Indexes

```sql
-- Drop and recreate indexes if needed
ALTER TABLE accounts DROP INDEX idx_provider_active;
ALTER TABLE accounts ADD INDEX idx_provider_active (provider_id, is_active);
```

## Troubleshooting

### Database Connection Errors

**Symptom**: "Error connecting to database"

**Solutions**:
1. Verify MySQL is running: `systemctl status mysql`
2. Check credentials in `config/config.yaml`
3. Test connection: `mysql -uroot -p -e "SELECT 1;"`

### Migration Failures

**Symptom**: "Auto-migration failed"

**Solutions**:
1. Check MySQL error log: `/var/log/mysql/error.log`
2. Verify user has sufficient permissions:
   ```sql
   GRANT ALL PRIVILEGES ON aigateway.* TO 'user'@'localhost';
   ```
3. Run migrations manually

### Seed Data Errors

**Symptom**: "Duplicate entry" when loading seed data

**Solution**: Database already has data, either:
1. Drop and recreate database:
   ```bash
   mysql -uroot -p -e "DROP DATABASE aigateway;"
   ./scripts/init.sh
   ```
2. Skip seed data loading

### Character Encoding Issues

**Symptom**: Garbled characters in database

**Solution**: Ensure UTF-8 encoding:
```sql
ALTER DATABASE aigateway CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
ALTER TABLE accounts CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

## Related Documentation

- [Database Schema](../architecture/database.md) - Detailed schema documentation
- [Architecture Overview](../architecture/README.md) - System architecture
- [Monitoring](monitoring.md) - Database metrics and monitoring
