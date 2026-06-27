package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Definition is the top-level structure of a workflow YAML file.
type Definition struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Steps       []StepDef `yaml:"steps"`
}

// StepDef describes a single step in a workflow definition.
type StepDef struct {
	ID        string   `yaml:"id"`
	Name      string   `yaml:"name"`
	Executor  string   `yaml:"executor"`
	DependsOn []string `yaml:"depends_on"`
	Rollback  bool     `yaml:"rollback"`
	Retry     RetryDef `yaml:"retry"`
}

// RetryDef holds raw (string) retry configuration from YAML.
type RetryDef struct {
	MaxAttempts int    `yaml:"max_attempts"`
	Strategy    string `yaml:"strategy"`    // "fixed" | "exponential_backoff"
	BaseDelay   string `yaml:"base_delay"`  // e.g. "2s", "500ms"
	MaxDelay    string `yaml:"max_delay"`
}

// ParsedRetry holds retry configuration with resolved time.Duration values.
type ParsedRetry struct {
	MaxAttempts int
	Strategy    string
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// loader holds the directory path used to resolve workflow YAML files.
var workflowDir = "../workflows"

// SetWorkflowDir configures the directory from which YAML definitions are loaded.
// Called once during application startup from the config.
func SetWorkflowDir(dir string) {
	workflowDir = dir
}

// Load reads and parses a workflow definition by name (e.g. "create-environment").
func Load(name string) (*Definition, error) {
	path := filepath.Join(workflowDir, name+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("workflow definition %q not found at %s: %w", name, path, err)
	}

	var def Definition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("parse workflow definition %q: %w", name, err)
	}
	return &def, nil
}

// TopoSort returns steps in a dependency-respecting sequential order using Kahn's algorithm.
// Returns an error if a dependency cycle is detected.
func TopoSort(steps []StepDef) ([]StepDef, error) {
	byID := make(map[string]StepDef, len(steps))
	inDegree := make(map[string]int, len(steps))
	dependents := make(map[string][]string) // step ID → IDs of steps that depend on it

	for _, s := range steps {
		byID[s.ID] = s
		if _, ok := inDegree[s.ID]; !ok {
			inDegree[s.ID] = 0
		}
		for _, dep := range s.DependsOn {
			dependents[dep] = append(dependents[dep], s.ID)
			inDegree[s.ID]++
		}
	}

	// Seed queue with all steps that have no dependencies
	queue := make([]string, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	sorted := make([]StepDef, 0, len(steps))
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		sorted = append(sorted, byID[curr])
		for _, next := range dependents[curr] {
			inDegree[next]--
			if inDegree[next] == 0 {
				queue = append(queue, next)
			}
		}
	}

	if len(sorted) != len(steps) {
		return nil, fmt.Errorf("dependency cycle detected in workflow definition")
	}
	return sorted, nil
}

// ParseRetry converts raw YAML retry config into a ParsedRetry with resolved durations.
func ParseRetry(r RetryDef) ParsedRetry {
	baseDelay, _ := time.ParseDuration(r.BaseDelay)
	maxDelay, _ := time.ParseDuration(r.MaxDelay)

	maxAttempts := r.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	return ParsedRetry{
		MaxAttempts: maxAttempts,
		Strategy:    r.Strategy,
		BaseDelay:   baseDelay,
		MaxDelay:    maxDelay,
	}
}
