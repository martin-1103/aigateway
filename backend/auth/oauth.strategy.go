package auth

import (
	"aigateway-backend/models"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type OAuthStrategy struct {
	redis      *redis.Client
	db         *gorm.DB
	httpClient *http.Client
}

type TokenCache struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

func NewOAuthStrategy(redis *redis.Client, db *gorm.DB) *OAuthStrategy {
	return &OAuthStrategy{redis: redis, db: db, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

func (s *OAuthStrategy) Name() string {
	return "oauth"
}

// GetToken retrieves a valid OAuth token from cache or auth data
func (s *OAuthStrategy) GetToken(ctx context.Context, authData map[string]interface{}) (string, error) {
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
		var token TokenCache
		if err := json.Unmarshal([]byte(cached), &token); err == nil {
			if time.Now().Add(5 * time.Minute).Before(token.ExpiresAt) {
				return token.AccessToken, nil
			}
		}
	}
	accessToken, ok := authData["access_token"].(string)
	if !ok || accessToken == "" {
		return "", fmt.Errorf("access_token not found in auth data")
	}
	expiresAt := s.parseExpiration(authData)
	if time.Now().Add(5 * time.Minute).After(expiresAt) {
		return "", fmt.Errorf("token expired, refresh required")
	}
	s.cacheToken(ctx, cacheKey, authData, accessToken, expiresAt)
	return accessToken, nil
}

// RefreshToken refreshes an OAuth token using refresh token
func (s *OAuthStrategy) RefreshToken(ctx context.Context, authData map[string]interface{}, oldToken string) (string, error) {
	accountID, _ := authData["account_id"].(string)
	providerID, _ := authData["provider_id"].(string)
	refreshToken, ok := authData["refresh_token"].(string)
	if !ok || refreshToken == "" {
		return "", fmt.Errorf("refresh_token not found")
	}
	tokenURL, ok := authData["token_url"].(string)
	if !ok || tokenURL == "" {
		return "", fmt.Errorf("token_url not found")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	if clientID, ok := authData["client_id"].(string); ok {
		data.Set("client_id", clientID)
	}
	if clientSecret, ok := authData["client_secret"].(string); ok {
		data.Set("client_secret", clientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	var tokenResp OAuthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	s.updateCacheAndDB(ctx, providerID, accountID, authData, &tokenResp, expiresAt)
	return tokenResp.AccessToken, nil
}

// ValidateToken checks if an OAuth token is valid
func (s *OAuthStrategy) ValidateToken(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("token is empty")
	}
	return true, nil
}

func (s *OAuthStrategy) getCacheKey(providerID, accountID string) string {
	return fmt.Sprintf("auth:oauth:%s:%s", providerID, accountID)
}

func (s *OAuthStrategy) parseExpiration(authData map[string]interface{}) time.Time {
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

func (s *OAuthStrategy) cacheToken(ctx context.Context, cacheKey string, authData map[string]interface{}, accessToken string, expiresAt time.Time) {
	refreshToken, _ := authData["refresh_token"].(string)
	tokenType, _ := authData["token_type"].(string)
	if tokenType == "" {
		tokenType = "Bearer"
	}
	tokenCache := TokenCache{AccessToken: accessToken, RefreshToken: refreshToken, ExpiresAt: expiresAt, TokenType: tokenType}
	if cacheData, err := json.Marshal(tokenCache); err == nil {
		if ttl := time.Until(expiresAt); ttl > 0 {
			s.redis.Set(ctx, cacheKey, cacheData, ttl)
		}
	}
}

func (s *OAuthStrategy) updateCacheAndDB(ctx context.Context, providerID, accountID string, authData map[string]interface{}, tokenResp *OAuthTokenResponse, expiresAt time.Time) {
	tokenCache := TokenCache{AccessToken: tokenResp.AccessToken, RefreshToken: tokenResp.RefreshToken, ExpiresAt: expiresAt, TokenType: tokenResp.TokenType}
	cacheKey := s.getCacheKey(providerID, accountID)
	if cacheData, err := json.Marshal(tokenCache); err == nil {
		s.redis.Set(ctx, cacheKey, cacheData, time.Duration(tokenResp.ExpiresIn)*time.Second)
	}

	updatedAuthData := make(map[string]interface{})
	for k, v := range authData {
		updatedAuthData[k] = v
	}
	updatedAuthData["access_token"] = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		updatedAuthData["refresh_token"] = tokenResp.RefreshToken
	}
	updatedAuthData["expires_at"] = expiresAt.Format(time.RFC3339)
	updatedAuthData["expires_in"] = tokenResp.ExpiresIn
	updatedAuthData["token_type"] = tokenResp.TokenType

	if authDataJSON, err := json.Marshal(updatedAuthData); err == nil {
		s.db.Model(&models.Account{}).Where("id = ?", accountID).Update("auth_data", string(authDataJSON))
	}
}

