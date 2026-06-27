package models

import "time"

// EnvironmentStatus tracks the lifecycle state of an environment.
type EnvironmentStatus string

const (
	EnvironmentStatusPending     EnvironmentStatus = "pending"
	EnvironmentStatusProvisioning EnvironmentStatus = "provisioning"
	EnvironmentStatusReady       EnvironmentStatus = "ready"
	EnvironmentStatusFailed      EnvironmentStatus = "failed"
	EnvironmentStatusDeleting    EnvironmentStatus = "deleting"
	EnvironmentStatusDeleted     EnvironmentStatus = "deleted"
)

// Environment represents a provisioned application environment.
type Environment struct {
	ID          string            `gorm:"primaryKey;type:text" json:"id"`
	ProjectID   string            `gorm:"not null;index" json:"project_id"`
	Name        string            `gorm:"not null" json:"name"`
	Status      EnvironmentStatus `gorm:"not null;default:'pending'" json:"status"`
	Config      string            `gorm:"type:json" json:"config"` // JSON blob
	CreatedBy   string            `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   *time.Time        `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Project   *Project            `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Workflows []WorkflowExecution `gorm:"foreignKey:EnvironmentID" json:"workflows,omitempty"`
}
