package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aigateway-backend/auth/claude"
	"aigateway-backend/auth/codex"
	"aigateway-backend/auth/manager"
	"aigateway-backend/handlers"
	"aigateway-backend/internal/config"
	"aigateway-backend/internal/database"
	"aigateway-backend/middleware"
	"aigateway-backend/providers"
	"aigateway-backend/providers/antigravity"
	"aigateway-backend/providers/glm"
	"aigateway-backend/providers/openai"
	"aigateway-backend/repositories"
	"aigateway-backend/routes"
	"aigateway-backend/services"

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

	// Skip migration if SKIP_MIGRATION env var is set
	if os.Getenv("SKIP_MIGRATION") != "true" {
		if err := database.AutoMigrate(db); err != nil {
			log.Fatalf("Failed to migrate database: %v", err)
		}
		log.Println("Database migration completed successfully")
	} else {
		log.Println("Skipping database migration (SKIP_MIGRATION=true)")
	}

	// Seed default admin user
	if err := database.SeedDefaultAdmin(db); err != nil {
		log.Fatalf("Failed to seed admin: %v", err)
	}

	redis, err := database.NewRedis(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Create background context for services
	ctx := context.Background()

	// Initialize repositories
	accountRepo := repositories.NewAccountRepository(db)
	proxyRepo := repositories.NewProxyRepository(db)
	providerRepo := repositories.NewProviderRepository(db)
	statsRepo := repositories.NewStatsRepository(db)
	modelMappingRepo := repositories.NewModelMappingRepository(db)
	userRepo := repositories.NewUserRepository(db)
	apiKeyRepo := repositories.NewAPIKeyRepository(db)
	quotaPatternRepo := repositories.NewQuotaPatternRepository(db)

	// Initialize core services
	httpClientService := services.NewHTTPClientService()
	errorLogService := services.NewErrorLogService(redis)
	errorLogService.StartCleanupRoutine() // Cleanup logs older than 24h

	accountService := services.NewAccountService(accountRepo, redis)
	proxyService := services.NewProxyService(proxyRepo, accountRepo, &cfg.Proxy)
	accountService.SetProxyService(proxyService) // Wire proxy service for availability checks
	oauthService := services.NewOAuthService(redis, accountRepo, httpClientService, errorLogService)
	oauthFlowService := services.NewOAuthFlowService(redis, accountService, accountRepo, proxyService)

	// Initialize and start token refresh service (legacy)
	tokenRefreshService := services.NewTokenRefreshService(accountRepo, redis)
	go tokenRefreshService.Start()

	proxyHealthService := services.NewProxyHealthService(proxyRepo, redis)
	statsTrackerService := services.NewStatsTrackerService(statsRepo, proxyRepo, redis, proxyHealthService)

	// Initialize proxy health check service (automatic recovery)
	proxyHealthCheckService := services.NewProxyHealthCheckService(proxyRepo, 5, 1440) // Check every 5 min, recover after 1 day down
	proxyHealthCheckService.Start(ctx)
	statsQueryService := services.NewStatsQueryService(statsRepo)
	quotaTrackerService := services.NewQuotaTrackerService(quotaPatternRepo, redis)
	tokenExtractor := services.NewTokenExtractor()
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
		providerRepo,
		accountService,
		proxyService,
		oauthService,
		statsTrackerService,
	)

	// ========================================
	// Initialize Auth Manager (new system)
	// ========================================
	authManager := manager.NewManager(accountRepo, redis)

	// Register token refreshers
	authManager.RegisterRefresher("claude", claude.NewRefresher())
	authManager.RegisterRefresher("codex", codex.NewRefresher())
	// Note: antigravity uses existing tokenRefreshService

	// Wire quota tracker to AuthManager
	authManager.SetQuotaTracker(quotaTrackerService, tokenExtractor)

	// Wire AuthManager to RouterService
	routerService.SetAuthManager(authManager)

	// Wire AuthManager to OAuthFlowService for hot-reload
	oauthFlowService.SetAuthManager(authManager)

	// Start background token refresh (for claude/codex)
	authManager.StartAutoRefresh(ctx, 30*time.Second)

	// Start periodic reconciliation for hot-reload recovery (deferred)
	providerIDs := []string{"antigravity", "claude", "codex"}
	authManager.StartPeriodicReconcile(ctx, 5*time.Minute, providerIDs)

	// Load accounts async after server starts
	go func() {
		time.Sleep(2 * time.Second)
		if err := authManager.LoadAccounts(ctx, "antigravity", "claude", "codex"); err != nil {
			log.Printf("Warning: Failed to load accounts into AuthManager: %v", err)
		}
	}()

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
	logsHandler := handlers.NewLogsHandler(errorLogService)
	modelsHandler := handlers.NewModelsHandler(modelsService)
	modelMappingHandler := handlers.NewModelMappingHandler(modelMappingService)
	authHandler := handlers.NewAuthHandler(authService, userService)
	userHandler := handlers.NewUserHandler(userService)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyService)
	oauthHandler := handlers.NewOAuthHandler(oauthFlowService)
	quotaHandler := handlers.NewQuotaHandler(quotaTrackerService, accountRepo, quotaPatternRepo)

	// Initialize auth status handler (for AuthManager dashboard)
	authStatusHandler := handlers.NewAuthStatusHandler(authManager, authManager.GetMetrics())

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(authService)

	// Setup routes
	r := gin.Default()
	routes.SetupRoutes(
		r,
		cfg,
		proxyHandler,
		accountHandler,
		proxyMgmtHandler,
		statsHandler,
		logsHandler,
		modelsHandler,
		modelMappingHandler,
		authHandler,
		userHandler,
		apiKeyHandler,
		oauthHandler,
		quotaHandler,
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
	authManager.StopPeriodicReconcile()

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
