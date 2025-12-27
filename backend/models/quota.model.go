package models

import "time"

// AccountQuotaPattern stores learned quota patterns per account per model
type AccountQuotaPattern struct {
	ID        int64  `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID string `gorm:"size:36;not null;uniqueIndex:idx_account_model" json:"account_id"`
	Model     string `gorm:"size:100;not null;uniqueIndex:idx_account_model" json:"model"`

	// Learned thresholds (nil = unknown)
	EstRequestLimit *int   `json:"est_request_limit"`
	EstTokenLimit   *int64 `json:"est_token_limit"`

	// Confidence tracking
	Confidence  float64 `gorm:"type:decimal(3,2);default:0" json:"confidence"`
	SampleCount int     `gorm:"default:0" json:"sample_count"`

	// State tracking
	LastExhaustedAt *time.Time `json:"last_exhausted_at"`
	LastResetAt     *time.Time `json:"last_reset_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Account *Account `gorm:"foreignKey:AccountID" json:"account,omitempty"`
}

func (AccountQuotaPattern) TableName() string {
	return "account_quota_pattern"
}

// QuotaStatus represents current quota state for an account+model
type QuotaStatus struct {
	AccountID       string     `json:"account_id"`
	Model           string     `json:"model"`
	RequestsUsed    int        `json:"requests_used"`
	TokensUsed      int64      `json:"tokens_used"`
	EstRequestLimit *int       `json:"est_request_limit"`
	EstTokenLimit   *int64     `json:"est_token_limit"`
	PercentUsed     *float64   `json:"percent_used"`
	Confidence      float64    `json:"confidence"`
	IsExhausted     bool       `json:"is_exhausted"`
	ResetsAt        *time.Time `json:"resets_at"`
}

// ProviderQuotaSummary represents quota summary for a provider
type ProviderQuotaSummary struct {
	ProviderID        string                       `json:"provider_id"`
	TotalAccounts     int                          `json:"total_accounts"`
	AvailableAccounts int                          `json:"available_accounts"`
	ExhaustedAccounts int                          `json:"exhausted_accounts"`
	Models            map[string]*ModelQuotaStatus `json:"models"`
	Health            string                       `json:"health"`
}

// ModelQuotaStatus represents quota status for a specific model
type ModelQuotaStatus struct {
	Total          int        `json:"total"`
	Available      int        `json:"available"`
	Exhausted      int        `json:"exhausted"`
	AvgPercentUsed *float64   `json:"avg_percent_used"`
	NextResetAt    *time.Time `json:"next_reset_at"`
}
