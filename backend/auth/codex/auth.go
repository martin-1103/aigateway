package codex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"aigateway-backend/auth/manager"
	"aigateway-backend/models"
)

// Refresher implements token refresh for Codex/OpenAI OAuth
type Refresher struct {
	httpClient  *http.Client
	refreshLead time.Duration
}

// NewRefresher creates a new Codex token refresher
func NewRefresher() *Refresher {
	return &Refresher{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		refreshLead: RefreshLeadDefault,
	}
}

// RefreshLead returns how long before expiry to start refresh
func (r *Refresher) RefreshLead() time.Duration {
	return r.refreshLead
}

// Refresh refreshes the OAuth token for a Codex account
func (r *Refresher) Refresh(ctx context.Context, account *models.Account) (*manager.TokenResult, error) {
	refreshToken, err := extractRefreshToken(account.AuthData)
	if err != nil {
		return nil, err
	}

	tokenResp, err := r.doRefresh(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	result := &manager.TokenResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresAt:    expiresAt,
		Metadata:     make(map[string]interface{}),
	}

	// Parse id_token to get account info
	if tokenResp.IDToken != "" {
		claims, err := ParseIDToken(tokenResp.IDToken)
		if err == nil {
			result.Metadata["account_id"] = claims.Sub
			result.Metadata["email"] = claims.Email
			result.Metadata["name"] = claims.Name
		}
	}

	return result, nil
}

func (r *Refresher) doRefresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// Codex uses form-urlencoded, not JSON
	data := url.Values{}
	data.Set("client_id", OAuthClientID)
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("scope", OAuthScope)

	req, err := http.NewRequestWithContext(ctx, "POST", OAuthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("refresh failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &tokenResp, nil
}

func extractRefreshToken(authData string) (string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(authData), &data); err != nil {
		return "", fmt.Errorf("failed to parse auth data: %w", err)
	}

	refreshToken, ok := data["refresh_token"].(string)
	if !ok || refreshToken == "" {
		return "", fmt.Errorf("no refresh token in auth data")
	}

	return refreshToken, nil
}
