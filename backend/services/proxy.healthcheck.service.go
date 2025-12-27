package services

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"aigateway-backend/models"
	"aigateway-backend/repositories"
)

// ProxyHealthCheckService performs periodic health checks on proxies
// and automatically recovers down proxies when they become available again
type ProxyHealthCheckService struct {
	repo          *repositories.ProxyRepository
	ticker        *time.Ticker
	done          chan struct{}
	wg            sync.WaitGroup
	checkInterval time.Duration
	recoveryDelay time.Duration
}

// NewProxyHealthCheckService creates a new health check service
func NewProxyHealthCheckService(
	repo *repositories.ProxyRepository,
	checkIntervalMin int,
	recoveryDelayMin int,
) *ProxyHealthCheckService {
	return &ProxyHealthCheckService{
		repo:          repo,
		done:          make(chan struct{}),
		checkInterval: time.Duration(checkIntervalMin) * time.Minute,
		recoveryDelay: time.Duration(recoveryDelayMin) * time.Minute,
	}
}

// Start begins the periodic health check service
func (s *ProxyHealthCheckService) Start(ctx context.Context) {
	s.wg.Add(1)
	go s.periodicHealthCheck(ctx)
	log.Printf("ProxyHealthCheckService started with interval: %v, recovery delay: %v",
		s.checkInterval, s.recoveryDelay)
}

// Stop stops the health check service
func (s *ProxyHealthCheckService) Stop() {
	close(s.done)
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.wg.Wait()
	log.Println("ProxyHealthCheckService stopped")
}

// periodicHealthCheck runs the health check loop
func (s *ProxyHealthCheckService) periodicHealthCheck(ctx context.Context) {
	defer s.wg.Done()

	// Run immediately on startup
	s.checkAllProxies()

	// Then run periodically
	s.ticker = time.NewTicker(s.checkInterval)
	defer s.ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ctx.Done():
			return
		case <-s.ticker.C:
			s.checkAllProxies()
		}
	}
}

// checkAllProxies checks health of all proxies
func (s *ProxyHealthCheckService) checkAllProxies() {
	proxies, _, err := s.repo.List(1000, 0) // Get all proxies
	if err != nil {
		log.Printf("Error fetching proxies: %v", err)
		return
	}

	if len(proxies) == 0 {
		return
	}

	log.Printf("Starting health check for %d proxies", len(proxies))

	// Check each proxy concurrently
	var wg sync.WaitGroup
	for _, proxy := range proxies {
		wg.Add(1)
		go func(p *models.Proxy) {
			defer wg.Done()
			s.checkProxy(p)
		}(proxy)
	}
	wg.Wait()
	log.Println("Health check completed")
}

// checkProxy performs health check on a single proxy
func (s *ProxyHealthCheckService) checkProxy(proxy *models.Proxy) {
	if !proxy.IsActive {
		return
	}

	// Tier 1: Quick TCP connectivity check
	if !s.quickTCPCheck(proxy) {
		// TCP failed - stay down or mark down
		if proxy.HealthStatus != models.HealthStatusDown {
			now := time.Now()
			s.repo.UpdateHealthWithDownTime(proxy.ID, models.HealthStatusDown, &now)
			log.Printf("Proxy %d marked DOWN (TCP failed)", proxy.ID)
		}
		return
	}

	// Tier 2: Full HTTP check through proxy with credentials
	if s.fullHTTPCheck(proxy) {
		// Success - mark healthy
		if proxy.HealthStatus != models.HealthStatusHealthy {
			s.repo.UpdateHealthWithDownTime(proxy.ID, models.HealthStatusHealthy, nil)
			log.Printf("Proxy %d recovered to HEALTHY", proxy.ID)
		}
	} else {
		// Failed - mark degraded (allow fallback use)
		if proxy.HealthStatus != models.HealthStatusDegraded {
			s.repo.UpdateHealthWithDownTime(proxy.ID, models.HealthStatusDegraded, nil)
			log.Printf("Proxy %d marked DEGRADED (HTTP check failed)", proxy.ID)
		}
	}
}

// quickTCPCheck performs a quick TCP connection test to proxy
func (s *ProxyHealthCheckService) quickTCPCheck(proxy *models.Proxy) bool {
	parsed, err := url.Parse(proxy.URL)
	if err != nil {
		return false
	}

	// Extract host:port from proxy URL
	host := parsed.Host
	if host == "" {
		host = parsed.Hostname()
		if host == "" {
			return false
		}
		if parsed.Port() != "" {
			host = host + ":" + parsed.Port()
		} else {
			// Default to proxy port 44444 if not specified
			host = host + ":44444"
		}
	}

	// Attempt TCP connection with timeout
	conn, err := net.DialTimeout("tcp", host, 5*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// fullHTTPCheck performs a full HTTP test through the proxy
func (s *ProxyHealthCheckService) fullHTTPCheck(proxy *models.Proxy) bool {
	// Build transport with proxy
	parsed, err := url.Parse(proxy.URL)
	if err != nil {
		return false
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(parsed),
		// Tight timeout for health check
		IdleConnTimeout:     5 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// Test with a simple external service
	req, err := http.NewRequest("HEAD", "http://httpbin.org/status/200", nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Accept any 2xx status
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// CheckProxyManual manually triggers a health check for a specific proxy
// Useful for admin operations
func (s *ProxyHealthCheckService) CheckProxyManual(proxyID int) error {
	proxy, err := s.repo.GetByID(proxyID)
	if err != nil {
		return fmt.Errorf("proxy not found: %w", err)
	}

	s.checkProxy(proxy)
	return nil
}

// GetProxyHealth returns the current health status of a proxy
func (s *ProxyHealthCheckService) GetProxyHealth(proxyID int) (models.HealthStatus, error) {
	proxy, err := s.repo.GetByID(proxyID)
	if err != nil {
		return models.HealthStatusDown, err
	}
	return proxy.HealthStatus, nil
}

// IsProxyRecoverable checks if a down proxy is eligible for recovery attempt
func (s *ProxyHealthCheckService) IsProxyRecoverable(proxy *models.Proxy) bool {
	if proxy.HealthStatus != models.HealthStatusDown {
		return false
	}

	// Don't retry if marked down too recently
	if proxy.MarkedDownAt == nil {
		return true // No marked_down_at means old entry, try recovery
	}

	return time.Since(*proxy.MarkedDownAt) > s.recoveryDelay
}
