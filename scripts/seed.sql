-- ============================================
-- AI Gateway Seed Data
-- ============================================
-- This script populates the database with initial test data:
-- - 3 AI providers (antigravity, openai, glm)
-- - Example accounts for each provider
-- - Example proxy servers
-- ============================================

-- ============================================
-- PROVIDERS
-- ============================================

-- Antigravity Provider (OAuth-based)
INSERT INTO providers (id, name, base_url, auth_type, is_active, config, created_at, updated_at)
VALUES (
    'antigravity',
    'Antigravity AI',
    'https://api.antigravity.ai/v1',
    'oauth',
    1,
    JSON_OBJECT(
        'oauth_config', JSON_OBJECT(
            'auth_url', 'https://oauth.antigravity.ai/authorize',
            'token_url', 'https://oauth.antigravity.ai/token',
            'scope', 'chat.completion'
        ),
        'rate_limits', JSON_OBJECT(
            'requests_per_minute', 60,
            'tokens_per_minute', 90000
        ),
        'supported_models', JSON_ARRAY('antigravity-fast', 'antigravity-smart')
    ),
    NOW(),
    NOW()
);

-- OpenAI Provider (API Key-based)
INSERT INTO providers (id, name, base_url, auth_type, is_active, config, created_at, updated_at)
VALUES (
    'openai',
    'OpenAI',
    'https://api.openai.com/v1',
    'api_key',
    1,
    JSON_OBJECT(
        'api_version', 'v1',
        'rate_limits', JSON_OBJECT(
            'requests_per_minute', 500,
            'tokens_per_minute', 150000
        ),
        'supported_models', JSON_ARRAY('gpt-4', 'gpt-4-turbo', 'gpt-3.5-turbo'),
        'headers', JSON_OBJECT(
            'OpenAI-Organization', 'optional_org_id'
        )
    ),
    NOW(),
    NOW()
);

-- GLM Provider (Bearer Token-based)
INSERT INTO providers (id, name, base_url, auth_type, is_active, config, created_at, updated_at)
VALUES (
    'glm',
    'GLM (ChatGLM)',
    'https://open.bigmodel.cn/api/paas/v4',
    'bearer',
    1,
    JSON_OBJECT(
        'api_version', 'v4',
        'rate_limits', JSON_OBJECT(
            'requests_per_minute', 100,
            'tokens_per_minute', 100000
        ),
        'supported_models', JSON_ARRAY('glm-4', 'glm-4-plus', 'glm-3-turbo'),
        'timeout_seconds', 30
    ),
    NOW(),
    NOW()
);

-- ============================================
-- PROXY POOL
-- ============================================

-- High-priority US proxy
INSERT INTO proxy_pool (
    url, protocol, is_active, health_status, max_accounts, current_accounts,
    priority, weight, max_failures, success_rate, avg_latency_ms,
    created_at, updated_at
)
VALUES (
    'http://proxy-us-01.example.com:8080',
    'http',
    1,
    'healthy',
    50,
    0,
    10,
    5,
    3,
    100.00,
    120,
    NOW(),
    NOW()
);

-- Medium-priority EU proxy
INSERT INTO proxy_pool (
    url, protocol, is_active, health_status, max_accounts, current_accounts,
    priority, weight, max_failures, success_rate, avg_latency_ms,
    created_at, updated_at
)
VALUES (
    'https://proxy-eu-01.example.com:443',
    'https',
    1,
    'healthy',
    30,
    0,
    5,
    3,
    3,
    98.50,
    180,
    NOW(),
    NOW()
);

-- Backup SOCKS5 proxy
INSERT INTO proxy_pool (
    url, protocol, is_active, health_status, max_accounts, current_accounts,
    priority, weight, max_failures, success_rate, avg_latency_ms,
    created_at, updated_at
)
VALUES (
    'socks5://proxy-backup.example.com:1080',
    'socks5',
    1,
    'healthy',
    20,
    0,
    1,
    1,
    5,
    95.00,
    250,
    NOW(),
    NOW()
);

-- ============================================
-- ACCOUNTS - Antigravity
-- ============================================

-- Antigravity Account 1 (OAuth with refresh token)
INSERT INTO accounts (
    id, provider_id, label, auth_data, metadata, is_active,
    proxy_id, usage_count, created_at, updated_at
)
VALUES (
    UUID(),
    'antigravity',
    'antigravity-prod-001',
    JSON_OBJECT(
        'access_token', 'ag_access_token_example_001',
        'refresh_token', 'ag_refresh_token_example_001',
        'expires_at', UNIX_TIMESTAMP(DATE_ADD(NOW(), INTERVAL 1 HOUR))
    ),
    JSON_OBJECT(
        'account_owner', 'team-alpha',
        'environment', 'production',
        'quota_limit', 100000
    ),
    1,
    1,
    0,
    NOW(),
    NOW()
);

