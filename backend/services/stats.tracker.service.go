package services

import (
	"aigateway-backend/models"
	"aigateway-backend/repositories"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// StatsTrackerService handles recording and tracking of request statistics
type StatsTrackerService struct {
	repo        *repositories.StatsRepository
	proxyRepo   *repositories.ProxyRepository
	redis       *redis.Client
	healthService *ProxyHealthService
}

// NewStatsTrackerService creates a new stats tracker service instance
func NewStatsTrackerService(
	repo *repositories.StatsRepository,
	proxyRepo *repositories.ProxyRepository,
	redis *redis.Client,
	healthService *ProxyHealthService,
) *StatsTrackerService {
	return &StatsTrackerService{
		repo:        repo,
		proxyRepo:   proxyRepo,
		redis:       redis,
		healthService: healthService,
	}
}

// RecordRequest records a successful or failed request with all relevant metrics
func (s *StatsTrackerService) RecordRequest(accountID *string, proxyID *int, providerID *string, model string, statusCode, latencyMs int) {
	// Create request log
	log := &models.RequestLog{
		AccountID:  accountID,
		ProxyID:    proxyID,
		ProviderID: providerID,
		Model:      model,
		StatusCode: statusCode,
		LatencyMs:  latencyMs,
		CreatedAt:  time.Now(),
	}

	// Store log in database
	go s.repo.CreateRequestLog(log)

	// Update proxy stats if proxy was used
	if proxyID != nil {
		success := statusCode >= 200 && statusCode < 300
		go s.repo.IncrementProxyStats(*proxyID, providerID, success, latencyMs)

		// Update proxy health status
		if success {
			go s.healthService.MarkHealthy(*proxyID, latencyMs)
		} else {
			go s.healthService.MarkDegraded(*proxyID, latencyMs)
		}

		// Update Redis counters for real-time stats
		s.updateRedisCounters(*proxyID, success)
	}
}

// RecordFailure records a failed request with error information
func (s *StatsTrackerService) RecordFailure(accountID *string, proxyID *int, latencyMs int, err error) {
	log := &models.RequestLog{
		AccountID: accountID,
		ProxyID:   proxyID,
		StatusCode: 0,
		LatencyMs: latencyMs,
		Error:     err.Error(),
		CreatedAt: time.Now(),
	}

	go s.repo.CreateRequestLog(log)

	// Mark proxy as down if failure occurred
	if proxyID != nil {
		go s.healthService.MarkDown(*proxyID, latencyMs)
	}
}

// updateRedisCounters updates Redis counters for today's requests and errors
func (s *StatsTrackerService) updateRedisCounters(proxyID int, success bool) {
	ctx := context.Background()

	// Increment request counter
	requestKey := fmt.Sprintf("stats:proxy:%d:requests:today", proxyID)
	s.redis.Incr(ctx, requestKey)
	s.redis.Expire(ctx, requestKey, 24*time.Hour)

	// Increment error counter if request failed
	if !success {
		errorKey := fmt.Sprintf("stats:proxy:%d:errors:today", proxyID)
		s.redis.Incr(ctx, errorKey)
		s.redis.Expire(ctx, errorKey, 24*time.Hour)
	}
}

// GetTodayRequestCount retrieves the request count for today from Redis
func (s *StatsTrackerService) GetTodayRequestCount(proxyID int) (int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("stats:proxy:%d:requests:today", proxyID)
	count, err := s.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// GetTodayErrorCount retrieves the error count for today from Redis
func (s *StatsTrackerService) GetTodayErrorCount(proxyID int) (int64, error) {
	ctx := context.Background()
	key := fmt.Sprintf("stats:proxy:%d:errors:today", proxyID)
	count, err := s.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// CleanupOldLogs removes request logs older than the specified number of days
func (s *StatsTrackerService) CleanupOldLogs(days int) error {
	before := time.Now().AddDate(0, 0, -days)
	return s.repo.DeleteOldLogs(before)
}

// RecordRequestWithRetry records a request with retry and account switch information
func (s *StatsTrackerService) RecordRequestWithRetry(
	accountID *string,
	proxyID *int,
	providerID *string,
	model string,
	statusCode, latencyMs int,
	retryCount int,
	switchedFromAccountID *string,
) {
	log := &models.RequestLog{
		AccountID:             accountID,
		ProxyID:               proxyID,
		ProviderID:            providerID,
		Model:                 model,
		StatusCode:            statusCode,
		LatencyMs:             latencyMs,
		RetryCount:            retryCount,
		SwitchedFromAccountID: switchedFromAccountID,
		CreatedAt:             time.Now(),
	}

	go s.repo.CreateRequestLog(log)

	if proxyID != nil {
		success := statusCode >= 200 && statusCode < 300
		go s.repo.IncrementProxyStats(*proxyID, providerID, success, latencyMs)

		if success {
			go s.healthService.MarkHealthy(*proxyID, latencyMs)
		} else {
			go s.healthService.MarkDegraded(*proxyID, latencyMs)
		}

		s.updateRedisCounters(*proxyID, success)
	}

	// Track retry and switch metrics in Redis
	if retryCount > 0 {
		s.incrementRetryCounter()
	}
	if switchedFromAccountID != nil {
		s.incrementSwitchCounter()
	}
}

// RecordFailureWithRetry records a failed request with retry information
func (s *StatsTrackerService) RecordFailureWithRetry(
	accountID *string,
	proxyID *int,
	latencyMs int,
	err error,
	retryCount int,
	switchedFromAccountID *string,
) {
	log := &models.RequestLog{
		AccountID:             accountID,
		ProxyID:               proxyID,
		StatusCode:            0,
		LatencyMs:             latencyMs,
		RetryCount:            retryCount,
		SwitchedFromAccountID: switchedFromAccountID,
		Error:                 err.Error(),
		CreatedAt:             time.Now(),
	}

	go s.repo.CreateRequestLog(log)

	if proxyID != nil {
		go s.healthService.MarkDown(*proxyID, latencyMs)
	}

	if retryCount > 0 {
		s.incrementRetryCounter()
	}
	if switchedFromAccountID != nil {
		s.incrementSwitchCounter()
	}
}

// incrementRetryCounter increments the global retry counter in Redis
func (s *StatsTrackerService) incrementRetryCounter() {
	ctx := context.Background()
	key := "stats:global:retries:today"
	s.redis.Incr(ctx, key)
	s.redis.Expire(ctx, key, 24*time.Hour)
}

// incrementSwitchCounter increments the global account switch counter in Redis
func (s *StatsTrackerService) incrementSwitchCounter() {
	ctx := context.Background()
	key := "stats:global:account_switches:today"
	s.redis.Incr(ctx, key)
	s.redis.Expire(ctx, key, 24*time.Hour)
}

// GetTodayRetryCount retrieves the retry count for today
func (s *StatsTrackerService) GetTodayRetryCount() (int64, error) {
	ctx := context.Background()
	key := "stats:global:retries:today"
	count, err := s.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// GetTodaySwitchCount retrieves the account switch count for today
func (s *StatsTrackerService) GetTodaySwitchCount() (int64, error) {
	ctx := context.Background()
	key := "stats:global:account_switches:today"
	count, err := s.redis.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}
