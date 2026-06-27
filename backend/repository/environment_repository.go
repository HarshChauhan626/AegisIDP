package repository

import (
	"github.com/HarshChauhan626/AegisIDP/backend/models"
	"gorm.io/gorm"
)

// EnvironmentRepository defines data access for environments.
type EnvironmentRepository interface {
	Create(env *models.Environment) error
	FindByID(id string) (*models.Environment, error)
	List(projectID string) ([]models.Environment, error)
	Update(env *models.Environment) error
	Delete(id string) error
}

type environmentRepository struct {
	db *gorm.DB
}

func NewEnvironmentRepository(db *gorm.DB) EnvironmentRepository {
	return &environmentRepository{db: db}
}

func (r *environmentRepository) Create(env *models.Environment) error {
	return r.db.Create(env).Error
}

func (r *environmentRepository) FindByID(id string) (*models.Environment, error) {
	var env models.Environment
	if err := r.db.First(&env, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &env, nil
}

func (r *environmentRepository) List(projectID string) ([]models.Environment, error) {
	var envs []models.Environment
	q := r.db
	if projectID != "" {
		q = q.Where("project_id = ?", projectID)
	}
	if err := q.Find(&envs).Error; err != nil {
		return nil, err
	}
	return envs, nil
}

func (r *environmentRepository) Update(env *models.Environment) error {
	return r.db.Save(env).Error
}

func (r *environmentRepository) Delete(id string) error {
	return r.db.Delete(&models.Environment{}, "id = ?", id).Error
}
