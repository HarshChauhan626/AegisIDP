package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/HarshChauhan626/AegisIDP/backend/middleware"
	"github.com/HarshChauhan626/AegisIDP/backend/models"
	"github.com/HarshChauhan626/AegisIDP/backend/repository"
	"github.com/HarshChauhan626/AegisIDP/backend/workers"
	"github.com/gin-gonic/gin"
)

// WorkflowHandler handles workflow execution routes.
type WorkflowHandler struct {
	workflowRepo repository.WorkflowRepository
	dispatcher   *workers.Dispatcher
}

func NewWorkflowHandler(workflowRepo repository.WorkflowRepository, dispatcher *workers.Dispatcher) *WorkflowHandler {
	return &WorkflowHandler{workflowRepo: workflowRepo, dispatcher: dispatcher}
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

// Get returns a single workflow execution with its steps ordered by step_order.
// GET /api/workflows/:id
func (h *WorkflowHandler) Get(c *gin.Context) {
	wf, err := h.workflowRepo.FindExecutionByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Workflow not found"))
		return
	}
	c.JSON(http.StatusOK, successResponse(wf))
}

// Retry re-dispatches a failed or rolled-back workflow as a new execution.
// POST /api/workflows/:id/retry
func (h *WorkflowHandler) Retry(c *gin.Context) {
	wf, err := h.workflowRepo.FindExecutionByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Workflow not found"))
		return
	}

	if wf.State != models.WorkflowStateFailed && wf.State != models.WorkflowStateRolledBack {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATE", "Only failed or rolled-back workflows can be retried"))
		return
	}

	newWF, err := h.dispatcher.Dispatch(c.Request.Context(), workers.DispatchRequest{
		ExecutionType: wf.Type,
		EnvironmentID: wf.EnvironmentID,
		CreatedBy:     middleware.GetUserID(c),
		Input:         parseJSONInput(wf.Input),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DISPATCH_ERROR", err.Error()))
		return
	}

	c.JSON(http.StatusAccepted, successResponse(gin.H{
		"message":     "Retry enqueued",
		"workflow_id": newWF.ID,
	}))
}

// Cancel signals cancellation of a running or queued workflow.
// POST /api/workflows/:id/cancel
func (h *WorkflowHandler) Cancel(c *gin.Context) {
	wf, err := h.workflowRepo.FindExecutionByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse("NOT_FOUND", "Workflow not found"))
		return
	}

	if wf.State != models.WorkflowStateRunning && wf.State != models.WorkflowStateQueued {
		c.JSON(http.StatusBadRequest, errorResponse("INVALID_STATE", "Only running or queued workflows can be cancelled"))
		return
	}

	if err := h.dispatcher.Cancel(wf.ID); err != nil {
		c.JSON(http.StatusConflict, errorResponse("NOT_RUNNING", err.Error()))
		return
	}

	c.JSON(http.StatusAccepted, successResponse(gin.H{
		"message":     "Cancellation signalled",
		"workflow_id": wf.ID,
	}))
}

// Stream opens an SSE connection for live workflow step updates (Phase 4).
// GET /api/workflows/:id/stream
func (h *WorkflowHandler) Stream(c *gin.Context) {
	id := c.Param("id")
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	wf, err := h.workflowRepo.FindExecutionByID(id)
	if err != nil {
		c.SSEvent("error", gin.H{"code": "NOT_FOUND", "message": "Workflow not found"})
		return
	}
	// Phase 4 will attach a real SSE broker; for now, send a single snapshot event
	c.SSEvent("snapshot", gin.H{"workflow_id": wf.ID, "state": wf.State, "steps": wf.Steps})
}

// ── Helpers ─────────────────────────────────────────────────────────────────

func queryInt(c *gin.Context, key string, defaultVal int) int {
	if s := c.Query(key); s != "" {
		if i, err := strconv.Atoi(s); err == nil && i >= 0 {
			return i
		}
	}
	return defaultVal
}

// parseJSONInput safely unmarshals a JSON string into map[string]any.
// Returns an empty map on error.
func parseJSONInput(raw string) map[string]any {
	if raw == "" || raw == "{}" {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return map[string]any{}
	}
	return m
}
