// repositories/apikey.repository.go
package repositories

import (
	"time"

	"aigateway/models"

	"gorm.io/gorm"
)

type APIKeyRepository struct {
	db *gorm.DB
}

func NewAPIKeyRepository(db *gorm.DB) *APIKeyRepository {
	return &APIKeyRepository{db: db}
}

func (r *APIKeyRepository) Create(key *models.APIKey) error {
	return r.db.Create(key).Error
}

func (r *APIKeyRepository) GetByID(id string) (*models.APIKey, error) {
	var key models.APIKey
	err := r.db.Preload("User").Where("id = ?", id).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *APIKeyRepository) GetByHash(hash string) (*models.APIKey, error) {
	var key models.APIKey
	err := r.db.Preload("User").Where("key_hash = ? AND is_active = ?", hash, true).First(&key).Error
	if err != nil {
		return nil, err
	}
	return &key, nil
}

func (r *APIKeyRepository) ListByUserID(userID string) ([]*models.APIKey, error) {
	var keys []*models.APIKey
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&keys).Error
	return keys, err
}

func (r *APIKeyRepository) ListAll(limit, offset int) ([]*models.APIKey, int64, error) {
	var keys []*models.APIKey
	var total int64

	r.db.Model(&models.APIKey{}).Count(&total)
	err := r.db.Preload("User").Limit(limit).Offset(offset).Order("created_at DESC").Find(&keys).Error

	return keys, total, err
}

func (r *APIKeyRepository) UpdateLastUsed(id string) error {
	now := time.Now()
	return r.db.Model(&models.APIKey{}).Where("id = ?", id).Update("last_used_at", &now).Error
}

func (r *APIKeyRepository) Revoke(id string) error {
	return r.db.Model(&models.APIKey{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *APIKeyRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.APIKey{}).Error
}
