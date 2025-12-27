package repositories

import (
	"aigateway/models"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type QuotaPatternRepository struct {
	db *gorm.DB
}

func NewQuotaPatternRepository(db *gorm.DB) *QuotaPatternRepository {
	return &QuotaPatternRepository{db: db}
}

func (r *QuotaPatternRepository) GetByAccountModel(accountID, model string) (*models.AccountQuotaPattern, error) {
	var pattern models.AccountQuotaPattern
	err := r.db.Where("account_id = ? AND model = ?", accountID, model).First(&pattern).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &pattern, err
}

func (r *QuotaPatternRepository) GetOrCreate(accountID, model string) (*models.AccountQuotaPattern, error) {
	pattern, err := r.GetByAccountModel(accountID, model)
	if err != nil {
		return nil, err
	}
	if pattern != nil {
		return pattern, nil
	}

	// Create new pattern
	pattern = &models.AccountQuotaPattern{
		AccountID:   accountID,
		Model:       model,
		Confidence:  0,
		SampleCount: 0,
	}
	err = r.db.Create(pattern).Error
	return pattern, err
}

func (r *QuotaPatternRepository) Save(pattern *models.AccountQuotaPattern) error {
	return r.db.Save(pattern).Error
}

func (r *QuotaPatternRepository) Upsert(pattern *models.AccountQuotaPattern) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "account_id"}, {Name: "model"}},
		DoUpdates: clause.AssignmentColumns([]string{"est_request_limit", "est_token_limit", "confidence", "sample_count", "last_exhausted_at", "last_reset_at", "updated_at"}),
	}).Create(pattern).Error
}

func (r *QuotaPatternRepository) UpdateLimits(accountID, model string, estRequestLimit int, estTokenLimit int64, confidence float64, sampleCount int) error {
	now := time.Now()
	return r.db.Model(&models.AccountQuotaPattern{}).
		Where("account_id = ? AND model = ?", accountID, model).
		Updates(map[string]interface{}{
			"est_request_limit":  estRequestLimit,
			"est_token_limit":    estTokenLimit,
			"confidence":         confidence,
			"sample_count":       sampleCount,
			"last_exhausted_at":  &now,
			"updated_at":         now,
		}).Error
}

func (r *QuotaPatternRepository) MarkExhausted(accountID, model string) error {
	now := time.Now()
	return r.db.Model(&models.AccountQuotaPattern{}).
		Where("account_id = ? AND model = ?", accountID, model).
		Updates(map[string]interface{}{
			"last_exhausted_at": &now,
			"sample_count":      gorm.Expr("sample_count + 1"),
			"updated_at":        now,
		}).Error
}

func (r *QuotaPatternRepository) MarkReset(accountID, model string) error {
	now := time.Now()
	return r.db.Model(&models.AccountQuotaPattern{}).
		Where("account_id = ? AND model = ?", accountID, model).
		Updates(map[string]interface{}{
			"last_reset_at": &now,
			"updated_at":    now,
		}).Error
}

func (r *QuotaPatternRepository) ListByAccount(accountID string) ([]*models.AccountQuotaPattern, error) {
	var patterns []*models.AccountQuotaPattern
	err := r.db.Where("account_id = ?", accountID).Find(&patterns).Error
	return patterns, err
}

func (r *QuotaPatternRepository) ListByProvider(providerID string) ([]*models.AccountQuotaPattern, error) {
	var patterns []*models.AccountQuotaPattern
	err := r.db.
		Joins("JOIN accounts ON accounts.id = account_quota_pattern.account_id").
		Where("accounts.provider_id = ?", providerID).
		Preload("Account").
		Find(&patterns).Error
	return patterns, err
}

func (r *QuotaPatternRepository) ListAll() ([]*models.AccountQuotaPattern, error) {
	var patterns []*models.AccountQuotaPattern
	err := r.db.Preload("Account").Find(&patterns).Error
	return patterns, err
}

func (r *QuotaPatternRepository) DeleteByAccount(accountID string) error {
	return r.db.Where("account_id = ?", accountID).Delete(&models.AccountQuotaPattern{}).Error
}

func (r *QuotaPatternRepository) DeleteStale(before time.Time) error {
	return r.db.Where("updated_at < ?", before).Delete(&models.AccountQuotaPattern{}).Error
}