-- Antigravity Account 2 (OAuth with refresh token)
INSERT INTO accounts (
    id, provider_id, label, auth_data, metadata, is_active,
    proxy_id, usage_count, created_at, updated_at
)
VALUES (
    UUID(),
    'antigravity',
    'antigravity-staging-001',
    JSON_OBJECT(
        'access_token', 'ag_access_token_example_002',
        'refresh_token', 'ag_refresh_token_example_002',
        'expires_at', UNIX_TIMESTAMP(DATE_ADD(NOW(), INTERVAL 1 HOUR))
    ),
    JSON_OBJECT(
        'account_owner', 'team-beta',
        'environment', 'staging',
        'quota_limit', 50000
    ),
    1,
    2,
    0,
    NOW(),
    NOW()
);

-- ============================================
-- ACCOUNTS - OpenAI
-- ============================================

-- OpenAI Account 1 (API Key)
INSERT INTO accounts (
    id, provider_id, label, auth_data, metadata, is_active,
    proxy_id, usage_count, created_at, updated_at
)
VALUES (
    UUID(),
    'openai',
    'openai-main-001',
    JSON_OBJECT(
        'api_key', 'sk-example-key-001-XXXXXXXXXXXXXXXXXXXXXXXX'
    ),
    JSON_OBJECT(
        'account_owner', 'engineering',
        'billing_tier', 'tier-4',
        'monthly_budget', 5000
    ),
    1,
    1,
    0,
    NOW(),
    NOW()
);

-- OpenAI Account 2 (API Key)
INSERT INTO accounts (
    id, provider_id, label, auth_data, metadata, is_active,
    proxy_id, usage_count, created_at, updated_at
)
VALUES (
    UUID(),
    'openai',
    'openai-backup-001',
    JSON_OBJECT(
        'api_key', 'sk-example-key-002-XXXXXXXXXXXXXXXXXXXXXXXX'
    ),
    JSON_OBJECT(
        'account_owner', 'product',
        'billing_tier', 'tier-3',
        'monthly_budget', 2000
    ),
    1,
    2,
    0,
    NOW(),
    NOW()
);

-- OpenAI Account 3 (API Key - inactive for testing)
INSERT INTO accounts (
    id, provider_id, label, auth_data, metadata, is_active,
    proxy_id, usage_count, created_at, updated_at
)
VALUES (
    UUID(),
    'openai',
    'openai-deprecated-001',
    JSON_OBJECT(
        'api_key', 'sk-example-key-003-XXXXXXXXXXXXXXXXXXXXXXXX'
    ),
    JSON_OBJECT(
        'account_owner', 'legacy',
        'deprecated_date', '2024-01-01',
        'reason', 'billing_expired'
    ),
    0,
    NULL,
    0,
    NOW(),
    NOW()
);

-- ============================================
-- ACCOUNTS - GLM
-- ============================================

-- GLM Account 1 (Bearer Token)
INSERT INTO accounts (
    id, provider_id, label, auth_data, metadata, is_active,
    proxy_id, usage_count, created_at, updated_at
)
VALUES (
    UUID(),
    'glm',
    'glm-primary-001',
    JSON_OBJECT(
        'bearer_token', 'glm_bearer_token_example_001_XXXXXXXXXX'
    ),
    JSON_OBJECT(
        'account_owner', 'research-team',
        'region', 'cn-north',
        'tier', 'enterprise'
    ),
    1,
    3,
    0,
    NOW(),
    NOW()
);

-- GLM Account 2 (Bearer Token)
INSERT INTO accounts (
    id, provider_id, label, auth_data, metadata, is_active,
    proxy_id, usage_count, created_at, updated_at
)
VALUES (
    UUID(),
    'glm',
    'glm-secondary-001',
    JSON_OBJECT(
        'bearer_token', 'glm_bearer_token_example_002_XXXXXXXXXX'
    ),
    JSON_OBJECT(
        'account_owner', 'data-science',
        'region', 'cn-south',
        'tier', 'standard'
    ),
    1,
    1,
    0,
    NOW(),
    NOW()
);

-- ============================================
-- Summary
-- ============================================
-- Providers: 3 (antigravity, openai, glm)
-- Proxies: 3 (US, EU, Backup)
-- Accounts: 8 total
--   - Antigravity: 2
--   - OpenAI: 3 (1 inactive)
--   - GLM: 2
-- ============================================
