package handlers

import (
	"aigateway-backend/middleware"
	"aigateway-backend/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

// OAuthHandler handles OAuth authorization flow endpoints
type OAuthHandler struct {
	service *services.OAuthFlowService
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(service *services.OAuthFlowService) *OAuthHandler {
	return &OAuthHandler{service: service}
}

// InitFlow starts OAuth authorization flow
// POST /api/v1/oauth/init
func (h *OAuthHandler) InitFlow(c *gin.Context) {
	var req services.InitFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := middleware.GetCurrentUser(c)
	if user != nil {
		req.CreatedBy = &user.ID
	}

	resp, err := h.service.InitFlow(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Callback handles OAuth callback for automatic flow
// GET /api/v1/oauth/callback
func (h *OAuthHandler) Callback(c *gin.Context) {
	callbackURL := c.Request.URL.String()

	resp, err := h.service.ExchangeCode(c.Request.Context(), callbackURL)
	if err != nil {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(h.errorHTML(err.Error())))
		return
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(h.successHTML(resp.Account.Label)))
}

// Exchange handles manual OAuth flow by parsing pasted callback URL
// POST /api/v1/oauth/exchange
func (h *OAuthHandler) Exchange(c *gin.Context) {
	var req services.ExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.ExchangeCode(c.Request.Context(), req.CallbackURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetProviders returns list of available OAuth providers
// GET /api/v1/oauth/providers
func (h *OAuthHandler) GetProviders(c *gin.Context) {
	providers := h.service.GetProviders()
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// RefreshToken manually refreshes an account's OAuth token
// POST /api/v1/oauth/refresh
func (h *OAuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		AccountID string `json:"account_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.RefreshToken(c.Request.Context(), req.AccountID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "token refreshed successfully"})
}

// successHTML returns HTML that closes popup and notifies parent
func (h *OAuthHandler) successHTML(accountName string) string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>OAuth Success</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            display: flex;
            align-items: center;
            justify-content: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        }
        .container {
            text-align: center;
            background: white;
            padding: 3rem;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
        }
        .icon {
            font-size: 4rem;
            margin-bottom: 1rem;
        }
        h1 {
            color: #2d3748;
            margin: 0 0 1rem;
            font-size: 1.5rem;
        }
        p {
            color: #718096;
            margin: 0;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">✓</div>
        <h1>Authentication Successful</h1>
        <p>Account "` + accountName + `" has been connected.</p>
        <p>This window will close automatically.</p>
    </div>
    <script>
        if (window.opener) {
            window.opener.postMessage({ type: 'oauth_success', account: '` + accountName + `' }, '*');
        }
        setTimeout(() => window.close(), 2000);
    </script>
</body>
</html>`
}

// errorHTML returns HTML that shows error and closes popup
func (h *OAuthHandler) errorHTML(errorMsg string) string {
	return `<!DOCTYPE html>
<html>
<head>
    <title>OAuth Error</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            display: flex;
            align-items: center;
            justify-content: center;
            height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
        }
        .container {
            text-align: center;
            background: white;
            padding: 3rem;
            border-radius: 12px;
            box-shadow: 0 10px 40px rgba(0,0,0,0.2);
            max-width: 500px;
        }
        .icon {
            font-size: 4rem;
            margin-bottom: 1rem;
        }
        h1 {
            color: #2d3748;
            margin: 0 0 1rem;
            font-size: 1.5rem;
        }
        p {
            color: #718096;
            margin: 0.5rem 0;
        }
        .error {
            color: #e53e3e;
            font-family: monospace;
            font-size: 0.875rem;
            background: #fff5f5;
            padding: 1rem;
            border-radius: 6px;
            margin-top: 1rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">✗</div>
        <h1>Authentication Failed</h1>
        <p>An error occurred during OAuth authentication.</p>
        <div class="error">` + errorMsg + `</div>
        <p style="margin-top: 1rem;">This window will close automatically.</p>
    </div>
    <script>
        if (window.opener) {
            window.opener.postMessage({ type: 'oauth_error', error: '` + errorMsg + `' }, '*');
        }
        setTimeout(() => window.close(), 3000);
    </script>
</body>
</html>`
}
