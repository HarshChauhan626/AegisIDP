package workflow

import (
	"context"
	"sync"
)

// StepExecutor is the interface every provisioning executor must implement.
// Phase 2 uses stub executors; Phase 3 replaces them with real simulators.
type StepExecutor interface {
	// Key returns the unique executor identifier referenced in workflow YAML files.
	Key() string
	// Execute runs the forward provisioning action for this step.
	Execute(ctx context.Context, sc *StepContext) (map[string]any, error)
	// Rollback reverses the provisioning action (Saga compensation).
	Rollback(ctx context.Context, sc *StepContext) error
}

// ExecutorRegistry is satisfied by any type that can resolve executor keys.
// Defined here (not in the executors package) to avoid import cycles.
type ExecutorRegistry interface {
	Get(key string) (StepExecutor, bool)
}

// StepContext carries shared mutable state across all steps in a single workflow run.
// Outputs from completed steps are stored here and available to subsequent steps.
type StepContext struct {
	WorkflowID    string
	EnvironmentID string
	Input         map[string]any

	mu      sync.RWMutex
	outputs map[string]map[string]any // step YAML ID → output map
}

// NewStepContext creates an initialised StepContext.
func NewStepContext(workflowID, environmentID string, input map[string]any) *StepContext {
	if input == nil {
		input = map[string]any{}
	}
	return &StepContext{
		WorkflowID:    workflowID,
		EnvironmentID: environmentID,
		Input:         input,
		outputs:       make(map[string]map[string]any),
	}
}

// GetOutput returns the output map written by a previously completed step.
func (sc *StepContext) GetOutput(stepID string) map[string]any {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.outputs[stepID]
}

// SetOutput records the output of a completed step so downstream steps can consume it.
func (sc *StepContext) SetOutput(stepID string, output map[string]any) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.outputs[stepID] = output
}
