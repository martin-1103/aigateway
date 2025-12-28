package models

import "time"

// Provider represents an AI service provider
type Provider struct {
	ID                 string      `gorm:"primaryKey;size:50" json:"id"`
	Name               string      `gorm:"size:100;not null" json:"name"`
	SupportedAuthTypes StringArray `gorm:"type:json;not null" json:"supported_auth_types"`
	BaseURL            string    `gorm:"size:255" json:"base_url"`
	AuthStrategy       string    `gorm:"size:50" json:"auth_strategy"`
	SupportedModels    string    `gorm:"type:json" json:"supported_models"`
	IsActive           bool      `gorm:"not null;default:1" json:"is_active"`
	Config             string    `gorm:"type:json" json:"config"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (Provider) TableName() string {
	return "providers"
}
