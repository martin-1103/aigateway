package handlers

import (
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
}

func (h *ModelMappingHandler) Create(c *gin.Context) {
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

	if err := h.service.Create(c.Request.Context(), mapping); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, mapping)
}

func (h *ModelMappingHandler) Get(c *gin.Context) {
	alias := c.Param("alias")

	mapping, err := h.service.GetByAlias(alias)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mapping not found"})
		return
	}

	c.JSON(http.StatusOK, mapping)
}

func (h *ModelMappingHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	mappings, total, err := h.service.List(limit, offset)
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
	alias := c.Param("alias")

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

	if err := h.service.Update(c.Request.Context(), alias, mapping); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mapping)
}

func (h *ModelMappingHandler) Delete(c *gin.Context) {
	alias := c.Param("alias")

	if err := h.service.Delete(c.Request.Context(), alias); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
