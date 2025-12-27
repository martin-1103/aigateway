package claude

import "time"

// OAuth configuration for Claude/Anthropic
const (
	OAuthTokenURL = "https://console.anthropic.com/v1/oauth/token"
	OAuthClientID = "9d1c250a-e61b-44d9-88ed-5944d1962f5e"

	// RefreshLeadDefault is how long before expiry to refresh token
	RefreshLeadDefault = 5 * time.Minute
)

// TokenRequest is the OAuth token refresh request body
type TokenRequest struct {
	ClientID     string `json:"client_id"`
	GrantType    string `json:"grant_type"`
	RefreshToken string `json:"refresh_token"`
}

// TokenResponse is the OAuth token refresh response
type TokenResponse struct {
	AccessToken  string         `json:"access_token"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    int            `json:"expires_in"`
	TokenType    string         `json:"token_type"`
	Scope        string         `json:"scope"`
	Account      *AccountInfo   `json:"account,omitempty"`
}

// AccountInfo contains account details from token response
type AccountInfo struct {
	EmailAddress string `json:"email_address"`
}
