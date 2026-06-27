package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HarshChauhan626/AegisIDP/backend/models"
	"github.com/HarshChauhan626/AegisIDP/backend/queue"
	"github.com/HarshChauhan626/AegisIDP/backend/repository"
	"github.com/google/uuid"
)

// DispatchRequest carries the parameters needed to submit a workflow execution.
type DispatchRequest struct {
	ExecutionType string
	EnvironmentID string
	CreatedBy     string
	Input         map[string]any
	Trigger       models.WorkflowTrigger
}

// Dispatcher creates WorkflowExecution records and enqueues them for processing.
// It also exposes Cancel to signal cancellation of a running workflow.
type Dispatcher struct {
	queue        *queue.Queue
	pool         *WorkerPool
	workflowRepo repository.WorkflowRepository
}

// NewDispatcher creates a Dispatcher wired to the given queue, pool, and repo.
func NewDispatcher(q *queue.Queue, pool *WorkerPool, workflowRepo repository.WorkflowRepository) *Dispatcher {
	return &Dispatcher{queue: q, pool: pool, workflowRepo: workflowRepo}
}

// Dispatch creates a WorkflowExecution record (pending → queued) and enqueues the job.
// Returns the newly created execution so the caller can surface the workflow_id.
func (d *Dispatcher) Dispatch(ctx context.Context, req DispatchRequest) (*models.WorkflowExecution, error) {
	now := time.Now()

	if req.Trigger == "" {
		req.Trigger = models.TriggerManual
	}

	inputJSON := "{}"
	if req.Input != nil {
		if b, err := json.Marshal(req.Input); err == nil {
			inputJSON = string(b)
		}
	}

	wf := &models.WorkflowExecution{
		ID:            uuid.NewString(),
		EnvironmentID: req.EnvironmentID,
		Type:          req.ExecutionType,
		State:         models.WorkflowStatePending,
		Trigger:       req.Trigger,
		Input:         inputJSON,
		CreatedBy:     req.CreatedBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := d.workflowRepo.CreateExecution(wf); err != nil {
		return nil, fmt.Errorf("create workflow execution: %w", err)
	}

	// Transition to queued
	wf.State = models.WorkflowStateQueued
	wf.UpdatedAt = time.Now()
	if err := d.workflowRepo.UpdateExecution(wf); err != nil {
		return nil, fmt.Errorf("update to queued: %w", err)
	}

	// Enqueue — on failure revert state to failed so it's visible
	job := queue.Job{
		WorkflowID:    wf.ID,
		ExecutionType: req.ExecutionType,
		EnvironmentID: req.EnvironmentID,
		Input:         req.Input,
	}
	if err := d.queue.Enqueue(ctx, job); err != nil {
		wf.State = models.WorkflowStateFailed
		wf.Error = err.Error()
		wf.UpdatedAt = time.Now()
		_ = d.workflowRepo.UpdateExecution(wf)
		return nil, fmt.Errorf("enqueue job: %w", err)
	}

	return wf, nil
}

// Cancel signals cancellation for a currently-running workflow via the worker pool.
// Returns an error if the workflow is not currently being processed.
func (d *Dispatcher) Cancel(workflowID string) error {
	if !d.pool.Cancel(workflowID) {
		return fmt.Errorf("workflow %q is not currently running in the worker pool", workflowID)
	}
	return nil
}
