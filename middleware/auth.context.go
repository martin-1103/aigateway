// middleware/auth.context.go
package middleware

import (
	"aigateway/models"

	"github.com/gin-gonic/gin"
)

const UserContextKey = "current_user"

func SetCurrentUser(c *gin.Context, user *models.User) {
	c.Set(UserContextKey, user)
}

func GetCurrentUser(c *gin.Context) *models.User {
	val, exists := c.Get(UserContextKey)
	if !exists {
		return nil
	}
	user, ok := val.(*models.User)
	if !ok {
		return nil
	}
	return user
}

func GetCurrentUserID(c *gin.Context) string {
	user := GetCurrentUser(c)
	if user == nil {
		return ""
	}
	return user.ID
}

func GetCurrentRole(c *gin.Context) models.Role {
	user := GetCurrentUser(c)
	if user == nil {
		return ""
	}
	return user.Role
}
