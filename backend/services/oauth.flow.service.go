package services

import (
	"aigateway-backend/auth/manager"
	"aigateway-backend/auth/oauth"
	"aigateway-backend/auth/pkce"
	"aigateway-backend/models"
	"aigateway-backend/repositories"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// OAuth session TTL in Redis
	OAuthSessionTTL = 10 * time.Minute

	// Default redirect URI
	DefaultRedirectURI = "http://172.235.254.157:8088/api/v1/oauth/callback"
)

// OAuthFlowService handles OAuth authorization flow
type OAuthFlowService struct {
	redis       *redis.Client
	accountSvc  *AccountService
	repo        *repositories.AccountRepository
	proxySvc    *ProxyService
	authManager *manager.Manager
}

// OAuthSession represents an OAuth flow session stored in Redis
type OAuthSession struct {
	Provider     string    `json:"provider"`
	ProjectID    string    `json:"project_id"`
	FlowType     string    `json:"flow_type"`
	RedirectURI  string    `json:"redirect_uri"`
	CodeVerifier string    `json:"code_verifier"`
	CreatedAt    time.Time `json:"created_at"`
	CreatedBy    *string   `json:"created_by,omitempty"`
}

// InitFlowRequest represents OAuth init request
type InitFlowRequest struct {
	Provider    string  `json:"provider" binding:"required"`
	ProjectID   string  `json:"project_id" binding:"required"`
	FlowType    string  `json:"flow_type" binding:"required"`
	RedirectURI string  `json:"redirect_uri"`
	CreatedBy   *string `json:"created_by,omitempty"`
}

