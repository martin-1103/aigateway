// services/auth.apikey.service.go
package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"aigateway-backend/models"
	"aigateway-backend/repositories"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type APIKeyService struct {
	repo  *repositories.APIKeyRepository
	redis *redis.Client
}

func NewAPIKeyService(repo *repositories.APIKeyRepository, redis *redis.Client) *APIKeyService {
	return &APIKeyService{repo: repo, redis: redis}
}

func (s *APIKeyService) Generate(userID, label string) (*models.APIKey, string, error) {
	rawKey := s.generateRawKey()
	hash := s.hashKey(rawKey)
	prefix := rawKey[:12]

	apiKey := &models.APIKey{
		ID:        uuid.New().String(),
		UserID:    userID,
		KeyHash:   hash,
		KeyPrefix: prefix,
		Label:     label,
		IsActive:  true,
	}

	if err := s.repo.Create(apiKey); err != nil {
		return nil, "", err
	}

	return apiKey, rawKey, nil
}

func (s *APIKeyService) Validate(rawKey string) (*models.APIKey, error) {
	hash := s.hashKey(rawKey)

	// Check cache first
	ctx := context.Background()
	cacheKey := fmt.Sprintf("apikey:%s", hash)

	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var key models.APIKey
		if json.Unmarshal([]byte(cached), &key) == nil {
			go s.repo.UpdateLastUsed(key.ID)
			return &key, nil
		}
	}

	// Lookup in DB
	key, err := s.repo.GetByHash(hash)
	if err != nil {
		return nil, err
	}

	// Cache for 5 minutes
	data, _ := json.Marshal(key)
	s.redis.Set(ctx, cacheKey, data, 5*time.Minute)

	go s.repo.UpdateLastUsed(key.ID)

	return key, nil
}

func (s *APIKeyService) ListByUser(userID string) ([]*models.APIKey, error) {
	return s.repo.ListByUserID(userID)
}

func (s *APIKeyService) ListAll(limit, offset int) ([]*models.APIKey, int64, error) {
	return s.repo.ListAll(limit, offset)
}

func (s *APIKeyService) Revoke(id string) error {
	return s.repo.Revoke(id)
}

func (s *APIKeyService) GetByID(id string) (*models.APIKey, error) {
	return s.repo.GetByID(id)
}

func (s *APIKeyService) generateRawKey() string {
	bytes := make([]byte, 24)
	rand.Read(bytes)
	return "ak_" + hex.EncodeToString(bytes)
}

func (s *APIKeyService) hashKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}
