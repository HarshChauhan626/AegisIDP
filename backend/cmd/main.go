package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HarshChauhan626/AegisIDP/backend/api"
	"github.com/HarshChauhan626/AegisIDP/backend/api/handlers"
	"github.com/HarshChauhan626/AegisIDP/backend/config"
	"github.com/HarshChauhan626/AegisIDP/backend/logger"
	"github.com/HarshChauhan626/AegisIDP/backend/repository"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	migrateOnly := flag.Bool("migrate-only", false, "Run migrations and exit")
	flag.Parse()

	// Load .env if present (local dev convenience)
	_ = godotenv.Load()

	// Load configuration
	cfg := config.Load()

	// Initialise logger
	if err := logger.Init(cfg.LogLevel); err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialise logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Log.Sync() // nolint:errcheck

	log := logger.Log

	// Initialise database
	db, err := repository.Init(cfg.DBPath)
	if err != nil {
		log.Fatal("failed to initialise database", zap.Error(err))
	}
	log.Info("database initialised", zap.String("path", cfg.DBPath))

	if *migrateOnly {
		log.Info("migration complete, exiting")
		return
	}

	// Instantiate repositories
	userRepo := repository.NewUserRepository(db)
	envRepo := repository.NewEnvironmentRepository(db)
	workflowRepo := repository.NewWorkflowRepository(db)
	eventRepo := repository.NewEventRepository(db)
	auditRepo := repository.NewAuditLogRepository(db)

	// Seed default admin user on first run
	if err := handlers.SeedAdminUser(userRepo); err != nil {
		log.Warn("failed to seed admin user", zap.Error(err))
	}

	// Build router
	router := api.NewRouter(api.RouterDeps{
		UserRepo:     userRepo,
		EnvRepo:      envRepo,
		WorkflowRepo: workflowRepo,
		EventRepo:    eventRepo,
		AuditRepo:    auditRepo,
		JWTSecret:    cfg.JWTSecret,
		FrontendURL:  cfg.FrontendURL,
	})

	// HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	go func() {
		log.Info("server starting", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}
	log.Info("server exited")
}
