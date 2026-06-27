package models

import "time"

// Schedule configures a cron-based trigger for a workflow template.
type Schedule struct {
	ID             string     `gorm:"primaryKey;type:text" json:"id"`
	TemplateID     string     `gorm:"not null;index" json:"template_id"`
	CronExpression string     `gorm:"not null" json:"cron_expression"`
	Enabled        bool       `gorm:"not null;default:true" json:"enabled"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time `json:"next_run_at,omitempty"`
	CreatedBy      string     `gorm:"not null" json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	Template *Template `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
}
