package handlers

import (
	"aigateway/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	service *services.StatsQueryService
}

func NewStatsHandler(service *services.StatsQueryService) *StatsHandler {
	return &StatsHandler{service: service}
}

func (h *StatsHandler) GetProxyStats(c *gin.Context) {
	proxyID, _ := strconv.Atoi(c.Param("id"))
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	stats, err := h.service.GetProxyStats(proxyID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

func (h *StatsHandler) GetRecentLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	logs, err := h.service.GetRecentLogs(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}
