# Permanent Proxy Assignment Design

## Problem

Saat ini proxy di-assign secara dinamis setiap request, menyebabkan:
1. IP tidak konsisten per account - bisa berubah antar request
2. Risiko ban dari provider karena IP anomaly (request AI dari IP A, refresh token dari IP B)
3. Tidak ada kontrol kapan account pakai proxy mana

## Solution

Assign proxy secara permanen saat OAuth registration. Proxy tidak berubah sepanjang hidup account.

## Design

### 1. Permanent Proxy Assignment

**Saat OAuth Registration (`ExchangeCode`):**
```
User complete OAuth → ExchangeCode()
   ↓
Select proxy dengan capacity tersedia (fill-first strategy)
   ↓
Create account dengan ProxyID + ProxyURL
   ↓
Increment proxy.CurrentAccounts
```

**Proxy Selection Criteria:**
- `is_active = true`
- `health_status != 'down'` ATAU `last_marked_down + recovery_delay < now`
- `current_accounts < max_accounts` (kalau max_accounts > 0)
- Order by: `priority DESC`, `current_accounts ASC`

### 2. Request Flow dengan Retry

```
Request masuk
   ↓
Round-robin select account (skip yang proxy down < recovery_delay)
   ↓
Execute dengan account.ProxyURL
   ↓
Gagal (connection/timeout)?
   ↓
Retry 2-3x dengan account + proxy yang sama
   ↓
Masih gagal?
   ↓
Mark proxy sebagai DOWN + timestamp
   ↓
Select account lain (round-robin, skip down proxies)
   ↓
Execute dengan account baru
```

### 3. Proxy Health States

| State | Behavior |
|-------|----------|
| `healthy` | Normal, masuk round-robin |
| `degraded` | Masuk round-robin, tapi priority lebih rendah |
| `down` | Skip dari round-robin selama `recovery_delay` |

**Recovery Flow:**
```
Proxy marked DOWN pada T0
   ↓
Skip selama recovery_delay (e.g., 24h)
   ↓
Setelah T0 + 24h → Eligible untuk round-robin lagi
   ↓
Request masuk, account dengan proxy ini dipilih
   ↓
Execute berhasil? → Mark HEALTHY
Execute gagal? → Mark DOWN lagi, reset timer
```

### 4. Configuration

```yaml
# config/config.yaml
proxy:
  # Retry sebelum switch account
  max_retries: 3

  # Delay sebelum retry
  retry_delay: 1s

  # Berapa lama proxy down di-skip
  down_recovery_delay: 24h

  # Timeout untuk proxy connection
  connect_timeout: 10s
```

### 5. Database Changes

**Tambah kolom di `proxy_pool`:**
```sql
ALTER TABLE proxy_pool
ADD COLUMN marked_down_at TIMESTAMP NULL;
```

**Model update:**
```go
type Proxy struct {
    // ... existing fields
    MarkedDownAt *time.Time `json:"marked_down_at"`
}
```

### 6. Reporting

**Tambah field di `request_logs`:**
```sql
ALTER TABLE request_logs
ADD COLUMN retry_count INT DEFAULT 0,
ADD COLUMN switched_from_account_id VARCHAR(36) NULL;
```

**Events yang di-log:**
| Event | Data |
|-------|------|
| Retry attempt | account_id, proxy_id, retry_count, error |
| Account switch | from_account_id, to_account_id, reason |
| Proxy marked down | proxy_id, marked_down_at, consecutive_failures |
| Proxy recovered | proxy_id, recovery_time |

### 7. Code Changes

| File | Change |
|------|--------|
| `services/oauth.flow.service.go` | Assign proxy di `ExchangeCode()` |
| `services/proxy.service.go` | Remove dynamic `AssignProxy()`, add `SelectProxyForNewAccount()` |
| `services/account.service.go` | Skip accounts dengan proxy down di round-robin |
| `services/router.retry.go` | Retry logic dengan same account, fallback ke account lain |
| `services/stats.tracker.service.go` | Log retry count dan account switch |
| `models/proxy.model.go` | Add `MarkedDownAt` field |
| `models/proxy.model.go` | Add `RetryCount`, `SwitchedFromAccountID` di RequestLog |
| `config/config.go` | Add proxy retry/recovery config |

### 8. Migration Path

1. Add `marked_down_at` column ke `proxy_pool`
2. Add `retry_count`, `switched_from_account_id` ke `request_logs`
3. Run `RecalculateAccountCounts()` untuk sync counter
4. Deploy new code
5. Existing accounts tanpa proxy akan di-assign saat request pertama (backward compat)

### 9. Edge Cases

| Case | Handling |
|------|----------|
| Semua proxy down | Return error 503 "All proxies unavailable" |
| Account tanpa proxy (legacy) | Assign on first request, log warning |
| Proxy dihapus | Account tetap ada, proxy_id jadi NULL, assign baru on request |
| Proxy capacity penuh saat registration | Return error, minta admin tambah proxy |

### 10. Observability

**Metrics untuk monitoring:**
- `proxy_down_events_total` - Counter proxy marked down
- `proxy_recovery_events_total` - Counter proxy recovered
- `request_retry_total` - Counter retry attempts
- `account_switch_total` - Counter account switches
- `proxy_capacity_utilization` - Gauge current/max per proxy

**Alerts:**
- Semua proxy down untuk provider tertentu
- Proxy sering down (> 3x dalam 24 jam)
- Capacity hampir penuh (> 90%)
