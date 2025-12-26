package handlers

import (
	"aigateway/models"
	"aigateway/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProxyManagementHandler struct {
	service *services.ProxyService
}

func NewProxyManagementHandler(service *services.ProxyService) *ProxyManagementHandler {
	return &ProxyManagementHandler{service: service}
}

func (h *ProxyManagementHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	proxies, total, err := h.service.List(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   proxies,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *ProxyManagementHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	proxy, err := h.service.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "proxy not found"})
		return
	}

	c.JSON(http.StatusOK, proxy)
}

func (h *ProxyManagementHandler) Create(c *gin.Context) {
	var proxy models.Proxy
	if err := c.ShouldBindJSON(&proxy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Create(&proxy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, proxy)
}

func (h *ProxyManagementHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var proxy models.Proxy
	if err := c.ShouldBindJSON(&proxy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	proxy.ID = id
	if err := h.service.Update(&proxy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, proxy)
}

func (h *ProxyManagementHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.service.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "proxy deleted"})
}

func (h *ProxyManagementHandler) GetAssignments(c *gin.Context) {
	assignments, err := h.service.GetAssignments()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"assignments": assignments})
}

func (h *ProxyManagementHandler) RecalculateCounts(c *gin.Context) {
	if err := h.service.RecalculateCounts(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "counts recalculated"})
}
