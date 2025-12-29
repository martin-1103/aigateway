package models

import "time"

// Account represents authentication credentials for a provider
type Account struct {
	ID         string     `gorm:"primaryKey;size:36" json:"id"`
	ProviderID string     `gorm:"size:50;not null;index:idx_provider_active" json:"provider_id"`
	Label      string     `gorm:"size:100;not null;index" json:"label"`
	AuthData   string     `gorm:"type:json;not null" json:"auth_data"`
	Metadata   string     `gorm:"type:json" json:"metadata"`
	IsActive   bool       `gorm:"default:true;index:idx_provider_active" json:"is_active"`
	ProxyURL   string     `gorm:"size:255" json:"proxy_url"`
	ProxyID    *int       `gorm:"index" json:"proxy_id"`
	ExpiresAt  *time.Time `gorm:"index" json:"expires_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	UsageCount int64      `gorm:"default:0" json:"usage_count"`

	// Health tracking
	HealthStatus   string     `gorm:"size:20;default:'healthy';index" json:"health_status"` // healthy, degraded, down
	FailureCount   int        `gorm:"default:0" json:"failure_count"`
	LastErrorAt    *time.Time `json:"last_error_at"`
	LastErrorMsg   string     `gorm:"type:text" json:"last_error_msg"`
	LastSuccessAt  *time.Time `json:"last_success_at"`

	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	CreatedBy  *string    `gorm:"type:varchar(36);index" json:"created_by,omitempty"`

	Provider *Provider `gorm:"foreignKey:ProviderID" json:"provider,omitempty"`
	Proxy    *Proxy    `gorm:"foreignKey:ProxyID" json:"proxy,omitempty"`
}

func (Account) TableName() string {
	return "accounts"
}
