# Security Considerations

This document outlines security best practices and recommendations for deploying and operating AIGateway.

## Credential Storage

### At Rest

**Current State**:
- Credentials stored in JSON format in MySQL `auth_data` column
- No encryption by default

**Recommendations**:

1. **Encrypt auth_data column**:
   ```sql
   -- Enable MySQL encryption at rest
   ALTER TABLE accounts ENCRYPTION='Y';
   ```

2. **Use MySQL Transparent Data Encryption (TDE)**:
   ```ini
   # my.cnf
   [mysqld]
   early-plugin-load=keyring_file.so
   keyring_file_data=/var/lib/mysql-keyring/keyring
   ```

3. **Application-level encryption**:
   ```go
   // Encrypt before storing
   encryptedData, _ := encrypt(authData, encryptionKey)
   account.AuthData = encryptedData

   // Decrypt when retrieving
   decryptedData, _ := decrypt(account.AuthData, encryptionKey)
   ```

### In Transit

**Required**:
- TLS for all provider API calls
- TLS for database connections
- TLS for Redis connections (if not localhost)

**Configuration**:

```yaml
# config/config.yaml
database:
  host: "localhost"
  port: 3306
  tls_mode: "required"
  tls_ca: "/path/to/ca.pem"

redis:
  host: "localhost"
  port: 6379
  tls_enabled: true

providers:
  verify_ssl: true
```

## Access Control

### API Authentication

**Current State**: No authentication on proxy endpoints

**Recommendations**:

1. **Add API Key Middleware**:
   ```go
   func APIKeyMiddleware() gin.HandlerFunc {
       return func(c *gin.Context) {
           apiKey := c.GetHeader("X-API-Key")
           if !validateAPIKey(apiKey) {
               c.JSON(401, gin.H{"error": "invalid api key"})
               c.Abort()
               return
           }
           c.Next()
       }
   }

   // Apply to routes
   router.Use(APIKeyMiddleware())
   ```

2. **Implement JWT Authentication**:
   ```go
   func JWTMiddleware() gin.HandlerFunc {
       return func(c *gin.Context) {
           token := c.GetHeader("Authorization")
           claims, err := validateJWT(token)
           if err != nil {
               c.JSON(401, gin.H{"error": "invalid token"})
               c.Abort()
               return
           }
           c.Set("user", claims.UserID)
           c.Next()
       }
   }
   ```

3. **Use OAuth for Management API**:
   - Implement OAuth 2.0 for `/api/v1/*` endpoints
   - Require authentication for account/proxy management
   - Use scopes for permission management

### Database Access

**Recommendations**:

1. **Restrict to localhost or private network**:
   ```sql
   -- Create dedicated user for application
   CREATE USER 'aigateway'@'localhost' IDENTIFIED BY 'strong_password';

   -- Grant only required permissions
   GRANT SELECT, INSERT, UPDATE ON aigateway.* TO 'aigateway'@'localhost';

   -- Revoke unnecessary privileges
   REVOKE DELETE ON aigateway.* FROM 'aigateway'@'localhost';
   ```

2. **Use strong passwords**:
   ```bash
   # Generate strong password
   openssl rand -base64 32
   ```

3. **Enable audit logging**:
   ```sql
   -- Enable audit log
   INSTALL PLUGIN audit_log SONAME 'audit_log.so';
   SET GLOBAL audit_log_policy = 'ALL';
   ```

### Redis Security

**Recommendations**:

1. **Enable AUTH password**:
   ```conf
   # redis.conf
   requirepass your_strong_password_here
   ```

2. **Bind to localhost only**:
   ```conf
   # redis.conf
   bind 127.0.0.1 ::1
   ```

3. **Disable dangerous commands**:
   ```conf
   # redis.conf
   rename-command FLUSHDB ""
   rename-command FLUSHALL ""
   rename-command CONFIG ""
   ```

## Token Security

### OAuth Tokens

**Best Practices**:

1. **Cached in Redis with TTL**:
   - TTL equals token expiry
   - Automatic eviction prevents stale tokens

