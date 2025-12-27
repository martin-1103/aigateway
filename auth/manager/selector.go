package manager

import (
	"context"
	"fmt"
	"time"
)

// AllBlockedError returned when all accounts are blocked
type AllBlockedError struct {
	WaitDuration time.Time // Earliest retry time
	Message      string
}

func (e *AllBlockedError) Error() string {
	return e.Message
}

// getCandidates returns accounts for given provider
func (m *Manager) getCandidates(providerID string) []*AccountState {
	candidates := make([]*AccountState, 0)

	for _, acc := range m.accounts {
		if acc.Account.ProviderID == providerID && !acc.Disabled {
			candidates = append(candidates, acc)
		}
	}

	return candidates
}

// selectBest selects best available account for model
func (m *Manager) selectBest(candidates []*AccountState, model string) (*AccountState, error) {
	now := time.Now()
	available := make([]*AccountState, 0)
	var earliestRetry time.Time

	// Filter available accounts
	for _, acc := range candidates {
		blocked, _ := acc.IsBlockedFor(model, now)
		if blocked {
			retryTime := acc.GetNextRetryTime(model)
			if !retryTime.IsZero() {
				if earliestRetry.IsZero() || retryTime.Before(earliestRetry) {
					earliestRetry = retryTime
				}
			}
			continue
		}
		available = append(available, acc)
	}

	if len(available) == 0 {
		return nil, &AllBlockedError{
			WaitDuration: earliestRetry,
			Message:      fmt.Sprintf("all accounts blocked, retry at %v", earliestRetry),
		}
	}

	// Round-robin selection
	return m.roundRobinSelect(available, model)
}

// roundRobinSelect picks next account using round-robin
func (m *Manager) roundRobinSelect(available []*AccountState, model string) (*AccountState, error) {
	if len(available) == 0 {
		return nil, fmt.Errorf("no available accounts")
	}

	if len(available) == 1 {
		return available[0], nil
	}

	// Get counter from Redis for fair distribution
	counter := m.getCounter(model)
	idx := int(counter) % len(available)

	return available[idx], nil
}

// getCounter gets and increments round-robin counter from Redis
func (m *Manager) getCounter(model string) int64 {
	if m.redis == nil {
		return 0
	}

	key := fmt.Sprintf("auth:rr:%s", model)
	ctx := context.Background()

	val, err := m.redis.Incr(ctx, key).Result()
	if err != nil {
		return 0
	}

	return val
}

// SelectWithRetry selects account with wait-and-retry for blocked accounts
func (m *Manager) SelectWithRetry(ctx context.Context, providerID, model string, maxWait time.Duration) (*AccountState, error) {
	acc, err := m.Select(ctx, providerID, model)
	if err == nil {
		return acc, nil
	}

	// Check if all blocked with wait time
	allBlocked, ok := err.(*AllBlockedError)
	if !ok {
		return nil, err
	}

	// Calculate wait duration
	waitDur := time.Until(allBlocked.WaitDuration)
	if waitDur <= 0 {
		// Retry immediately
		return m.Select(ctx, providerID, model)
	}

	if waitDur > maxWait {
		return nil, fmt.Errorf("wait time %v exceeds max %v", waitDur, maxWait)
	}

	// Wait and retry
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(waitDur):
		return m.Select(ctx, providerID, model)
	}
}
