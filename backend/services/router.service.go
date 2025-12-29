package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"aigateway-backend/auth/manager"
	"aigateway-backend/models"
	"aigateway-backend/providers"
	"aigateway-backend/repositories"
)

// Request represents a unified request structure for the router
type Request struct {
	ProviderID string
	Model      string
	Payload    []byte
	Stream     bool
	AccountID  string // Optional: override account selection for testing
}

// Response represents a unified response structure from the router
type Response struct {
	StatusCode int
	Payload    []byte
}

// RouterConfig holds configuration for the router
type RouterConfig struct {
	UseAuthManager bool
	MaxRetries     int
	MaxRetryWait   time.Duration
}

// DefaultRouterConfig returns default configuration
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		UseAuthManager: false,
		MaxRetries:     3,
		MaxRetryWait:   30 * time.Second,
	}
}

// RouterService handles model-to-provider routing and orchestrates the execution pipeline
type RouterService struct {
	registry            *providers.Registry
	providerRepo        *repositories.ProviderRepository
	accountService      *AccountService
	accountRepo         *repositories.AccountRepository
	proxyService        *ProxyService
	oauthService        *OAuthService
	statsTrackerService *StatsTrackerService

	// Auth manager for health-aware selection
	authManager *manager.Manager
	config      RouterConfig
}

// NewRouterService creates a new router service instance
func NewRouterService(
	registry *providers.Registry,
	providerRepo *repositories.ProviderRepository,
	accountService *AccountService,
	accountRepo *repositories.AccountRepository,
	proxyService *ProxyService,
	oauthService *OAuthService,
	statsTrackerService *StatsTrackerService,
) *RouterService {
	return &RouterService{
		registry:            registry,
		providerRepo:        providerRepo,
		accountService:      accountService,
		accountRepo:         accountRepo,
		proxyService:        proxyService,
		oauthService:        oauthService,
		statsTrackerService: statsTrackerService,
		config:              DefaultRouterConfig(),
	}
}

// SetAuthManager sets the auth manager for health-aware selection
func (s *RouterService) SetAuthManager(m *manager.Manager) {
	s.authManager = m
}

// SetConfig sets the router configuration
func (s *RouterService) SetConfig(config RouterConfig) {
	s.config = config
}

// EnableAuthManager enables the auth manager for account selection
func (s *RouterService) EnableAuthManager(enabled bool) {
	s.config.UseAuthManager = enabled
}

// Route determines the appropriate provider for a given model
func (s *RouterService) Route(model string) (providers.Provider, string, error) {
	provider, resolvedModel, err := s.registry.GetByModel(model)
	if err != nil {
		return nil, "", fmt.Errorf("failed to route model %s: %w", model, err)
	}
	return provider, resolvedModel, nil
}

// Execute orchestrates the complete request pipeline with optional retry
func (s *RouterService) Execute(ctx context.Context, req Request) (Response, error) {
	if s.config.UseAuthManager && s.authManager != nil {
		return s.executeWithAuthManager(ctx, req, 0)
	}
	return s.executeLegacy(ctx, req)
}

// selectAccount selects account using configured method
func (s *RouterService) selectAccount(ctx context.Context, providerID, model string) (*models.Account, *manager.AccountState, error) {
	if s.config.UseAuthManager && s.authManager != nil {
		accState, err := s.authManager.Select(ctx, providerID, model)
		if err != nil {
			return nil, nil, err
		}
		return accState.Account, accState, nil
	}

	account, err := s.accountService.SelectAccount(providerID, model)
	if err != nil {
		return nil, nil, err
	}
	return account, nil, nil
}

// executeLegacy is the original execution path without AuthManager
func (s *RouterService) executeLegacy(ctx context.Context, req Request) (Response, error) {
	provider, resolvedModel, err := s.Route(req.Model)
	if err != nil {
		return Response{}, err
	}

	providerID := provider.ID()

	account, err := s.accountService.SelectAccount(providerID, resolvedModel)
	if err != nil {
		return Response{}, fmt.Errorf("failed to select account: %w", err)
	}

	return s.executeWithAccount(ctx, provider, account, resolvedModel, req)
}

// executeWithAccount executes request with given account using permanent proxy
func (s *RouterService) executeWithAccount(
	ctx context.Context,
	provider providers.Provider,
	account *models.Account,
	resolvedModel string,
	req Request,
) (Response, error) {
	providerID := provider.ID()

	// Use account's permanent proxy (no dynamic assignment)
	token, err := s.oauthService.GetAccessToken(account)
	if err != nil {
		return Response{}, fmt.Errorf("failed to get access token: %w", err)
	}

	executeReq := &providers.ExecuteRequest{
		Model:    resolvedModel,
		Payload:  req.Payload,
		Stream:   req.Stream,
		Account:  account,
		ProxyURL: account.ProxyURL,
		Token:    token,
	}

	executeResp, err := provider.Execute(ctx, executeReq)
	if err != nil {
		s.statsTrackerService.RecordFailure(&account.ID, account.ProxyID, 0, err)
		// Track health failure
		go s.accountRepo.UpdateHealthFailure(account.ID, err.Error())
		return Response{}, fmt.Errorf("provider execution failed: %w", err)
	}

	statusCode := executeResp.StatusCode
	providerIDPtr := &providerID

	go s.statsTrackerService.RecordRequest(
		&account.ID,
		account.ProxyID,
		providerIDPtr,
		resolvedModel,
		statusCode,
		executeResp.LatencyMs,
	)

	if statusCode < 200 || statusCode >= 300 {
		// Track health failure for non-2xx
		go s.accountRepo.UpdateHealthFailure(account.ID, fmt.Sprintf("HTTP %d", statusCode))
		return Response{
			StatusCode: statusCode,
			Payload:    executeResp.Payload,
		}, fmt.Errorf("upstream error: %d", statusCode)
	}

	// Track health success
	go s.accountRepo.UpdateHealthSuccess(account.ID)

	return Response{
		StatusCode: statusCode,
		Payload:    executeResp.Payload,
	}, nil
}

// ListProviders returns all active providers from database
func (s *RouterService) ListProviders() []ProviderInfo {
	dbProviders, err := s.providerRepo.ListActive()
	if err != nil {
		log.Printf("error fetching providers from database: %v", err)
		return []ProviderInfo{}
	}

	result := make([]ProviderInfo, 0, len(dbProviders))
	for _, p := range dbProviders {
		result = append(result, ProviderInfo{
			ID:                 p.ID,
			Name:               p.Name,
			BaseURL:            p.BaseURL,
			SupportedAuthTypes: p.SupportedAuthTypes,
			IsActive:           p.IsActive,
		})
	}

	return result
}

// ProviderInfo represents provider information for API responses
type ProviderInfo struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	BaseURL            string   `json:"base_url"`
	SupportedAuthTypes []string `json:"supported_auth_types"`
	IsActive           bool     `json:"is_active"`
}
