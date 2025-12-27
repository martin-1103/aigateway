package manager

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics collects auth manager metrics
type Metrics struct {
	// Rotation counts per provider
	rotationCounts sync.Map // map[string]*int64

	// Cooldown events per reason
	cooldownEvents sync.Map // map[BlockReason]*int64

	// Account health snapshots
	healthSnapshots sync.Map // map[string]*AccountHealth

	// Retry counts
	retryTotal   int64
	retrySuccess int64

	// Selection metrics
	selectTotal   int64
	selectSuccess int64
	selectBlocked int64

	mu sync.RWMutex
}

// AccountHealth represents health status of an account
type AccountHealth struct {
	AccountID    string                 `json:"account_id"`
	ProviderID   string                 `json:"provider_id"`
	Label        string                 `json:"label"`
	IsDisabled   bool                   `json:"is_disabled"`
	ModelStates  map[string]ModelHealth `json:"model_states"`
	LastUpdated  time.Time              `json:"last_updated"`
}

// ModelHealth represents health status for a specific model
type ModelHealth struct {
	Model          string    `json:"model"`
	IsBlocked      bool      `json:"is_blocked"`
	BlockReason    string    `json:"block_reason,omitempty"`
	NextRetryAfter time.Time `json:"next_retry_after,omitempty"`
	SuccessCount   int64     `json:"success_count"`
	FailureCount   int64     `json:"failure_count"`
	LastUsedAt     time.Time `json:"last_used_at,omitempty"`
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{}
}

// RecordRotation records a rotation event for a provider
func (m *Metrics) RecordRotation(providerID string) {
	val, _ := m.rotationCounts.LoadOrStore(providerID, new(int64))
	atomic.AddInt64(val.(*int64), 1)
}

// RecordCooldown records a cooldown event
func (m *Metrics) RecordCooldown(reason BlockReason) {
	val, _ := m.cooldownEvents.LoadOrStore(reason, new(int64))
	atomic.AddInt64(val.(*int64), 1)
}

// RecordSelect records a selection attempt
func (m *Metrics) RecordSelect(success bool, allBlocked bool) {
	atomic.AddInt64(&m.selectTotal, 1)
	if success {
		atomic.AddInt64(&m.selectSuccess, 1)
	}
	if allBlocked {
		atomic.AddInt64(&m.selectBlocked, 1)
	}
}

// RecordRetry records a retry attempt
func (m *Metrics) RecordRetry(success bool) {
	atomic.AddInt64(&m.retryTotal, 1)
	if success {
		atomic.AddInt64(&m.retrySuccess, 1)
	}
}

// UpdateAccountHealth updates health snapshot for an account
func (m *Metrics) UpdateAccountHealth(acc *AccountState) {
	now := time.Now()

	health := &AccountHealth{
		AccountID:   acc.Account.ID,
		ProviderID:  acc.Account.ProviderID,
		Label:       acc.Account.Label,
		IsDisabled:  acc.Disabled,
		ModelStates: make(map[string]ModelHealth),
		LastUpdated: now,
	}

	acc.mu.RLock()
	for model, ms := range acc.ModelStates {
		health.ModelStates[model] = ModelHealth{
			Model:          ms.Model,
			IsBlocked:      ms.IsBlocked(now),
			BlockReason:    string(ms.BlockReason),
			NextRetryAfter: ms.NextRetryAfter,
			SuccessCount:   ms.SuccessCount,
			FailureCount:   ms.FailureCount,
			LastUsedAt:     ms.LastUsedAt,
		}
	}
	acc.mu.RUnlock()

	m.healthSnapshots.Store(acc.Account.ID, health)
}

// GetRotationCounts returns rotation counts per provider
func (m *Metrics) GetRotationCounts() map[string]int64 {
	result := make(map[string]int64)
	m.rotationCounts.Range(func(key, value interface{}) bool {
		result[key.(string)] = atomic.LoadInt64(value.(*int64))
		return true
	})
	return result
}

// GetCooldownEvents returns cooldown events per reason
func (m *Metrics) GetCooldownEvents() map[string]int64 {
	result := make(map[string]int64)
	m.cooldownEvents.Range(func(key, value interface{}) bool {
		result[string(key.(BlockReason))] = atomic.LoadInt64(value.(*int64))
		return true
	})
	return result
}

// GetAccountHealths returns all account health snapshots
func (m *Metrics) GetAccountHealths() []*AccountHealth {
	var result []*AccountHealth
	m.healthSnapshots.Range(func(key, value interface{}) bool {
		result = append(result, value.(*AccountHealth))
		return true
	})
	return result
}

// GetAccountHealth returns health snapshot for specific account
func (m *Metrics) GetAccountHealth(accountID string) *AccountHealth {
	if val, ok := m.healthSnapshots.Load(accountID); ok {
		return val.(*AccountHealth)
	}
	return nil
}

// GetSelectionStats returns selection statistics
func (m *Metrics) GetSelectionStats() map[string]int64 {
	return map[string]int64{
		"total":   atomic.LoadInt64(&m.selectTotal),
		"success": atomic.LoadInt64(&m.selectSuccess),
		"blocked": atomic.LoadInt64(&m.selectBlocked),
	}
}

// GetRetryStats returns retry statistics
func (m *Metrics) GetRetryStats() map[string]int64 {
	return map[string]int64{
		"total":   atomic.LoadInt64(&m.retryTotal),
		"success": atomic.LoadInt64(&m.retrySuccess),
	}
}

// Summary returns a summary of all metrics
func (m *Metrics) Summary() map[string]interface{} {
	return map[string]interface{}{
		"rotation_counts":  m.GetRotationCounts(),
		"cooldown_events":  m.GetCooldownEvents(),
		"selection_stats":  m.GetSelectionStats(),
		"retry_stats":      m.GetRetryStats(),
		"account_count":    len(m.GetAccountHealths()),
	}
}
