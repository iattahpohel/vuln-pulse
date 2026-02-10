package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/atta/vulnpulse/internal/repository"
	"github.com/atta/vulnpulse/internal/service"
	"github.com/atta/vulnpulse/internal/worker"
	"github.com/atta/vulnpulse/pkg/cache"
	"github.com/atta/vulnpulse/pkg/config"
	"github.com/atta/vulnpulse/pkg/database"
	"github.com/atta/vulnpulse/pkg/logger"
	"github.com/atta/vulnpulse/pkg/queue"
)

func main() {
	// Initialize logger
	log := logger.New("worker")
	log.Info("starting VulnPulse worker")

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
	vulnRepo := repository.NewVulnerabilityRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	alertRepo := repository.NewAlertRepository(db)

	// Initialize services
	matchService := service.NewMatchService(vulnRepo, assetRepo, alertRepo, queueClient, log)

	// Create worker
	w := worker.NewWorker(queueClient, matchService, log, cfg.RabbitMQ.QueueName)

	// Start worker in goroutine
	go func() {
		if err := w.Start(ctx); err != nil {
			log.Error("worker error", "error", err)
			os.Exit(1)
		}
	}()

	log.Info("worker is running")

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down worker...")
	log.Info("worker exited")
}
