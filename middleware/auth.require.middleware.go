// middleware/auth.require.middleware.go
package middleware

import (
	"net/http"

	"aigateway/models"

	"github.com/gin-gonic/gin"
)

func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}
		c.Next()
	}
}

func RequireRole(roles ...models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := GetCurrentUser(c)
		if user == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}

		for _, role := range roles {
			if user.Role == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "insufficient permissions",
		})
	}
}

func RequireAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin)
}

func RequireAIAccess() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin, models.RoleUser)
}

func RequireAccountAccess() gin.HandlerFunc {
	return RequireRole(models.RoleAdmin, models.RoleProvider, models.RoleUser)
}
