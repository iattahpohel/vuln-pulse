package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atta/vulnpulse/internal/api/handlers"
	"github.com/atta/vulnpulse/internal/api/routes"
	"github.com/atta/vulnpulse/internal/repository"
	"github.com/atta/vulnpulse/pkg/auth"
	"github.com/atta/vulnpulse/pkg/cache"
	"github.com/atta/vulnpulse/pkg/config"
	"github.com/atta/vulnpulse/pkg/database"
	"github.com/atta/vulnpulse/pkg/logger"
	"github.com/atta/vulnpulse/pkg/queue"
)

func main() {
	// Initialize logger
	log := logger.New("api")
	log.Info("starting VulnPulse API server")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// Connect to database
	db, err := database.Connect(
		ctx,
		cfg.Database.URL,
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
		cfg.Database.ConnMaxLifetime,
	)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("connected to database")

	// Run migrations
	if err := database.RunMigrations(ctx, db); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	log.Info("migrations completed")

	// Connect to Redis
	redisClient, err := cache.Connect(ctx, cfg.Redis.URL, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	defer redisClient.Close()
	log.Info("connected to Redis")

	// Connect to RabbitMQ
	queueClient, err := queue.Connect(cfg.RabbitMQ.URL)
	if err != nil {
		log.Error("failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer queueClient.Close()
	log.Info("connected to RabbitMQ")

	// Declare queue
	if err := queueClient.DeclareQueue(cfg.RabbitMQ.QueueName); err != nil {
		log.Error("failed to declare queue", "error", err)
		os.Exit(1)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	vulnRepo := repository.NewVulnerabilityRepository(db)
	alertRepo := repository.NewAlertRepository(db)

	// Initialize services
	authService := auth.NewService(cfg.Auth.JWTSecret, cfg.Auth.TokenDuration)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, authService, log)
	assetHandler := handlers.NewAssetHandler(assetRepo, queueClient, log)
	vulnHandler := handlers.NewVulnerabilityHandler(vulnRepo, queueClient, log)
	alertHandler := handlers.NewAlertHandler(alertRepo, log)

	// Setup routes
	router := routes.SetupRoutes(
		authHandler,
		assetHandler,
		vulnHandler,
		alertHandler,
		authService,
		log,
	)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info("API server listening", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", "error", err)
	}

	log.Info("server exited")
}
