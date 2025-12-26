package services

import (
	"aigateway/models"
	"aigateway/repositories"
	"time"
)

// StatsQueryService handles querying and retrieving statistics data
type StatsQueryService struct {
	repo *repositories.StatsRepository
}

// NewStatsQueryService creates a new stats query service instance
func NewStatsQueryService(repo *repositories.StatsRepository) *StatsQueryService {
	return &StatsQueryService{
		repo: repo,
	}
}

// GetProxyStats retrieves aggregated stats for a proxy over the specified number of days
func (s *StatsQueryService) GetProxyStats(proxyID int, days int) ([]*models.ProxyStats, error) {
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	return s.repo.GetProxyStatsRange(proxyID, startDate, endDate)
}

// GetProxyStatsRange retrieves aggregated stats for a proxy within a date range
func (s *StatsQueryService) GetProxyStatsRange(proxyID int, startDate, endDate string) ([]*models.ProxyStats, error) {
	return s.repo.GetProxyStatsRange(proxyID, startDate, endDate)
}

// GetRecentLogs retrieves the most recent request logs up to the specified limit
func (s *StatsQueryService) GetRecentLogs(limit int) ([]*models.RequestLog, error) {
	return s.repo.GetRecentRequestLogs(limit)
}

// GetLogsByAccount retrieves request logs for a specific account
func (s *StatsQueryService) GetLogsByAccount(accountID string, limit int) ([]*models.RequestLog, error) {
	// This would require adding a method to the repository
	// For now, return an error indicating this is not implemented
	return nil, nil
}

// GetLogsByProxy retrieves request logs for a specific proxy
func (s *StatsQueryService) GetLogsByProxy(proxyID int, limit int) ([]*models.RequestLog, error) {
	// This would require adding a method to the repository
	// For now, return an error indicating this is not implemented
	return nil, nil
}

// GetLogsByProvider retrieves request logs for a specific provider
func (s *StatsQueryService) GetLogsByProvider(providerID string, limit int) ([]*models.RequestLog, error) {
	// This would require adding a method to the repository
	// For now, return an error indicating this is not implemented
	return nil, nil
}

// GetSuccessRate calculates the success rate for a proxy over the specified number of days
func (s *StatsQueryService) GetSuccessRate(proxyID int, days int) (float64, error) {
	stats, err := s.GetProxyStats(proxyID, days)
	if err != nil {
		return 0, err
	}

	if len(stats) == 0 {
		return 0, nil
	}

	var totalRequests int64
	var successfulRequests int64

	for _, stat := range stats {
		totalRequests += stat.TotalRequests
		successfulRequests += stat.SuccessfulRequests
	}

	if totalRequests == 0 {
		return 0, nil
	}

	return float64(successfulRequests) / float64(totalRequests) * 100, nil
}

// GetAverageLatency calculates the average latency for a proxy over the specified number of days
func (s *StatsQueryService) GetAverageLatency(proxyID int, days int) (float64, error) {
	stats, err := s.GetProxyStats(proxyID, days)
	if err != nil {
		return 0, err
	}

	if len(stats) == 0 {
		return 0, nil
	}

	var totalLatency int64
	var count int64

	for _, stat := range stats {
		totalLatency += int64(stat.AverageLatencyMs) * stat.TotalRequests
		count += stat.TotalRequests
	}

	if count == 0 {
		return 0, nil
	}

	return float64(totalLatency) / float64(count), nil
}
