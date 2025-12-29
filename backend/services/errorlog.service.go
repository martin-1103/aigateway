package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	ErrorLogKey = "error_logs"
	ErrorLogTTL = 24 * time.Hour
)

type ErrorLogEntry struct {
	ID        string                 `json:"id"`
	Service   string                 `json:"service"`
	Operation string                 `json:"operation"`
	Message   string                 `json:"message"`
	Context   map[string]interface{} `json:"context,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type ErrorLogService struct {
	redis *redis.Client
}

func NewErrorLogService(redis *redis.Client) *ErrorLogService {
	return &ErrorLogService{redis: redis}
}

// Log adds an error entry to Redis sorted set
func (s *ErrorLogService) Log(service, operation, message string, ctx map[string]interface{}) error {
	entry := ErrorLogEntry{
		ID:        uuid.New().String(),
		Service:   service,
		Operation: operation,
		Message:   message,
		Context:   ctx,
		CreatedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	score := float64(entry.CreatedAt.UnixMilli())
	return s.redis.ZAdd(context.Background(), ErrorLogKey, redis.Z{
		Score:  score,
		Member: string(data),
	}).Err()
}

// LogError is a convenience method for logging errors
func (s *ErrorLogService) LogError(service, operation string, err error, ctx map[string]interface{}) {
	if ctx == nil {
		ctx = make(map[string]interface{})
	}
	ctx["error_type"] = fmt.Sprintf("%T", err)

	_ = s.Log(service, operation, err.Error(), ctx)
}

// GetRecent returns recent error logs (newest first)
func (s *ErrorLogService) GetRecent(limit int64) ([]ErrorLogEntry, error) {
	// Get entries from sorted set (highest score = newest)
	results, err := s.redis.ZRevRange(context.Background(), ErrorLogKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]ErrorLogEntry, 0, len(results))
	for _, r := range results {
		var entry ErrorLogEntry
		if json.Unmarshal([]byte(r), &entry) == nil {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// GetByTimeRange returns error logs within a time range
func (s *ErrorLogService) GetByTimeRange(from, to time.Time, limit int64) ([]ErrorLogEntry, error) {
	min := fmt.Sprintf("%d", from.UnixMilli())
	max := fmt.Sprintf("%d", to.UnixMilli())

	results, err := s.redis.ZRevRangeByScore(context.Background(), ErrorLogKey, &redis.ZRangeBy{
		Min:   min,
		Max:   max,
		Count: limit,
	}).Result()
	if err != nil {
		return nil, err
	}

	entries := make([]ErrorLogEntry, 0, len(results))
	for _, r := range results {
		var entry ErrorLogEntry
		if json.Unmarshal([]byte(r), &entry) == nil {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// Cleanup removes entries older than TTL
func (s *ErrorLogService) Cleanup() error {
	cutoff := time.Now().Add(-ErrorLogTTL).UnixMilli()
	return s.redis.ZRemRangeByScore(context.Background(), ErrorLogKey, "-inf", fmt.Sprintf("%d", cutoff)).Err()
}

// StartCleanupRoutine starts a background goroutine to cleanup old logs
func (s *ErrorLogService) StartCleanupRoutine() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			_ = s.Cleanup()
		}
	}()
}
