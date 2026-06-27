package models

import "time"

// AuditLog records every user-initiated action for compliance and auditing.
type AuditLog struct {
	ID           string    `gorm:"primaryKey;type:text" json:"id"`
	UserID       string    `gorm:"not null;index" json:"user_id"`
	Action       string    `gorm:"not null;index" json:"action"` // e.g. environment.create
	ResourceType string    `gorm:"not null" json:"resource_type"`
	ResourceID   string    `gorm:"not null" json:"resource_id"`
	Metadata     string    `gorm:"type:json" json:"metadata,omitempty"` // JSON
	OccurredAt   time.Time `gorm:"not null;index" json:"occurred_at"`
}
