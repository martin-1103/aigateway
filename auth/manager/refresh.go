package manager

import (
	"context"
	"log"
	"time"

	"aigateway/models"
)

// TokenRefresher interface for provider-specific token refresh
type TokenRefresher interface {
	// RefreshLead returns how long before expiry to refresh
	RefreshLead() time.Duration

	// Refresh refreshes token for account
	Refresh(ctx context.Context, account *models.Account) (*TokenResult, error)
}

// TokenResult contains refreshed token data
type TokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
	Metadata     map[string]interface{}
}

// Default refresh settings
const (
	DefaultRefreshInterval = 30 * time.Second
	RefreshFailureBackoff  = 30 * time.Second
)

// StartAutoRefresh starts background token refresh loop
func (m *Manager) StartAutoRefresh(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = DefaultRefreshInterval
	}

	// Cancel previous refresh loop if any
	if m.refreshCancel != nil {
		m.refreshCancel()
	}

	refreshCtx, cancel := context.WithCancel(ctx)
	m.refreshCancel = cancel

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run immediately on start
		m.checkRefreshes(refreshCtx)

		for {
			select {
			case <-refreshCtx.Done():
				return
			case <-ticker.C:
				m.checkRefreshes(refreshCtx)
			}
		}
	}()
}

// StopAutoRefresh stops background refresh loop
func (m *Manager) StopAutoRefresh() {
	if m.refreshCancel != nil {
		m.refreshCancel()
		m.refreshCancel = nil
	}
}

// checkRefreshes checks all accounts for needed refreshes
func (m *Manager) checkRefreshes(ctx context.Context) {
	accounts := m.GetAllAccounts()
	now := time.Now()

	for _, acc := range accounts {
		if acc.Disabled {
			continue
		}

		refresher := m.getRefresher(acc.Account.ProviderID)
		if refresher == nil {
			continue // No refresh needed for this provider
		}

		if !m.shouldRefresh(acc, refresher, now) {
			continue
		}

		// Refresh in goroutine to not block
		go m.refreshAccount(ctx, acc, refresher)
	}
}

// shouldRefresh checks if account needs token refresh
func (m *Manager) shouldRefresh(acc *AccountState, refresher TokenRefresher, now time.Time) bool {
	// Check if refresh already pending
	if !acc.NextRefreshAfter.IsZero() && now.Before(acc.NextRefreshAfter) {
		return false
	}

	// Get expiry from account auth data
	expiresAt := getExpiryFromAccount(acc.Account)
	if expiresAt.IsZero() {
		return false
	}

	// Check if within refresh lead time
	lead := refresher.RefreshLead()
	return time.Until(expiresAt) <= lead
}

// refreshAccount performs token refresh for account
func (m *Manager) refreshAccount(ctx context.Context, acc *AccountState, refresher TokenRefresher) {
	refreshCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	result, err := refresher.Refresh(refreshCtx, acc.Account)
	now := time.Now()

	if err != nil {
		log.Printf("Token refresh failed for %s: %v", acc.Account.ID, err)
		acc.mu.Lock()
		acc.NextRefreshAfter = now.Add(RefreshFailureBackoff)
		acc.mu.Unlock()
		return
	}

	// Update account with new token
	m.updateAccountToken(acc, result, now)
	log.Printf("Token refreshed for %s", acc.Account.ID)
}

func (m *Manager) getRefresher(providerID string) TokenRefresher {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.refreshers[providerID]
}

func (m *Manager) updateAccountToken(acc *AccountState, result *TokenResult, now time.Time) {
	acc.mu.Lock()
	defer acc.mu.Unlock()

	acc.LastRefreshedAt = now
	acc.NextRefreshAfter = time.Time{}

	// Note: Actual auth data update should be done via repository
	// This is a placeholder for the integration
}

// getExpiryFromAccount extracts token expiry from account auth data
func getExpiryFromAccount(account *models.Account) time.Time {
	// Parse auth data to get expires_at
	// Implementation depends on auth data format
	// This is a simplified version
	return time.Time{}
}
