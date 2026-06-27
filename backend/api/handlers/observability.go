package handlers

import (
	"net/http"

	"github.com/HarshChauhan626/AegisIDP/backend/repository"
	"github.com/gin-gonic/gin"
)

// ObservabilityHandler handles metrics, events, audit, and logs endpoints.
type ObservabilityHandler struct {
	eventRepo    repository.EventRepository
	auditLogRepo repository.AuditLogRepository
}

func NewObservabilityHandler(
	eventRepo repository.EventRepository,
	auditLogRepo repository.AuditLogRepository,
) *ObservabilityHandler {
	return &ObservabilityHandler{
		eventRepo:    eventRepo,
		auditLogRepo: auditLogRepo,
	}
}

// GetMetrics returns in-memory operational metrics.
// Full implementation in Phase 4 — returns a stub for Phase 1.
// GET /api/metrics
func (h *ObservabilityHandler) GetMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, successResponse(gin.H{
		"running_workflows":        0,
		"completed_workflows":      0,
		"failed_workflows":         0,
		"success_rate":             0.0,
		"queue_depth":              0,
		"active_workers":           0,
		"avg_provisioning_time_ms": 0,
		"total_retries":            0,
	}))
}

// ListEvents returns workflow events with pagination.
// GET /api/events
func (h *ObservabilityHandler) ListEvents(c *gin.Context) {
	limit := queryInt(c, "limit", 50)
	offset := queryInt(c, "offset", 0)

	events, total, err := h.eventRepo.List(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, paginatedResponse(events, total, limit, offset))
}

// ListAuditLogs returns audit log entries with pagination.
// GET /api/audit
func (h *ObservabilityHandler) ListAuditLogs(c *gin.Context) {
	limit := queryInt(c, "limit", 50)
	offset := queryInt(c, "offset", 0)

	logs, total, err := h.auditLogRepo.List(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("DB_ERROR", err.Error()))
		return
	}
	c.JSON(http.StatusOK, paginatedResponse(logs, total, limit, offset))
}

// ListLogs is a stub for structured workflow log streaming (Phase 4).
// GET /api/logs
func (h *ObservabilityHandler) ListLogs(c *gin.Context) {
	c.JSON(http.StatusOK, successResponse([]interface{}{}))
}
