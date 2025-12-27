package services

import (
	"aigateway-backend/models"
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	modelsAvailableKey = "models:available"
	modelsCacheTTL     = 5 * time.Minute
)

type ModelsService struct {
	db    *gorm.DB
	redis *redis.Client
}

type ProviderModels struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Models []string `json:"models"`
}

type ModelsResponse struct {
	Providers []ProviderModels `json:"providers"`
}

func NewModelsService(db *gorm.DB, redis *redis.Client) *ModelsService {
	return &ModelsService{
		db:    db,
		redis: redis,
	}
}

func (s *ModelsService) GetAvailableModels(ctx context.Context) (*ModelsResponse, error) {
	// Check Redis cache
	cached, err := s.redis.Get(ctx, modelsAvailableKey).Result()
	if err == nil {
		var response ModelsResponse
		if json.Unmarshal([]byte(cached), &response) == nil {
			return &response, nil
		}
	}

	// Cache miss - query database
	var providers []models.Provider
	if err := s.db.Where("is_active = ?", true).Find(&providers).Error; err != nil {
		return nil, err
	}

	// Build response
	response := &ModelsResponse{
		Providers: make([]ProviderModels, 0, len(providers)),
	}

	for _, p := range providers {
		var modelsList []string
		if p.SupportedModels != "" {
			json.Unmarshal([]byte(p.SupportedModels), &modelsList)
		}

		response.Providers = append(response.Providers, ProviderModels{
			ID:     p.ID,
			Name:   p.Name,
			Models: modelsList,
		})
	}

	// Cache result
	if data, err := json.Marshal(response); err == nil {
		s.redis.Set(ctx, modelsAvailableKey, data, modelsCacheTTL)
	}

	return response, nil
}

func (s *ModelsService) InvalidateCache(ctx context.Context) error {
	return s.redis.Del(ctx, modelsAvailableKey).Err()
}