2. **Encryption in Redis**:
   ```go
   // Encrypt before caching
   encryptedToken := encrypt(tokenData)
   redis.Set(ctx, key, encryptedToken, ttl)
   ```

3. **Secure transmission**:
   - Always use HTTPS for OAuth endpoints
   - Verify SSL certificates
   - Use TLS 1.2 or higher

### API Keys

**Best Practices**:

1. **Rotation schedule**:
   - Rotate API keys every 90 days
   - Automate rotation where possible
   - Maintain key version history

2. **Least privilege**:
   - Use minimum required API key scopes
   - Create separate keys for different purposes
   - Revoke unused keys immediately

3. **Monitoring**:
   - Log all API key usage
   - Alert on unusual activity
   - Track key age and expiry

## Network Security

### Firewall Rules

**Recommended Configuration**:

```bash
# Allow only necessary ports
ufw allow 8080/tcp  # Application port
ufw allow 3306/tcp from 127.0.0.1  # MySQL (localhost only)
ufw allow 6379/tcp from 127.0.0.1  # Redis (localhost only)
ufw deny 3306/tcp  # Block external MySQL
ufw deny 6379/tcp  # Block external Redis
ufw enable
```

### TLS/SSL Configuration

**Minimum TLS Version**:
```go
// Configure HTTP client
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
            CipherSuites: []uint16{
                tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
                tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
            },
        },
    },
}
```

### Proxy Security

**Recommendations**:

1. **Use authenticated proxies**:
   ```
   http://username:password@proxy.example.com:8080
   ```

2. **Verify proxy SSL certificates**:
   ```go
   transport := &http.Transport{
       Proxy: http.ProxyURL(proxyURL),
       TLSClientConfig: &tls.Config{
           InsecureSkipVerify: false,  // Verify certificates
       },
    }
   ```

3. **Rotate proxy credentials regularly**

## Application Security

### Input Validation

**Required Validation**:

1. **Model names**:
   ```go
   func validateModel(model string) error {
       if model == "" {
           return errors.New("model is required")
       }
       if len(model) > 100 {
           return errors.New("model name too long")
       }
       if !regexp.MustCompile(`^[a-z0-9-_]+$`).MatchString(model) {
           return errors.New("invalid model format")
       }
       return nil
   }
   ```

2. **Request payloads**:
   ```go
   func validateRequest(req Request) error {
       if len(req.Messages) == 0 {
           return errors.New("messages cannot be empty")
       }
       if req.MaxTokens < 1 || req.MaxTokens > 100000 {
           return errors.New("max_tokens out of range")
       }
       return nil
   }
   ```

3. **SQL injection prevention**:
   ```go
   // Use parameterized queries (GORM does this automatically)
   db.Where("provider_id = ?", providerID).Find(&accounts)

   // NEVER do this:
   // db.Where(fmt.Sprintf("provider_id = '%s'", providerID))
   ```

### Rate Limiting

**Implement Application-Level Rate Limiting**:

```go
import "golang.org/x/time/rate"

var limiters = make(map[string]*rate.Limiter)

func rateLimitMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")

        limiter, exists := limiters[apiKey]
        if !exists {
            limiter = rate.NewLimiter(rate.Limit(100), 200) // 100 req/sec, burst 200
            limiters[apiKey] = limiter
        }

        if !limiter.Allow() {
            c.JSON(429, gin.H{"error": "rate limit exceeded"})
            c.Abort()
            return
        }

        c.Next()
    }
}
```

### Error Handling

**Prevent Information Disclosure**:

```go
// DON'T expose internal errors
c.JSON(500, gin.H{"error": err.Error()})  // May expose sensitive info

// DO return generic errors
log.Printf("Internal error: %v", err)  // Log full error
c.JSON(500, gin.H{"error": "internal server error"})  // Return generic message
```

## Logging and Monitoring

### Audit Logging

**Log Security Events**:

