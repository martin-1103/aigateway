package antigravity

import (
	"aigateway/providers"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// AntigravityProvider implements the Provider interface for Antigravity (Google Cloud Code) API
type AntigravityProvider struct {
	httpClients map[string]*http.Client
	clientMu    sync.RWMutex
	executor    *Executor
}

// NewAntigravityProvider creates a new Antigravity provider instance
func NewAntigravityProvider() *AntigravityProvider {
	return &AntigravityProvider{
		httpClients: make(map[string]*http.Client),
		executor:    NewExecutor(),
	}
}

// ID returns the provider identifier
func (p *AntigravityProvider) ID() string {
	return ProviderID
}

// Name returns the human-readable provider name
func (p *AntigravityProvider) Name() string {
	return "Antigravity"
}

// AuthStrategy returns the authentication strategy
func (p *AntigravityProvider) AuthStrategy() string {
	return AuthType
}

// SupportedModels returns the list of supported models
func (p *AntigravityProvider) SupportedModels() []string {
	return SupportedModels
}

// TranslateRequest converts a request from Claude format to Antigravity format
func (p *AntigravityProvider) TranslateRequest(format string, payload []byte, model string) ([]byte, error) {
	// Currently only supporting Claude format
	if format != "claude" && format != "anthropic" {
		return nil, fmt.Errorf("unsupported input format: %s", format)
	}

	// Validate model is supported
	supported := false
	for _, m := range SupportedModels {
		if m == model {
			supported = true
			break
		}
	}
	if !supported {
		return nil, fmt.Errorf("unsupported model: %s", model)
	}

	translated := TranslateClaudeToAntigravity(payload, model)
	return translated, nil
}

// TranslateResponse converts Antigravity response to Claude format
func (p *AntigravityProvider) TranslateResponse(payload []byte) ([]byte, error) {
	translated := TranslateAntigravityToClaude(payload)
	return translated, nil
}

// Execute performs the API call to Antigravity
func (p *AntigravityProvider) Execute(ctx context.Context, req *providers.ExecuteRequest) (*providers.ExecuteResponse, error) {
	if req.Account == nil {
		return nil, fmt.Errorf("account is required")
	}

	// Extract access token
	accessToken := req.Token
	if accessToken == "" {
		// Try to extract from account auth data
		var authData map[string]interface{}
		if err := json.Unmarshal([]byte(req.Account.AuthData), &authData); err != nil {
			return nil, fmt.Errorf("failed to parse auth data: %w", err)
		}

		token, ok := authData["access_token"].(string)
		if !ok || token == "" {
			return nil, fmt.Errorf("no access token found in account")
		}
		accessToken = token
	}

	// Get or create HTTP client for this proxy
	httpClient := p.getHTTPClient(req.ProxyURL)

	// Create executor request
	execReq := &ExecuteRequest{
		Model:       req.Model,
		Payload:     req.Payload,
		Stream:      req.Stream,
		AccessToken: accessToken,
		HTTPClient:  httpClient,
	}

	// Execute the request
	execResp, err := p.executor.Execute(ctx, execReq)
	if err != nil {
		return nil, err
	}

	// Convert to provider response format
	return &providers.ExecuteResponse{
		StatusCode: execResp.StatusCode,
		Payload:    execResp.Body,
		LatencyMs:  int(execResp.Latency),
	}, nil
}

// getHTTPClient retrieves or creates an HTTP client for the given proxy URL
func (p *AntigravityProvider) getHTTPClient(proxyURL string) *http.Client {
	cacheKey := proxyURL
	if cacheKey == "" {
		cacheKey = "default"
	}

	// Try read lock first
	p.clientMu.RLock()
	if client, exists := p.httpClients[cacheKey]; exists {
		p.clientMu.RUnlock()
		return client
	}
	p.clientMu.RUnlock()

	// Create new client with write lock
	p.clientMu.Lock()
	defer p.clientMu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := p.httpClients[cacheKey]; exists {
		return client
	}

	// Create new HTTP client
	client := &http.Client{
		Timeout: 120 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	// TODO: Add proxy support if proxyURL is provided
	// This would require parsing the proxyURL and configuring the transport

	p.httpClients[cacheKey] = client
	return client
}
