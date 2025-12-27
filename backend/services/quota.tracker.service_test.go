package services

import (
	"aigateway-backend/repositories"
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// Create table directly (SQLite doesn't support ENUM)
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS account_quota_pattern (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id TEXT NOT NULL,
			model TEXT NOT NULL,
			est_request_limit INTEGER,
			est_token_limit INTEGER,
			confidence REAL DEFAULT 0,
			sample_count INTEGER DEFAULT 0,
			last_exhausted_at DATETIME,
			last_reset_at DATETIME,
			created_at DATETIME,
			updated_at DATETIME
		)
	`).Error
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}

	// Create unique index
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_account_model ON account_quota_pattern(account_id, model)`)

	return db
}

// setupTestRedis creates a miniredis instance for testing
func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return mr, client
}

func TestRecordUsage(t *testing.T) {
	db := setupTestDB(t)
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()

	repo := repositories.NewQuotaPatternRepository(db)
	service := NewQuotaTrackerService(repo, redisClient)

	accountID := "test-account-1"
	model := "gemini-2.5-pro"
	tokens := int64(1000)

	// Record usage
	service.RecordUsage(accountID, model, tokens)

	// Verify Redis counters
	keys := QuotaKeys{}

	reqCount, err := redisClient.Get(context.Background(), keys.RequestsKey(accountID, model)).Int()
	if err != nil {
		t.Fatalf("failed to get request count: %v", err)
	}
	if reqCount != 1 {
		t.Errorf("expected request count 1, got %d", reqCount)
	}

	tokenCount, err := redisClient.Get(context.Background(), keys.TokensKey(accountID, model)).Int64()
	if err != nil {
		t.Fatalf("failed to get token count: %v", err)
	}
	if tokenCount != tokens {
		t.Errorf("expected token count %d, got %d", tokens, tokenCount)
	}

	// Record more usage
	service.RecordUsage(accountID, model, 500)

	reqCount, _ = redisClient.Get(context.Background(), keys.RequestsKey(accountID, model)).Int()
	if reqCount != 2 {
		t.Errorf("expected request count 2, got %d", reqCount)
	}

	tokenCount, _ = redisClient.Get(context.Background(), keys.TokensKey(accountID, model)).Int64()
	if tokenCount != 1500 {
		t.Errorf("expected token count 1500, got %d", tokenCount)
	}
}

func TestMarkExhausted(t *testing.T) {
	db := setupTestDB(t)
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()

	repo := repositories.NewQuotaPatternRepository(db)
	service := NewQuotaTrackerService(repo, redisClient)

	accountID := "test-account-2"
	model := "claude-sonnet-4.5"

	// Record some usage first
	service.RecordUsage(accountID, model, 5000)
	service.RecordUsage(accountID, model, 3000)
	service.RecordUsage(accountID, model, 2000)

	// Mark as exhausted
	service.MarkExhausted(accountID, model)

	// Give goroutine time to complete
	time.Sleep(100 * time.Millisecond)

	// Verify exhausted flag in Redis
	keys := QuotaKeys{}
	exhausted, err := redisClient.Get(context.Background(), keys.ExhaustedKey(accountID, model)).Bool()
	if err != nil {
		t.Fatalf("failed to get exhausted flag: %v", err)
	}
	if !exhausted {
		t.Error("expected exhausted to be true")
	}

	// Verify pattern was learned in DB
	pattern, err := repo.GetByAccountModel(accountID, model)
	if err != nil {
		t.Fatalf("failed to get pattern: %v", err)
	}
	if pattern == nil {
		t.Fatal("expected pattern to be created")
	}

	// First exhaustion should set limits directly
	if pattern.EstRequestLimit == nil {
		t.Error("expected EstRequestLimit to be set")
	} else if *pattern.EstRequestLimit != 3 {
		t.Errorf("expected EstRequestLimit 3, got %d", *pattern.EstRequestLimit)
	}

	if pattern.EstTokenLimit == nil {
		t.Error("expected EstTokenLimit to be set")
	} else if *pattern.EstTokenLimit != 10000 {
		t.Errorf("expected EstTokenLimit 10000, got %d", *pattern.EstTokenLimit)
	}

	if pattern.SampleCount != 1 {
		t.Errorf("expected SampleCount 1, got %d", pattern.SampleCount)
	}
}

