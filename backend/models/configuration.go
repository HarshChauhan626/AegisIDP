package models

import "time"

// Configuration stores key-value settings for a project or environment.
type Configuration struct {
	ID            string    `gorm:"primaryKey;type:text" json:"id"`
	Scope         string    `gorm:"not null;index" json:"scope"`      // project | environment | global
	ScopeID       string    `gorm:"not null;index" json:"scope_id"`
	Key           string    `gorm:"not null" json:"key"`
	Value         string    `json:"value"`
	Encrypted     bool      `gorm:"not null;default:false" json:"encrypted"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
