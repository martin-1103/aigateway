package auth

import (
	"context"
	"fmt"
)

// APIKeyStrategy implements static API key authentication
// API keys are long-lived credentials that don't require refresh
type APIKeyStrategy struct{}

// NewAPIKeyStrategy creates a new API key authentication strategy
func NewAPIKeyStrategy() *APIKeyStrategy {
	return &APIKeyStrategy{}
}

// Name returns the strategy identifier
func (s *APIKeyStrategy) Name() string {
	return "api_key"
}

// GetToken extracts the API key from account auth data
// For API key authentication, the token is the API key itself stored in auth_data JSON
func (s *APIKeyStrategy) GetToken(ctx context.Context, authData map[string]interface{}) (string, error) {
	// Try common API key field names
	keyFields := []string{"api_key", "apiKey", "key", "token", "access_token"}

	for _, field := range keyFields {
		if apiKey, ok := authData[field].(string); ok && apiKey != "" {
			return apiKey, nil
		}
	}

	return "", fmt.Errorf("API key not found in auth data (looked for: %v)", keyFields)
}

// RefreshToken is a no-op for API keys
// API keys are static credentials that don't expire or require refresh
func (s *APIKeyStrategy) RefreshToken(ctx context.Context, authData map[string]interface{}, oldToken string) (string, error) {
	// API keys don't need refresh, return the same key
	return s.GetToken(ctx, authData)
}

// ValidateToken validates an API key
// API keys are considered always valid (they don't expire)
// Actual validation happens when the key is used with the provider
func (s *APIKeyStrategy) ValidateToken(ctx context.Context, token string) (bool, error) {
	if token == "" {
		return false, fmt.Errorf("API key is empty")
	}

	// API keys don't have built-in expiration
	// Validation happens at the provider level when the key is used
	return true, nil
}
