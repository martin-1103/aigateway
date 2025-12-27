package services

import (
	"fmt"
	"time"
)

// Redis key patterns for quota tracking
const (
	// Key prefixes
	quotaKeyPrefix = "quota"

	// TTL for quota window (5 hours for Antigravity)
	QuotaWindowTTL = 5 * time.Hour
)

// QuotaKeys provides Redis key generation for quota tracking
type QuotaKeys struct{}

// RequestsKey returns the key for tracking request count
// Format: quota:{account_id}:{model}:requests
func (QuotaKeys) RequestsKey(accountID, model string) string {
	return fmt.Sprintf("%s:%s:%s:requests", quotaKeyPrefix, accountID, model)
}

// TokensKey returns the key for tracking token count
// Format: quota:{account_id}:{model}:tokens
func (QuotaKeys) TokensKey(accountID, model string) string {
	return fmt.Sprintf("%s:%s:%s:tokens", quotaKeyPrefix, accountID, model)
}

// ExhaustedKey returns the key for marking account+model as exhausted
// Format: quota:{account_id}:{model}:exhausted
func (QuotaKeys) ExhaustedKey(accountID, model string) string {
	return fmt.Sprintf("%s:%s:%s:exhausted", quotaKeyPrefix, accountID, model)
}

// WindowStartKey returns the key for tracking window start time
// Format: quota:{account_id}:{model}:window_start
func (QuotaKeys) WindowStartKey(accountID, model string) string {
	return fmt.Sprintf("%s:%s:%s:window_start", quotaKeyPrefix, accountID, model)
}

// AllKeysPattern returns pattern to match all quota keys for an account+model
// Format: quota:{account_id}:{model}:*
func (QuotaKeys) AllKeysPattern(accountID, model string) string {
	return fmt.Sprintf("%s:%s:%s:*", quotaKeyPrefix, accountID, model)
}

// AccountPattern returns pattern to match all quota keys for an account
// Format: quota:{account_id}:*
func (QuotaKeys) AccountPattern(accountID string) string {
	return fmt.Sprintf("%s:%s:*", quotaKeyPrefix, accountID)
}
