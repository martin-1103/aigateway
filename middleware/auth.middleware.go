// middleware/auth.middleware.go
package middleware

import (
	"strings"

	"aigateway/services"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	authService *services.AuthService
}

func NewAuthMiddleware(authService *services.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

func (m *AuthMiddleware) ExtractAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try X-API-Key header first
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			user, err := m.authService.ValidateAPIKey(apiKey)
			if err == nil {
				SetCurrentUser(c, user)
			}
			c.Next()
			return
		}

		// Try Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 {
			c.Next()
			return
		}

		scheme := strings.ToLower(parts[0])
		token := parts[1]

		switch scheme {
		case "bearer":
			// Could be JWT or API key
			if strings.HasPrefix(token, "ak_") {
				user, err := m.authService.ValidateAPIKey(token)
				if err == nil {
					SetCurrentUser(c, user)
				}
			} else {
				user, err := m.authService.ValidateJWT(token)
				if err == nil {
					SetCurrentUser(c, user)
				}
			}
		}

		c.Next()
	}
}
