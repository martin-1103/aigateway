package openai

import (
	"context"
	"encoding/json"
	"fmt"

	"aigateway/providers"
)

// OpenAIProvider implements the Provider interface for OpenAI API
type OpenAIProvider struct{}

// NewOpenAIProvider creates a new OpenAI provider instance
func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{}
}

// ID returns the unique identifier for OpenAI provider
func (p *OpenAIProvider) ID() string {
	return ProviderID
}

// Name returns the human-readable name
func (p *OpenAIProvider) Name() string {
	return "OpenAI"
}

// AuthStrategy returns the authentication strategy identifier
func (p *OpenAIProvider) AuthStrategy() string {
	return AuthType
}

// SupportedModels returns the list of supported model identifiers
func (p *OpenAIProvider) SupportedModels() []string {
	return SupportedModels
}

// TranslateRequest converts Claude format to OpenAI format
func (p *OpenAIProvider) TranslateRequest(format string, payload []byte, model string) ([]byte, error) {
	if format == "claude" || format == "anthropic" {
		return ClaudeToOpenAI(payload, model)
	}

	// If already in OpenAI format or unknown, pass through
	return payload, nil
}

// TranslateResponse converts OpenAI response to Claude format
func (p *OpenAIProvider) TranslateResponse(payload []byte) ([]byte, error) {
	return OpenAIToClaude(payload)
}

// Execute performs the API call to OpenAI
func (p *OpenAIProvider) Execute(ctx context.Context, req *providers.ExecuteRequest) (*providers.ExecuteResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("execute request cannot be nil")
	}

	if req.Account == nil {
		return nil, fmt.Errorf("account cannot be nil")
	}

	// Extract API key from account auth data
	var authData map[string]interface{}
	if err := json.Unmarshal([]byte(req.Account.AuthData), &authData); err != nil {
		return nil, fmt.Errorf("failed to parse auth data: %w", err)
	}

	apiKey, ok := authData["api_key"].(string)
	if !ok || apiKey == "" {
		// Fall back to Token field if api_key not in AuthData
		if req.Token != "" {
			apiKey = req.Token
		} else {
			return nil, fmt.Errorf("api_key not found in auth data")
		}
	}

	// Determine proxy URL
	proxyURL := req.ProxyURL
	if proxyURL == "" && req.Account.ProxyURL != "" {
		proxyURL = req.Account.ProxyURL
	}

	// Execute HTTP request
	return executeHTTP(ctx, &HTTPRequest{
		Model:    req.Model,
		Payload:  req.Payload,
		Stream:   req.Stream,
		APIKey:   apiKey,
		ProxyURL: proxyURL,
	})
}

// ExecuteStream performs a streaming API call to OpenAI
func (p *OpenAIProvider) ExecuteStream(ctx context.Context, req *providers.ExecuteRequest) (*providers.StreamResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("execute request cannot be nil")
	}

	if req.Account == nil {
		return nil, fmt.Errorf("account cannot be nil")
	}

	// Extract API key from account auth data
	var authData map[string]interface{}
	if err := json.Unmarshal([]byte(req.Account.AuthData), &authData); err != nil {
		return nil, fmt.Errorf("failed to parse auth data: %w", err)
	}

	apiKey, ok := authData["api_key"].(string)
	if !ok || apiKey == "" {
		if req.Token != "" {
			apiKey = req.Token
		} else {
			return nil, fmt.Errorf("api_key not found in auth data")
		}
	}

	// Determine proxy URL
	proxyURL := req.ProxyURL
	if proxyURL == "" && req.Account.ProxyURL != "" {
		proxyURL = req.Account.ProxyURL
	}

	// Execute streaming HTTP request
	return executeHTTPStream(ctx, &HTTPRequest{
		Model:    req.Model,
		Payload:  req.Payload,
		Stream:   true,
		APIKey:   apiKey,
		ProxyURL: proxyURL,
	})
}

// SupportsStreaming indicates that OpenAI supports streaming
func (p *OpenAIProvider) SupportsStreaming() bool {
	return true
}
