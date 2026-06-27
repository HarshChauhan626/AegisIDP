package repository

import (
	"github.com/HarshChauhan626/AegisIDP/backend/models"
	"gorm.io/gorm"
)

// EventRepository persists workflow events.
type EventRepository interface {
	Create(event *models.Event) error
	ListByWorkflow(workflowID string) ([]models.Event, error)
	List(limit, offset int) ([]models.Event, int64, error)
}

type eventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) EventRepository {
	return &eventRepository{db: db}
}

func (r *eventRepository) Create(event *models.Event) error {
	return r.db.Create(event).Error
}

func (r *eventRepository) ListByWorkflow(workflowID string) ([]models.Event, error) {
	var events []models.Event
	if err := r.db.Where("workflow_id = ?", workflowID).Order("occurred_at ASC").Find(&events).Error; err != nil {
		return nil, err
	}
	return events, nil
}

func (r *eventRepository) List(limit, offset int) ([]models.Event, int64, error) {
	var events []models.Event
	var count int64
	if err := r.db.Model(&models.Event{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Order("occurred_at DESC").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		return nil, 0, err
	}
	return events, count, nil
}

// AuditLogRepository persists user-action audit records.
type AuditLogRepository interface {
	Create(log *models.AuditLog) error
	List(limit, offset int) ([]models.AuditLog, int64, error)
}

type auditLogRepository struct {
	db *gorm.DB
}

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

func (r *auditLogRepository) List(limit, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var count int64
	if err := r.db.Model(&models.AuditLog{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Order("occurred_at DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, count, nil
}
