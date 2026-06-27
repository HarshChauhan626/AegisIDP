package repository

import (
	"github.com/HarshChauhan626/AegisIDP/backend/models"
	"gorm.io/gorm"
)

// WorkflowRepository defines data access for workflow executions and steps.
type WorkflowRepository interface {
	CreateExecution(wf *models.WorkflowExecution) error
	FindExecutionByID(id string) (*models.WorkflowExecution, error)
	ListExecutions(environmentID string, limit, offset int) ([]models.WorkflowExecution, int64, error)
	UpdateExecution(wf *models.WorkflowExecution) error

	CreateStep(step *models.WorkflowStep) error
	FindStepByID(id string) (*models.WorkflowStep, error)
	ListStepsByWorkflow(workflowID string) ([]models.WorkflowStep, error)
	UpdateStep(step *models.WorkflowStep) error
}

type workflowRepository struct {
	db *gorm.DB
}

func NewWorkflowRepository(db *gorm.DB) WorkflowRepository {
	return &workflowRepository{db: db}
}

func (r *workflowRepository) CreateExecution(wf *models.WorkflowExecution) error {
	return r.db.Create(wf).Error
}

func (r *workflowRepository) FindExecutionByID(id string) (*models.WorkflowExecution, error) {
	var wf models.WorkflowExecution
	if err := r.db.Preload("Steps").First(&wf, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &wf, nil
}

func (r *workflowRepository) ListExecutions(environmentID string, limit, offset int) ([]models.WorkflowExecution, int64, error) {
	var wfs []models.WorkflowExecution
	var count int64
	q := r.db.Model(&models.WorkflowExecution{})
	if environmentID != "" {
		q = q.Where("environment_id = ?", environmentID)
	}
	if err := q.Count(&count).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Preload("Steps").Order("created_at DESC").Limit(limit).Offset(offset).Find(&wfs).Error; err != nil {
		return nil, 0, err
	}
	return wfs, count, nil
}

func (r *workflowRepository) UpdateExecution(wf *models.WorkflowExecution) error {
	return r.db.Save(wf).Error
}

func (r *workflowRepository) CreateStep(step *models.WorkflowStep) error {
	return r.db.Create(step).Error
}

func (r *workflowRepository) FindStepByID(id string) (*models.WorkflowStep, error) {
	var step models.WorkflowStep
	if err := r.db.First(&step, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &step, nil
}

func (r *workflowRepository) ListStepsByWorkflow(workflowID string) ([]models.WorkflowStep, error) {
	var steps []models.WorkflowStep
	if err := r.db.Where("workflow_id = ?", workflowID).Order("created_at ASC").Find(&steps).Error; err != nil {
		return nil, err
	}
	return steps, nil
}

func (r *workflowRepository) UpdateStep(step *models.WorkflowStep) error {
	return r.db.Save(step).Error
}
