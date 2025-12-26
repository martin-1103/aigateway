package handlers

import (
	"aigateway/middleware"
	"aigateway/models"
	"aigateway/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ModelMappingHandler struct {
	service *services.ModelMappingService
}

func NewModelMappingHandler(service *services.ModelMappingService) *ModelMappingHandler {
	return &ModelMappingHandler{service: service}
}

type CreateMappingRequest struct {
	Alias       string `json:"alias" binding:"required"`
	ProviderID  string `json:"provider_id" binding:"required"`
	ModelName   string `json:"model_name" binding:"required"`
	Description string `json:"description"`
	Enabled     *bool  `json:"enabled"`
	Priority    int    `json:"priority"`
	IsGlobal    bool   `json:"is_global"` // Admin only: create global mapping
}

func (h *ModelMappingHandler) Create(c *gin.Context) {
	user := middleware.GetCurrentUser(c)

	var req CreateMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	mapping := &models.ModelMapping{
		Alias:       req.Alias,
		ProviderID:  req.ProviderID,
		ModelName:   req.ModelName,
		Description: req.Description,
		Enabled:     enabled,
		Priority:    req.Priority,
	}

	// Set owner: admin can create global (nil), user creates owned
	if user != nil {
		if user.Role == models.RoleAdmin && req.IsGlobal {
			mapping.OwnerID = nil // Global mapping
		} else {
			mapping.OwnerID = &user.ID // User-owned mapping
		}
	}

	if err := h.service.Create(c.Request.Context(), mapping); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, mapping)
}

func (h *ModelMappingHandler) Get(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	alias := c.Param("alias")

	mapping, err := h.service.GetByAliasWithOwner(alias)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mapping not found"})
		return
	}

	// Check access: admin sees all, user sees global + own
	if user != nil && user.Role != models.RoleAdmin {
		if mapping.OwnerID != nil && *mapping.OwnerID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	c.JSON(http.StatusOK, mapping)
}

func (h *ModelMappingHandler) List(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var mappings []*models.ModelMapping
	var total int64
	var err error

	// Admin sees all, user sees global + own
	if user != nil && user.Role == models.RoleAdmin {
		mappings, total, err = h.service.List(limit, offset)
	} else if user != nil {
		mappings, total, err = h.service.ListForUser(user.ID, limit, offset)
	} else {
		mappings, total, err = h.service.List(limit, offset)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  mappings,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ModelMappingHandler) Update(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	alias := c.Param("alias")

	// Get existing to check ownership
	existing, err := h.service.GetByAliasWithOwner(alias)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mapping not found"})
		return
	}

	// Check ownership: admin can update all, user can only update own
	if user != nil && user.Role != models.RoleAdmin {
		if existing.OwnerID == nil || *existing.OwnerID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	var req CreateMappingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	mapping := &models.ModelMapping{
		Alias:       req.Alias,
		ProviderID:  req.ProviderID,
		ModelName:   req.ModelName,
		Description: req.Description,
		Enabled:     enabled,
		Priority:    req.Priority,
		OwnerID:     existing.OwnerID, // Preserve owner
	}

	if err := h.service.Update(c.Request.Context(), alias, mapping); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapping)
}

func (h *ModelMappingHandler) Delete(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	alias := c.Param("alias")

	// Get existing to check ownership
	existing, err := h.service.GetByAliasWithOwner(alias)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mapping not found"})
		return
	}

	// Check ownership: admin can delete all, user can only delete own
	if user != nil && user.Role != models.RoleAdmin {
		if existing.OwnerID == nil || *existing.OwnerID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	}

	if err := h.service.Delete(c.Request.Context(), alias); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
