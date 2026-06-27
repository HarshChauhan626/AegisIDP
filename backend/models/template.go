package models

import "time"

// Template stores a saved workflow definition authored via the Monaco Editor.
type Template struct {
	ID         string    `gorm:"primaryKey;type:text" json:"id"`
	Name       string    `gorm:"not null;uniqueIndex" json:"name"`
	Definition string    `gorm:"not null;type:text" json:"definition"` // YAML source
	Version    int       `gorm:"not null;default:1" json:"version"`
	CreatedBy  string    `gorm:"not null" json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relations
	Schedules []Schedule `gorm:"foreignKey:TemplateID" json:"schedules,omitempty"`
}
