package codex

import "time"

// OAuth configuration for Codex/OpenAI Auth
const (
	OAuthTokenURL = "https://auth.openai.com/oauth/token"
	OAuthClientID = "app_EMoamEEZ73f0CkXaXp7hrann"
	OAuthScope    = "openid profile email"

	// RefreshLeadDefault is how long before expiry to refresh token
	RefreshLeadDefault = 5 * time.Minute
)

// TokenResponse is the OAuth token refresh response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
}

// IDTokenClaims contains parsed claims from the id_token JWT
type IDTokenClaims struct {
	Sub          string `json:"sub"`           // Account ID
	Email        string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name         string `json:"name"`
	Picture      string `json:"picture"`
	Iat          int64  `json:"iat"`           // Issued at
	Exp          int64  `json:"exp"`           // Expires at
	Iss          string `json:"iss"`           // Issuer
	Aud          string `json:"aud"`           // Audience (client_id)
}
