// handlers/apikey.handler.go
package handlers

import (
	"net/http"
	"strconv"

	"aigateway/middleware"
	"aigateway/models"
	"aigateway/services"

	"github.com/gin-gonic/gin"
)

type APIKeyHandler struct {
	apiKeyService *services.APIKeyService
}

func NewAPIKeyHandler(apiKeyService *services.APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{apiKeyService: apiKeyService}
}

func (h *APIKeyHandler) List(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	// Admin can filter by user_id
	if user.Role == models.RoleAdmin {
		userID := c.Query("user_id")
		if userID != "" {
			keys, err := h.apiKeyService.ListByUser(userID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"data": keys})
			return
		}

		// List all
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		keys, total, err := h.apiKeyService.ListAll(limit, offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": keys, "total": total})
		return
	}

	// User can only list own keys
	keys, err := h.apiKeyService.ListByUser(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": keys})
}

type CreateAPIKeyRequest struct {
	Label string `json:"label"`
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req CreateAPIKeyRequest
	c.ShouldBindJSON(&req)

	apiKey, rawKey, err := h.apiKeyService.Generate(user.ID, req.Label)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         apiKey.ID,
		"key":        rawKey,
		"key_prefix": apiKey.KeyPrefix,
		"label":      apiKey.Label,
		"message":    "Save this key - it will not be shown again",
	})
}

func (h *APIKeyHandler) Revoke(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	id := c.Param("id")

	// Check ownership unless admin
	apiKey, err := h.apiKeyService.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "api key not found"})
		return
	}

	if user.Role != models.RoleAdmin && apiKey.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not your api key"})
		return
	}

	if err := h.apiKeyService.Revoke(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "api key revoked"})
}
