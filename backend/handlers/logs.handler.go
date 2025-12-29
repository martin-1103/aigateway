package handlers

import (
	"net/http"
	"strconv"
	"time"

	"aigateway-backend/services"

	"github.com/gin-gonic/gin"
)

type LogsHandler struct {
	errorLogService *services.ErrorLogService
}

func NewLogsHandler(errorLogService *services.ErrorLogService) *LogsHandler {
	return &LogsHandler{errorLogService: errorLogService}
}

func (h *LogsHandler) GetRecentErrors(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	entries, err := h.errorLogService.GetRecent(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  entries,
		"total": len(entries),
	})
}

func (h *LogsHandler) GetErrorsByTimeRange(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")
	limitStr := c.DefaultQuery("limit", "100")

	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	var from, to time.Time
	var err error

	if fromStr != "" {
		from, err = time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' date format, use RFC3339"})
			return
		}
	} else {
		from = time.Now().Add(-24 * time.Hour)
	}

	if toStr != "" {
		to, err = time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' date format, use RFC3339"})
			return
		}
	} else {
		to = time.Now()
	}

	entries, err := h.errorLogService.GetByTimeRange(from, to, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":  entries,
		"total": len(entries),
		"from":  from.Format(time.RFC3339),
		"to":    to.Format(time.RFC3339),
	})
}

func (h *LogsHandler) CleanupOldLogs(c *gin.Context) {
	if err := h.errorLogService.Cleanup(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cleanup completed"})
}
