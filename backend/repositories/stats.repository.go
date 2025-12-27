package repositories

import (
	"aigateway-backend/models"
	"time"

	"gorm.io/gorm"
)

type StatsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

func (r *StatsRepository) CreateRequestLog(log *models.RequestLog) error {
	return r.db.Create(log).Error
}

func (r *StatsRepository) IncrementProxyStats(proxyID int, providerID *string, success bool, latencyMs int) error {
	date := time.Now().Format("2006-01-02")

	updates := map[string]interface{}{
		"request_count":    gorm.Expr("request_count + 1"),
		"total_latency_ms": gorm.Expr("total_latency_ms + ?", latencyMs),
	}

	if success {
		updates["success_count"] = gorm.Expr("success_count + 1")
	} else {
		updates["error_count"] = gorm.Expr("error_count + 1")
	}

	result := r.db.Model(&models.ProxyStats{}).
		Where("proxy_id = ? AND date = ?", proxyID, date).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		stats := &models.ProxyStats{
			ProxyID:        proxyID,
			ProviderID:     providerID,
			Date:           parseDate(date),
			RequestCount:   1,
			TotalLatencyMs: int64(latencyMs),
		}
		if success {
			stats.SuccessCount = 1
		} else {
			stats.ErrorCount = 1
		}
		return r.db.Create(stats).Error
	}

	return nil
}

func (r *StatsRepository) GetProxyStatsByDate(proxyID int, date string) (*models.ProxyStats, error) {
	var stats models.ProxyStats
	err := r.db.Where("proxy_id = ? AND date = ?", proxyID, date).First(&stats).Error
	return &stats, err
}

func (r *StatsRepository) GetProxyStatsRange(proxyID int, startDate, endDate string) ([]*models.ProxyStats, error) {
	var stats []*models.ProxyStats
	err := r.db.Where("proxy_id = ? AND date BETWEEN ? AND ?", proxyID, startDate, endDate).
		Order("date DESC").
		Find(&stats).Error
	return stats, err
}

func (r *StatsRepository) GetRecentRequestLogs(limit int) ([]*models.RequestLog, error) {
	var logs []*models.RequestLog
	err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *StatsRepository) GetRequestLogsByAccount(accountID string, limit int) ([]*models.RequestLog, error) {
	var logs []*models.RequestLog
	err := r.db.Where("account_id = ?", accountID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

func (r *StatsRepository) DeleteOldLogs(before time.Time) error {
	return r.db.Where("created_at < ?", before).Delete(&models.RequestLog{}).Error
}

func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}
