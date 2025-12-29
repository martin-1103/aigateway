package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"aigateway-backend/services"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type ProxyHandler struct {
	executor      *services.ExecutorService
	routerService *services.RouterService
	startTime     time.Time
	version       string
	authManagerEnabled bool
}

func NewProxyHandler(executor *services.ExecutorService, routerService *services.RouterService) *ProxyHandler {
	return &ProxyHandler{
		executor:      executor,
		routerService: routerService,
		startTime:     time.Now(),
	}
}

func (h *ProxyHandler) SetBuildInfo(version string, authManagerEnabled bool) {
	h.version = version
	h.authManagerEnabled = authManagerEnabled
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

	accountID := c.Query("account_id")

	req := services.Request{
		Model:     model,
		Payload:   body,
		Stream:    stream,
		AccountID: accountID,
	}

	ctx := context.Background()

	// Handle streaming vs non-streaming
	if stream {
		h.handleStreaming(c, ctx, req)
	} else {
		h.handleNonStreaming(c, ctx, req)
	}
}

// handleNonStreaming handles regular non-streaming requests
func (h *ProxyHandler) handleNonStreaming(c *gin.Context, ctx context.Context, req services.Request) {
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

// handleStreaming handles streaming requests
func (h *ProxyHandler) handleStreaming(c *gin.Context, ctx context.Context, req services.Request) {
	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Execute streaming request
	streamResp, err := h.executor.ExecuteStream(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check status code
	if streamResp.StatusCode < 200 || streamResp.StatusCode >= 300 {
		c.JSON(streamResp.StatusCode, gin.H{"error": "upstream error"})
		return
	}

	// Forward stream to client
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming not supported"})
		return
	}

	// Forward all chunks
	for {
		select {
		case data, ok := <-streamResp.DataCh:
			if !ok {
				return
			}

			// Write chunk directly (already in SSE format from translator)
			if _, err := c.Writer.Write(data); err != nil {
				return
			}
			flusher.Flush()

		case err := <-streamResp.ErrCh:
			if err != nil {
				c.Writer.Write([]byte(fmt.Sprintf("event: error\ndata: {\"error\": \"%s\"}\n\n", err.Error())))
				flusher.Flush()
			}
			return

		case <-streamResp.Done:
			return

		case <-c.Request.Context().Done():
			return
		}
	}
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
	uptime := time.Since(h.startTime)

	response := gin.H{
		"status":  "ok",
		"service": "aigateway",
		"started_at": h.startTime.Format(time.RFC3339),
		"uptime_seconds": int(uptime.Seconds()),
		"auth_manager_enabled": h.authManagerEnabled,
	}

	if h.version != "" {
		response["version"] = h.version
	}

	c.JSON(http.StatusOK, response)
}
