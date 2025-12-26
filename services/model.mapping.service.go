package services

import (
	"aigateway/models"
	"aigateway/providers"
	"aigateway/repositories"
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

const modelMappingKeyPrefix = "model:mapping:"

type ModelMappingService struct {
	repo  *repositories.ModelMappingRepository
	redis *redis.Client
}

// cachedMapping is the Redis cache format
type cachedMapping struct {
	ProviderID string `json:"provider_id"`
	ModelName  string `json:"model_name"`
}

func NewModelMappingService(repo *repositories.ModelMappingRepository, redis *redis.Client) *ModelMappingService {
	return &ModelMappingService{
		repo:  repo,
		redis: redis,
	}
}

// Resolve implements providers.MappingResolver interface
func (s *ModelMappingService) Resolve(ctx context.Context, alias string) *providers.ResolvedMapping {
	key := modelMappingKeyPrefix + alias

	// Check Redis cache
	cached, err := s.redis.Get(ctx, key).Result()
	if err == nil {
		var cm cachedMapping
		if json.Unmarshal([]byte(cached), &cm) == nil {
			return &providers.ResolvedMapping{
				ProviderID: cm.ProviderID,
				ModelName:  cm.ModelName,
			}
		}
	}

	// Cache miss - query DB
	mapping, err := s.repo.GetByAlias(alias)
	if err != nil {
		return nil
	}

	// Cache result (no expiry - invalidated on write)
	s.cacheMapping(ctx, alias, &cachedMapping{
		ProviderID: mapping.ProviderID,
		ModelName:  mapping.ModelName,
	})

	return &providers.ResolvedMapping{
		ProviderID: mapping.ProviderID,
		ModelName:  mapping.ModelName,
	}
}

func (s *ModelMappingService) Create(ctx context.Context, mapping *models.ModelMapping) error {
	if err := s.repo.Create(mapping); err != nil {
		return err
	}
	return s.cacheMapping(ctx, mapping.Alias, &cachedMapping{
		ProviderID: mapping.ProviderID,
		ModelName:  mapping.ModelName,
	})
}

func (s *ModelMappingService) GetByAlias(alias string) (*models.ModelMapping, error) {
	return s.repo.GetByAlias(alias)
}

func (s *ModelMappingService) List(limit, offset int) ([]*models.ModelMapping, int64, error) {
	return s.repo.List(limit, offset)
}

func (s *ModelMappingService) Update(ctx context.Context, oldAlias string, mapping *models.ModelMapping) error {
	if err := s.repo.Update(oldAlias, mapping); err != nil {
		return err
	}

	// Invalidate old key if alias changed
	if oldAlias != mapping.Alias {
		s.redis.Del(ctx, modelMappingKeyPrefix+oldAlias)
	}

	// Cache new mapping
	return s.cacheMapping(ctx, mapping.Alias, &cachedMapping{
		ProviderID: mapping.ProviderID,
		ModelName:  mapping.ModelName,
	})
}

func (s *ModelMappingService) Delete(ctx context.Context, alias string) error {
	if err := s.repo.Delete(alias); err != nil {
		return err
	}
	return s.redis.Del(ctx, modelMappingKeyPrefix+alias).Err()
}

func (s *ModelMappingService) cacheMapping(ctx context.Context, alias string, resolved *cachedMapping) error {
	key := modelMappingKeyPrefix + alias
	val, err := json.Marshal(resolved)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}
	return s.redis.Set(ctx, key, val, 0).Err() // 0 = no expiry
}
