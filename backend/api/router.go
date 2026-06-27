package api

import (
	"net/http"

	"github.com/HarshChauhan626/AegisIDP/backend/api/handlers"
	"github.com/HarshChauhan626/AegisIDP/backend/middleware"
	"github.com/HarshChauhan626/AegisIDP/backend/repository"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RouterDeps holds all handler dependencies injected at router setup.
type RouterDeps struct {
	UserRepo     repository.UserRepository
	EnvRepo      repository.EnvironmentRepository
	WorkflowRepo repository.WorkflowRepository
	EventRepo    repository.EventRepository
	AuditRepo    repository.AuditLogRepository
	JWTSecret    string
	FrontendURL  string
}

// NewRouter creates and configures the Gin router with all routes registered.
func NewRouter(deps RouterDeps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// CORS — allow frontend origin
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{deps.FrontendURL, "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
	}))

	// Instantiate handlers
	authH := handlers.NewAuthHandler(deps.UserRepo, deps.JWTSecret)
	envH := handlers.NewEnvironmentHandler(deps.EnvRepo)
	wfH := handlers.NewWorkflowHandler(deps.WorkflowRepo)
	userH := handlers.NewUserHandler(deps.UserRepo)
	obsH := handlers.NewObservabilityHandler(deps.EventRepo, deps.AuditRepo)

	// Health check (public)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "platform-orchestrator"})
	})

	// Auth routes (public)
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", authH.Login)
		authGroup.POST("/refresh", authH.Refresh)
	}

	// Protected routes — require valid JWT
	api := r.Group("/api", middleware.JWTAuth(deps.JWTSecret))
	{
		// Environment routes
		envGroup := api.Group("/environments")
		{
			envGroup.POST("", middleware.RequirePermission(middleware.PermEnvironmentCreate), envH.Create)
			envGroup.GET("", middleware.RequirePermission(middleware.PermWorkflowView), envH.List)
			envGroup.GET("/:id", middleware.RequirePermission(middleware.PermWorkflowView), envH.Get)
			envGroup.DELETE("/:id", middleware.RequirePermission(middleware.PermEnvironmentDelete), envH.Delete)
		}

		// Workflow routes
		wfGroup := api.Group("/workflows")
		{
			wfGroup.GET("", middleware.RequirePermission(middleware.PermWorkflowView), wfH.List)
			wfGroup.GET("/:id", middleware.RequirePermission(middleware.PermWorkflowView), wfH.Get)
			wfGroup.GET("/:id/stream", middleware.RequirePermission(middleware.PermWorkflowView), wfH.Stream)
			wfGroup.POST("/:id/retry", middleware.RequirePermission(middleware.PermWorkflowRetry), wfH.Retry)
			wfGroup.POST("/:id/cancel", middleware.RequirePermission(middleware.PermWorkflowCancel), wfH.Cancel)
		}

		// Observability routes
		api.GET("/metrics", middleware.RequirePermission(middleware.PermMetricsView), obsH.GetMetrics)
		api.GET("/events", middleware.RequirePermission(middleware.PermWorkflowView), obsH.ListEvents)
		api.GET("/audit", middleware.RequirePermission(middleware.PermAuditView), obsH.ListAuditLogs)
		api.GET("/logs", middleware.RequirePermission(middleware.PermWorkflowView), obsH.ListLogs)

		// User management routes (Admin only)
		userGroup := api.Group("/users", middleware.RequirePermission(middleware.PermUserManage))
		{
			userGroup.GET("", userH.List)
			userGroup.POST("", userH.Create)
			userGroup.PATCH("/:id", userH.Update)
		}
	}

	return r
}
