package handlers

import (
	"net/http"
	"time"

	"github.com/HarshChauhan626/AegisIDP/backend/middleware"
	"github.com/HarshChauhan626/AegisIDP/backend/models"
	"github.com/HarshChauhan626/AegisIDP/backend/repository"
	"github.com/HarshChauhan626/AegisIDP/backend/workers"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// EnvironmentHandler handles environment CRUD routes.
type EnvironmentHandler struct {
	envRepo    repository.EnvironmentRepository
	dispatcher *workers.Dispatcher
}

func NewEnvironmentHandler(envRepo repository.EnvironmentRepository, dispatcher *workers.Dispatcher) *EnvironmentHandler {
	return &EnvironmentHandler{envRepo: envRepo, dispatcher: dispatcher}
}

type createEnvironmentRequest struct {
	ProjectID string      `json:"project_id" binding:"required"`
	Name      string      `json:"name" binding:"required,min=2,max=64"`
	Config    interface{} `json:"config"`
}

// Create provisions a new environment and dispatches a create-environment workflow.
// POST /api/environments
func (h *EnvironmentHandler) Create(c *gin.Context) {
	var req createEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse("VALIDATION_ERROR", err.Error()))
		return
	}

	env := &models.Environment{
		ID:        uuid.NewString(),
		ProjectID: req.ProjectID,
		Name:      req.Name,
		Status:    models.EnvironmentStatusPending,
		CreatedBy: middleware.GetUserID(c),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.envRepo.Create(env); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", err.Error()))
		return
	}

	// Dispatch create-environment workflow asynchronously
	wf, err := h.dispatcher.Dispatch(c.Request.Context(), workers.DispatchRequest{
		ExecutionType: "create-environment",
		EnvironmentID: env.ID,
		CreatedBy:     middleware.GetUserID(c),
		Input: map[string]any{
			"environment_name": env.Name,
			"project_id":       env.ProjectID,
			"environment_id":   env.ID,
		},
	})
	if err != nil {
		// Dispatch failure is non-fatal: environment record exists, operator can retry
		c.JSON(http.StatusCreated, successResponse(gin.H{
			"environment": env,
			"workflow_id": nil,
			"warning":     "workflow dispatch failed: " + err.Error(),
		}))
		return
	}

	// Update environment to provisioning status now that workflow is queued
	env.Status = models.EnvironmentStatusProvisioning
	env.UpdatedAt = time.Now()
	_ = h.envRepo.Update(env)

	c.JSON(http.StatusCreated, successResponse(gin.H{
		"environment": env,
		"workflow_id": wf.ID,
	}))
}

// List returns all environments, optionally filtered by project.
// GET /api/environments
func (h *EnvironmentHandler) List(c *gin.Context) {
	projectID := c.Query("project_id")
	envs, err := h.envRepo.List(projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, successResponse(envs))
}

// Get returns a single environment by ID.
// GET /api/environments/:id
func (h *EnvironmentHandler) Get(c *gin.Context) {
	env, err := h.envRepo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Environment not found"))
		return
	}
	c.JSON(http.StatusOK, successResponse(env))
}

// Delete marks an environment for deletion and dispatches a delete-environment workflow.
// DELETE /api/environments/:id
func (h *EnvironmentHandler) Delete(c *gin.Context) {
	env, err := h.envRepo.FindByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Environment not found"))
		return
	}

	env.Status = models.EnvironmentStatusDeleting
	env.UpdatedAt = time.Now()
	if err := h.envRepo.Update(env); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", err.Error()))
		return
	}

	// Dispatch delete-environment workflow
	wf, _ := h.dispatcher.Dispatch(c.Request.Context(), workers.DispatchRequest{
		ExecutionType: "delete-environment",
		EnvironmentID: env.ID,
		CreatedBy:     middleware.GetUserID(c),
		Input: map[string]any{
			"environment_id":   env.ID,
			"environment_name": env.Name,
		},
	})

	resp := gin.H{"message": "Environment deletion initiated", "id": env.ID}
	if wf != nil {
		resp["workflow_id"] = wf.ID
	}
	c.JSON(http.StatusOK, successResponse(resp))
}
