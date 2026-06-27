package models

import "time"

// EventType classifies the kind of event emitted by the workflow engine.
type EventType string

const (
	EventWorkflowStarted    EventType = "WorkflowStarted"
	EventWorkflowCompleted  EventType = "WorkflowCompleted"
	EventWorkflowFailed     EventType = "WorkflowFailed"
	EventWorkflowCancelled  EventType = "WorkflowCancelled"
	EventStepStarted        EventType = "StepStarted"
	EventStepSucceeded      EventType = "StepSucceeded"
	EventStepFailed         EventType = "StepFailed"
	EventRetryTriggered     EventType = "RetryTriggered"
	EventRollbackStarted    EventType = "RollbackStarted"
	EventRollbackCompleted  EventType = "RollbackCompleted"
	EventResourceCreated    EventType = "ResourceCreated"
	EventResourceDeleted    EventType = "ResourceDeleted"
)

// Event is an immutable record of something that happened during workflow execution.
type Event struct {
	ID         string    `gorm:"primaryKey;type:text" json:"id"`
	Type       EventType `gorm:"not null;index" json:"type"`
	WorkflowID string    `gorm:"not null;index" json:"workflow_id"`
	StepID     string    `gorm:"index" json:"step_id,omitempty"`
	Payload    string    `gorm:"type:json" json:"payload,omitempty"` // JSON
	OccurredAt time.Time `gorm:"not null;index" json:"occurred_at"`
}
