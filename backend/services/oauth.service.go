package services

import (
	"aigateway-backend/models"
	"aigateway-backend/providers/antigravity"
	"aigateway-backend/repositories"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenCache struct {
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type OAuthService struct {
	redis             *redis.Client
	repo              *repositories.AccountRepository
	httpClientService *HTTPClientService
}

func NewOAuthService(redis *redis.Client, repo *repositories.AccountRepository, httpClientService *HTTPClientService) *OAuthService {
	return &OAuthService{
		redis:             redis,
		repo:              repo,
		httpClientService: httpClientService,
	}
}

func (s *OAuthService) GetAccessToken(account *models.Account) (string, error) {
	cacheKey := fmt.Sprintf("auth:%s:%s", account.ProviderID, account.ID)
	ctx := context.Background()

	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && cached != "" {
		var token TokenCache
		if json.Unmarshal([]byte(cached), &token) == nil {
			// Use UTC for consistent timezone comparison
			if time.Now().UTC().Add(antigravity.RefreshSkew).Before(token.ExpiresAt.UTC()) {
				return token.AccessToken, nil
			}
		}
	}

	var authData map[string]interface{}
	if err := json.Unmarshal([]byte(account.AuthData), &authData); err != nil {
		return "", err
	}

	accessToken, _ := authData["access_token"].(string)
	// Default expiry: 1 hour from now in UTC
	expiresAt := time.Now().UTC().Add(3600 * time.Second)

	if exp, ok := authData["expires_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, exp); err == nil {
			// Convert to UTC for consistent comparison
			expiresAt = t.UTC()
		}
	}

	// Use UTC for consistent timezone comparison
	if time.Now().UTC().Add(antigravity.RefreshSkew).After(expiresAt) {
		refreshToken, ok := authData["refresh_token"].(string)
		if !ok || refreshToken == "" {
			return "", fmt.Errorf("token expired and no refresh token available")
		}

		newAccessToken, newExpiresAt, err := s.refreshToken(account.ProviderID, refreshToken, account.ProxyURL)
		if err != nil {
			return "", fmt.Errorf("token refresh failed: %w", err)
		}

		authData["access_token"] = newAccessToken
		authData["expires_at"] = newExpiresAt.Format(time.RFC3339)

		updatedAuth, _ := json.Marshal(authData)
		account.AuthData = string(updatedAuth)
		if err := s.repo.UpdateAuthData(account.ID, string(updatedAuth)); err != nil {
			return "", fmt.Errorf("failed to save refreshed token: %w", err)
		}

		accessToken = newAccessToken
		expiresAt = newExpiresAt

		s.redis.Del(ctx, cacheKey)
	}

	if accessToken == "" {
		return "", fmt.Errorf("no access token available")
	}

	tokenCache := TokenCache{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
	}
	cacheData, _ := json.Marshal(tokenCache)
	s.redis.Set(ctx, cacheKey, cacheData, time.Until(expiresAt))

	return accessToken, nil
}

func (s *OAuthService) refreshToken(providerID string, refreshToken string, proxyURL string) (string, time.Time, error) {
	var clientID, clientSecret, tokenURL string

	switch providerID {
	case "antigravity":
		clientID = antigravity.OAuthClientID
		clientSecret = antigravity.OAuthClientSecret
		tokenURL = antigravity.OAuthTokenURL
	default:
		return "", time.Time{}, fmt.Errorf("token refresh not supported for provider: %s", providerID)
	}

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpClient := s.httpClientService.GetClient(proxyURL)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, err
	}

	if resp.StatusCode != 200 {
		return "", time.Time{}, fmt.Errorf("token refresh failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", time.Time{}, err
	}

	// Always use UTC for consistent timezone handling
	expiresAt := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return tokenResp.AccessToken, expiresAt, nil
}

func (s *OAuthService) InvalidateCache(account *models.Account) error {
	cacheKey := fmt.Sprintf("auth:%s:%s", account.ProviderID, account.ID)
	ctx := context.Background()
	return s.redis.Del(ctx, cacheKey).Err()
}
