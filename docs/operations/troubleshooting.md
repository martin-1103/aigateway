# Troubleshooting Guide

This document provides solutions to common issues you may encounter when running AIGateway.

## Connection Issues

### Issue: "No available accounts for provider"

**Symptom**: Error message "no available accounts for provider X"

**Cause**: No active accounts configured for the requested provider/model.

**Solution**:

1. Check if accounts exist for the provider:
   ```sql
   SELECT * FROM accounts WHERE provider_id = 'antigravity' AND is_active = true;
   ```

2. If no accounts exist, create one:
   ```bash
   curl -X POST http://localhost:8080/api/v1/accounts \
     -H "Content-Type: application/json" \
     -d '{
       "provider_id": "antigravity",
       "label": "Account 1",
       "auth_data": {...},
       "is_active": true
     }'
   ```

3. If accounts exist but are inactive, activate them:
   ```sql
   UPDATE accounts SET is_active = true WHERE provider_id = 'antigravity';
   ```

### Issue: Proxy connection errors

**Symptom**: "Failed to connect to provider" or timeout errors

**Cause**: Invalid proxy URL or proxy server unreachable.

**Solutions**:

1. Verify proxy URL format:
   ```
   http://username:password@proxy.example.com:8080
   https://proxy.example.com:8080
   socks5://proxy.example.com:1080
   ```

2. Test proxy connectivity:
   ```bash
   curl -x http://proxy.example.com:8080 https://api.provider.com/health
   ```

3. Check proxy status in database:
   ```sql
   SELECT * FROM proxy_pool WHERE is_active = true;
   ```

4. Disable problematic proxy:
   ```sql
   UPDATE proxy_pool SET is_active = false WHERE id = 1;
   ```

5. Recalculate proxy assignments:
   ```bash
   curl -X POST http://localhost:8080/api/v1/proxies/recalculate
   ```

## Authentication Issues

### Issue: OAuth token refresh failures

**Symptom**: "Failed to refresh OAuth token" or "401 Unauthorized" from provider

**Cause**: Invalid refresh token or expired credentials.

**Solutions**:

1. Check token expiry:
   ```sql
   SELECT id, label, JSON_EXTRACT(auth_data, '$.expires_at') as expires_at
   FROM accounts
   WHERE provider_id = 'antigravity';
   ```

2. Verify refresh token is valid:
   ```sql
   SELECT JSON_EXTRACT(auth_data, '$.refresh_token')
   FROM accounts
   WHERE id = 'account-id';
   ```

3. Update account with fresh OAuth credentials:
   ```bash
   curl -X PUT http://localhost:8080/api/v1/accounts/{account-id} \
     -H "Content-Type: application/json" \
     -d '{
       "auth_data": {
         "access_token": "new-token",
         "refresh_token": "new-refresh-token",
         "expires_at": "2024-12-26T10:00:00Z",
         "expires_in": 3600,
         ...
       }
     }'
   ```

4. Check Redis cache:
   ```bash
   redis-cli GET "auth:oauth:antigravity:account-id"
   ```

5. Clear Redis cache if stale:
   ```bash
   redis-cli DEL "auth:oauth:antigravity:account-id"
   ```

### Issue: API key authentication failures

**Symptom**: "401 Unauthorized" with API key providers

**Solutions**:

1. Verify API key format:
   - OpenAI: starts with `sk-proj-` or `sk-`
   - Anthropic: starts with `sk-ant-`
   - Check provider documentation for correct format

2. Check API key in database:
   ```sql
   SELECT JSON_EXTRACT(auth_data, '$.api_key')
   FROM accounts
   WHERE id = 'account-id';
   ```

3. Test API key directly with provider:
   ```bash
   curl https://api.openai.com/v1/models \
     -H "Authorization: Bearer sk-proj-xxx"
   ```

4. Update API key if invalid:
   ```bash
   curl -X PUT http://localhost:8080/api/v1/accounts/{account-id} \
     -H "Content-Type: application/json" \
     -d '{"auth_data": {"api_key": "new-key"}}'
   ```

## Rate Limiting Issues

### Issue: "429 Too Many Requests"

**Symptom**: Provider returns 429 status code

**Cause**: Exceeded provider rate limits.

**Solutions**:

1. Check current account count:
   ```sql
   SELECT provider_id, COUNT(*) as account_count
   FROM accounts
   WHERE is_active = true
   GROUP BY provider_id;
   ```

2. Add more accounts for load distribution:
   ```bash
   curl -X POST http://localhost:8080/api/v1/accounts \
     -H "Content-Type: application/json" \
     -d '{
       "provider_id": "openai",
       "label": "OpenAI Account 2",
       "auth_data": {"api_key": "sk-proj-xxx"},
       "is_active": true
     }'
   ```

3. Verify round-robin is working:
   ```bash
   redis-cli GET "account:rr:openai:gpt-4"
   ```

4. Contact provider for higher limits

5. Implement request throttling at application level

## Performance Issues

### Issue: High latency

**Symptom**: Requests taking longer than expected

**Causes and Solutions**:

1. **Proxy performance issues**:
   ```sql
   SELECT proxy_id, AVG(latency_ms) as avg_latency
   FROM request_logs
   WHERE created_at >= NOW() - INTERVAL 1 HOUR
   GROUP BY proxy_id
   ORDER BY avg_latency DESC;
   ```
   - Disable slow proxies
   - Add faster proxies
   - Redistribute accounts

2. **Provider API slowness**:
   ```sql
   SELECT provider_id, AVG(latency_ms) as avg_latency
   FROM request_logs
   WHERE created_at >= NOW() - INTERVAL 1 HOUR
   GROUP BY provider_id;
   ```
   - Check provider status page
   - Switch to alternative providers
   - Report to provider support

