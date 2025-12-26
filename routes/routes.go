package routes

import (
	"aigateway/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	proxyHandler *handlers.ProxyHandler,
	accountHandler *handlers.AccountHandler,
	proxyMgmtHandler *handlers.ProxyManagementHandler,
	statsHandler *handlers.StatsHandler,
) {
	// Health check endpoint
	r.GET("/health", proxyHandler.HealthCheck)

	// AI model proxy endpoints
	r.POST("/v1/messages", proxyHandler.HandleProxy)
	r.POST("/v1/chat/completions", proxyHandler.HandleProxy)

	api := r.Group("/api/v1")
	{
		// Provider endpoints
		api.GET("/providers", proxyHandler.GetProviders)

		accounts := api.Group("/accounts")
		{
			accounts.GET("", accountHandler.List)
			accounts.GET("/:id", accountHandler.Get)
			accounts.POST("", accountHandler.Create)
			accounts.PUT("/:id", accountHandler.Update)
			accounts.DELETE("/:id", accountHandler.Delete)
		}

		proxies := api.Group("/proxies")
		{
			proxies.GET("", proxyMgmtHandler.List)
			proxies.GET("/:id", proxyMgmtHandler.Get)
			proxies.POST("", proxyMgmtHandler.Create)
			proxies.PUT("/:id", proxyMgmtHandler.Update)
			proxies.DELETE("/:id", proxyMgmtHandler.Delete)
			proxies.GET("/assignments", proxyMgmtHandler.GetAssignments)
			proxies.POST("/recalculate", proxyMgmtHandler.RecalculateCounts)
		}

		stats := api.Group("/stats")
		{
			stats.GET("/proxies/:id", statsHandler.GetProxyStats)
			stats.GET("/logs", statsHandler.GetRecentLogs)
		}
	}
}
