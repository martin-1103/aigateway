package providers

import (
	"context"

	"aigateway/models"
)

// Provider defines the interface that all AI provider implementations must satisfy.
// Each provider translates between standard formats and provider-specific APIs.
type Provider interface {
	// ID returns the unique identifier for this provider (e.g., "openai", "anthropic")
	ID() string

	// Name returns the human-readable name of the provider
	Name() string

	// AuthStrategy returns the authentication strategy identifier for this provider
	// (e.g., "api_key", "oauth", "bearer_token")
	AuthStrategy() string

	// SupportedModels returns the list of model identifiers supported by this provider
	SupportedModels() []string

	// TranslateRequest converts a request from the specified format to the provider's format
	// format: the input format (e.g., "openai", "anthropic")
	// payload: the request payload in the input format
	// model: the target model identifier
	// Returns the translated payload or an error
	TranslateRequest(format string, payload []byte, model string) ([]byte, error)

	// TranslateResponse converts the provider's response format to a standard format
	// payload: the response payload from the provider
	// Returns the translated payload or an error
	TranslateResponse(payload []byte) ([]byte, error)

	// Execute performs the actual API call to the provider
	// ctx: context for cancellation and timeout control
	// req: the execution request with all necessary parameters
	// Returns the execution response or an error
	Execute(ctx context.Context, req *ExecuteRequest) (*ExecuteResponse, error)
}

// ExecuteRequest contains all parameters needed to execute a provider API call
type ExecuteRequest struct {
	// Model is the model identifier to use for the request
	Model string

	// Payload is the request body in the provider's expected format
	Payload []byte

	// Stream indicates whether to use streaming response mode
	Stream bool

	// Account contains the authentication credentials and metadata
	Account *models.Account

	// ProxyURL is the optional proxy server URL to use for the request
	ProxyURL string

	// Token is the authentication token (may be pre-fetched or from Account.AuthData)
	Token string
}

// ExecuteResponse contains the result of a provider API call
type ExecuteResponse struct {
	// StatusCode is the HTTP status code from the provider
	StatusCode int

	// Payload is the response body from the provider
	Payload []byte

	// LatencyMs is the request latency in milliseconds
	LatencyMs int
}
