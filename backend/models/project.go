package models

import "time"

// Project groups environments and workflows.
type Project struct {
	ID          string     `gorm:"primaryKey;type:text" json:"id"`
	Name        string     `gorm:"not null;uniqueIndex" json:"name"`
	Description string     `json:"description"`
	CreatedBy   string     `gorm:"not null" json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `gorm:"index" json:"deleted_at,omitempty"`

	// Relations
	Environments []Environment `gorm:"foreignKey:ProjectID" json:"environments,omitempty"`
}
