# Operations Monitoring

This document provides operational guidance for monitoring AIGateway in production.

For detailed metrics collection and observability architecture, see [Architecture: Monitoring and Observability](../architecture/monitoring.md).

## Quick Health Checks

### Application Health

```bash
# Check if application is running
curl http://localhost:8080/health

# Expected response
{"status": "ok"}
```

### Database Health

```bash
# Check database connection
mysql -uroot -p -e "SELECT 1;"

# Check table row counts
mysql -uroot -p aigateway -e "
SELECT
  'providers' as table_name, COUNT(*) as rows FROM providers
UNION ALL
SELECT 'accounts', COUNT(*) FROM accounts
UNION ALL
SELECT 'proxy_pool', COUNT(*) FROM proxy_pool
UNION ALL
SELECT 'request_logs', COUNT(*) FROM request_logs WHERE created_at >= NOW() - INTERVAL 1 HOUR;"
```

### Redis Health

```bash
# Check Redis connection
redis-cli PING

# Check key counts
redis-cli DBSIZE

# Check memory usage
redis-cli INFO memory | grep used_memory_human
```

## Real-Time Monitoring

### Request Rate

```bash
# Requests per minute (last hour)
mysql -uroot -p aigateway -e "
SELECT
  DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:00') as minute,
  COUNT(*) as requests
FROM request_logs
WHERE created_at >= NOW() - INTERVAL 1 HOUR
GROUP BY minute
ORDER BY minute DESC
LIMIT 10;"
```

### Error Rate

```bash
# Error rate (last hour)
mysql -uroot -p aigateway -e "
SELECT
  COUNT(*) as total_requests,
  SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as errors,
  ROUND(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) * 100.0 / COUNT(*), 2) as error_rate_pct
FROM request_logs
WHERE created_at >= NOW() - INTERVAL 1 HOUR;"
```

### Latency

```bash
# Latency percentiles (last hour)
mysql -uroot -p aigateway -e "
SELECT
  MIN(latency_ms) as min,
  AVG(latency_ms) as avg,
  MAX(latency_ms) as max,
  (SELECT latency_ms FROM request_logs WHERE created_at >= NOW() - INTERVAL 1 HOUR ORDER BY latency_ms LIMIT 1 OFFSET (SELECT FLOOR(COUNT(*) * 0.50) FROM request_logs WHERE created_at >= NOW() - INTERVAL 1 HOUR)) as p50,
  (SELECT latency_ms FROM request_logs WHERE created_at >= NOW() - INTERVAL 1 HOUR ORDER BY latency_ms LIMIT 1 OFFSET (SELECT FLOOR(COUNT(*) * 0.95) FROM request_logs WHERE created_at >= NOW() - INTERVAL 1 HOUR)) as p95
FROM request_logs
WHERE created_at >= NOW() - INTERVAL 1 HOUR;"
```

## Dashboard Setup

### Grafana Dashboard (Recommended)

**Data Source**: MySQL

**Key Panels**:

1. **Request Rate** (Time series):
   ```sql
   SELECT
     UNIX_TIMESTAMP(created_at) as time,
     COUNT(*) / 60 as rps
   FROM request_logs
   WHERE created_at >= FROM_UNIXTIME($__from / 1000)
     AND created_at <= FROM_UNIXTIME($__to / 1000)
   GROUP BY UNIX_TIMESTAMP(created_at) DIV 60
   ```

2. **Error Rate** (Gauge):
   ```sql
   SELECT
     SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as error_rate
   FROM request_logs
   WHERE created_at >= NOW() - INTERVAL 5 MINUTE
   ```

3. **Latency** (Graph):
   ```sql
   SELECT
     created_at as time,
     latency_ms as value
   FROM request_logs
   WHERE created_at >= FROM_UNIXTIME($__from / 1000)
   ORDER BY created_at
   ```

4. **Provider Breakdown** (Pie chart):
   ```sql
   SELECT
     provider_id as metric,
     COUNT(*) as value
   FROM request_logs
   WHERE created_at >= NOW() - INTERVAL 1 HOUR
   GROUP BY provider_id
   ```

### Prometheus Integration (Future)

**Planned metrics endpoint**: `GET /metrics`

Example metrics:
```
# TYPE aigateway_requests_total counter
aigateway_requests_total{provider="antigravity",status="200"} 12345

# TYPE aigateway_request_duration_seconds histogram
aigateway_request_duration_seconds_bucket{provider="antigravity",le="0.5"} 8000
aigateway_request_duration_seconds_sum{provider="antigravity"} 2500
aigateway_request_duration_seconds_count{provider="antigravity"} 10000
```

## Alerting

### Alert Rules

**Critical Alerts**:

1. **High Error Rate** (>5% in 5 minutes):
   ```sql
   SELECT
     SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) * 100.0 / COUNT(*) as error_rate
   FROM request_logs
   WHERE created_at >= NOW() - INTERVAL 5 MINUTE
   HAVING error_rate > 5;
   ```

