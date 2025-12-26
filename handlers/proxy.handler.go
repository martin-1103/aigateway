package handlers

import (
	"context"
	"io"
	"net/http"

	"aigateway/services"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type ProxyHandler struct {
	executor      *services.ExecutorService
	routerService *services.RouterService
}

func NewProxyHandler(executor *services.ExecutorService, routerService *services.RouterService) *ProxyHandler {
	return &ProxyHandler{
		executor:      executor,
		routerService: routerService,
	}
}

// HandleProxy processes incoming AI model requests and routes them to appropriate providers
func (h *ProxyHandler) HandleProxy(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	model := gjson.GetBytes(body, "model").String()
	if model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "model is required"})
		return
	}

	stream := c.Query("stream") == "true"
	if !stream {
		streamField := gjson.GetBytes(body, "stream")
		if streamField.Exists() {
			stream = streamField.Bool()
		}
	}

	req := services.Request{
		Model:   model,
		Payload: body,
		Stream:  stream,
	}

	ctx := context.Background()
	resp, err := h.executor.Execute(ctx, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if resp.StatusCode > 0 {
			statusCode = resp.StatusCode
		}

		if len(resp.Payload) > 0 {
			c.Data(statusCode, "application/json", resp.Payload)
		} else {
			c.JSON(statusCode, gin.H{"error": err.Error()})
		}
		return
	}

	c.Data(resp.StatusCode, "application/json", resp.Payload)
}

// GetProviders returns list of all registered providers
func (h *ProxyHandler) GetProviders(c *gin.Context) {
	providers := h.routerService.ListProviders()
	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
		"total":     len(providers),
	})
}

// HealthCheck returns service health status
func (h *ProxyHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"service": "aigateway",
	})
}
