package main

import (
	"aigateway/config"
	"aigateway/database"
	"aigateway/handlers"
	"aigateway/providers"
	"aigateway/providers/antigravity"
	"aigateway/providers/glm"
	"aigateway/providers/openai"
	"aigateway/repositories"
	"aigateway/routes"
	"aigateway/services"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	redis, err := database.NewRedis(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Initialize repositories
	accountRepo := repositories.NewAccountRepository(db)
	proxyRepo := repositories.NewProxyRepository(db)
	statsRepo := repositories.NewStatsRepository(db)
	modelMappingRepo := repositories.NewModelMappingRepository(db)

	// Initialize core services
	accountService := services.NewAccountService(accountRepo, redis)
	proxyService := services.NewProxyService(proxyRepo, accountRepo)
	oauthService := services.NewOAuthService(redis, accountRepo)

	// Initialize and start token refresh service
	tokenRefreshService := services.NewTokenRefreshService(accountRepo, redis)
	go tokenRefreshService.Start()

	proxyHealthService := services.NewProxyHealthService(proxyRepo, redis)
	statsTrackerService := services.NewStatsTrackerService(statsRepo, proxyRepo, redis, proxyHealthService)
	statsQueryService := services.NewStatsQueryService(statsRepo)
	modelsService := services.NewModelsService(db, redis)
	modelMappingService := services.NewModelMappingService(modelMappingRepo, redis)

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

	// Setup routes
	r := gin.Default()
	routes.SetupRoutes(r, proxyHandler, accountHandler, proxyMgmtHandler, statsHandler, modelsHandler, modelMappingHandler)

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

	log.Println("Server exited")
}
