package models

import "time"

// StepState represents the lifecycle state of an individual workflow step.
type StepState string

const (
	StepStatePending    StepState = "pending"
	StepStateRunning    StepState = "running"
	StepStateSucceeded  StepState = "succeeded"
	StepStateFailed     StepState = "failed"
	StepStateSkipped    StepState = "skipped"
	StepStateRolledBack StepState = "rolled_back"
	StepStateRollbackFailed StepState = "rollback_failed"
)

// WorkflowStep is a single step within a workflow execution.
type WorkflowStep struct {
	ID          string     `gorm:"primaryKey;type:text" json:"id"`
	WorkflowID  string     `gorm:"not null;index" json:"workflow_id"`
	Name        string     `gorm:"not null" json:"name"`
	ExecutorKey string     `gorm:"not null" json:"executor_key"`
	State       StepState  `gorm:"not null;default:'pending'" json:"state"`
	Attempt     int        `gorm:"not null;default:0" json:"attempt"`
	Input       string     `gorm:"type:json" json:"input,omitempty"`   // JSON
	Output      string     `gorm:"type:json" json:"output,omitempty"`  // JSON
	Error       string     `json:"error,omitempty"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
