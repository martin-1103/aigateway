package services

import (
	"context"
	"fmt"
	"time"

	autherrors "aigateway-backend/auth/errors"
	"aigateway-backend/auth/manager"
	"aigateway-backend/models"
	"aigateway-backend/providers"
)

// RetryContext tracks retry state across attempts
type RetryContext struct {
	OriginalAccountID  string
	CurrentAccountID   string
	RetryCount         int
	SwitchedFromAccID  *string
	ProxyMarkedDown    bool
}

// executeWithAuthManager executes request with health-aware account selection and retry
func (s *RouterService) executeWithAuthManager(ctx context.Context, req Request, attempt int) (Response, error) {
	retryCtx := &RetryContext{}
	return s.executeWithRetry(ctx, req, attempt, retryCtx)
}

// executeWithRetry handles retry logic with same account before switching
func (s *RouterService) executeWithRetry(ctx context.Context, req Request, attempt int, retryCtx *RetryContext) (Response, error) {
	if attempt >= s.config.MaxRetries*2 { // Allow retries for both original and fallback account
		return Response{}, fmt.Errorf("max retries (%d) exceeded", s.config.MaxRetries*2)
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

	// Track original account
	if retryCtx.OriginalAccountID == "" {
		retryCtx.OriginalAccountID = accState.Account.ID
	}
	retryCtx.CurrentAccountID = accState.Account.ID

	// Execute and track result
	resp, statusCode, payload, execErr := s.executeWithPermanentProxy(ctx, provider, accState.Account, resolvedModel, req, retryCtx)

	// Mark result in AuthManager
	s.authManager.MarkResult(accState.Account.ID, resolvedModel, statusCode, payload)

	// Handle retry logic
	if execErr != nil && s.shouldRetry(statusCode, attempt) {
		retryCtx.RetryCount++

		// Check if we've exhausted retries for current account
		if retryCtx.RetryCount >= s.config.MaxRetries {
			// Mark proxy as down if we have one
			if accState.Account.ProxyID != nil && !retryCtx.ProxyMarkedDown {
				s.proxyService.MarkProxyDown(*accState.Account.ProxyID)
				retryCtx.ProxyMarkedDown = true
			}

			// Try to switch to a different account
			altAccount, switchErr := s.accountService.SelectAccountExcluding(providerID, resolvedModel, accState.Account.ID)
			if switchErr == nil {
				// Track that we switched accounts
				retryCtx.SwitchedFromAccID = &accState.Account.ID
				retryCtx.RetryCount = 0 // Reset retry count for new account

				// Execute with new account
				return s.executeWithSwitchedAccount(ctx, provider, altAccount, resolvedModel, req, retryCtx)
			}
			// No alternative account available, return the error
			return resp, execErr
		}

		// Retry with same account after delay
		time.Sleep(time.Duration(s.config.MaxRetries*100) * time.Millisecond)
		return s.executeWithRetry(ctx, req, attempt+1, retryCtx)
	}

	return resp, execErr
}

// executeWithSwitchedAccount executes with a different account after retry failure
func (s *RouterService) executeWithSwitchedAccount(
	ctx context.Context,
	provider providers.Provider,
	account *models.Account,
	resolvedModel string,
	req Request,
	retryCtx *RetryContext,
) (Response, error) {
	retryCtx.CurrentAccountID = account.ID

	resp, statusCode, payload, execErr := s.executeWithPermanentProxy(ctx, provider, account, resolvedModel, req, retryCtx)

	// Mark result
	s.authManager.MarkResult(account.ID, resolvedModel, statusCode, payload)

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
		retryCtx := &RetryContext{}
		return s.executeWithRetry(ctx, req, attempt+1, retryCtx)
	}

	if waitDur > s.config.MaxRetryWait {
		return Response{}, fmt.Errorf("all accounts blocked, wait time %v exceeds max %v", waitDur, s.config.MaxRetryWait)
	}

	// Wait and retry
	select {
	case <-ctx.Done():
		return Response{}, ctx.Err()
	case <-time.After(waitDur):
		retryCtx := &RetryContext{}
		return s.executeWithRetry(ctx, req, attempt+1, retryCtx)
	}
}

// executeWithPermanentProxy executes request using account's permanent proxy
func (s *RouterService) executeWithPermanentProxy(
	ctx context.Context,
	provider providers.Provider,
	account *models.Account,
	resolvedModel string,
	req Request,
	retryCtx *RetryContext,
) (Response, int, []byte, error) {
	providerID := provider.ID()

	// Get token (uses account's permanent proxy for refresh if needed)
	token, err := s.oauthService.GetAccessToken(account)
	if err != nil {
		return Response{}, 0, nil, fmt.Errorf("failed to get access token: %w", err)
	}

	// Execute request with account's permanent proxy
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
		s.statsTrackerService.RecordFailureWithRetry(&account.ID, account.ProxyID, 0, err, retryCtx.RetryCount, retryCtx.SwitchedFromAccID)
		// Track health failure (defensive: check accountRepo exists)
		if s.accountRepo != nil {
			go s.accountRepo.UpdateHealthFailure(account.ID, err.Error())
		}
		return Response{}, 0, nil, fmt.Errorf("provider execution failed: %w", err)
	}

	statusCode := executeResp.StatusCode
	payload := executeResp.Payload

	// Record stats async with retry info
	providerIDPtr := &providerID
	go s.statsTrackerService.RecordRequestWithRetry(
		&account.ID,
		account.ProxyID,
		providerIDPtr,
		resolvedModel,
		statusCode,
		executeResp.LatencyMs,
		retryCtx.RetryCount,
		retryCtx.SwitchedFromAccID,
	)

	// Check success
	if statusCode < 200 || statusCode >= 300 {
		// Track health failure for non-2xx (defensive: check accountRepo exists)
		if s.accountRepo != nil {
			go s.accountRepo.UpdateHealthFailure(account.ID, fmt.Sprintf("HTTP %d", statusCode))
		}
		return Response{
			StatusCode: statusCode,
			Payload:    payload,
		}, statusCode, payload, fmt.Errorf("upstream error: %d", statusCode)
	}

	// Track health success (defensive: check accountRepo exists)
	if s.accountRepo != nil {
		go s.accountRepo.UpdateHealthSuccess(account.ID)
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
