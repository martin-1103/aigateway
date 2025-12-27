package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"aigateway-backend/auth/manager"
	"aigateway-backend/models"
)

// Refresher implements token refresh for Claude/Anthropic OAuth
type Refresher struct {
	httpClient  *http.Client
	refreshLead time.Duration
}

// NewRefresher creates a new Claude token refresher
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

// Refresh refreshes the OAuth token for a Claude account
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

	if tokenResp.Account != nil {
		result.Metadata["email"] = tokenResp.Account.EmailAddress
	}

	return result, nil
}

func (r *Refresher) doRefresh(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	reqBody := TokenRequest{
		ClientID:     OAuthClientID,
		GrantType:    "refresh_token",
		RefreshToken: refreshToken,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", OAuthTokenURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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
