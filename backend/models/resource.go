package models

import "time"

// Resource tracks an individual provisioned resource within an environment.
type Resource struct {
	ID            string     `gorm:"primaryKey;type:text" json:"id"`
	EnvironmentID string     `gorm:"not null;index" json:"environment_id"`
	WorkflowID    string     `gorm:"not null;index" json:"workflow_id"`
	StepID        string     `gorm:"not null" json:"step_id"`
	Type          string     `gorm:"not null" json:"type"` // namespace | postgresql | redis | rabbitmq | deployment | dns
	Name          string     `gorm:"not null" json:"name"`
	Status        string     `gorm:"not null;default:'active'" json:"status"` // active | deleted
	Metadata      string     `gorm:"type:json" json:"metadata,omitempty"`     // JSON
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
