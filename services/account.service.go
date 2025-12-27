package services

import (
	"aigateway/models"
	"aigateway/repositories"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type AccountService struct {
	repo     *repositories.AccountRepository
	redis    *redis.Client
	proxySvc *ProxyService
}

func NewAccountService(repo *repositories.AccountRepository, redis *redis.Client) *AccountService {
	return &AccountService{
		repo:  repo,
		redis: redis,
	}
}

// SetProxyService sets the proxy service (to avoid circular dependency)
func (s *AccountService) SetProxyService(proxySvc *ProxyService) {
	s.proxySvc = proxySvc
}

func (s *AccountService) SelectAccount(providerID, model string) (*models.Account, error) {
	key := fmt.Sprintf("account:rr:%s:%s", providerID, model)
	ctx := context.Background()

	accounts, err := s.repo.GetActiveByProvider(providerID)
	if err != nil || len(accounts) == 0 {
		return nil, fmt.Errorf("no available accounts for provider %s", providerID)
	}

	// Filter accounts with available proxies
	availableAccounts := s.filterAvailableAccounts(accounts)
	if len(availableAccounts) == 0 {
		return nil, fmt.Errorf("no accounts with available proxies for provider %s", providerID)
	}

	idx, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		idx = 1
	}

	selected := availableAccounts[(idx-1)%int64(len(availableAccounts))]

	go s.repo.UpdateLastUsed(selected.ID)

	return selected, nil
}

// SelectAccountExcluding selects an account excluding the specified account ID
// Used for fallback when retry fails
func (s *AccountService) SelectAccountExcluding(providerID, model, excludeAccountID string) (*models.Account, error) {
	key := fmt.Sprintf("account:rr:%s:%s", providerID, model)
	ctx := context.Background()

	accounts, err := s.repo.GetActiveByProvider(providerID)
	if err != nil || len(accounts) == 0 {
		return nil, fmt.Errorf("no available accounts for provider %s", providerID)
	}

	// Filter accounts with available proxies, excluding the specified account
	var availableAccounts []*models.Account
	for _, acc := range accounts {
		if acc.ID == excludeAccountID {
			continue
		}
		if s.isAccountProxyAvailable(acc) {
			availableAccounts = append(availableAccounts, acc)
		}
	}

	if len(availableAccounts) == 0 {
		return nil, fmt.Errorf("no alternative accounts with available proxies for provider %s", providerID)
	}

	idx, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		idx = 1
	}

	selected := availableAccounts[(idx-1)%int64(len(availableAccounts))]

	go s.repo.UpdateLastUsed(selected.ID)

	return selected, nil
}

// filterAvailableAccounts filters accounts whose proxy is available
func (s *AccountService) filterAvailableAccounts(accounts []*models.Account) []*models.Account {
	var available []*models.Account
	for _, acc := range accounts {
		if s.isAccountProxyAvailable(acc) {
			available = append(available, acc)
		}
	}
	return available
}

// isAccountProxyAvailable checks if account's proxy is available for requests
func (s *AccountService) isAccountProxyAvailable(acc *models.Account) bool {
	// Account without proxy is available (legacy or direct connection)
	if acc.ProxyID == nil {
		return true
	}

	// Check proxy availability via ProxyService
	if s.proxySvc != nil {
		return s.proxySvc.IsProxyAvailableForRequest(*acc.ProxyID)
	}

	return true
}

func (s *AccountService) Create(account *models.Account) error {
	return s.repo.Create(account)
}

func (s *AccountService) GetByID(id string) (*models.Account, error) {
	return s.repo.GetByID(id)
}

func (s *AccountService) List(limit, offset int) ([]*models.Account, int64, error) {
	return s.repo.List(limit, offset)
}

func (s *AccountService) Update(account *models.Account) error {
	return s.repo.Update(account)
}

func (s *AccountService) Delete(id string) error {
	return s.repo.Delete(id)
}

func (s *AccountService) ListByCreator(creatorID string, limit, offset int) ([]*models.Account, int64, error) {
	return s.repo.ListByCreator(creatorID, limit, offset)
}
