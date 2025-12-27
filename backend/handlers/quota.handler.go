package handlers

import (
	"aigateway-backend/models"
	"aigateway-backend/repositories"
	"aigateway-backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type QuotaHandler struct {
	quotaService *services.QuotaTrackerService
	accountRepo  *repositories.AccountRepository
	patternRepo  *repositories.QuotaPatternRepository
}

func NewQuotaHandler(
	quotaService *services.QuotaTrackerService,
	accountRepo *repositories.AccountRepository,
	patternRepo *repositories.QuotaPatternRepository,
) *QuotaHandler {
	return &QuotaHandler{
		quotaService: quotaService,
		accountRepo:  accountRepo,
		patternRepo:  patternRepo,
	}
}

// ListAccountsQuota returns quota status for all accounts
func (h *QuotaHandler) ListAccountsQuota(c *gin.Context) {
	providerID := c.Query("provider")

	var accounts []*models.Account
	var err error

	if providerID != "" {
		accounts, err = h.accountRepo.GetActiveByProvider(providerID)
	} else {
		accounts, _, err = h.accountRepo.List(1000, 0)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get patterns for learned limits
	patterns, _ := h.patternRepo.ListAll()
	patternMap := make(map[string]*models.AccountQuotaPattern)
	for _, p := range patterns {
		key := p.AccountID + ":" + p.Model
		patternMap[key] = p
	}

	// Build response with quota status per account
	type AccountQuotaResponse struct {
		AccountID  string                          `json:"account_id"`
		Label      string                          `json:"label"`
		ProviderID string                          `json:"provider_id"`
		Models     map[string]*models.QuotaStatus  `json:"models"`
	}

	result := make([]*AccountQuotaResponse, 0)

	for _, acc := range accounts {
		resp := &AccountQuotaResponse{
			AccountID:  acc.ID,
			Label:      acc.Label,
			ProviderID: acc.ProviderID,
			Models:     make(map[string]*models.QuotaStatus),
		}

		// Get quota status for each model this account has patterns for
		for _, p := range patterns {
			if p.AccountID == acc.ID {
				status := h.quotaService.GetQuotaStatus(acc.ID, p.Model)
				resp.Models[p.Model] = status
			}
		}

		result = append(result, resp)
	}

	c.JSON(http.StatusOK, gin.H{"accounts": result})
}

// GetAccountQuota returns quota status for a specific account
func (h *QuotaHandler) GetAccountQuota(c *gin.Context) {
	accountID := c.Param("id")

	account, err := h.accountRepo.GetByID(accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}

	patterns, _ := h.patternRepo.ListByAccount(accountID)

	modelsStatus := make(map[string]*models.QuotaStatus)
	for _, p := range patterns {
		status := h.quotaService.GetQuotaStatus(accountID, p.Model)
		modelsStatus[p.Model] = status
	}

	c.JSON(http.StatusOK, gin.H{
		"account_id":  account.ID,
		"label":       account.Label,
		"provider_id": account.ProviderID,
		"models":      modelsStatus,
	})
}

// GetProviderSummary returns quota summary for a provider
func (h *QuotaHandler) GetProviderSummary(c *gin.Context) {
	providerID := c.Param("provider")

	accounts, err := h.accountRepo.GetActiveByProvider(providerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	patterns, _ := h.patternRepo.ListByProvider(providerID)

	// Group by model
	modelStats := make(map[string]*models.ModelQuotaStatus)
	totalExhausted := 0

	for _, acc := range accounts {
		for _, p := range patterns {
			if p.AccountID != acc.ID {
				continue
			}

			if _, exists := modelStats[p.Model]; !exists {
				modelStats[p.Model] = &models.ModelQuotaStatus{}
			}

			ms := modelStats[p.Model]
			ms.Total++

			status := h.quotaService.GetQuotaStatus(acc.ID, p.Model)
			if status.IsExhausted {
				ms.Exhausted++
				totalExhausted++
				if status.ResetsAt != nil {
					if ms.NextResetAt == nil || status.ResetsAt.Before(*ms.NextResetAt) {
						ms.NextResetAt = status.ResetsAt
					}
				}
			} else {
				ms.Available++
			}

			if status.PercentUsed != nil {
				if ms.AvgPercentUsed == nil {
					pct := *status.PercentUsed
					ms.AvgPercentUsed = &pct
				} else {
					*ms.AvgPercentUsed = (*ms.AvgPercentUsed + *status.PercentUsed) / 2
				}
			}
		}
	}

	// Calculate health
	totalAccounts := len(accounts)
	availableAccounts := totalAccounts - totalExhausted
	health := calculateHealth(availableAccounts, totalAccounts)

	c.JSON(http.StatusOK, gin.H{
		"provider_id":        providerID,
		"total_accounts":     totalAccounts,
		"available_accounts": availableAccounts,
		"exhausted_accounts": totalExhausted,
		"models":             modelStats,
		"health":             health,
	})
}

// ClearAccountQuota manually clears quota for an account+model
func (h *QuotaHandler) ClearAccountQuota(c *gin.Context) {
	accountID := c.Param("id")
	model := c.Query("model")

	if model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model query parameter required"})
		return
	}

	if err := h.quotaService.ClearQuota(accountID, model); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "quota cleared", "account_id": accountID, "model": model})
}

func calculateHealth(available, total int) string {
	if total == 0 {
		return "unknown"
	}
	ratio := float64(available) / float64(total)
	switch {
	case ratio >= 0.5:
		return "healthy"
	case ratio >= 0.2:
		return "degraded"
	default:
		return "critical"
	}
}
