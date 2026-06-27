package models

import "time"

// WorkflowState represents the lifecycle state of a workflow execution.
type WorkflowState string

const (
	WorkflowStatePending     WorkflowState = "pending"
	WorkflowStateQueued      WorkflowState = "queued"
	WorkflowStateRunning     WorkflowState = "running"
	WorkflowStateCompleted   WorkflowState = "completed"
	WorkflowStateFailed      WorkflowState = "failed"
	WorkflowStateRollingBack WorkflowState = "rolling_back"
	WorkflowStateRolledBack  WorkflowState = "rolled_back"
	WorkflowStateCancelled   WorkflowState = "cancelled"
)

// WorkflowTrigger indicates how a workflow was started.
type WorkflowTrigger string

const (
	TriggerManual    WorkflowTrigger = "manual"
	TriggerScheduled WorkflowTrigger = "scheduled"
)

// WorkflowExecution is a single run of a workflow definition.
type WorkflowExecution struct {
	ID            string          `gorm:"primaryKey;type:text" json:"id"`
	EnvironmentID string          `gorm:"index" json:"environment_id"`
	TemplateID    string          `gorm:"index" json:"template_id,omitempty"`
	Type          string          `gorm:"not null" json:"type"` // e.g. create-environment
	State         WorkflowState   `gorm:"not null;default:'pending'" json:"state"`
	Trigger       WorkflowTrigger `gorm:"not null;default:'manual'" json:"trigger"`
	Input         string          `gorm:"type:json" json:"input"`  // JSON
	Error         string          `json:"error,omitempty"`
	StartedAt     *time.Time      `json:"started_at,omitempty"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	CreatedBy     string          `gorm:"not null" json:"created_by"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`

	// Relations
	Steps []WorkflowStep `gorm:"foreignKey:WorkflowID" json:"steps,omitempty"`
}

// ValidTransitions defines allowed state machine transitions for workflow executions.
var ValidWorkflowTransitions = map[WorkflowState][]WorkflowState{
	WorkflowStatePending:     {WorkflowStateQueued},
	WorkflowStateQueued:      {WorkflowStateRunning, WorkflowStateCancelled},
	WorkflowStateRunning:     {WorkflowStateCompleted, WorkflowStateFailed, WorkflowStateCancelled},
	WorkflowStateFailed:      {WorkflowStateRollingBack},
	WorkflowStateRollingBack: {WorkflowStateRolledBack},
}

// CanTransition checks if a state transition is valid.
func CanTransition(from, to WorkflowState) bool {
	allowed, ok := ValidWorkflowTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}
