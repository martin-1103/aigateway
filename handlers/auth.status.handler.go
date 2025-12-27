package handlers

import (
	"net/http"
	"time"

	"aigateway/auth/manager"

	"github.com/gin-gonic/gin"
)

// AuthStatusHandler handles auth manager status endpoints
type AuthStatusHandler struct {
	manager *manager.Manager
	metrics *manager.Metrics
}

// NewAuthStatusHandler creates a new auth status handler
func NewAuthStatusHandler(m *manager.Manager, metrics *manager.Metrics) *AuthStatusHandler {
	return &AuthStatusHandler{
		manager: m,
		metrics: metrics,
	}
}

// GetAccountsStatus returns status of all accounts
// GET /api/v1/auth/accounts
func (h *AuthStatusHandler) GetAccountsStatus(c *gin.Context) {
	if h.manager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "auth manager not initialized",
		})
		return
	}

	accounts := h.manager.GetAllAccounts()
	now := time.Now()

	result := make([]AccountStatusResponse, 0, len(accounts))
	for _, acc := range accounts {
		status := h.buildAccountStatus(acc, now)
		result = append(result, status)
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts":   result,
		"total":      len(result),
		"checked_at": now.Format(time.RFC3339),
	})
}

// GetAccountStatus returns status of a specific account
// GET /api/v1/auth/accounts/:id
func (h *AuthStatusHandler) GetAccountStatus(c *gin.Context) {
	accountID := c.Param("id")

	if h.manager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "auth manager not initialized",
		})
		return
	}

	acc := h.manager.GetAccount(accountID)
	if acc == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "account not found",
		})
		return
	}

	now := time.Now()
	status := h.buildAccountStatus(acc, now)

	c.JSON(http.StatusOK, status)
}

// GetMetrics returns auth manager metrics
// GET /api/v1/auth/metrics
func (h *AuthStatusHandler) GetMetrics(c *gin.Context) {
	if h.metrics == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "metrics not initialized",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metrics":    h.metrics.Summary(),
		"checked_at": time.Now().Format(time.RFC3339),
	})
}

// GetHealthSummary returns a health summary
// GET /api/v1/auth/health
func (h *AuthStatusHandler) GetHealthSummary(c *gin.Context) {
	if h.manager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"error":  "auth manager not initialized",
		})
		return
	}

	accounts := h.manager.GetAllAccounts()
	now := time.Now()

	var total, healthy, blocked, disabled int
	providerStats := make(map[string]*ProviderHealthStats)

	for _, acc := range accounts {
		total++

		stats, ok := providerStats[acc.Account.ProviderID]
		if !ok {
			stats = &ProviderHealthStats{ProviderID: acc.Account.ProviderID}
			providerStats[acc.Account.ProviderID] = stats
		}
		stats.Total++

		if acc.Disabled {
			disabled++
			stats.Disabled++
			continue
		}

		isBlocked := false
		for model := range acc.ModelStates {
			if b, _ := acc.IsBlockedFor(model, now); b {
				isBlocked = true
				break
			}
		}

		if isBlocked {
			blocked++
			stats.Blocked++
		} else {
			healthy++
			stats.Healthy++
		}
	}

	status := "healthy"
	if healthy == 0 && total > 0 {
		status = "degraded"
	}
	if disabled == total && total > 0 {
		status = "critical"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         status,
		"total":          total,
		"healthy":        healthy,
		"blocked":        blocked,
		"disabled":       disabled,
		"provider_stats": providerStats,
		"checked_at":     now.Format(time.RFC3339),
	})
}

func (h *AuthStatusHandler) buildAccountStatus(acc *manager.AccountState, now time.Time) AccountStatusResponse {
	modelStatuses := make(map[string]ModelStatusResponse)

	acc.Account = acc.Account // ensure not nil

	for model, ms := range acc.ModelStates {
		blocked, reason := acc.IsBlockedFor(model, now)
		modelStatuses[model] = ModelStatusResponse{
			Model:          model,
			IsBlocked:      blocked,
			BlockReason:    string(reason),
			NextRetryAfter: formatTime(ms.NextRetryAfter),
			SuccessCount:   ms.SuccessCount,
			FailureCount:   ms.FailureCount,
			LastUsedAt:     formatTime(ms.LastUsedAt),
		}
	}

	return AccountStatusResponse{
		ID:          acc.Account.ID,
		ProviderID:  acc.Account.ProviderID,
		Label:       acc.Account.Label,
		IsDisabled:  acc.Disabled,
		ModelStates: modelStatuses,
		UpdatedAt:   formatTime(acc.UpdatedAt),
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// Response types

// AccountStatusResponse represents account status in API response
type AccountStatusResponse struct {
	ID          string                       `json:"id"`
	ProviderID  string                       `json:"provider_id"`
	Label       string                       `json:"label"`
	IsDisabled  bool                         `json:"is_disabled"`
	ModelStates map[string]ModelStatusResponse `json:"model_states"`
	UpdatedAt   string                       `json:"updated_at"`
}

// ModelStatusResponse represents model status in API response
type ModelStatusResponse struct {
	Model          string `json:"model"`
	IsBlocked      bool   `json:"is_blocked"`
	BlockReason    string `json:"block_reason,omitempty"`
	NextRetryAfter string `json:"next_retry_after,omitempty"`
	SuccessCount   int64  `json:"success_count"`
	FailureCount   int64  `json:"failure_count"`
	LastUsedAt     string `json:"last_used_at,omitempty"`
}

// ProviderHealthStats represents health stats per provider
type ProviderHealthStats struct {
	ProviderID string `json:"provider_id"`
	Total      int    `json:"total"`
	Healthy    int    `json:"healthy"`
	Blocked    int    `json:"blocked"`
	Disabled   int    `json:"disabled"`
}
