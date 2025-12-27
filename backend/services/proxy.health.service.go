package services

import (
	"aigateway-backend/models"
	"aigateway-backend/repositories"

	"github.com/redis/go-redis/v9"
)

// ProxyHealthService handles proxy health monitoring and status updates
type ProxyHealthService struct {
	repo  *repositories.ProxyRepository
	redis *redis.Client
}

// NewProxyHealthService creates a new proxy health service instance
func NewProxyHealthService(repo *repositories.ProxyRepository, redis *redis.Client) *ProxyHealthService {
	return &ProxyHealthService{
		repo:  repo,
		redis: redis,
	}
}

// UpdateHealth updates the health status and latency of a proxy
func (s *ProxyHealthService) UpdateHealth(proxyID int, status models.HealthStatus, latencyMs int) error {
	return s.repo.UpdateHealth(proxyID, status, latencyMs)
}

// GetHealthStatus retrieves the current health status of a proxy
func (s *ProxyHealthService) GetHealthStatus(proxyID int) (models.HealthStatus, error) {
	proxy, err := s.repo.GetByID(proxyID)
	if err != nil {
		return models.HealthStatusDown, err
	}
	return proxy.HealthStatus, nil
}

// MarkHealthy marks a proxy as healthy with the given latency
func (s *ProxyHealthService) MarkHealthy(proxyID int, latencyMs int) error {
	return s.UpdateHealth(proxyID, models.HealthStatusHealthy, latencyMs)
}

// MarkDegraded marks a proxy as degraded with the given latency
func (s *ProxyHealthService) MarkDegraded(proxyID int, latencyMs int) error {
	return s.UpdateHealth(proxyID, models.HealthStatusDegraded, latencyMs)
}

// MarkDown marks a proxy as down
func (s *ProxyHealthService) MarkDown(proxyID int, latencyMs int) error {
	return s.UpdateHealth(proxyID, models.HealthStatusDown, latencyMs)
}

// IsHealthy checks if a proxy is in healthy status
func (s *ProxyHealthService) IsHealthy(proxyID int) (bool, error) {
	status, err := s.GetHealthStatus(proxyID)
	if err != nil {
		return false, err
	}
	return status == models.HealthStatusHealthy, nil
}

// GetLatency retrieves the current latency of a proxy
func (s *ProxyHealthService) GetLatency(proxyID int) (int, error) {
	proxy, err := s.repo.GetByID(proxyID)
	if err != nil {
		return 0, err
	}
	return proxy.AvgLatencyMs, nil
}
