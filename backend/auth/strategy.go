package auth

import (
	"context"
)

// Strategy defines the interface for authentication mechanisms used by providers.
// Different providers may use different authentication strategies (API keys, OAuth, etc.)
type Strategy interface {
	// Name returns the unique identifier for this authentication strategy
	// (e.g., "api_key", "oauth2", "bearer_token")
	Name() string

	// GetToken retrieves a valid authentication token
	// ctx: context for cancellation and timeout control
	// authData: provider-specific authentication data (e.g., API key, client credentials)
	// Returns the authentication token or an error
	GetToken(ctx context.Context, authData map[string]interface{}) (string, error)

	// RefreshToken refreshes an expired or expiring authentication token
	// ctx: context for cancellation and timeout control
	// authData: provider-specific authentication data
	// oldToken: the token to refresh
	// Returns the new authentication token or an error
	RefreshToken(ctx context.Context, authData map[string]interface{}, oldToken string) (string, error)

	// ValidateToken checks if a token is valid and not expired
	// ctx: context for cancellation and timeout control
	// token: the token to validate
	// Returns true if valid, false otherwise, or an error
	ValidateToken(ctx context.Context, token string) (bool, error)
}

// TokenResponse represents a standardized token response structure
type TokenResponse struct {
	// Token is the authentication token
	Token string

	// ExpiresIn is the token validity duration in seconds
	ExpiresIn int64

	// RefreshToken is an optional refresh token for token renewal
	RefreshToken string

	// TokenType is the type of token (e.g., "Bearer")
	TokenType string
}

// StrategyConfig contains common configuration for authentication strategies
type StrategyConfig struct {
	// Name is the strategy identifier
	Name string

	// CacheTTL is the token cache duration in seconds
	CacheTTL int64

	// RetryAttempts is the number of retry attempts for token operations
	RetryAttempts int

	// RetryDelayMs is the delay between retry attempts in milliseconds
	RetryDelayMs int
}