2. **Application Down**:
   ```bash
   curl -f http://localhost:8080/health || alert
   ```

3. **OAuth Refresh Failures**:
   ```sql
   SELECT COUNT(*) as failures
   FROM request_logs
   WHERE error LIKE '%refresh failed%'
     AND created_at >= NOW() - INTERVAL 5 MINUTE
   HAVING failures > 0;
   ```

**Warning Alerts**:

1. **High Latency** (p95 > 1000ms):
   ```sql
   SELECT latency_ms FROM request_logs
   WHERE created_at >= NOW() - INTERVAL 5 MINUTE
   ORDER BY latency_ms
   LIMIT 1 OFFSET (SELECT FLOOR(COUNT(*) * 0.95) FROM request_logs WHERE created_at >= NOW() - INTERVAL 5 MINUTE)
   HAVING latency_ms > 1000;
   ```

2. **Low Cache Hit Rate** (<90%):
   ```bash
   # Calculate from Redis stats
   hits=$(redis-cli GET oauth:cache:hits)
   misses=$(redis-cli GET oauth:cache:misses)
   rate=$(echo "scale=2; $hits * 100 / ($hits + $misses)" | bc)
   if (( $(echo "$rate < 90" | bc -l) )); then
     echo "Cache hit rate low: $rate%"
   fi
   ```

### Alert Channels

**Recommended Integrations**:
- Email (for critical alerts)
- Slack/Discord (for all alerts)
- PagerDuty (for on-call rotation)
- SMS (for critical alerts)

## Log Management

### Application Logs

**Log Format**:
```json
{
  "timestamp": "2024-12-26T10:00:00Z",
  "level": "INFO",
  "message": "Request completed",
  "request_id": "req-123",
  "provider": "antigravity",
  "model": "claude-sonnet-4-5",
  "latency_ms": 234,
  "status": 200
}
```

**Log Rotation**:
```bash
# /etc/logrotate.d/aigateway
/var/log/aigateway/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0640 aigateway aigateway
    sharedscripts
    postrotate
        systemctl reload aigateway
    endscript
}
```

### Centralized Logging

**ELK Stack** (Elasticsearch, Logstash, Kibana):

```conf
# Logstash input
input {
  file {
    path => "/var/log/aigateway/*.log"
    start_position => "beginning"
    codec => json
  }
}

output {
  elasticsearch {
    hosts => ["localhost:9200"]
    index => "aigateway-%{+YYYY.MM.dd}"
  }
}
```

**Loki** (with Grafana):
```yaml
# promtail config
clients:
  - url: http://localhost:3100/loki/api/v1/push

scrape_configs:
  - job_name: aigateway
    static_configs:
      - targets:
          - localhost
        labels:
          job: aigateway
          __path__: /var/log/aigateway/*.log
```

## Capacity Planning

### Resource Monitoring

**CPU Usage**:
```bash
top -b -n 1 | grep aigateway
```

**Memory Usage**:
```bash
ps aux | grep aigateway | awk '{print $4 "% " $6/1024 "MB"}'
```

**Disk Space**:
```bash
df -h /var/lib/mysql
du -sh /var/log/aigateway
```

### Scaling Indicators

**Scale Up When**:
- CPU usage > 70% sustained
- Memory usage > 80%
- Request latency p95 > 1000ms
- Error rate > 5%

**Scale Out When**:
- Single instance can't handle load
- Geographic distribution needed
- High availability required

## Backup Monitoring

### Database Backups

```bash
# Check last backup
ls -lh /backups/mysql/*.sql | tail -1

# Verify backup integrity
mysql -uroot -p aigateway_backup < /backups/mysql/latest.sql 2>&1 | grep -i error
```

### Redis Persistence

```bash
# Check RDB save time
redis-cli LASTSAVE

# Check AOF status
redis-cli INFO persistence | grep aof
```

## Performance Benchmarking

### Load Testing

```bash
# Using Apache Bench
ab -n 1000 -c 10 -p request.json -T application/json http://localhost:8080/v1/messages

# Using wrk
wrk -t4 -c100 -d30s --latency http://localhost:8080/health
```

### Baseline Metrics

Track baseline performance for comparison:
- Throughput: X req/s
- Latency p50: Y ms
- Latency p95: Z ms
- Error rate: <1%

## On-Call Runbook

### First Response

1. **Check application health**:
   ```bash
   curl http://localhost:8080/health
   ```

2. **Check recent errors**:
   ```sql
   SELECT * FROM request_logs
   WHERE status_code >= 400
   ORDER BY created_at DESC
   LIMIT 10;
   ```

3. **Check resource usage**:
   ```bash
   top
   df -h
   ```

### Common Incidents

See [Troubleshooting Guide](troubleshooting.md) for detailed incident resolution.

## Related Documentation

- [Architecture: Monitoring](../architecture/monitoring.md) - Detailed metrics architecture
- [Troubleshooting](troubleshooting.md) - Common issues and solutions
- [Security](security.md) - Security monitoring
- [Database Setup](database.md) - Database operations
