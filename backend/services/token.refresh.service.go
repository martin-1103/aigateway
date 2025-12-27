package services

import (
	"aigateway-backend/models"
	"aigateway-backend/providers/antigravity"
	"aigateway-backend/repositories"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokenRefreshService struct {
	accountRepo   *repositories.AccountRepository
	redis         *redis.Client
	httpClient    *http.Client
	interval      time.Duration
	refreshBefore time.Duration
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewTokenRefreshService(
	accountRepo *repositories.AccountRepository,
	redisClient *redis.Client,
) *TokenRefreshService {
	ctx, cancel := context.WithCancel(context.Background())
	return &TokenRefreshService{
		accountRepo:   accountRepo,
		redis:         redisClient,
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		interval:      5 * time.Minute,
		refreshBefore: 10 * time.Minute,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (s *TokenRefreshService) Start() {
	log.Println("TokenRefreshService: Starting background token refresh job")
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			log.Println("TokenRefreshService: Shutting down")
			return
		case <-ticker.C:
			s.refreshExpiring()
		}
	}
}

func (s *TokenRefreshService) Stop() {
	s.cancel()
}

func (s *TokenRefreshService) refreshExpiring() {
	accounts, err := s.accountRepo.GetExpiringAccounts("antigravity", s.refreshBefore)
	if err != nil {
		log.Printf("TokenRefreshService: Failed to get expiring accounts: %v", err)
		return
	}

	if len(accounts) == 0 {
		return
	}

	log.Printf("TokenRefreshService: Found %d accounts needing refresh", len(accounts))

	for _, account := range accounts {
		if err := s.refreshAccount(account); err != nil {
			log.Printf("TokenRefreshService: Failed to refresh account %s: %v", account.ID, err)
			continue
		}
		log.Printf("TokenRefreshService: Refreshed token for account %s", account.ID)
	}
}

func (s *TokenRefreshService) refreshAccount(account *models.Account) error {
	var authData map[string]interface{}
	if err := json.Unmarshal([]byte(account.AuthData), &authData); err != nil {
		return fmt.Errorf("failed to parse auth data: %w", err)
	}

	refreshToken, ok := authData["refresh_token"].(string)
	if !ok || refreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// Retry with exponential backoff
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*attempt) * time.Second)
		}

		newToken, expiresAt, err := s.doRefresh(refreshToken)
		if err != nil {
			lastErr = err
			continue
		}

		// Update auth data
		authData["access_token"] = newToken
		authData["expires_at"] = expiresAt.Format(time.RFC3339)

		updatedAuth, _ := json.Marshal(authData)
		if err := s.accountRepo.UpdateAuthDataWithExpiry(account.ID, string(updatedAuth), expiresAt); err != nil {
			return fmt.Errorf("failed to update auth data: %w", err)
		}

		// Flush Redis cache
		s.flushCache(account.ProviderID, account.ID)

		return nil
	}

	return fmt.Errorf("failed after 3 attempts: %w", lastErr)
}

func (s *TokenRefreshService) doRefresh(refreshToken string) (string, time.Time, error) {
	data := url.Values{}
	data.Set("client_id", antigravity.OAuthClientID)
	data.Set("client_secret", antigravity.OAuthClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", antigravity.OAuthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, err
	}

	if resp.StatusCode != 200 {
		return "", time.Time{}, fmt.Errorf("refresh failed: %s", string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", time.Time{}, err
	}

	// Always use UTC for consistent timezone handling
	expiresAt := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	return tokenResp.AccessToken, expiresAt, nil
}

func (s *TokenRefreshService) flushCache(providerID, accountID string) {
	cacheKey := fmt.Sprintf("auth:%s:%s", providerID, accountID)
	s.redis.Del(s.ctx, cacheKey)
}
