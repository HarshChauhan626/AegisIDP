package executors

import (
	"context"
	"time"

	"github.com/HarshChauhan626/AegisIDP/backend/workflow"
)

// StubExecutor is a Phase 2 placeholder that simulates a provisioning step
// by sleeping briefly and returning a success result. Phase 3 replaces each
// stub with a real simulator that models latency, failure rates, and outputs.
type StubExecutor struct {
	key string
}

// NewStub creates a StubExecutor for the given executor key.
func NewStub(key string) *StubExecutor {
	return &StubExecutor{key: key}
}

// Key returns the executor identifier as referenced in YAML workflow definitions.
func (s *StubExecutor) Key() string { return s.key }

// Execute simulates ~200ms of provisioning work and returns a basic output map.
func (s *StubExecutor) Execute(ctx context.Context, sc *workflow.StepContext) (map[string]any, error) {
	select {
	case <-time.After(200 * time.Millisecond):
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return map[string]any{
		"status":        "provisioned",
		"executor":      s.key,
		"environment_id": sc.EnvironmentID,
	}, nil
}

// Rollback simulates ~100ms of cleanup work.
func (s *StubExecutor) Rollback(ctx context.Context, sc *workflow.StepContext) error {
	select {
	case <-time.After(100 * time.Millisecond):
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
