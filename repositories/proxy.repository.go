package repositories

import (
	"aigateway/models"
	"time"

	"gorm.io/gorm"
)

type ProxyRepository struct {
	db *gorm.DB
}

func NewProxyRepository(db *gorm.DB) *ProxyRepository {
	return &ProxyRepository{db: db}
}

func (r *ProxyRepository) Create(proxy *models.Proxy) error {
	return r.db.Create(proxy).Error
}

func (r *ProxyRepository) GetByID(id int) (*models.Proxy, error) {
	var proxy models.Proxy
	err := r.db.First(&proxy, id).Error
	return &proxy, err
}

func (r *ProxyRepository) GetHealthyProxies() ([]*models.Proxy, error) {
	var proxies []*models.Proxy
	err := r.db.Where("is_active = ? AND health_status = ?", true, models.HealthStatusHealthy).
		Order("priority DESC, current_accounts ASC").
		Find(&proxies).Error
	return proxies, err
}

func (r *ProxyRepository) GetActiveByProvider(providerID string) ([]*models.Proxy, error) {
	var proxies []*models.Proxy
	err := r.db.Where("is_active = ? AND health_status != ?", true, models.HealthStatusDown).
		Order("priority DESC, current_accounts ASC").
		Find(&proxies).Error
	return proxies, err
}

func (r *ProxyRepository) List(limit, offset int) ([]*models.Proxy, int64, error) {
	var proxies []*models.Proxy
	var total int64

	if err := r.db.Model(&models.Proxy{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Limit(limit).Offset(offset).Find(&proxies).Error
	return proxies, total, err
}

func (r *ProxyRepository) Update(proxy *models.Proxy) error {
	return r.db.Save(proxy).Error
}

func (r *ProxyRepository) Delete(id int) error {
	return r.db.Delete(&models.Proxy{}, id).Error
}

func (r *ProxyRepository) UpdateLastUsed(id int) error {
	now := time.Now()
	return r.db.Model(&models.Proxy{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"last_used_at": &now,
			"usage_count":  gorm.Expr("usage_count + 1"),
		}).Error
}

func (r *ProxyRepository) IncrementAccountCount(id int) error {
	return r.db.Model(&models.Proxy{}).
		Where("id = ?", id).
		Update("current_accounts", gorm.Expr("current_accounts + 1")).Error
}

func (r *ProxyRepository) DecrementAccountCount(id int) error {
	return r.db.Model(&models.Proxy{}).
		Where("id = ? AND current_accounts > 0", id).
		Update("current_accounts", gorm.Expr("current_accounts - 1")).Error
}

func (r *ProxyRepository) UpdateHealth(id int, status models.HealthStatus, latencyMs int) error {
	now := time.Now()
	updates := map[string]interface{}{
		"health_status":   status,
		"avg_latency_ms":  latencyMs,
		"last_checked_at": &now,
	}

	if status == models.HealthStatusHealthy {
		updates["consecutive_failures"] = 0
	} else {
		updates["consecutive_failures"] = gorm.Expr("consecutive_failures + 1")
	}

	return r.db.Model(&models.Proxy{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ProxyRepository) UpdateSuccessRate(id int, successRate float64) error {
	return r.db.Model(&models.Proxy{}).
		Where("id = ?", id).
		Update("success_rate", successRate).Error
}

func (r *ProxyRepository) RecalculateAccountCounts() error {
	return r.db.Exec(`
		UPDATE proxy_pool p
		SET current_accounts = (
			SELECT COUNT(*)
			FROM accounts a
			WHERE a.proxy_id = p.id AND a.is_active = true
		)
	`).Error
}
