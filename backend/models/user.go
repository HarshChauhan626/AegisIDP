package models

import (
	"time"
)

// User represents a platform user with a role for RBAC.
type User struct {
	ID           string     `gorm:"primaryKey;type:text" json:"id"`
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	Name         string     `gorm:"not null" json:"name"`
	Role         string     `gorm:"not null;default:'viewer'" json:"role"` // admin | developer | viewer
	Active       bool       `gorm:"not null;default:true" json:"active"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}

// Role constants
const (
	RoleAdmin     = "admin"
	RoleDeveloper = "developer"
	RoleViewer    = "viewer"
)
