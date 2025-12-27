package services

import (
	"aigateway-backend/internal/config"
	"aigateway-backend/models"
	"aigateway-backend/repositories"
	"fmt"
	"sync"
	"time"
)

// ProxyService handles proxy assignment and management operations
type ProxyService struct {
	repo                 *repositories.ProxyRepository
	accountRepo          *repositories.AccountRepository
	mu                   sync.RWMutex
	downRecoveryDelay    time.Duration
}

// NewProxyService creates a new proxy service instance
func NewProxyService(repo *repositories.ProxyRepository, accountRepo *repositories.AccountRepository, cfg *config.ProxyConfig) *ProxyService {
	recoveryDelay := 24 * time.Hour // default 24h
	if cfg != nil && cfg.DownRecoveryDelayMin > 0 {
		recoveryDelay = time.Duration(cfg.DownRecoveryDelayMin) * time.Minute
	}
	return &ProxyService{
		repo:              repo,
		accountRepo:       accountRepo,
		downRecoveryDelay: recoveryDelay,
	}
}

// AssignProxy assigns an available proxy to an account based on capacity and health
func (s *ProxyService) AssignProxy(account *models.Account, providerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if current proxy is still valid
	if account.ProxyID != nil && s.isProxyValid(*account.ProxyID) {
		return nil
	}

	// Clear invalid proxy
	if account.ProxyID != nil {
		s.repo.DecrementAccountCount(*account.ProxyID)
	}

	// Get active proxies for provider
	proxies, err := s.repo.GetActiveByProvider(providerID)
	if err != nil || len(proxies) == 0 {
		// No proxies available, clear proxy assignment
		account.ProxyURL = ""
		account.ProxyID = nil
		s.accountRepo.ClearProxy(account.ID)
		return nil
	}

	// Find proxy with available capacity
	for _, proxy := range proxies {
		if s.hasCapacity(proxy) {
			account.ProxyURL = proxy.URL
			account.ProxyID = &proxy.ID

			s.accountRepo.UpdateProxy(account.ID, proxy.ID, proxy.URL)
			s.repo.IncrementAccountCount(proxy.ID)

			return nil
		}
	}

	// No capacity available, clear proxy assignment
	account.ProxyURL = ""
	account.ProxyID = nil
	s.accountRepo.ClearProxy(account.ID)
	return nil
}

// hasCapacity checks if a proxy has available capacity for more accounts
func (s *ProxyService) hasCapacity(proxy *models.Proxy) bool {
	if proxy.MaxAccounts <= 0 {
		return true
	}
	return proxy.CurrentAccounts < proxy.MaxAccounts
}

// isProxyValid checks if a proxy is active and healthy
func (s *ProxyService) isProxyValid(proxyID int) bool {
	proxy, err := s.repo.GetByID(proxyID)
	if err != nil {
		return false
	}
	return proxy.IsActive && proxy.HealthStatus != models.HealthStatusDown
}

// Create creates a new proxy
func (s *ProxyService) Create(proxy *models.Proxy) error {
	return s.repo.Create(proxy)
}

// GetByID retrieves a proxy by ID
func (s *ProxyService) GetByID(id int) (*models.Proxy, error) {
	return s.repo.GetByID(id)
}

// List retrieves a paginated list of proxies
func (s *ProxyService) List(limit, offset int) ([]*models.Proxy, int64, error) {
	return s.repo.List(limit, offset)
}

// Update updates an existing proxy
func (s *ProxyService) Update(proxy *models.Proxy) error {
	return s.repo.Update(proxy)
}

// Delete deletes a proxy by ID
func (s *ProxyService) Delete(id int) error {
	return s.repo.Delete(id)
}

// GetAssignments returns a map of proxy IDs to assigned account IDs
func (s *ProxyService) GetAssignments() (map[int][]string, error) {
	accounts, err := s.accountRepo.GetByProvider("antigravity")
	if err != nil {
		return nil, err
	}

	assignments := make(map[int][]string)
	for _, acc := range accounts {
		if acc.ProxyID != nil {
			assignments[*acc.ProxyID] = append(assignments[*acc.ProxyID], acc.ID)
		}
	}

	return assignments, nil
}

// RecalculateCounts recalculates the account counts for all proxies
func (s *ProxyService) RecalculateCounts() error {
	return s.repo.RecalculateAccountCounts()
}

// SelectProxyForNewAccount selects and assigns a proxy for a new account during registration
// Returns the selected proxy or error if no proxy available
func (s *ProxyService) SelectProxyForNewAccount(providerID string) (*models.Proxy, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	proxies, err := s.repo.GetActiveByProvider(providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proxies: %w", err)
	}

	if len(proxies) == 0 {
		return nil, fmt.Errorf("no active proxies available for provider %s", providerID)
	}

	// Find proxy with capacity, prefer healthy ones
	for _, proxy := range proxies {
		if s.hasCapacity(proxy) && s.isProxyAvailableForAssignment(proxy) {
			s.repo.IncrementAccountCount(proxy.ID)
			return proxy, nil
		}
	}

	return nil, fmt.Errorf("all proxies at capacity for provider %s", providerID)
}

// IsProxyAvailableForRequest checks if account's proxy is available for making requests
// Returns true if proxy is usable, false if should skip this account
func (s *ProxyService) IsProxyAvailableForRequest(proxyID int) bool {
	proxy, err := s.repo.GetByID(proxyID)
	if err != nil {
		return false
	}

	if !proxy.IsActive {
		return false
	}

	// Healthy or degraded proxies are available
	if proxy.HealthStatus != models.HealthStatusDown {
		return true
	}

	// Down proxies: check if recovery delay has passed
	return s.hasRecoveryDelayPassed(proxy)
}

// isProxyAvailableForAssignment checks if proxy can be assigned to new accounts
func (s *ProxyService) isProxyAvailableForAssignment(proxy *models.Proxy) bool {
	if !proxy.IsActive {
		return false
	}

	// Only assign to healthy proxies for new accounts
	if proxy.HealthStatus == models.HealthStatusDown {
		return false
	}

	return true
}

// hasRecoveryDelayPassed checks if enough time has passed since proxy was marked down
func (s *ProxyService) hasRecoveryDelayPassed(proxy *models.Proxy) bool {
	if proxy.MarkedDownAt == nil {
		return true
	}
	return time.Since(*proxy.MarkedDownAt) >= s.downRecoveryDelay
}

// MarkProxyDown marks a proxy as down with timestamp
func (s *ProxyService) MarkProxyDown(proxyID int) error {
	now := time.Now()
	return s.repo.UpdateHealthWithDownTime(proxyID, models.HealthStatusDown, &now)
}

// MarkProxyHealthy marks a proxy as healthy and clears down timestamp
func (s *ProxyService) MarkProxyHealthy(proxyID int, latencyMs int) error {
	return s.repo.UpdateHealthWithDownTime(proxyID, models.HealthStatusHealthy, nil)
}

// ReleaseProxyAssignment decrements account count when account creation fails
func (s *ProxyService) ReleaseProxyAssignment(proxyID int) error {
	return s.repo.DecrementAccountCount(proxyID)
}
