package repositories

import (
	"aigateway-backend/models"
	"gorm.io/gorm"
)

type ProviderRepository struct {
	db *gorm.DB
}

func NewProviderRepository(db *gorm.DB) *ProviderRepository {
	return &ProviderRepository{db: db}
}

// GetByID retrieves a provider by ID
func (r *ProviderRepository) GetByID(id string) (*models.Provider, error) {
	var provider models.Provider
	if err := r.db.First(&provider, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &provider, nil
}

// GetActive retrieves active providers with pagination
func (r *ProviderRepository) GetActive(limit, offset int) ([]models.Provider, int64, error) {
	var providers []models.Provider
	var total int64

	if err := r.db.Where("is_active = ?", true).Count(&total).
		Limit(limit).Offset(offset).Find(&providers).Error; err != nil {
		return nil, 0, err
	}
	return providers, total, nil
}

// List retrieves all providers
func (r *ProviderRepository) List() ([]models.Provider, error) {
	var providers []models.Provider
	if err := r.db.Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}

// ListActive retrieves all active providers
func (r *ProviderRepository) ListActive() ([]models.Provider, error) {
	var providers []models.Provider
	if err := r.db.Where("is_active = ?", true).Find(&providers).Error; err != nil {
		return nil, err
	}
	return providers, nil
}
