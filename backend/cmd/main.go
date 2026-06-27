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
	"github.com/HarshChauhan626/AegisIDP/backend/executors"
	"github.com/HarshChauhan626/AegisIDP/backend/logger"
	"github.com/HarshChauhan626/AegisIDP/backend/queue"
	"github.com/HarshChauhan626/AegisIDP/backend/repository"
	"github.com/HarshChauhan626/AegisIDP/backend/workflow"
	"github.com/HarshChauhan626/AegisIDP/backend/workers"
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

	// Configure workflow definition directory
	workflow.SetWorkflowDir(cfg.WorkflowDir)
	log.Info("workflow definitions loaded from", zap.String("dir", cfg.WorkflowDir))

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

	// ── Workflow Engine wiring ──────────────────────────────────────────────
	// 1. Executor registry with all stub executors (Phase 3 replaces with simulators)
	registry := executors.New()
	executors.RegisterDefaults(registry)

	// 2. Engine
	engine := workflow.NewEngine(registry, workflowRepo, log)

	// 3. Job queue
	jobQueue := queue.New(cfg.QueueCapacity)

	// 4. Worker pool
	pool := workers.NewWorkerPool(cfg.WorkerCount, jobQueue, engine, log)

	// 5. Dispatcher (used by HTTP handlers)
	dispatcher := workers.NewDispatcher(jobQueue, pool, workflowRepo)

	// Start worker pool with a cancellable context
	poolCtx, poolCancel := context.WithCancel(context.Background())
	pool.Start(poolCtx)
	log.Info("worker pool running",
		zap.Int("workers", cfg.WorkerCount),
		zap.Int("queue_capacity", cfg.QueueCapacity),
	)

	// ── HTTP Server ─────────────────────────────────────────────────────────
	router := api.NewRouter(api.RouterDeps{
		UserRepo:     userRepo,
		EnvRepo:      envRepo,
		WorkflowRepo: workflowRepo,
		EventRepo:    eventRepo,
		AuditRepo:    auditRepo,
		Dispatcher:   dispatcher,
		JWTSecret:    cfg.JWTSecret,
		FrontendURL:  cfg.FrontendURL,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Info("server starting", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error", zap.Error(err))
		}
	}()

	// ── Graceful shutdown ───────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down — draining in-flight workflows...")

	// Stop accepting new HTTP requests
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("HTTP server forced shutdown", zap.Error(err))
	}

	// Signal workers to stop and wait for in-flight jobs to complete
	poolCancel()
	pool.Stop()

	log.Info("server exited cleanly")
}