// InitFlowResponse represents OAuth init response
type InitFlowResponse struct {
	AuthURL   string    `json:"auth_url"`
	State     string    `json:"state"`
	FlowType  string    `json:"flow_type"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ExchangeRequest represents OAuth exchange request
type ExchangeRequest struct {
	CallbackURL string `json:"callback_url" binding:"required"`
}

// ExchangeResponse represents OAuth exchange response
type ExchangeResponse struct {
	Success bool            `json:"success"`
	Account *models.Account `json:"account"`
}

// OAuthProviderInfo represents OAuth provider info
type OAuthProviderInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// NewOAuthFlowService creates a new OAuth flow service
func NewOAuthFlowService(redis *redis.Client, accountSvc *AccountService, repo *repositories.AccountRepository, proxySvc *ProxyService) *OAuthFlowService {
	return &OAuthFlowService{
		redis:      redis,
		accountSvc: accountSvc,
		repo:       repo,
		proxySvc:   proxySvc,
	}
}

// SetAuthManager sets the auth manager for hot-reload
func (s *OAuthFlowService) SetAuthManager(m *manager.Manager) {
	s.authManager = m
}

// InitFlow starts OAuth authorization flow
func (s *OAuthFlowService) InitFlow(ctx context.Context, req *InitFlowRequest) (*InitFlowResponse, error) {
	if req.FlowType != "auto" && req.FlowType != "manual" {
		return nil, fmt.Errorf("invalid flow_type: must be 'auto' or 'manual'")
	}

	redirectURI := req.RedirectURI
	if redirectURI == "" {
		redirectURI = DefaultRedirectURI
	}

	providerOAuth, err := oauth.GetProviderOAuth(req.Provider, redirectURI)
	if err != nil {
		return nil, err
	}

	pkceCodes, err := pkce.GeneratePKCECodes()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE codes: %w", err)
	}

	state := uuid.New().String()

	authURL, err := providerOAuth.BuildAuthURL(state, pkceCodes)
	if err != nil {
		return nil, fmt.Errorf("failed to build auth URL: %w", err)
	}

	session := OAuthSession{
		Provider:     req.Provider,
		ProjectID:    req.ProjectID,
		FlowType:     req.FlowType,
		RedirectURI:  redirectURI,
		CodeVerifier: pkceCodes.CodeVerifier,
		CreatedAt:    time.Now(),
		CreatedBy:    req.CreatedBy,
	}

	sessionJSON, err := json.Marshal(session)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal session: %w", err)
	}

	sessionKey := fmt.Sprintf("oauth:session:%s", state)
	if err := s.redis.Set(ctx, sessionKey, sessionJSON, OAuthSessionTTL).Err(); err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return &InitFlowResponse{
		AuthURL:   authURL,
		State:     state,
		FlowType:  req.FlowType,
		ExpiresAt: time.Now().Add(OAuthSessionTTL),
	}, nil
}


// ExchangeCode exchanges authorization code from callback URL
func (s *OAuthFlowService) ExchangeCode(ctx context.Context, callbackURL string) (*ExchangeResponse, error) {
	parsedURL, err := url.Parse(callbackURL)
	if err != nil {
		return nil, fmt.Errorf("invalid callback URL: %w", err)
	}

	query := parsedURL.Query()
	code := query.Get("code")
	state := query.Get("state")

	if code == "" || state == "" {
		return nil, fmt.Errorf("missing code or state parameter")
	}

	sessionKey := fmt.Sprintf("oauth:session:%s", state)
	sessionJSON, err := s.redis.Get(ctx, sessionKey).Result()
	if err != nil {
		return nil, fmt.Errorf("session not found or expired")
	}

	var session OAuthSession
	if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	providerOAuth, err := oauth.GetProviderOAuth(session.Provider, session.RedirectURI)
	if err != nil {
		return nil, err
	}

	pkceCodes := &pkce.PKCECodes{
		CodeVerifier: session.CodeVerifier,
	}

	tokenResp, err := providerOAuth.ExchangeCode(ctx, code, pkceCodes)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	authData := map[string]interface{}{
		"access_token":  tokenResp.AccessToken,
		"refresh_token": tokenResp.RefreshToken,
		"token_type":    tokenResp.TokenType,
		"expires_at":    expiresAt.Format(time.RFC3339),
		"expires_in":    tokenResp.ExpiresIn,
	}

	authDataJSON, err := json.Marshal(authData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth data: %w", err)
	}

	metadata := map[string]interface{}{
		"project_id": session.ProjectID,
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Fetch user info from Google's userinfo endpoint to get email
	userInfo, err := providerOAuth.GetUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	email, ok := userInfo["email"].(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("email not found in user info")
	}

	account := &models.Account{
		ID:         uuid.New().String(),
		ProviderID: session.Provider,
		Label:      email,
		AuthData:   string(authDataJSON),
		Metadata:   string(metadataJSON),
		IsActive:   true,
		ExpiresAt:  &expiresAt,
		CreatedBy:  session.CreatedBy,
	}

	// Assign proxy permanently during registration
	if s.proxySvc != nil {
		proxy, err := s.proxySvc.SelectProxyForNewAccount(session.Provider)
		if err != nil {
			// Log warning but don't fail - account can work without proxy
			// Admin should add more proxies if this happens frequently
		} else {
			account.ProxyID = &proxy.ID
			account.ProxyURL = proxy.URL
		}
	}

	if err := s.repo.Create(account); err != nil {
		// Rollback proxy assignment if account creation fails
		if account.ProxyID != nil {
			s.proxySvc.ReleaseProxyAssignment(*account.ProxyID)
		}
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	// Hot-reload: Add to AuthManager immediately (non-blocking)
	if s.authManager != nil {
		s.authManager.AddAccount(account)
		log.Printf("[OAuth] Hot-reload: Added account %s to AuthManager", account.ID)
	}

	s.redis.Del(ctx, sessionKey)

	return &ExchangeResponse{
		Success: true,
		Account: account,
	}, nil
}

// GetProviders returns list of available OAuth providers
func (s *OAuthFlowService) GetProviders() []OAuthProviderInfo {
	providers := oauth.ListProviders(DefaultRedirectURI)
	result := make([]OAuthProviderInfo, len(providers))

	for i, p := range providers {
		result[i] = OAuthProviderInfo{
			ID:   p.ProviderID,
			Name: p.Name,
		}
	}

	return result
}

// RefreshToken manually refreshes an account's OAuth token
func (s *OAuthFlowService) RefreshToken(ctx context.Context, accountID string) error {
	account, err := s.accountSvc.GetByID(accountID)
	if err != nil {
		return fmt.Errorf("account not found: %w", err)
	}

	var authData map[string]interface{}
	if err := json.Unmarshal([]byte(account.AuthData), &authData); err != nil {
		return fmt.Errorf("invalid auth data: %w", err)
	}

	refreshToken, ok := authData["refresh_token"].(string)
	if !ok || refreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	providerOAuth, err := oauth.GetProviderOAuth(account.ProviderID, DefaultRedirectURI)
	if err != nil {
		return err
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {providerOAuth.ClientID},
	}
	if providerOAuth.ClientSecret != "" {
		data.Set("client_secret", providerOAuth.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", providerOAuth.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("refresh failed with status %d", resp.StatusCode)
	}

	var tokenResp oauth.TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	authData["access_token"] = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		authData["refresh_token"] = tokenResp.RefreshToken
	}
	authData["expires_at"] = expiresAt.Format(time.RFC3339)
	authData["expires_in"] = tokenResp.ExpiresIn

	updatedAuthData, err := json.Marshal(authData)
	if err != nil {
		return fmt.Errorf("failed to marshal auth data: %w", err)
	}

	return s.repo.UpdateAuthDataWithExpiry(accountID, string(updatedAuthData), expiresAt)
}
