package handlers

import (
	"net/http"
	"strconv"

	"github.com/HarshChauhan626/AegisIDP/backend/repository"
	"github.com/gin-gonic/gin"
)

// WorkflowHandler handles workflow execution routes.
type WorkflowHandler struct {
	workflowRepo repository.WorkflowRepository
}

func NewWorkflowHandler(workflowRepo repository.WorkflowRepository) *WorkflowHandler {
	return &WorkflowHandler{workflowRepo: workflowRepo}
}

// List returns workflow executions with optional environment filter and pagination.
// GET /api/workflows
func (h *WorkflowHandler) List(c *gin.Context) {
	environmentID := c.Query("environment_id")
	limit := queryInt(c, "limit", 20)
	offset := queryInt(c, "offset", 0)

	wfs, total, err := h.workflowRepo.ListExecutions(environmentID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, paginatedResponse(wfs, total, limit, offset))
}

// Get returns a single workflow execution with its steps.
// GET /api/workflows/:id
func (h *WorkflowHandler) Get(c *gin.Context) {
	wf, err := h.workflowRepo.FindExecutionByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Workflow not found"))
		return
	}
	c.JSON(http.StatusOK, successResponse(wf))
}

// Retry re-enqueues a failed workflow execution (full implementation in Phase 2).
// POST /api/workflows/:id/retry
func (h *WorkflowHandler) Retry(c *gin.Context) {
	wf, err := h.workflowRepo.FindExecutionByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Workflow not found"))
		return
	}

	if wf.State != "failed" && wf.State != "rolled_back" {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATE", "Only failed or rolled-back workflows can be retried"))
		return
	}

	// Phase 2: enqueue retry job
	c.JSON(http.StatusAccepted, successResponse(gin.H{
		"message":     "Retry enqueued",
		"workflow_id": wf.ID,
	}))
}

// Cancel requests cancellation of a running workflow (full implementation in Phase 2).
// POST /api/workflows/:id/cancel
func (h *WorkflowHandler) Cancel(c *gin.Context) {
	wf, err := h.workflowRepo.FindExecutionByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Workflow not found"))
		return
	}

	if wf.State != "running" && wf.State != "queued" {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATE", "Only running or queued workflows can be cancelled"))
		return
	}

	// Phase 2: signal cancellation via context
	c.JSON(http.StatusAccepted, successResponse(gin.H{
		"message":     "Cancellation requested",
		"workflow_id": wf.ID,
	}))
}

// Stream opens an SSE connection for live workflow updates (full implementation in Phase 4).
// GET /api/workflows/:id/stream
func (h *WorkflowHandler) Stream(c *gin.Context) {
	id := c.Param("id")
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Phase 4: attach SSE broker; for now send a single status event
	wf, err := h.workflowRepo.FindExecutionByID(id)
	if err != nil {
		c.SSEvent("error", gin.H{"code": "NOT_FOUND", "message": "Workflow not found"})
		return
	}
	c.SSEvent("status", gin.H{"workflow_id": wf.ID, "state": wf.State})
}

func queryInt(c *gin.Context, key string, defaultVal int) int {
	if s := c.Query(key); s != "" {
		if i, err := strconv.Atoi(s); err == nil && i >= 0 {
			return i
		}
	}
	return defaultVal
}
