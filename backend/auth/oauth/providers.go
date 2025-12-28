package oauth

import (
	"aigateway-backend/auth/pkce"
	"aigateway-backend/providers/antigravity"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// Antigravity OAuth config (Google Cloud Code)
	AntigravityAuthURL  = "https://accounts.google.com/o/oauth2/v2/auth"
	AntigravityTokenURL = antigravity.OAuthTokenURL
	AntigravityClientID = antigravity.OAuthClientID
	AntigravitySecret   = antigravity.OAuthClientSecret
	AntigravityScope    = "https://www.googleapis.com/auth/cloud-platform https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile https://www.googleapis.com/auth/cclog https://www.googleapis.com/auth/experimentsandconfigs"
	GoogleUserInfoURL   = "https://www.googleapis.com/oauth2/v1/userinfo?alt=json"

	// OpenAI Codex OAuth config
	CodexAuthURL  = "https://auth.openai.com/oauth/authorize"
	CodexTokenURL = "https://auth.openai.com/oauth/token"
	CodexClientID = "app_EMoamEEZ73f0CkXaXp7hrann"
	CodexScope    = "openid email profile offline_access"

	// Claude OAuth config
	ClaudeAuthURL  = "https://claude.ai/oauth/authorize"
	ClaudeTokenURL = "https://console.anthropic.com/v1/oauth/token"
	ClaudeClientID = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"
	ClaudeScope    = "org:create_api_key user:profile user:inference"
)

// ProviderOAuth defines OAuth configuration for a provider
type ProviderOAuth struct {
	ProviderID   string
	Name         string
	AuthURL      string
	TokenURL     string
	ClientID     string
	ClientSecret string
	Scope        string
	RedirectURI  string
	httpClient   *http.Client
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// NewAntigravityOAuth creates OAuth config for Antigravity provider
func NewAntigravityOAuth(redirectURI string) *ProviderOAuth {
	return &ProviderOAuth{
		ProviderID:   "antigravity",
		Name:         "Google Cloud Code (Antigravity)",
		AuthURL:      AntigravityAuthURL,
		TokenURL:     AntigravityTokenURL,
		ClientID:     AntigravityClientID,
		ClientSecret: AntigravitySecret,
		Scope:        AntigravityScope,
		RedirectURI:  redirectURI,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// NewCodexOAuth creates OAuth config for OpenAI Codex provider
func NewCodexOAuth(redirectURI string) *ProviderOAuth {
	return &ProviderOAuth{
		ProviderID:  "codex",
		Name:        "OpenAI Codex",
		AuthURL:     CodexAuthURL,
		TokenURL:    CodexTokenURL,
		ClientID:    CodexClientID,
		Scope:       CodexScope,
		RedirectURI: redirectURI,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// NewClaudeOAuth creates OAuth config for Claude provider
func NewClaudeOAuth(redirectURI string) *ProviderOAuth {
	return &ProviderOAuth{
		ProviderID:  "claude",
		Name:        "Anthropic Claude",
		AuthURL:     ClaudeAuthURL,
		TokenURL:    ClaudeTokenURL,
		ClientID:    ClaudeClientID,
		Scope:       ClaudeScope,
		RedirectURI: redirectURI,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// BuildAuthURL constructs the OAuth authorization URL with PKCE
func (p *ProviderOAuth) BuildAuthURL(state string, pkceCodes *pkce.PKCECodes) (string, error) {
	if pkceCodes == nil {
		return "", fmt.Errorf("PKCE codes are required")
	}

	params := url.Values{
		"client_id":             {p.ClientID},
		"response_type":         {"code"},
		"redirect_uri":          {p.RedirectURI},
		"scope":                 {p.Scope},
		"state":                 {state},
		"code_challenge":        {pkceCodes.CodeChallenge},
		"code_challenge_method": {"S256"},
		"access_type":           {"offline"},
		"prompt":                {"consent"},
	}

	// Provider-specific params
	switch p.ProviderID {
	case "codex":
		params.Set("id_token_add_organizations", "true")
		params.Set("codex_cli_simplified_flow", "true")
	case "claude":
		params.Set("code", "true")
	}

	return fmt.Sprintf("%s?%s", p.AuthURL, params.Encode()), nil
}

// ExchangeCode exchanges authorization code for access token
func (p *ProviderOAuth) ExchangeCode(ctx context.Context, code string, pkceCodes *pkce.PKCECodes) (*TokenResponse, error) {
	if pkceCodes == nil {
		return nil, fmt.Errorf("PKCE codes are required")
	}

	var body io.Reader

	// Claude uses JSON, others use form
	if p.ProviderID == "claude" {
		reqBody := map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     p.ClientID,
			"code":          code,
			"redirect_uri":  p.RedirectURI,
			"code_verifier": pkceCodes.CodeVerifier,
		}
		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}
		body = strings.NewReader(string(jsonBody))
	} else {
		data := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {code},
			"redirect_uri":  {p.RedirectURI},
			"code_verifier": {pkceCodes.CodeVerifier},
			"client_id":     {p.ClientID},
		}
		if p.ClientSecret != "" {
			data.Set("client_secret", p.ClientSecret)
		}
		body = strings.NewReader(data.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.TokenURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if p.ProviderID == "claude" {
		req.Header.Set("Content-Type", "application/json")
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo fetches user info from Google's userinfo endpoint
func (p *ProviderOAuth) GetUserInfo(ctx context.Context, accessToken string) (map[string]interface{}, error) {
	if p.ProviderID != "antigravity" {
		return nil, fmt.Errorf("userinfo endpoint only supported for antigravity provider")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", GoogleUserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read userinfo response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var userInfo map[string]interface{}
	if err := json.Unmarshal(respBody, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse userinfo response: %w", err)
	}

	return userInfo, nil
}

// GetProviderOAuth returns OAuth config for a provider ID
func GetProviderOAuth(providerID, redirectURI string) (*ProviderOAuth, error) {
	switch providerID {
	case "antigravity":
		return NewAntigravityOAuth(redirectURI), nil
	case "codex":
		return NewCodexOAuth(redirectURI), nil
	case "claude":
		return NewClaudeOAuth(redirectURI), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerID)
	}
}

// ListProviders returns all available OAuth providers
func ListProviders(redirectURI string) []*ProviderOAuth {
	return []*ProviderOAuth{
		NewAntigravityOAuth(redirectURI),
		NewCodexOAuth(redirectURI),
		NewClaudeOAuth(redirectURI),
	}
}
