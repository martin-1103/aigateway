package handlers

import (
	"fmt"
	"net/http"

	"aigateway-backend/middleware"
	"aigateway-backend/services"

	"github.com/gin-gonic/gin"
)

type LiteHandler struct {
	accountService *services.AccountService
	apiKeyService  *services.APIKeyService
	oauthService   *services.OAuthFlowService
}

func NewLiteHandler(
	accountService *services.AccountService,
	apiKeyService *services.APIKeyService,
	oauthService *services.OAuthFlowService,
) *LiteHandler {
	return &LiteHandler{
		accountService: accountService,
		apiKeyService:  apiKeyService,
		oauthService:   oauthService,
	}
}

func (h *LiteHandler) Me(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	})
}

func (h *LiteHandler) ListAccounts(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	accounts, total, err := h.accountService.ListByCreator(user.ID, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  accounts,
		"total": total,
	})
}

func (h *LiteHandler) ListAPIKeys(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	if !user.Role.CanManageAPIKeys() {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	keys, err := h.apiKeyService.ListByUser(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": keys})
}

func (h *LiteHandler) InitOAuth(c *gin.Context) {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req services.InitFlowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.CreatedBy = &user.ID

	// Store access key in state for callback redirect
	accessKey := c.Query("key")
	if accessKey == "" {
		accessKey = c.GetHeader("X-Access-Key")
	}
	req.LiteAccessKey = accessKey

	resp, err := h.oauthService.InitFlow(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *LiteHandler) OAuthCallback(c *gin.Context) {
	callbackURL := c.Request.URL.String()

	resp, err := h.oauthService.ExchangeCode(c.Request.Context(), callbackURL)
	if err != nil {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(h.errorHTML(err.Error())))
		return
	}

	// Extract access key from state to redirect back to lite dashboard
	accessKey := h.oauthService.GetAccessKeyFromState(c.Request.Context(), c.Query("state"))

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(h.successHTML(resp.Account.Label, accessKey)))
}

func (h *LiteHandler) GetOAuthProviders(c *gin.Context) {
	providers := h.oauthService.GetProviders()
	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

func (h *LiteHandler) successHTML(accountName, accessKey string) string {
	redirectURL := "/lite"
	if accessKey != "" {
		redirectURL = fmt.Sprintf("/lite?key=%s", accessKey)
	}

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
        .icon { font-size: 4rem; margin-bottom: 1rem; }
        h1 { color: #2d3748; margin: 0 0 1rem; font-size: 1.5rem; }
        p { color: #718096; margin: 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">✓</div>
        <h1>Authentication Successful</h1>
        <p>Account "` + accountName + `" has been connected.</p>
        <p>Redirecting back to dashboard...</p>
    </div>
    <script>
        setTimeout(() => {
            window.location.href = '` + redirectURL + `';
        }, 2000);
    </script>
</body>
</html>`
}

func (h *LiteHandler) errorHTML(errorMsg string) string {
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
        .icon { font-size: 4rem; margin-bottom: 1rem; }
        h1 { color: #2d3748; margin: 0 0 1rem; font-size: 1.5rem; }
        p { color: #718096; margin: 0.5rem 0; }
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
        <div class="error">` + errorMsg + `</div>
    </div>
</body>
</html>`
}
