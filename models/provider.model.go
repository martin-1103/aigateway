package models

import "time"

// AuthType represents authentication method
type AuthType string

const (
	AuthTypeOAuth   AuthType = "oauth"
	AuthTypeAPIKey  AuthType = "api_key"
	AuthTypeBearer  AuthType = "bearer"
)

// Provider represents an AI service provider
type Provider struct {
	ID              string    `gorm:"primaryKey;size:50" json:"id"`
	Name            string    `gorm:"size:100;not null" json:"name"`
	BaseURL         string    `gorm:"size:255" json:"base_url"`
	AuthType        AuthType  `gorm:"type:enum('oauth','api_key','bearer');not null" json:"auth_type"`
	AuthStrategy    string    `gorm:"size:50" json:"auth_strategy"`
	SupportedModels string    `gorm:"type:json" json:"supported_models"`
	IsActive        bool      `gorm:"default:true" json:"is_active"`
	Config          string    `gorm:"type:json" json:"config"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (Provider) TableName() string {
	return "providers"
}
