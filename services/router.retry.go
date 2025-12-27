package services

import (
	"context"
	"fmt"
	"time"

	autherrors "aigateway/auth/errors"
	"aigateway/auth/manager"
	"aigateway/models"
	"aigateway/providers"
)

// executeWithAuthManager executes request with health-aware account selection and retry
func (s *RouterService) executeWithAuthManager(ctx context.Context, req Request, attempt int) (Response, error) {
	if attempt >= s.config.MaxRetries {
		return Response{}, fmt.Errorf("max retries (%d) exceeded", s.config.MaxRetries)
	}

	provider, resolvedModel, err := s.Route(req.Model)
	if err != nil {
		return Response{}, err
	}

	providerID := provider.ID()

	// Select account using AuthManager
	accState, err := s.authManager.Select(ctx, providerID, resolvedModel)
	if err != nil {
		if allBlocked, ok := err.(*manager.AllBlockedError); ok {
			return s.handleAllBlocked(ctx, req, attempt, allBlocked)
		}
		return Response{}, fmt.Errorf("failed to select account: %w", err)
	}

	// Execute and track result
	resp, statusCode, payload, execErr := s.executeAndTrack(ctx, provider, accState.Account, resolvedModel, req)

	// Mark result in AuthManager
	s.authManager.MarkResult(accState.Account.ID, resolvedModel, statusCode, payload)

	// Retry if needed
	if execErr != nil && s.shouldRetry(statusCode, attempt) {
		return s.executeWithAuthManager(ctx, req, attempt+1)
	}

	return resp, execErr
}

// handleAllBlocked handles the case when all accounts are blocked
func (s *RouterService) handleAllBlocked(
	ctx context.Context,
	req Request,
	attempt int,
	allBlocked *manager.AllBlockedError,
) (Response, error) {
	waitDur := time.Until(allBlocked.WaitDuration)
	if waitDur <= 0 {
		// Retry immediately
		return s.executeWithAuthManager(ctx, req, attempt+1)
	}

	if waitDur > s.config.MaxRetryWait {
		return Response{}, fmt.Errorf("all accounts blocked, wait time %v exceeds max %v", waitDur, s.config.MaxRetryWait)
	}

	// Wait and retry
	select {
	case <-ctx.Done():
		return Response{}, ctx.Err()
	case <-time.After(waitDur):
		return s.executeWithAuthManager(ctx, req, attempt+1)
	}
}

// executeAndTrack executes request and returns status code and payload for result tracking
func (s *RouterService) executeAndTrack(
	ctx context.Context,
	provider providers.Provider,
	account *models.Account,
	resolvedModel string,
	req Request,
) (Response, int, []byte, error) {
	providerID := provider.ID()

	// Assign proxy
	if err := s.proxyService.AssignProxy(account, providerID); err != nil {
		return Response{}, 0, nil, fmt.Errorf("failed to assign proxy: %w", err)
	}

	// Get token
	token, err := s.oauthService.GetAccessToken(account)
	if err != nil {
		return Response{}, 0, nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Execute request
	executeReq := &providers.ExecuteRequest{
		Model:    resolvedModel,
		Payload:  req.Payload,
		Stream:   req.Stream,
		Account:  account,
		ProxyURL: account.ProxyURL,
		Token:    token,
	}

	executeResp, err := provider.Execute(ctx, executeReq)

	// Handle connection errors
	if err != nil && executeResp == nil {
		s.statsTrackerService.RecordFailure(&account.ID, account.ProxyID, 0, err)
		return Response{}, 0, nil, fmt.Errorf("provider execution failed: %w", err)
	}

	statusCode := executeResp.StatusCode
	payload := executeResp.Payload

	// Record stats async
	providerIDPtr := &providerID
	go s.statsTrackerService.RecordRequest(
		&account.ID,
		account.ProxyID,
		providerIDPtr,
		resolvedModel,
		statusCode,
		executeResp.LatencyMs,
	)

	// Check success
	if statusCode < 200 || statusCode >= 300 {
		return Response{
			StatusCode: statusCode,
			Payload:    payload,
		}, statusCode, payload, fmt.Errorf("upstream error: %d", statusCode)
	}

	return Response{
		StatusCode: statusCode,
		Payload:    payload,
	}, statusCode, payload, nil
}

// shouldRetry determines if request should be retried based on status code
func (s *RouterService) shouldRetry(statusCode int, attempt int) bool {
	if attempt >= s.config.MaxRetries-1 {
		return false
	}
	return autherrors.IsRetryableStatus(statusCode)
}
