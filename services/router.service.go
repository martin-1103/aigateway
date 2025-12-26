package services

import (
	"context"
	"fmt"

	"aigateway/models"
	"aigateway/providers"
)

// Request represents a unified request structure for the router
type Request struct {
	ProviderID string
	Model      string
	Payload    []byte
	Stream     bool
}

// Response represents a unified response structure from the router
type Response struct {
	StatusCode int
	Payload    []byte
}

// RouterService handles model-to-provider routing and orchestrates the execution pipeline
type RouterService struct {
	registry          *providers.Registry
	accountService    *AccountService
	proxyService      *ProxyService
	oauthService      *OAuthService
	statsTrackerService *StatsTrackerService
}

// NewRouterService creates a new router service instance
func NewRouterService(
	registry *providers.Registry,
	accountService *AccountService,
	proxyService *ProxyService,
	oauthService *OAuthService,
	statsTrackerService *StatsTrackerService,
) *RouterService {
	return &RouterService{
		registry:          registry,
		accountService:    accountService,
		proxyService:      proxyService,
		oauthService:      oauthService,
		statsTrackerService: statsTrackerService,
	}
}

// Route determines the appropriate provider for a given model
func (s *RouterService) Route(model string) (providers.Provider, error) {
	providerModel, err := s.registry.GetByModel(model)
	if err != nil {
		return nil, fmt.Errorf("failed to route model %s: %w", model, err)
	}

	return providerModel.Implementation, nil
}

// Execute orchestrates the complete request pipeline: route → account select → proxy assign → auth → provider execute → stats
func (s *RouterService) Execute(ctx context.Context, req Request) (Response, error) {
	// Step 1: Route to appropriate provider
	provider, err := s.Route(req.Model)
	if err != nil {
		return Response{}, err
	}

	providerID := provider.ID()

	// Step 2: Select account using round-robin
	account, err := s.accountService.SelectAccount(providerID, req.Model)
	if err != nil {
		return Response{}, fmt.Errorf("failed to select account: %w", err)
	}

	// Step 3: Assign proxy to account
	if err := s.proxyService.AssignProxy(account, providerID); err != nil {
		return Response{}, fmt.Errorf("failed to assign proxy: %w", err)
	}

	// Step 4: Get authentication token
	token, err := s.oauthService.GetAccessToken(account)
	if err != nil {
		return Response{}, fmt.Errorf("failed to get access token: %w", err)
	}

	// Step 5: Execute provider request
	executeReq := &providers.ExecuteRequest{
		Model:    req.Model,
		Payload:  req.Payload,
		Stream:   req.Stream,
		Account:  account,
		ProxyURL: account.ProxyURL,
		Token:    token,
	}

	executeResp, err := provider.Execute(ctx, executeReq)
	if err != nil {
		// Record failure in stats
		s.statsTrackerService.RecordFailure(&account.ID, account.ProxyID, 0, err)
		return Response{}, fmt.Errorf("provider execution failed: %w", err)
	}

	// Step 6: Record success stats
	statusCode := executeResp.StatusCode
	latencyMs := executeResp.LatencyMs

	providerIDPtr := &providerID
	go s.statsTrackerService.RecordRequest(
		&account.ID,
		account.ProxyID,
		providerIDPtr,
		req.Model,
		statusCode,
		latencyMs,
	)

	// Check if request was successful
	if statusCode < 200 || statusCode >= 300 {
		return Response{
			StatusCode: statusCode,
			Payload:    executeResp.Payload,
		}, fmt.Errorf("upstream error: %d", statusCode)
	}

	return Response{
		StatusCode: statusCode,
		Payload:    executeResp.Payload,
	}, nil
}

// ListProviders returns all registered providers
func (s *RouterService) ListProviders() []ProviderInfo {
	providers := s.registry.ListActive()
	result := make([]ProviderInfo, 0, len(providers))

	for _, p := range providers {
		result = append(result, ProviderInfo{
			ID:       p.ID,
			Name:     p.Name,
			BaseURL:  p.BaseURL,
			AuthType: string(p.AuthType),
			IsActive: p.IsActive,
		})
	}

	return result
}

// ProviderInfo represents provider information for API responses
type ProviderInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	BaseURL  string `json:"base_url"`
	AuthType string `json:"auth_type"`
	IsActive bool   `json:"is_active"`
}
