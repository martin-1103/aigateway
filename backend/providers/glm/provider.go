package glm

import (
	"context"
	"fmt"

	"aigateway-backend/providers"
)

// Provider implements the providers.Provider interface for Zhipu AI (GLM)
type Provider struct{}

// NewProvider creates a new GLM provider instance
func NewProvider() *Provider {
	return &Provider{}
}

// ID returns the unique identifier for the GLM provider
func (p *Provider) ID() string {
	return ProviderID
}

// Name returns the human-readable name of the provider
func (p *Provider) Name() string {
	return "Zhipu AI (GLM)"
}

// AuthStrategy returns the authentication strategy identifier
func (p *Provider) AuthStrategy() string {
	return AuthType
}

// SupportedModels returns the list of models supported by GLM
func (p *Provider) SupportedModels() []string {
	return SupportedModels
}

// TranslateRequest converts a request from the specified format to GLM format
// Supports conversion from Claude format to GLM OpenAI-compatible format
func (p *Provider) TranslateRequest(format string, payload []byte, model string) ([]byte, error) {
	switch format {
	case "claude", "anthropic":
		return TranslateClaudeToGLM(payload, model), nil
	case "openai":
		// GLM uses OpenAI-compatible format, minimal translation needed
		return TranslateOpenAIToGLM(payload, model), nil
	default:
		return nil, fmt.Errorf("unsupported input format: %s", format)
	}
}

// TranslateResponse converts GLM response format to Claude format
// GLM returns OpenAI-compatible responses that need to be converted to Claude format
func (p *Provider) TranslateResponse(payload []byte) ([]byte, error) {
	return TranslateGLMToClaude(payload), nil
}

// Execute performs the actual API call to GLM
// Orchestrates the flow: validation -> HTTP request -> response handling
func (p *Provider) Execute(ctx context.Context, req *providers.ExecuteRequest) (*providers.ExecuteResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("execute request cannot be nil")
	}

	if req.Account == nil {
		return nil, fmt.Errorf("account cannot be nil")
	}

	if req.Payload == nil {
		return nil, fmt.Errorf("payload cannot be nil")
	}

	// Validate model is supported
	if !p.isModelSupported(req.Model) {
		return nil, fmt.Errorf("unsupported model: %s", req.Model)
	}

	// Execute HTTP request to GLM API
	resp, err := executeHTTP(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("http execution failed: %w", err)
	}

	return resp, nil
}

// ExecuteStream performs a streaming API call to GLM
func (p *Provider) ExecuteStream(ctx context.Context, req *providers.ExecuteRequest) (*providers.StreamResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("execute request cannot be nil")
	}

	if req.Account == nil {
		return nil, fmt.Errorf("account cannot be nil")
	}

	if req.Payload == nil {
		return nil, fmt.Errorf("payload cannot be nil")
	}

	// Validate model is supported
	if !p.isModelSupported(req.Model) {
		return nil, fmt.Errorf("unsupported model: %s", req.Model)
	}

	// Execute streaming HTTP request to GLM API
	return executeHTTPStream(ctx, req)
}

// SupportsStreaming indicates that GLM supports streaming
func (p *Provider) SupportsStreaming() bool {
	return true
}

// isModelSupported checks if the given model is in the supported models list
func (p *Provider) isModelSupported(model string) bool {
	for _, supported := range SupportedModels {
		if supported == model {
			return true
		}
	}
	return false
}

// Config constants
const (
	// ProviderID is the unique identifier for GLM provider
	ProviderID = "glm"

	// AuthType defines the authentication method (API key)
	AuthType = "api_key"

	// BaseURL is the GLM API base URL
	BaseURL = "https://open.bigmodel.cn/api/paas/v4"

	// EndpointChatCompletions is the chat completions endpoint
	EndpointChatCompletions = "/chat/completions"

	// ContentType is the HTTP Content-Type header value
	ContentType = "application/json"
)

// SupportedModels returns the list of models supported by GLM
var SupportedModels = []string{
	"glm-4",
	"glm-4-flash",
	"glm-4-air",
	"glm-4-plus",
	"glm-4-0520",
	"glm-4-airx",
}
