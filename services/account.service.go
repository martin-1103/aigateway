package services

import (
	"aigateway/models"
	"aigateway/repositories"
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type AccountService struct {
	repo  *repositories.AccountRepository
	redis *redis.Client
}

func NewAccountService(repo *repositories.AccountRepository, redis *redis.Client) *AccountService {
	return &AccountService{
		repo:  repo,
		redis: redis,
	}
}

func (s *AccountService) SelectAccount(providerID, model string) (*models.Account, error) {
	key := fmt.Sprintf("account:rr:%s:%s", providerID, model)
	ctx := context.Background()

	accounts, err := s.repo.GetActiveByProvider(providerID)
	if err != nil || len(accounts) == 0 {
		return nil, fmt.Errorf("no available accounts for provider %s", providerID)
	}

	idx, err := s.redis.Incr(ctx, key).Result()
	if err != nil {
		idx = 1
	}

	selected := accounts[(idx-1)%int64(len(accounts))]

	go s.repo.UpdateLastUsed(selected.ID)

	return selected, nil
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
