package repositories

import (
	"aigateway/models"

	"gorm.io/gorm"
)

type ModelMappingRepository struct {
	db *gorm.DB
}

func NewModelMappingRepository(db *gorm.DB) *ModelMappingRepository {
	return &ModelMappingRepository{db: db}
}

func (r *ModelMappingRepository) Create(mapping *models.ModelMapping) error {
	return r.db.Create(mapping).Error
}

func (r *ModelMappingRepository) GetByAlias(alias string) (*models.ModelMapping, error) {
	var mapping models.ModelMapping
	err := r.db.Where("alias = ? AND enabled = ?", alias, true).First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}

func (r *ModelMappingRepository) List(limit, offset int) ([]*models.ModelMapping, int64, error) {
	var mappings []*models.ModelMapping
	var total int64

	if err := r.db.Model(&models.ModelMapping{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Order("priority DESC, alias ASC").Limit(limit).Offset(offset).Find(&mappings).Error
	return mappings, total, err
}

func (r *ModelMappingRepository) Update(alias string, mapping *models.ModelMapping) error {
	return r.db.Where("alias = ?", alias).Updates(mapping).Error
}

func (r *ModelMappingRepository) Delete(alias string) error {
	return r.db.Where("alias = ?", alias).Delete(&models.ModelMapping{}).Error
}

func (r *ModelMappingRepository) ListForUser(userID string, limit, offset int) ([]*models.ModelMapping, int64, error) {
	var mappings []*models.ModelMapping
	var total int64

	// Show global (owner_id IS NULL) + user's own mappings
	query := r.db.Model(&models.ModelMapping{}).Where("owner_id IS NULL OR owner_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Where("owner_id IS NULL OR owner_id = ?", userID).
		Order("priority DESC, alias ASC").
		Limit(limit).Offset(offset).
		Find(&mappings).Error
	return mappings, total, err
}

func (r *ModelMappingRepository) GetByAliasWithOwner(alias string) (*models.ModelMapping, error) {
	var mapping models.ModelMapping
	err := r.db.Where("alias = ?", alias).First(&mapping).Error
	if err != nil {
		return nil, err
	}
	return &mapping, nil
}
