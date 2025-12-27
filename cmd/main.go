package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aigateway/auth/claude"
	"aigateway/auth/codex"
	"aigateway/auth/manager"
	"aigateway/config"
	"aigateway/database"
	"aigateway/handlers"
	"aigateway/middleware"
	"aigateway/providers"
	"aigateway/providers/antigravity"
	"aigateway/providers/glm"
	"aigateway/providers/openai"
	"aigateway/repositories"
	"aigateway/routes"
	"aigateway/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewMySQL(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}

	if err := database.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Seed default admin user
	if err := database.SeedDefaultAdmin(db); err != nil {
		log.Fatalf("Failed to seed admin: %v", err)
	}

	redis, err := database.NewRedis(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize repositories
	accountRepo := repositories.NewAccountRepository(db)
	proxyRepo := repositories.NewProxyRepository(db)
	statsRepo := repositories.NewStatsRepository(db)
	modelMappingRepo := repositories.NewModelMappingRepository(db)
	userRepo := repositories.NewUserRepository(db)
	apiKeyRepo := repositories.NewAPIKeyRepository(db)

	// Initialize core services
	httpClientService := services.NewHTTPClientService()
	accountService := services.NewAccountService(accountRepo, redis)
	proxyService := services.NewProxyService(proxyRepo, accountRepo, &cfg.Proxy)
	accountService.SetProxyService(proxyService) // Wire proxy service for availability checks
	oauthService := services.NewOAuthService(redis, accountRepo, httpClientService)
	oauthFlowService := services.NewOAuthFlowService(redis, accountService, accountRepo, proxyService)

	// Initialize and start token refresh service (legacy)
	tokenRefreshService := services.NewTokenRefreshService(accountRepo, redis)
	go tokenRefreshService.Start()

	proxyHealthService := services.NewProxyHealthService(proxyRepo, redis)
	statsTrackerService := services.NewStatsTrackerService(statsRepo, proxyRepo, redis, proxyHealthService)
	statsQueryService := services.NewStatsQueryService(statsRepo)
	modelsService := services.NewModelsService(db, redis)
	modelMappingService := services.NewModelMappingService(modelMappingRepo, redis)

	// Initialize RBAC services
	passwordService := services.NewPasswordService()
	jwtService := services.NewJWTService(cfg.Server.JWTSecret)
	userService := services.NewUserService(userRepo, passwordService)
	apiKeyService := services.NewAPIKeyService(apiKeyRepo, redis)
	authService := services.NewAuthService(userService, jwtService, apiKeyService)

	// Initialize providers
	antigravityProvider := antigravity.NewAntigravityProvider()
	openaiProvider := openai.NewOpenAIProvider()
	glmProvider := glm.NewProvider()

	// Initialize provider registry
	registry := providers.NewRegistry()
	registry.Register("antigravity", antigravityProvider)
	registry.Register("openai", openaiProvider)
	registry.Register("glm", glmProvider)

	// Set custom model mapping resolver
	registry.SetMappingResolver(modelMappingService)

	// Initialize router service
	routerService := services.NewRouterService(
		registry,
		accountService,
		proxyService,
		oauthService,
		statsTrackerService,
	)

	// ========================================
	// Initialize Auth Manager (new system)
	// ========================================
	ctx := context.Background()
	authManager := manager.NewManager(accountRepo, redis)

	// Register token refreshers
	authManager.RegisterRefresher("claude", claude.NewRefresher())
	authManager.RegisterRefresher("codex", codex.NewRefresher())
	// Note: antigravity uses existing tokenRefreshService

	// Load accounts for all OAuth providers
	if err := authManager.LoadAccounts(ctx, "antigravity", "claude", "codex"); err != nil {
		log.Printf("Warning: Failed to load accounts into AuthManager: %v", err)
	}

	// Start background token refresh (for claude/codex)
	authManager.StartAutoRefresh(ctx, 30*time.Second)

	// Wire AuthManager to RouterService
	routerService.SetAuthManager(authManager)

	// Enable AuthManager for account selection (feature flag)
	// Set to true to use health-aware selection with retry
	useAuthManager := os.Getenv("USE_AUTH_MANAGER") == "true"
	routerService.EnableAuthManager(useAuthManager)
	if useAuthManager {
		log.Println("AuthManager enabled for health-aware account selection")
	}

	// ========================================

	// Initialize executor service with router
	executorService := services.NewExecutorService(
		routerService,
		accountService,
		proxyService,
		oauthService,
		statsTrackerService,
	)

	// Initialize handlers
	proxyHandler := handlers.NewProxyHandler(executorService, routerService)
	accountHandler := handlers.NewAccountHandler(accountService)
	proxyMgmtHandler := handlers.NewProxyManagementHandler(proxyService)
	statsHandler := handlers.NewStatsHandler(statsQueryService)
	modelsHandler := handlers.NewModelsHandler(modelsService)
	modelMappingHandler := handlers.NewModelMappingHandler(modelMappingService)
	authHandler := handlers.NewAuthHandler(authService, userService)
	userHandler := handlers.NewUserHandler(userService)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService)
	oauthHandler := handlers.NewOAuthHandler(oauthFlowService)

	// Initialize auth status handler (for AuthManager dashboard)
	authStatusHandler := handlers.NewAuthStatusHandler(authManager, authManager.GetMetrics())

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Setup routes
	r := gin.Default()
	routes.SetupRoutes(
		r,
		proxyHandler,
		accountHandler,
		proxyMgmtHandler,
		statsHandler,
		modelsHandler,
		modelMappingHandler,
		authHandler,
		userHandler,
		apiKeyHandler,
		oauthHandler,
		authMiddleware,
	)

	// Setup AuthManager status routes
	setupAuthStatusRoutes(r, authStatusHandler, authMiddleware)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server starting on %s", addr)

	go func() {
		if err := r.Run(addr); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-quit
	log.Println("Shutting down server...")

	// Stop background services
	tokenRefreshService.Stop()
	authManager.StopAutoRefresh()

	log.Println("Server exited")
}

// setupAuthStatusRoutes registers AuthManager status endpoints
func setupAuthStatusRoutes(r *gin.Engine, h *handlers.AuthStatusHandler, authMiddleware *middleware.AuthMiddleware) {
	authStatus := r.Group("/api/v1/auth-manager")
	authStatus.Use(authMiddleware.ExtractAuth())
	authStatus.Use(middleware.RequireAdmin())
	{
		authStatus.GET("/accounts", h.GetAccountsStatus)
		authStatus.GET("/accounts/:id", h.GetAccountStatus)
		authStatus.GET("/metrics", h.GetMetrics)
		authStatus.GET("/health", h.GetHealthSummary)
	}
}
