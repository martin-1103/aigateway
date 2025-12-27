package services

import (
	"context"
	"fmt"

	"aigateway-backend/providers"
)

// ExecutorService orchestrates the complete execution pipeline for AI provider requests
type ExecutorService struct {
	routerService     *RouterService
	accountService    *AccountService
	proxyService      *ProxyService
	oauthService      *OAuthService
	statsTrackerService *StatsTrackerService
}

// NewExecutorService creates a new executor service instance
func NewExecutorService(
	routerService *RouterService,
	accountService *AccountService,
	proxyService *ProxyService,
	oauthService *OAuthService,
	statsTrackerService *StatsTrackerService,
) *ExecutorService {
	return &ExecutorService{
		routerService:     routerService,
		accountService:    accountService,
		proxyService:      proxyService,
		oauthService:      oauthService,
		statsTrackerService: statsTrackerService,
	}
}

// Execute processes a request through the complete pipeline: route → account → proxy → auth → execute → stats
func (s *ExecutorService) Execute(ctx context.Context, req Request) (Response, error) {
	// Step 1: Route to appropriate provider (may resolve alias to actual model)
	provider, resolvedModel, err := s.routerService.Route(req.Model)
	if err != nil {
		return Response{}, err
	}

	providerID := provider.ID()

	// Step 2: Select account using round-robin
	account, err := s.accountService.SelectAccount(providerID, resolvedModel)
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

	// Step 5: Execute provider request (use resolved model name)
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
		resolvedModel,
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

// ExecuteStream processes a streaming request through the complete pipeline
func (s *ExecutorService) ExecuteStream(ctx context.Context, req Request) (*providers.StreamResponse, error) {
	// Step 1: Route to appropriate provider (may resolve alias to actual model)
	provider, resolvedModel, err := s.routerService.Route(req.Model)
	if err != nil {
		return nil, err
	}

	// Check if provider supports streaming
	if !provider.SupportsStreaming() {
		return nil, fmt.Errorf("provider %s does not support streaming", provider.ID())
	}

	providerID := provider.ID()

	// Step 2: Select account using round-robin
	account, err := s.accountService.SelectAccount(providerID, resolvedModel)
	if err != nil {
		return nil, fmt.Errorf("failed to select account: %w", err)
	}

	// Step 3: Assign proxy to account
	if err := s.proxyService.AssignProxy(account, providerID); err != nil {
		return nil, fmt.Errorf("failed to assign proxy: %w", err)
	}

	// Step 4: Get authentication token
	token, err := s.oauthService.GetAccessToken(account)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Step 5: Execute provider streaming request (use resolved model name)
	executeReq := &providers.ExecuteRequest{
		Model:    resolvedModel,
		Payload:  req.Payload,
		Stream:   true,
		Account:  account,
		ProxyURL: account.ProxyURL,
		Token:    token,
	}

	streamResp, err := provider.ExecuteStream(ctx, executeReq)
	if err != nil {
		// Record failure in stats
		s.statsTrackerService.RecordFailure(&account.ID, account.ProxyID, 0, err)
		return nil, fmt.Errorf("provider streaming execution failed: %w", err)
	}

	// Step 6: Record success stats (asynchronously after stream completes)
	go func() {
		<-streamResp.Done
		providerIDPtr := &providerID
		s.statsTrackerService.RecordRequest(
			&account.ID,
			account.ProxyID,
			providerIDPtr,
			resolvedModel,
			streamResp.StatusCode,
			0,
		)
	}()

	return streamResp, nil
}
