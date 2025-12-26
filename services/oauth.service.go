package services

import (
	"aigateway/models"
	"aigateway/repositories"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenCache struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type OAuthService struct {
	redis *redis.Client
	repo  *repositories.AccountRepository
}

func NewOAuthService(redis *redis.Client, repo *repositories.AccountRepository) *OAuthService {
	return &OAuthService{
		redis: redis,
		repo:  repo,
	}
}

func (s *OAuthService) GetAccessToken(account *models.Account) (string, error) {
	cacheKey := fmt.Sprintf("auth:%s:%s", account.ProviderID, account.ID)
	ctx := context.Background()

	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var token TokenCache
		if json.Unmarshal([]byte(cached), &token) == nil {
			if time.Now().Add(5 * time.Minute).Before(token.ExpiresAt) {
				return token.AccessToken, nil
			}
		}
	}

	var authData map[string]interface{}
	if err := json.Unmarshal([]byte(account.AuthData), &authData); err != nil {
		return "", err
	}

	accessToken, ok := authData["access_token"].(string)
	if !ok || accessToken == "" {
		return "", fmt.Errorf("no access token in auth data")
	}

	expiresAt := time.Now().Add(3600 * time.Second)
	if exp, ok := authData["expires_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, exp); err == nil {
			expiresAt = t
		}
	}

	if time.Now().Add(5 * time.Minute).After(expiresAt) {
		return "", fmt.Errorf("token expired, refresh needed")
	}

	tokenCache := TokenCache{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
	}
	cacheData, _ := json.Marshal(tokenCache)
	s.redis.Set(ctx, cacheKey, cacheData, time.Until(expiresAt))

	return accessToken, nil
}

func (s *OAuthService) InvalidateCache(account *models.Account) error {
	cacheKey := fmt.Sprintf("auth:%s:%s", account.ProviderID, account.ID)
	ctx := context.Background()
	return s.redis.Del(ctx, cacheKey).Err()
}
