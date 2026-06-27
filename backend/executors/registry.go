package executors

import (
	"github.com/HarshChauhan626/AegisIDP/backend/workflow"
)

// Registry maps executor keys to their StepExecutor implementations.
// It satisfies the workflow.ExecutorRegistry interface.
type Registry struct {
	m map[string]workflow.StepExecutor
}

// New creates an empty Registry.
func New() *Registry {
	return &Registry{m: make(map[string]workflow.StepExecutor)}
}

// Register adds a StepExecutor to the registry under its Key().
func (r *Registry) Register(e workflow.StepExecutor) {
	r.m[e.Key()] = e
}

// Get looks up a StepExecutor by key. Returns (executor, true) or (nil, false).
func (r *Registry) Get(key string) (workflow.StepExecutor, bool) {
	e, ok := r.m[key]
	return e, ok
}

// RegisterDefaults registers stub executors for all known executor keys.
// Phase 3 will replace each stub with a real simulator.
func RegisterDefaults(r *Registry) {
	keys := []string{
		"validate",
		"reserve",
		"namespace",
		"postgresql",
		"redis",
		"rabbitmq",
		"deployment",
		"dns",
		"healthcheck",
	}
	for _, k := range keys {
		r.Register(NewStub(k))
	}
}