func TestMarkExhausted_LearningAlgorithm(t *testing.T) {
	db := setupTestDB(t)
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()

	repo := repositories.NewQuotaPatternRepository(db)
	service := NewQuotaTrackerService(repo, redisClient)

	accountID := "test-account-3"
	model := "gemini-2.5-pro"

	// First exhaustion: 100 requests
	for i := 0; i < 100; i++ {
		service.RecordUsage(accountID, model, 100)
	}
	service.MarkExhausted(accountID, model)
	time.Sleep(100 * time.Millisecond)

	// Clear Redis for second test
	mr.FlushAll()

	// Second exhaustion: 120 requests (weighted average should adjust)
	for i := 0; i < 120; i++ {
		service.RecordUsage(accountID, model, 100)
	}
	service.MarkExhausted(accountID, model)
	time.Sleep(100 * time.Millisecond)

	pattern, _ := repo.GetByAccountModel(accountID, model)

	if pattern.SampleCount != 2 {
		t.Errorf("expected SampleCount 2, got %d", pattern.SampleCount)
	}

	// Weighted average: (100 * 0.1 + 120) / 1.1 ≈ 118
	// But first hit sets directly, so second uses weighted avg
	if pattern.EstRequestLimit != nil {
		limit := *pattern.EstRequestLimit
		// Should be somewhere between 100 and 120 due to weighted average
		if limit < 100 || limit > 120 {
			t.Errorf("expected EstRequestLimit between 100-120, got %d", limit)
		}
	}

	// Confidence should increase
	if pattern.Confidence < 0.1 {
		t.Errorf("expected Confidence >= 0.1, got %f", pattern.Confidence)
	}
}

func TestIsAvailable(t *testing.T) {
	db := setupTestDB(t)
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()

	repo := repositories.NewQuotaPatternRepository(db)
	service := NewQuotaTrackerService(repo, redisClient)

	accountID := "test-account-4"
	model := "gemini-2.5-pro"

	// Initially should be available
	if !service.IsAvailable(accountID, model) {
		t.Error("expected account to be available initially")
	}

	// Record usage and mark exhausted
	service.RecordUsage(accountID, model, 1000)
	service.MarkExhausted(accountID, model)

	// Should not be available after exhaustion
	if service.IsAvailable(accountID, model) {
		t.Error("expected account to not be available after exhaustion")
	}

	// Different model should still be available
	if !service.IsAvailable(accountID, "different-model") {
		t.Error("expected different model to be available")
	}

	// Different account should still be available
	if !service.IsAvailable("different-account", model) {
		t.Error("expected different account to be available")
	}
}

func TestGetQuotaStatus(t *testing.T) {
	db := setupTestDB(t)
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()

	repo := repositories.NewQuotaPatternRepository(db)
	service := NewQuotaTrackerService(repo, redisClient)

	accountID := "test-account-5"
	model := "gemini-2.5-pro"

	// Record some usage
	service.RecordUsage(accountID, model, 500)
	service.RecordUsage(accountID, model, 300)

	status := service.GetQuotaStatus(accountID, model)

	if status.AccountID != accountID {
		t.Errorf("expected AccountID %s, got %s", accountID, status.AccountID)
	}

	if status.Model != model {
		t.Errorf("expected Model %s, got %s", model, status.Model)
	}

	if status.RequestsUsed != 2 {
		t.Errorf("expected RequestsUsed 2, got %d", status.RequestsUsed)
	}

	if status.TokensUsed != 800 {
		t.Errorf("expected TokensUsed 800, got %d", status.TokensUsed)
	}

	if status.IsExhausted {
		t.Error("expected IsExhausted to be false")
	}

	// No learned limits yet
	if status.EstRequestLimit != nil {
		t.Error("expected EstRequestLimit to be nil")
	}
}

