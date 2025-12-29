-- Migration: Add health tracking columns to accounts table
-- Date: 2025-12-29

ALTER TABLE accounts
ADD COLUMN health_status VARCHAR(20) DEFAULT 'healthy' AFTER usage_count,
ADD COLUMN failure_count INT DEFAULT 0 AFTER health_status,
ADD COLUMN last_error_at TIMESTAMP NULL AFTER failure_count,
ADD COLUMN last_error_msg TEXT NULL AFTER last_error_at,
ADD COLUMN last_success_at TIMESTAMP NULL AFTER last_error_msg,
ADD INDEX idx_health_status (health_status);

-- Update existing accounts to healthy status
UPDATE accounts SET health_status = 'healthy' WHERE health_status IS NULL;

-- Rollback script (save for reference):
-- ALTER TABLE accounts
-- DROP COLUMN health_status,
-- DROP COLUMN failure_count,
-- DROP COLUMN last_error_at,
-- DROP COLUMN last_error_msg,
-- DROP COLUMN last_success_at,
-- DROP INDEX idx_health_status;
