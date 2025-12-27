package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"aigateway/auth/errors"
	"aigateway/models"
	"aigateway/repositories"

	"github.com/redis/go-redis/v9"
)

// QuotaTracker interface for quota tracking (avoid circular import)
type QuotaTracker interface {
	RecordUsage(accountID, model string, tokens int64)
	MarkExhausted(accountID, model string)
	IsAvailable(accountID, model string) bool
	GetEarliestReset(accountIDs []string, model string) *time.Time
}

// TokenExtractor interface for extracting tokens from response
type TokenExtractor interface {
	ExtractTokens(providerID string, payload []byte) int64
}

// Manager manages account states for all providers
type Manager struct {
	accounts map[string]*AccountState // key: account ID
	mu       sync.RWMutex

	// Dependencies
	accountRepo *repositories.AccountRepository
	redis       *redis.Client

	// Error parsers per provider
	errorParsers map[string]errors.ErrorParser

	// Token refreshers per provider
	refreshers map[string]TokenRefresher

	// Quota tracking
	quotaTracker   QuotaTracker
	tokenExtractor TokenExtractor

	// Background refresh control
	refreshCancel context.CancelFunc

	// Observability
	metrics *Metrics
	logger  *StateLogger
}

// NewManager creates a new auth manager
func NewManager(
	accountRepo *repositories.AccountRepository,
	redisClient *redis.Client,
) *Manager {
	m := &Manager{
		accounts:     make(map[string]*AccountState),
		accountRepo:  accountRepo,
		redis:        redisClient,
		errorParsers: make(map[string]errors.ErrorParser),
		refreshers:   make(map[string]TokenRefresher),
		metrics:      NewMetrics(),
		logger:       NewStateLogger(true),
	}

	// Register default error parsers
	m.errorParsers["claude"] = &errors.ClaudeParser{}
	m.errorParsers["anthropic"] = &errors.ClaudeParser{}
	m.errorParsers["codex"] = &errors.CodexParser{}
	m.errorParsers["openai"] = &errors.CodexParser{}
	m.errorParsers["antigravity"] = &errors.AntigravityParser{}

	return m
}

// GetMetrics returns the metrics collector
func (m *Manager) GetMetrics() *Metrics {
	return m.metrics
}

// SetLogging enables or disables state change logging
func (m *Manager) SetLogging(enabled bool) {
	m.logger.SetEnabled(enabled)
}

// RegisterParser registers error parser for provider
func (m *Manager) RegisterParser(providerID string, parser errors.ErrorParser) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorParsers[providerID] = parser
}

// RegisterRefresher registers token refresher for provider
func (m *Manager) RegisterRefresher(providerID string, refresher TokenRefresher) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.refreshers[providerID] = refresher
}

// SetQuotaTracker sets the quota tracker for usage tracking
func (m *Manager) SetQuotaTracker(tracker QuotaTracker, extractor TokenExtractor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.quotaTracker = tracker
	m.tokenExtractor = extractor
}

// LoadAccounts loads accounts from database into manager
func (m *Manager) LoadAccounts(ctx context.Context, providerIDs ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, providerID := range providerIDs {
		accounts, err := m.accountRepo.GetActiveByProvider(providerID)
		if err != nil {
			return fmt.Errorf("failed to load accounts for %s: %w", providerID, err)
		}

		for _, acc := range accounts {
			m.accounts[acc.ID] = NewAccountState(acc)
		}

		m.logger.LogAccountLoaded(providerID, len(accounts))
	}

	return nil
}

// Select picks best available account for provider and model
func (m *Manager) Select(ctx context.Context, providerID, model string) (*AccountState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	candidates := m.getCandidates(providerID)
	if len(candidates) == 0 {
		m.metrics.RecordSelect(false, false)
		return nil, fmt.Errorf("no accounts for provider %s", providerID)
	}

	acc, err := m.selectBest(candidates, model)
	if err != nil {
		if _, ok := err.(*AllBlockedError); ok {
			m.metrics.RecordSelect(false, true)
			m.logger.LogAllBlocked(providerID, model, time.Now())
		} else {
			m.metrics.RecordSelect(false, false)
		}
		return nil, err
	}

	m.metrics.RecordSelect(true, false)
	m.metrics.RecordRotation(providerID)
	m.logger.LogAccountSelected(acc.Account.ID, providerID, model)

	return acc, nil
}

// MarkResult updates account state based on execution result
func (m *Manager) MarkResult(accountID, model string, statusCode int, body []byte) {
	m.mu.RLock()
	acc, exists := m.accounts[accountID]
	m.mu.RUnlock()

	if !exists {
		return
	}

	now := time.Now()

	// Success case
	if statusCode >= 200 && statusCode < 300 {
		acc.MarkSuccess(model, now)
		m.logger.LogSuccess(accountID, model)
		m.metrics.UpdateAccountHealth(acc)

		// Track quota usage (extract tokens from response)
		if m.quotaTracker != nil && m.tokenExtractor != nil {
			tokens := m.tokenExtractor.ExtractTokens(acc.Account.ProviderID, body)
			m.quotaTracker.RecordUsage(accountID, model, tokens)
		}
		return
	}

	// Parse error
	parser := m.getParser(acc.Account.ProviderID)
	parsed := parser.Parse(statusCode, body)
	acc.MarkFailure(model, parsed, now)

	// Check for quota exhaustion
	if parsed.Type == errors.ErrTypeQuotaExceeded && m.quotaTracker != nil {
		m.quotaTracker.MarkExhausted(accountID, model)
		m.logger.LogQuotaExhausted(accountID, model)
	}

	// Log and record metrics
	m.logger.LogFailure(accountID, model, parsed)
	m.metrics.UpdateAccountHealth(acc)

	// Record cooldown event
	ms := acc.GetModelState(model)
	if ms.BlockReason != BlockReasonNone {
		m.metrics.RecordCooldown(ms.BlockReason)
		m.logger.LogAccountBlocked(accountID, model, ms.BlockReason, ms.NextRetryAfter)
	}

	// Check if account was disabled
	if acc.Disabled {
		m.logger.LogAccountDisabled(accountID, string(parsed.Type))
	}
}

// GetAccount returns account state by ID
func (m *Manager) GetAccount(accountID string) *AccountState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.accounts[accountID]
}

// GetAllAccounts returns snapshot of all account states
func (m *Manager) GetAllAccounts() []*AccountState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*AccountState, 0, len(m.accounts))
	for _, acc := range m.accounts {
		result = append(result, acc)
	}
	return result
}

// AddAccount adds new account to manager
func (m *Manager) AddAccount(account *models.Account) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.accounts[account.ID] = NewAccountState(account)
}

// RemoveAccount removes account from manager
func (m *Manager) RemoveAccount(accountID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.accounts, accountID)
}

func (m *Manager) getParser(providerID string) errors.ErrorParser {
	if parser, exists := m.errorParsers[providerID]; exists {
		return parser
	}
	return &errors.DefaultParser{}
}
