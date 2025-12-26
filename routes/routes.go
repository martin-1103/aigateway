package routes

import (
	"aigateway/handlers"
	"aigateway/middleware"
	"aigateway/models"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	proxyHandler *handlers.ProxyHandler,
	accountHandler *handlers.AccountHandler,
	proxyMgmtHandler *handlers.ProxyManagementHandler,
	statsHandler *handlers.StatsHandler,
	modelsHandler *handlers.ModelsHandler,
	modelMappingHandler *handlers.ModelMappingHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	apiKeyHandler *handlers.APIKeyHandler,
	oauthHandler *handlers.OAuthHandler,
	authMiddleware *middleware.AuthMiddleware,
) {
	// Apply global auth extraction
	r.Use(authMiddleware.ExtractAuth())

	// Health check endpoint (public)
	r.GET("/health", proxyHandler.HealthCheck)

	// Public models endpoint
	r.GET("/v1/models", modelsHandler.GetModels)

	// AI model proxy endpoints (require auth with AI access)
	r.POST("/v1/messages", middleware.RequireAIAccess(), proxyHandler.HandleProxy)
	r.POST("/v1/chat/completions", middleware.RequireAIAccess(), proxyHandler.HandleProxy)

	api := r.Group("/api/v1")
	{
		// Auth endpoints (public)
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", middleware.RequireAuth(), authHandler.Me)
			auth.PUT("/password", middleware.RequireAuth(), authHandler.ChangePassword)
		}

		// User endpoints (admin only)
		users := api.Group("/users")
		users.Use(middleware.RequireAdmin())
		{
			users.GET("", userHandler.List)
			users.GET("/:id", userHandler.Get)
			users.POST("", userHandler.Create)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
		}

		// API Key endpoints (admin + user)
		apiKeys := api.Group("/api-keys")
		apiKeys.Use(middleware.RequireRole(models.RoleAdmin, models.RoleUser))
		{
			apiKeys.GET("", apiKeyHandler.List)
			apiKeys.POST("", apiKeyHandler.Create)
			apiKeys.DELETE("/:id", apiKeyHandler.Revoke)
		}

		// Provider endpoints (public for now)
		api.GET("/providers", proxyHandler.GetProviders)

		// Account endpoints (admin + provider)
		accounts := api.Group("/accounts")
		accounts.Use(middleware.RequireAccountAccess())
		{
			accounts.GET("", accountHandler.List)
			accounts.GET("/:id", accountHandler.Get)
			accounts.POST("", accountHandler.Create)
			accounts.PUT("/:id", accountHandler.Update)
			accounts.DELETE("/:id", accountHandler.Delete)
		}

		// Proxy endpoints (admin only)
		proxies := api.Group("/proxies")
		proxies.Use(middleware.RequireAdmin())
		{
			proxies.GET("", proxyMgmtHandler.List)
			proxies.GET("/:id", proxyMgmtHandler.Get)
			proxies.POST("", proxyMgmtHandler.Create)
			proxies.PUT("/:id", proxyMgmtHandler.Update)
			proxies.DELETE("/:id", proxyMgmtHandler.Delete)
			proxies.GET("/assignments", proxyMgmtHandler.GetAssignments)
			proxies.POST("/recalculate", proxyMgmtHandler.RecalculateCounts)
		}

		// Stats endpoints (admin + user, filtered by role in handler)
		stats := api.Group("/stats")
		stats.Use(middleware.RequireRole(models.RoleAdmin, models.RoleUser))
		{
			stats.GET("/proxies/:id", statsHandler.GetProxyStats)
			stats.GET("/logs", statsHandler.GetRecentLogs)
		}

		// Model mapping endpoints (admin + user)
		mappings := api.Group("/model-mappings")
		mappings.Use(middleware.RequireRole(models.RoleAdmin, models.RoleUser))
		{
			mappings.GET("", modelMappingHandler.List)
			mappings.GET("/:alias", modelMappingHandler.Get)
			mappings.POST("", modelMappingHandler.Create)
			mappings.PUT("/:alias", modelMappingHandler.Update)
			mappings.DELETE("/:alias", modelMappingHandler.Delete)
		}

		// OAuth endpoints
		oauth := api.Group("/oauth")
		{
			// Public endpoints
			oauth.GET("/providers", oauthHandler.GetProviders)
			oauth.GET("/callback", oauthHandler.Callback)

			// Protected endpoints (admin + user)
			oauth.POST("/init", middleware.RequireAccountAccess(), oauthHandler.InitFlow)
			oauth.POST("/exchange", middleware.RequireAccountAccess(), oauthHandler.Exchange)
			oauth.POST("/refresh", middleware.RequireAccountAccess(), oauthHandler.RefreshToken)
		}
	}
}
