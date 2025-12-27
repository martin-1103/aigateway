package services

import (
	"aigateway-backend/models"
	"aigateway-backend/repositories"
	"context"
	"log"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

// QuotaTrackerService tracks quota usage and learns limits from exhaustion events
type QuotaTrackerService struct {
	repo      *repositories.QuotaPatternRepository
	redis     *redis.Client
	keys      QuotaKeys
	windowTTL time.Duration
}

// NewQuotaTrackerService creates a new quota tracker service
func NewQuotaTrackerService(
	repo *repositories.QuotaPatternRepository,
	redisClient *redis.Client,
) *QuotaTrackerService {
	return &QuotaTrackerService{
		repo:      repo,
		redis:     redisClient,
		keys:      QuotaKeys{},
		windowTTL: QuotaWindowTTL,
	}
}

// RecordUsage records successful request usage (requests + tokens)
func (s *QuotaTrackerService) RecordUsage(accountID, model string, tokens int64) {
	ctx := context.Background()

	// Increment request counter
	reqKey := s.keys.RequestsKey(accountID, model)
	pipe := s.redis.Pipeline()
	pipe.Incr(ctx, reqKey)
	pipe.Expire(ctx, reqKey, s.windowTTL)

	// Increment token counter
	tokenKey := s.keys.TokensKey(accountID, model)
	pipe.IncrBy(ctx, tokenKey, tokens)
	pipe.Expire(ctx, tokenKey, s.windowTTL)

	// Set window start if not exists
	windowKey := s.keys.WindowStartKey(accountID, model)
	pipe.SetNX(ctx, windowKey, time.Now().Unix(), s.windowTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("[QuotaTracker] Failed to record usage: %v", err)
	}
}

// MarkExhausted marks account+model as exhausted and learns from the pattern
func (s *QuotaTrackerService) MarkExhausted(accountID, model string) {
	ctx := context.Background()

	// Get current usage before marking exhausted
	requests, _ := s.redis.Get(ctx, s.keys.RequestsKey(accountID, model)).Int()
	tokens, _ := s.redis.Get(ctx, s.keys.TokensKey(accountID, model)).Int64()

	// Mark as exhausted in Redis
	exhaustedKey := s.keys.ExhaustedKey(accountID, model)
	s.redis.Set(ctx, exhaustedKey, true, s.windowTTL)

	// Learn from this exhaustion event (async)
	go s.learnFromExhaustion(accountID, model, requests, tokens)
}

// learnFromExhaustion updates learned limits based on exhaustion event
func (s *QuotaTrackerService) learnFromExhaustion(accountID, model string, requests int, tokens int64) {
	if requests == 0 && tokens == 0 {
		return // Nothing to learn
	}

	pattern, err := s.repo.GetOrCreate(accountID, model)
	if err != nil {
		log.Printf("[QuotaTracker] Failed to get/create pattern: %v", err)
		return
	}

	now := time.Now()

	if pattern.EstRequestLimit == nil {
		// First time hitting limit - set directly
		pattern.EstRequestLimit = &requests
		pattern.EstTokenLimit = &tokens
		pattern.Confidence = 0.1
	} else {
		// Update with weighted average
		pattern.EstRequestLimit = weightedAvgInt(*pattern.EstRequestLimit, requests, pattern.Confidence)
		pattern.EstTokenLimit = weightedAvgInt64(*pattern.EstTokenLimit, tokens, pattern.Confidence)
		pattern.Confidence = math.Min(1.0, float64(pattern.SampleCount+1)/10.0)
	}

	pattern.SampleCount++
	pattern.LastExhaustedAt = &now

	if err := s.repo.Save(pattern); err != nil {
		log.Printf("[QuotaTracker] Failed to save pattern: %v", err)
	}
}

// IsAvailable checks if account+model has available quota
func (s *QuotaTrackerService) IsAvailable(accountID, model string) bool {
	ctx := context.Background()

	// Check if marked as exhausted
	exhaustedKey := s.keys.ExhaustedKey(accountID, model)
	exhausted, err := s.redis.Get(ctx, exhaustedKey).Bool()
	if err == redis.Nil {
		return true // Key doesn't exist = available
	}
	if err != nil {
		// Redis error - fail open (optimistic)
		log.Printf("[QuotaTracker] Redis error, assuming available: %v", err)
		return true
	}

	return !exhausted
}

// GetQuotaStatus returns current quota status for account+model
func (s *QuotaTrackerService) GetQuotaStatus(accountID, model string) *models.QuotaStatus {
	ctx := context.Background()

	status := &models.QuotaStatus{
		AccountID: accountID,
		Model:     model,
	}

	// Get current usage from Redis
	requests, _ := s.redis.Get(ctx, s.keys.RequestsKey(accountID, model)).Int()
	tokens, _ := s.redis.Get(ctx, s.keys.TokensKey(accountID, model)).Int64()
	status.RequestsUsed = requests
	status.TokensUsed = tokens

	// Check exhausted status
	exhausted, _ := s.redis.Get(ctx, s.keys.ExhaustedKey(accountID, model)).Bool()
	status.IsExhausted = exhausted

	// Get window reset time
	windowStart, err := s.redis.Get(ctx, s.keys.WindowStartKey(accountID, model)).Int64()
	if err == nil && windowStart > 0 {
		resetAt := time.Unix(windowStart, 0).Add(s.windowTTL)
		status.ResetsAt = &resetAt
	}

	// Get learned limits from MySQL
	pattern, err := s.repo.GetByAccountModel(accountID, model)
	if err == nil && pattern != nil {
		status.EstRequestLimit = pattern.EstRequestLimit
		status.EstTokenLimit = pattern.EstTokenLimit
		status.Confidence = s.getDecayedConfidence(pattern)

		// Calculate percent used if we have learned limits
		if pattern.EstRequestLimit != nil && *pattern.EstRequestLimit > 0 {
			pct := float64(requests) / float64(*pattern.EstRequestLimit) * 100
			status.PercentUsed = &pct
		}
	}

	return status
}

// GetEarliestReset returns the earliest reset time among exhausted accounts for a provider+model
func (s *QuotaTrackerService) GetEarliestReset(accountIDs []string, model string) *time.Time {
	ctx := context.Background()
	var earliest *time.Time

	for _, accID := range accountIDs {
		windowStart, err := s.redis.Get(ctx, s.keys.WindowStartKey(accID, model)).Int64()
		if err != nil || windowStart == 0 {
			continue
		}

		resetAt := time.Unix(windowStart, 0).Add(s.windowTTL)
		if earliest == nil || resetAt.Before(*earliest) {
			earliest = &resetAt
		}
	}

	return earliest
}

// ClearQuota clears quota tracking for account+model (e.g., on manual reset)
func (s *QuotaTrackerService) ClearQuota(accountID, model string) error {
	ctx := context.Background()

	keys := []string{
		s.keys.RequestsKey(accountID, model),
		s.keys.TokensKey(accountID, model),
		s.keys.ExhaustedKey(accountID, model),
		s.keys.WindowStartKey(accountID, model),
	}

	return s.redis.Del(ctx, keys...).Err()
}

// getDecayedConfidence returns confidence with decay for stale data
func (s *QuotaTrackerService) getDecayedConfidence(pattern *models.AccountQuotaPattern) float64 {
	if pattern.LastExhaustedAt == nil {
		return 0
	}

	daysSinceLastHit := time.Since(*pattern.LastExhaustedAt).Hours() / 24
	confidence := pattern.Confidence

	// Decay confidence if data is old (>7 days = 50% decay per week)
	if daysSinceLastHit > 7 {
		weeks := daysSinceLastHit / 7
		confidence *= math.Pow(0.5, weeks)
	}

	return confidence
}

// Helper: weighted average for int
func weightedAvgInt(old, new int, weight float64) *int {
	result := int((float64(old)*weight + float64(new)) / (weight + 1))
	return &result
}

// Helper: weighted average for int64
func weightedAvgInt64(old, new int64, weight float64) *int64 {
	result := int64((float64(old)*weight + float64(new)) / (weight + 1))
	return &result
}
