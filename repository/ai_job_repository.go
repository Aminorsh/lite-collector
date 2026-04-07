package repository

import (
	"lite-collector/models"

	"gorm.io/gorm"
)

// AIJobRepository defines the interface for AI job data access
type AIJobRepository interface {
	Create(job *models.AIJob) error
	FindByID(id uint64) (*models.AIJob, error)
}

// aiJobRepository implements AIJobRepository using GORM
type aiJobRepository struct {
	db *gorm.DB
}

// NewAIJobRepository creates a new AI job repository instance
func NewAIJobRepository(db *gorm.DB) AIJobRepository {
	return &aiJobRepository{db: db}
}

// Create creates a new AI job
func (r *aiJobRepository) Create(job *models.AIJob) error {
	return r.db.Create(job).Error
}

// FindByID finds an AI job by ID
func (r *aiJobRepository) FindByID(id uint64) (*models.AIJob, error) {
	var job models.AIJob
	result := r.db.First(&job, id)
	return &job, result.Error
}
