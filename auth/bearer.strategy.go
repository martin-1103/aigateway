package auth

import (
	"aigateway/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// BearerStrategy implements Bearer token authentication with provider-specific refresh
type BearerStrategy struct {
	redis      *redis.Client
	db         *gorm.DB
	httpClient *http.Client
}

// BearerTokenCache represents cached bearer token data
type BearerTokenCache struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// BearerTokenResponse represents provider-specific token response
type BearerTokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
}

// NewBearerStrategy creates a new bearer token authentication strategy
func NewBearerStrategy(redis *redis.Client, db *gorm.DB) *BearerStrategy {
	return &BearerStrategy{
		redis: redis,
		db:    db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Name returns the strategy identifier
func (s *BearerStrategy) Name() string {
	return "bearer"
}

// GetToken retrieves a valid bearer token from cache or auth data
func (s *BearerStrategy) GetToken(ctx context.Context, authData map[string]interface{}) (string, error) {
	accountID, ok := authData["account_id"].(string)
	if !ok {
		return "", fmt.Errorf("account_id not found in auth data")
	}
	providerID, ok := authData["provider_id"].(string)
	if !ok {
		return "", fmt.Errorf("provider_id not found in auth data")
	}

	cacheKey := s.getCacheKey(providerID, accountID)
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var tokenCache BearerTokenCache
		if err := json.Unmarshal([]byte(cached), &tokenCache); err == nil {
			if time.Now().Add(5 * time.Minute).Before(tokenCache.ExpiresAt) {
				return tokenCache.Token, nil
			}
		}
	}

	token, ok := authData["token"].(string)
	if !ok || token == "" {
		token, ok = authData["access_token"].(string)
		if !ok || token == "" {
			return "", fmt.Errorf("token not found in auth data")
		}
	}

	expiresAt := s.parseExpiration(authData)
	if time.Now().Add(5 * time.Minute).After(expiresAt) {
		return "", fmt.Errorf("token expired, refresh required")
	}

	s.cacheToken(ctx, cacheKey, token, expiresAt)
	return token, nil
}

// RefreshToken refreshes a bearer token using provider-specific endpoint
func (s *BearerStrategy) RefreshToken(ctx context.Context, authData map[string]interface{}, oldToken string) (string, error) {
	accountID, _ := authData["account_id"].(string)
	providerID, _ := authData["provider_id"].(string)
	refreshURL, ok := authData["refresh_url"].(string)
	if !ok || refreshURL == "" {
		return "", fmt.Errorf("refresh_url not found")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", refreshURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create refresh request: %w", err)
	}

	if apiKey, ok := authData["api_key"].(string); ok && apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
	if authHeader, ok := authData["auth_header"].(string); ok && authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read refresh response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp BearerTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	s.updateCacheAndDB(ctx, providerID, accountID, authData, &tokenResp, expiresAt)
	return tokenResp.Token, nil
}

// ValidateToken checks if a bearer token is valid
func (s *BearerStrategy) ValidateToken(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("token is empty")
	}
	return true, nil
}

func (s *BearerStrategy) getCacheKey(providerID, accountID string) string {
	return fmt.Sprintf("auth:bearer:%s:%s", providerID, accountID)
}

func (s *BearerStrategy) parseExpiration(authData map[string]interface{}) time.Time {
	expiresAt := time.Now().Add(3600 * time.Second)
	if expiresAtStr, ok := authData["expires_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, expiresAtStr); err == nil {
			expiresAt = t
		}
	} else if expiresIn, ok := authData["expires_in"].(float64); ok {
		expiresAt = time.Now().Add(time.Duration(expiresIn) * time.Second)
	}
	return expiresAt
}

func (s *BearerStrategy) cacheToken(ctx context.Context, cacheKey, token string, expiresAt time.Time) {
	tokenCache := BearerTokenCache{Token: token, ExpiresAt: expiresAt}
	if cacheData, err := json.Marshal(tokenCache); err == nil {
		if ttl := time.Until(expiresAt); ttl > 0 {
			s.redis.Set(ctx, cacheKey, cacheData, ttl)
		}
	}
}

func (s *BearerStrategy) updateCacheAndDB(ctx context.Context, providerID, accountID string, authData map[string]interface{}, tokenResp *BearerTokenResponse, expiresAt time.Time) {
	tokenCache := BearerTokenCache{Token: tokenResp.Token, ExpiresAt: expiresAt}
	cacheKey := s.getCacheKey(providerID, accountID)
	if cacheData, err := json.Marshal(tokenCache); err == nil {
		s.redis.Set(ctx, cacheKey, cacheData, time.Duration(tokenResp.ExpiresIn)*time.Second)
	}

	updatedAuthData := make(map[string]interface{})
	for k, v := range authData {
		updatedAuthData[k] = v
	}
	updatedAuthData["token"] = tokenResp.Token
	updatedAuthData["expires_at"] = expiresAt.Format(time.RFC3339)
	updatedAuthData["expires_in"] = tokenResp.ExpiresIn

	if authDataJSON, err := json.Marshal(updatedAuthData); err == nil {
		s.db.Model(&models.Account{}).Where("id = ?", accountID).Update("auth_data", string(authDataJSON))
	}
}

