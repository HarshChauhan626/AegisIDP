package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HarshChauhan626/AegisIDP/backend/models"
	"github.com/HarshChauhan626/AegisIDP/backend/repository"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// completedRollbackStep pairs a step definition with its persisted model
// so the rollback phase can call the right executor and update the right record.
type completedRollbackStep struct {
	def   StepDef
	model *models.WorkflowStep
}

// Engine is the core workflow orchestrator. It loads YAML definitions, executes
// steps in topological order with per-step retry policies, and triggers Saga-style
// rollback on permanent failure.
type Engine struct {
	registry     ExecutorRegistry
	workflowRepo repository.WorkflowRepository
	log          *zap.Logger
}

// NewEngine creates an Engine with the provided executor registry and repository.
func NewEngine(registry ExecutorRegistry, workflowRepo repository.WorkflowRepository, log *zap.Logger) *Engine {
	return &Engine{
		registry:     registry,
		workflowRepo: workflowRepo,
		log:          log,
	}
}

// Run executes a workflow to completion. Called by a worker goroutine.
// workflowID must already exist in the database in "queued" state.
func (e *Engine) Run(ctx context.Context, workflowID, executionType, environmentID string, input map[string]any) error {
	log := e.log.With(zap.String("workflow_id", workflowID), zap.String("type", executionType))

	// ── Load YAML definition ────────────────────────────────────────────────
	def, err := Load(executionType)
	if err != nil {
		return e.markFailed(workflowID, fmt.Errorf("load definition: %w", err))
	}

	// ── Topological sort ────────────────────────────────────────────────────
	ordered, err := TopoSort(def.Steps)
	if err != nil {
		return e.markFailed(workflowID, err)
	}

	// ── Transition to running ───────────────────────────────────────────────
	wf, err := e.workflowRepo.FindExecutionByID(workflowID)
	if err != nil {
		return fmt.Errorf("find workflow: %w", err)
	}
	now := time.Now()
	wf.State = models.WorkflowStateRunning
	wf.StartedAt = &now
	wf.UpdatedAt = now
	if err := e.workflowRepo.UpdateExecution(wf); err != nil {
		return fmt.Errorf("update to running: %w", err)
	}
	log.Info("workflow running")

	// ── Create step records ─────────────────────────────────────────────────
	stepModels := make([]*models.WorkflowStep, 0, len(ordered))
	for i, stepDef := range ordered {
		sm := &models.WorkflowStep{
			ID:          uuid.NewString(),
			WorkflowID:  workflowID,
			Name:        stepDef.Name,
			ExecutorKey: stepDef.Executor,
			StepOrder:   i,
			State:       models.StepStatePending,
			Attempt:     0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := e.workflowRepo.CreateStep(sm); err != nil {
			return e.markFailed(workflowID, fmt.Errorf("create step record: %w", err))
		}
		stepModels = append(stepModels, sm)
	}

	// ── Execute steps in topological order ──────────────────────────────────
	sc := NewStepContext(workflowID, environmentID, input)
	var rollbackable []completedRollbackStep

	for i, stepDef := range ordered {
		sm := stepModels[i]

		log.Info("executing step", zap.String("step", stepDef.ID), zap.Int("order", i))

		output, execErr := e.executeStep(ctx, stepDef, sm, sc)
		if execErr != nil {
			// Context cancelled → mark cancelled and return
			if ctx.Err() != nil {
				e.markCancelled(workflowID)
				return ctx.Err()
			}

			log.Error("step failed — initiating rollback", zap.String("step", stepDef.ID), zap.Error(execErr))
			e.rollback(context.Background(), rollbackable, sc)
			return e.markFailed(workflowID, execErr)
		}

		if output != nil {
			sc.SetOutput(stepDef.ID, output)
		}
		if stepDef.Rollback {
			rollbackable = append(rollbackable, completedRollbackStep{def: stepDef, model: sm})
		}
	}

	// ── All steps succeeded ─────────────────────────────────────────────────
	completedAt := time.Now()
	wf, _ = e.workflowRepo.FindExecutionByID(workflowID)
	wf.State = models.WorkflowStateCompleted
	wf.CompletedAt = &completedAt
	wf.UpdatedAt = completedAt
	_ = e.workflowRepo.UpdateExecution(wf)

	log.Info("workflow completed")
	return nil
}

// executeStep runs a single step with its retry policy, persisting state after each transition.
func (e *Engine) executeStep(ctx context.Context, stepDef StepDef, sm *models.WorkflowStep, sc *StepContext) (map[string]any, error) {
	executor, ok := e.registry.Get(stepDef.Executor)
	if !ok {
		return nil, fmt.Errorf("executor %q not registered", stepDef.Executor)
	}

	policy := BuildRetryPolicy(ParseRetry(stepDef.Retry))

	var lastErr error
	for attempt := 1; ; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Persist: running
		now := time.Now()
		sm.State = models.StepStateRunning
		sm.Attempt = attempt
		sm.StartedAt = &now
		sm.UpdatedAt = now
		_ = e.workflowRepo.UpdateStep(sm)

		output, err := executor.Execute(ctx, sc)
		if err == nil {
			// Persist: succeeded
			done := time.Now()
			sm.State = models.StepStateSucceeded
			sm.CompletedAt = &done
			sm.UpdatedAt = done
			if output != nil {
				if b, jsonErr := json.Marshal(output); jsonErr == nil {
					sm.Output = string(b)
				}
			}
			_ = e.workflowRepo.UpdateStep(sm)
			return output, nil
		}

		lastErr = err
		e.log.Warn("step attempt failed",
			zap.String("step", stepDef.ID),
			zap.Int("attempt", attempt),
			zap.Int("max_attempts", stepDef.Retry.MaxAttempts),
			zap.Error(err),
		)

		if !policy.ShouldRetry(attempt) {
			break
		}

		delay := policy.Delay(attempt)
		e.log.Info("retrying step", zap.String("step", stepDef.ID), zap.Duration("delay", delay))
		if sleepErr := sleep(ctx, delay); sleepErr != nil {
			return nil, sleepErr
		}
	}

	// All attempts exhausted — persist: failed
	done := time.Now()
	sm.State = models.StepStateFailed
	sm.Error = lastErr.Error()
	sm.CompletedAt = &done
	sm.UpdatedAt = done
	_ = e.workflowRepo.UpdateStep(sm)

	return nil, lastErr
}

