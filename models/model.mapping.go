package models

import "time"

type ModelMapping struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Alias       string    `gorm:"uniqueIndex;size:100;not null" json:"alias"`
	ProviderID  string    `gorm:"size:50;not null" json:"provider_id"`
	ModelName   string    `gorm:"size:100;not null" json:"model_name"`
	Description string    `gorm:"size:255" json:"description,omitempty"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	Priority    int       `gorm:"default:0" json:"priority"`
	OwnerID     *string   `gorm:"type:varchar(36);index" json:"owner_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ModelMapping) TableName() string {
	return "model_mappings"
}
