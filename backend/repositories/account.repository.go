package repositories

import (
	"aigateway-backend/models"
	"time"

	"gorm.io/gorm"
)

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(account *models.Account) error {
	return r.db.Create(account).Error
}

func (r *AccountRepository) GetByID(id string) (*models.Account, error) {
	var account models.Account
	err := r.db.Preload("Provider").Preload("Proxy").First(&account, "id = ?", id).Error
	return &account, err
}

func (r *AccountRepository) GetActiveByProvider(providerID string) ([]*models.Account, error) {
	var accounts []*models.Account
	err := r.db.Where("provider_id = ? AND is_active = ?", providerID, true).
		Order("id ASC").
		Find(&accounts).Error
	return accounts, err
}

func (r *AccountRepository) GetByProvider(providerID string) ([]*models.Account, error) {
	var accounts []*models.Account
	err := r.db.Where("provider_id = ?", providerID).Find(&accounts).Error
	return accounts, err
}

func (r *AccountRepository) List(limit, offset int) ([]*models.Account, int64, error) {
	var accounts []*models.Account
	var total int64

	if err := r.db.Model(&models.Account{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Provider").Preload("Proxy").
		Limit(limit).Offset(offset).
		Find(&accounts).Error

	return accounts, total, err
}

func (r *AccountRepository) Update(account *models.Account) error {
	return r.db.Save(account).Error
}

func (r *AccountRepository) Delete(id string) error {
	return r.db.Delete(&models.Account{}, "id = ?", id).Error
}

func (r *AccountRepository) UpdateLastUsed(id string) error {
	now := time.Now()
	return r.db.Model(&models.Account{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_used_at": &now,
			"usage_count":  gorm.Expr("usage_count + 1"),
		}).Error
}

func (r *AccountRepository) UpdateProxy(accountID string, proxyID int, proxyURL string) error {
	return r.db.Model(&models.Account{}).
		Where("id = ?", accountID).
		Updates(map[string]interface{}{
			"proxy_id":  proxyID,
			"proxy_url": proxyURL,
		}).Error
}

func (r *AccountRepository) ClearProxy(accountID string) error {
	return r.db.Model(&models.Account{}).
		Where("id = ?", accountID).
		Updates(map[string]interface{}{
			"proxy_id":  nil,
			"proxy_url": "",
		}).Error
}

func (r *AccountRepository) UpdateAuthData(id string, authData string) error {
	return r.db.Model(&models.Account{}).
		Where("id = ?", id).
		Update("auth_data", authData).Error
}

func (r *AccountRepository) UpdateAuthDataWithExpiry(id string, authData string, expiresAt time.Time) error {
	return r.db.Model(&models.Account{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"auth_data":  authData,
			"expires_at": expiresAt,
		}).Error
}

func (r *AccountRepository) GetExpiringAccounts(providerID string, withinDuration time.Duration) ([]*models.Account, error) {
	var accounts []*models.Account
	threshold := time.Now().Add(withinDuration)
	err := r.db.Where("provider_id = ? AND is_active = ? AND expires_at IS NOT NULL AND expires_at < ?",
		providerID, true, threshold).
		Find(&accounts).Error
	return accounts, err
}

func (r *AccountRepository) ListByCreator(creatorID string, limit, offset int) ([]*models.Account, int64, error) {
	var accounts []*models.Account
	var total int64

	if err := r.db.Model(&models.Account{}).Where("created_by = ?", creatorID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Preload("Provider").Preload("Proxy").
		Where("created_by = ?", creatorID).
		Limit(limit).Offset(offset).
		Find(&accounts).Error

	return accounts, total, err
}