```go
// Authentication failures
log.Printf("AUTH_FAIL: API key invalid from IP %s", clientIP)

// Account creation/modification
log.Printf("ACCOUNT_CREATED: %s by user %s", accountID, userID)

// Proxy changes
log.Printf("PROXY_DISABLED: %d reason: %s", proxyID, reason)

// Unusual activity
log.Printf("ANOMALY: High error rate from account %s", accountID)
```

### Security Monitoring

**Alert on**:
- Failed authentication attempts (>5 in 1 minute)
- Account modifications
- Proxy configuration changes
- Unusual traffic patterns
- OAuth token refresh failures
- High error rates (>10%)

## Secrets Management

### Environment Variables

**Don't hardcode secrets**:

```go
// DON'T
const dbPassword = "mypassword"

// DO
dbPassword := os.Getenv("DB_PASSWORD")
if dbPassword == "" {
    log.Fatal("DB_PASSWORD not set")
}
```

### Secret Management Services

**Recommended**:

1. **HashiCorp Vault**:
   ```go
   import "github.com/hashicorp/vault/api"

   client, _ := api.NewClient(api.DefaultConfig())
   secret, _ := client.Logical().Read("secret/data/aigateway/db")
   dbPassword := secret.Data["password"].(string)
   ```

2. **AWS Secrets Manager**:
   ```go
   import "github.com/aws/aws-sdk-go/service/secretsmanager"

   svc := secretsmanager.New(session.New())
   result, _ := svc.GetSecretValue(&secretsmanager.GetSecretValueInput{
       SecretId: aws.String("aigateway/db"),
   })
   dbPassword := *result.SecretString
   ```

3. **Environment-specific config files**:
   ```yaml
   # config/production.yaml
   database:
     password: ${DB_PASSWORD}  # Load from environment

   # Start with: DB_PASSWORD=xxx ./aigateway
   ```

## Compliance

### Data Protection

**GDPR/Privacy Considerations**:

1. **Personal data in logs**:
   - Don't log full API keys or tokens
   - Mask sensitive data (show only last 4 chars)
   - Anonymize IP addresses if required

2. **Data retention**:
   - Delete request logs after 30-90 days
   - Archive if needed for compliance
   - Provide data deletion mechanism

3. **User consent**:
   - Document data collection practices
   - Obtain consent for logging/monitoring
   - Provide data access/deletion APIs

### Security Compliance

**Industry Standards**:

1. **SOC 2 Type II**:
   - Implement audit logging
   - Regular security reviews
   - Incident response procedures

2. **ISO 27001**:
   - Information security policies
   - Risk assessment
   - Access control procedures

## Security Checklist

### Deployment Checklist

- [ ] Database credentials encrypted
- [ ] TLS enabled for all connections
- [ ] API authentication implemented
- [ ] Redis AUTH password set
- [ ] Firewall rules configured
- [ ] Rate limiting enabled
- [ ] Audit logging configured
- [ ] Secrets in environment variables or secret manager
- [ ] Input validation on all endpoints
- [ ] Error messages don't expose sensitive info
- [ ] Regular security updates applied
- [ ] Backup and recovery procedures tested

### Ongoing Security

- [ ] Rotate API keys every 90 days
- [ ] Review access logs weekly
- [ ] Update dependencies monthly
- [ ] Security audit quarterly
- [ ] Penetration testing annually
- [ ] Incident response plan documented
- [ ] Backup verification monthly

## Incident Response

### Security Incident Procedure

1. **Detect**: Monitor logs and alerts
2. **Contain**: Disable compromised accounts/proxies
3. **Investigate**: Analyze logs, determine scope
4. **Remediate**: Fix vulnerability, rotate credentials
5. **Document**: Record incident details
6. **Review**: Post-mortem, update procedures

### Emergency Contacts

Maintain list of:
- Security team contacts
- Provider security contacts
- Incident response team
- Legal/compliance team

## Related Documentation

- [Authentication Strategies](../providers/authentication.md) - Auth implementation
- [Database Setup](database.md) - Database security
- [Monitoring](monitoring.md) - Security monitoring
- [Architecture](../architecture/README.md) - System architecture
