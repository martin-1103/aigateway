package manager

import (
	"time"

	"aigateway/auth/errors"
)

// BlockReason indicates why an account is blocked
type BlockReason string

const (
	BlockReasonNone     BlockReason = ""
	BlockReasonDisabled BlockReason = "disabled"   // Manually or permanently disabled
	BlockReasonCooldown BlockReason = "cooldown"   // Temporary cooldown (rate limit)
	BlockReasonQuota    BlockReason = "quota"      // Quota exceeded
	BlockReasonAuth     BlockReason = "auth_failed" // Authentication failed
)

// ModelState tracks the state of an account for a specific model
type ModelState struct {
	Model          string              // Model name (e.g., "claude-3-opus")
	Disabled       bool                // Permanently disabled for this model
	BlockReason    BlockReason         // Current block reason
	NextRetryAfter time.Time           // When to retry after block
	LastError      *errors.ParsedError // Last error encountered
	LastUsedAt     time.Time           // Last successful use
	SuccessCount   int64               // Total successful requests
	FailureCount   int64               // Total failed requests
}

// IsBlocked returns true if model is currently blocked
func (ms *ModelState) IsBlocked(now time.Time) bool {
	if ms.Disabled {
		return true
	}
	if !ms.NextRetryAfter.IsZero() && now.Before(ms.NextRetryAfter) {
		return true
	}
	return false
}

// ClearBlock clears the block state
func (ms *ModelState) ClearBlock() {
	ms.BlockReason = BlockReasonNone
	ms.NextRetryAfter = time.Time{}
	ms.LastError = nil
}

// QuotaState tracks quota with exponential backoff
type QuotaState struct {
	BackoffMultiplier int           // Current multiplier (1, 2, 4, 8...)
	BaseBackoff       time.Duration // Base backoff duration (default: 1s)
	MaxBackoff        time.Duration // Maximum backoff duration (default: 30m)
	LastQuotaError    time.Time     // When last quota error occurred
}

// NewQuotaState creates a new QuotaState with defaults
func NewQuotaState() *QuotaState {
	return &QuotaState{
		BackoffMultiplier: 0,
		BaseBackoff:       time.Second,
		MaxBackoff:        30 * time.Minute,
	}
}

// NextBackoff calculates next backoff duration
func (q *QuotaState) NextBackoff() time.Duration {
	if q.BackoffMultiplier == 0 {
		return q.BaseBackoff
	}
	backoff := q.BaseBackoff * time.Duration(q.BackoffMultiplier)
	if backoff > q.MaxBackoff {
		return q.MaxBackoff
	}
	return backoff
}

// Increment increases backoff multiplier
func (q *QuotaState) Increment() {
	if q.BackoffMultiplier == 0 {
		q.BackoffMultiplier = 1
	} else {
		q.BackoffMultiplier *= 2
	}
	// Cap at reasonable max (30 min with 1s base = 1800)
	if q.BackoffMultiplier > 1800 {
		q.BackoffMultiplier = 1800
	}
	q.LastQuotaError = time.Now()
}

// Reset resets backoff to initial state
func (q *QuotaState) Reset() {
	q.BackoffMultiplier = 0
}
