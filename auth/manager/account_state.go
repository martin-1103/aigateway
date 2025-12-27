package manager

import (
	"sync"
	"time"

	"aigateway/auth/errors"
	"aigateway/models"
)

// AccountState wraps account with state tracking
type AccountState struct {
	Account     *models.Account         // Underlying account
	ModelStates map[string]*ModelState  // State per model
	QuotaState  *QuotaState             // Quota backoff state
	Disabled    bool                    // Account-level disable
	LastError   *errors.ParsedError     // Last account-level error
	UpdatedAt   time.Time               // Last state update

	// Token refresh tracking
	LastRefreshedAt  time.Time // When token was last refreshed
	NextRefreshAfter time.Time // Backoff for refresh failures

	mu sync.RWMutex // Protects state mutations
}

// NewAccountState creates a new AccountState from account
func NewAccountState(account *models.Account) *AccountState {
	return &AccountState{
		Account:     account,
		ModelStates: make(map[string]*ModelState),
		QuotaState:  NewQuotaState(),
		UpdatedAt:   time.Now(),
	}
}

// GetModelState returns ModelState for given model, creating if needed
func (a *AccountState) GetModelState(model string) *ModelState {
	a.mu.Lock()
	defer a.mu.Unlock()

	if ms, exists := a.ModelStates[model]; exists {
		return ms
	}

	ms := &ModelState{Model: model}
	a.ModelStates[model] = ms
	return ms
}

// IsBlockedFor checks if account is blocked for specific model
func (a *AccountState) IsBlockedFor(model string, now time.Time) (bool, BlockReason) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Check account-level disable
	if a.Disabled {
		return true, BlockReasonDisabled
	}

	// Check model-level state
	ms, exists := a.ModelStates[model]
	if !exists {
		return false, BlockReasonNone
	}

	if ms.Disabled {
		return true, BlockReasonDisabled
	}

	if ms.IsBlocked(now) {
		return true, ms.BlockReason
	}

	return false, BlockReasonNone
}

// MarkSuccess records successful request
func (a *AccountState) MarkSuccess(model string, now time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()

	ms := a.getOrCreateModelState(model)
	ms.LastUsedAt = now
	ms.SuccessCount++
	ms.ClearBlock()

	// Reset quota backoff on success
	a.QuotaState.Reset()
	a.UpdatedAt = now
}

// MarkFailure records failed request with parsed error
func (a *AccountState) MarkFailure(model string, err *errors.ParsedError, now time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()

	ms := a.getOrCreateModelState(model)
	ms.LastUsedAt = now
	ms.FailureCount++
	ms.LastError = err

	switch err.Type {
	case errors.ErrTypeAuthentication, errors.ErrTypePermission:
		ms.BlockReason = BlockReasonAuth
		ms.NextRetryAfter = now.Add(err.CooldownDur)
		a.Disabled = true // Disable entire account

	case errors.ErrTypeQuotaExceeded:
		ms.BlockReason = BlockReasonQuota
		a.QuotaState.Increment()
		ms.NextRetryAfter = now.Add(a.QuotaState.NextBackoff())

	case errors.ErrTypeRateLimit:
		ms.BlockReason = BlockReasonCooldown
		ms.NextRetryAfter = now.Add(err.CooldownDur)

	case errors.ErrTypeOverloaded, errors.ErrTypeTransient:
		ms.BlockReason = BlockReasonCooldown
		ms.NextRetryAfter = now.Add(err.CooldownDur)

	case errors.ErrTypeNotFound:
		ms.Disabled = true
		ms.BlockReason = BlockReasonDisabled
		ms.NextRetryAfter = now.Add(err.CooldownDur)
	}

	a.UpdatedAt = now
}

// GetNextRetryTime returns earliest retry time for given model
func (a *AccountState) GetNextRetryTime(model string) time.Time {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if ms, exists := a.ModelStates[model]; exists {
		return ms.NextRetryAfter
	}
	return time.Time{}
}

func (a *AccountState) getOrCreateModelState(model string) *ModelState {
	if ms, exists := a.ModelStates[model]; exists {
		return ms
	}
	ms := &ModelState{Model: model}
	a.ModelStates[model] = ms
	return ms
}