3. **Database bottleneck**:
   ```sql
   SHOW PROCESSLIST;
   ```
   - Optimize slow queries
   - Add missing indexes
   - Increase connection pool size

4. **Redis bottleneck**:
   ```bash
   redis-cli INFO stats
   redis-cli SLOWLOG GET 10
   ```
   - Check for slow commands
   - Increase Redis memory
   - Use Redis cluster

### Issue: Request logs not appearing

**Symptom**: No entries in `request_logs` table

**Cause**: Async logging failure or database issues.

**Solutions**:

1. Check application logs for errors
2. Verify database connection:
   ```bash
   mysql -uroot -p -e "SELECT COUNT(*) FROM aigateway.request_logs;"
   ```
3. Check disk space:
   ```bash
   df -h
   ```
4. Verify table exists:
   ```sql
   SHOW TABLES LIKE 'request_logs';
   ```

## Redis Issues

### Issue: Round-robin not working

**Symptom**: Same account used repeatedly instead of rotating

**Solutions**:

1. Check Redis counter:
   ```bash
   redis-cli GET "account:rr:antigravity:claude-sonnet-4-5"
   ```

2. Verify Redis is running:
   ```bash
   redis-cli PING
   ```

3. Check Redis connection in config:
   ```yaml
   redis:
     host: "localhost"
     port: 6379
     password: ""
     db: 0
   ```

4. Reset counter if needed:
   ```bash
   redis-cli DEL "account:rr:antigravity:claude-sonnet-4-5"
   ```

### Issue: OAuth cache issues

**Symptom**: Token refresh happening too frequently

**Solutions**:

1. Check cache TTL:
   ```bash
   redis-cli TTL "auth:oauth:antigravity:account-id"
   ```

2. Verify token data:
   ```bash
   redis-cli GET "auth:oauth:antigravity:account-id"
   ```

3. Clear cache to force refresh:
   ```bash
   redis-cli DEL "auth:oauth:antigravity:account-id"
   ```

## Database Issues

### Issue: "Too many connections"

**Symptom**: Database connection pool exhausted

**Solutions**:

1. Check current connections:
   ```sql
   SHOW PROCESSLIST;
   SELECT * FROM information_schema.PROCESSLIST;
   ```

2. Increase max connections:
   ```sql
   SET GLOBAL max_connections = 500;
   ```

3. Configure connection pool in application:
   ```go
   db.DB().SetMaxOpenConns(100)
   db.DB().SetMaxIdleConns(10)
   db.DB().SetConnMaxLifetime(time.Hour)
   ```

### Issue: Slow queries

**Symptom**: Database queries taking too long

**Solutions**:

1. Enable slow query log:
   ```sql
   SET GLOBAL slow_query_log = 'ON';
   SET GLOBAL long_query_time = 1;
   ```

2. Analyze slow queries:
   ```bash
   cat /var/log/mysql/slow-query.log
   ```

3. Add missing indexes:
   ```sql
   SHOW INDEX FROM accounts;
   CREATE INDEX idx_last_used ON accounts(last_used_at);
   ```

4. Optimize queries:
   ```sql
   EXPLAIN SELECT * FROM accounts WHERE provider_id = 'antigravity';
   ```

## Application Errors

### Issue: "Provider not found"

**Symptom**: Error "provider not found: X"

**Solutions**:

1. Check provider exists:
   ```sql
   SELECT * FROM providers WHERE id = 'antigravity';
   ```

2. Verify provider is active:
   ```sql
   UPDATE providers SET is_active = true WHERE id = 'antigravity';
   ```

3. Check provider configuration:
   ```sql
   SELECT * FROM providers WHERE id = 'antigravity'\G
   ```

### Issue: Format translation errors

**Symptom**: "Invalid JSON" or malformed responses

**Solutions**:

1. Enable debug logging in translator service
2. Capture request/response payloads
3. Verify provider API hasn't changed format
4. Check translation logic in `services/translator.service.go`
5. Test with provider's official API documentation

### Issue: Panic or crashes

**Symptom**: Application crashes unexpectedly

**Solutions**:

1. Check application logs:
   ```bash
   tail -f /var/log/aigateway/app.log
   ```

2. Check for nil pointer dereferences
3. Verify all required fields are present
4. Add defensive nil checks
5. Enable panic recovery middleware

## Debugging Tips

### Enable Debug Logging

```yaml
# config/config.yaml
logging:
  level: "debug"
  format: "json"
```

### Check Application Health

```bash
curl http://localhost:8080/health
```

### Monitor Real-Time Metrics

```bash
# Watch Redis counters
watch -n 1 'redis-cli --scan --pattern "stats:*"'

# Monitor database connections
watch -n 1 'mysql -uroot -p -e "SHOW PROCESSLIST;" | wc -l'

# Check request logs
watch -n 1 'mysql -uroot -p -e "SELECT COUNT(*) FROM aigateway.request_logs WHERE created_at >= NOW() - INTERVAL 1 MINUTE;"'
```

### Trace Request Flow

Add request ID to logs for tracing:
```go
requestID := uuid.New().String()
log.Printf("[%s] Starting request: %s", requestID, model)
```

## Getting Help

If you can't resolve the issue:

1. Check GitHub issues: https://github.com/yourorg/aigateway/issues
2. Review documentation: `/docs/`
3. Enable debug logging and capture logs
4. Collect error messages and stack traces
5. Open a new issue with:
   - Clear description of problem
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant logs and configurations

## Related Documentation

- [Database Setup](database.md) - Database configuration
- [Monitoring](monitoring.md) - Metrics and observability
- [Architecture](../architecture/README.md) - System architecture
- [Provider Documentation](../providers/README.md) - Provider-specific issues
