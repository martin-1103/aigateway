package services

import (
	"aigateway/models"
	"aigateway/repositories"
	"sync"
)

// ProxyService handles proxy assignment and management operations
type ProxyService struct {
	repo        *repositories.ProxyRepository
	accountRepo *repositories.AccountRepository
	mu          sync.RWMutex
}

// NewProxyService creates a new proxy service instance
func NewProxyService(repo *repositories.ProxyRepository, accountRepo *repositories.AccountRepository) *ProxyService {
	return &ProxyService{
		repo:        repo,
		accountRepo: accountRepo,
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