func TestGetQuotaStatus_WithLearnedLimits(t *testing.T) {
	db := setupTestDB(t)
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()

	repo := repositories.NewQuotaPatternRepository(db)
	service := NewQuotaTrackerService(repo, redisClient)

	accountID := "test-account-6"
	model := "gemini-2.5-pro"

	// Record usage and exhaust
	for i := 0; i < 50; i++ {
		service.RecordUsage(accountID, model, 100)
	}
	service.MarkExhausted(accountID, model)
	time.Sleep(100 * time.Millisecond)

	// Clear and record new usage
	mr.FlushAll()
	service.RecordUsage(accountID, model, 100)
	service.RecordUsage(accountID, model, 100)

	status := service.GetQuotaStatus(accountID, model)

	if status.EstRequestLimit == nil {
		t.Fatal("expected EstRequestLimit to be set")
	}

	if *status.EstRequestLimit != 50 {
		t.Errorf("expected EstRequestLimit 50, got %d", *status.EstRequestLimit)
	}

	// PercentUsed should be calculated: 2/50 * 100 = 4%
	if status.PercentUsed == nil {
		t.Fatal("expected PercentUsed to be set")
	}

	if *status.PercentUsed != 4.0 {
		t.Errorf("expected PercentUsed 4.0, got %f", *status.PercentUsed)
	}
}

func TestClearQuota(t *testing.T) {
	db := setupTestDB(t)
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()

	repo := repositories.NewQuotaPatternRepository(db)
	service := NewQuotaTrackerService(repo, redisClient)

	accountID := "test-account-7"
	model := "gemini-2.5-pro"

	// Record usage and exhaust
	service.RecordUsage(accountID, model, 1000)
	service.MarkExhausted(accountID, model)

	// Should not be available
	if service.IsAvailable(accountID, model) {
		t.Error("expected account to not be available")
	}

	// Clear quota
	err := service.ClearQuota(accountID, model)
	if err != nil {
		t.Fatalf("failed to clear quota: %v", err)
	}

	// Should be available again
	if !service.IsAvailable(accountID, model) {
		t.Error("expected account to be available after clear")
	}

	// Usage counters should be reset
	status := service.GetQuotaStatus(accountID, model)
	if status.RequestsUsed != 0 {
		t.Errorf("expected RequestsUsed 0 after clear, got %d", status.RequestsUsed)
	}
}

func TestGetEarliestReset(t *testing.T) {
	db := setupTestDB(t)
	mr, redisClient := setupTestRedis(t)
	defer mr.Close()

	repo := repositories.NewQuotaPatternRepository(db)
	service := NewQuotaTrackerService(repo, redisClient)

	model := "gemini-2.5-pro"

	// Record usage for multiple accounts
	service.RecordUsage("account-1", model, 100)
	time.Sleep(10 * time.Millisecond)
	service.RecordUsage("account-2", model, 100)
	time.Sleep(10 * time.Millisecond)
	service.RecordUsage("account-3", model, 100)

	accountIDs := []string{"account-1", "account-2", "account-3"}

	resetAt := service.GetEarliestReset(accountIDs, model)

	if resetAt == nil {
		t.Fatal("expected resetAt to be set")
	}

	// Reset should be approximately 5 hours from now (window TTL)
	expectedReset := time.Now().Add(QuotaWindowTTL)
	diff := resetAt.Sub(expectedReset)

	// Allow 1 second tolerance
	if diff < -time.Second || diff > time.Second {
		t.Errorf("reset time off by %v", diff)
	}
}