// rollback executes Saga compensation actions in reverse order.
func (e *Engine) rollback(ctx context.Context, steps []completedRollbackStep, sc *StepContext) {
	// Transition workflow to rolling_back
	if wf, err := e.workflowRepo.FindExecutionByID(sc.WorkflowID); err == nil {
		wf.State = models.WorkflowStateRollingBack
		wf.UpdatedAt = time.Now()
		_ = e.workflowRepo.UpdateExecution(wf)
	}

	// Execute rollbacks in reverse insertion order
	for i := len(steps) - 1; i >= 0; i-- {
		s := steps[i]
		executor, ok := e.registry.Get(s.def.Executor)
		if !ok {
			e.log.Warn("rollback: executor not found", zap.String("step", s.def.ID))
			continue
		}

		e.log.Info("rolling back step", zap.String("step", s.def.ID))
		sm := s.model
		rbErr := executor.Rollback(ctx, sc)
		now := time.Now()
		if rbErr != nil {
			e.log.Error("rollback step failed", zap.String("step", s.def.ID), zap.Error(rbErr))
			sm.State = models.StepStateRollbackFailed
			sm.Error = rbErr.Error()
		} else {
			sm.State = models.StepStateRolledBack
		}
		sm.UpdatedAt = now
		_ = e.workflowRepo.UpdateStep(sm)
	}

	// Transition workflow to rolled_back
	if wf, err := e.workflowRepo.FindExecutionByID(sc.WorkflowID); err == nil {
		wf.State = models.WorkflowStateRolledBack
		wf.UpdatedAt = time.Now()
		_ = e.workflowRepo.UpdateExecution(wf)
	}

	e.log.Info("rollback complete", zap.String("workflow_id", sc.WorkflowID))
}

// markFailed sets the workflow state to failed with an error message.
func (e *Engine) markFailed(workflowID string, err error) error {
	if wf, dbErr := e.workflowRepo.FindExecutionByID(workflowID); dbErr == nil {
		now := time.Now()
		wf.State = models.WorkflowStateFailed
		wf.Error = err.Error()
		wf.CompletedAt = &now
		wf.UpdatedAt = now
		_ = e.workflowRepo.UpdateExecution(wf)
	}
	return err
}

// markCancelled sets the workflow state to cancelled.
func (e *Engine) markCancelled(workflowID string) {
	if wf, err := e.workflowRepo.FindExecutionByID(workflowID); err == nil {
		now := time.Now()
		wf.State = models.WorkflowStateCancelled
		wf.CompletedAt = &now
		wf.UpdatedAt = now
		_ = e.workflowRepo.UpdateExecution(wf)
	}
}
