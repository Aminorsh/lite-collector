package repository

import (
	"lite-collector/models"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// AIJobRepository defines the interface for AI job data access
type AIJobRepository interface {
	Create(job *models.AIJob) error
	FindByID(id uint64) (*models.AIJob, error)
	// ClaimQueued atomically finds one queued job and sets its status to processing.
	// Returns nil if no queued jobs exist.
	ClaimQueued() (*models.AIJob, error)
	Update(job *models.AIJob) error
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

// ClaimQueued atomically finds one queued job and sets its status to processing.
func (r *aiJobRepository) ClaimQueued() (*models.AIJob, error) {
	var job models.AIJob
	// Silence "record not found" — it's expected when the queue is empty
	silent := r.db.Session(&gorm.Session{Logger: r.db.Logger.LogMode(logger.Silent)})
	result := silent.Where("status = 0").Order("created_at ASC").First(&job)
	if result.Error != nil {
		return nil, result.Error
	}
	job.Status = 1 // processing
	if err := r.db.Save(&job).Error; err != nil {
		return nil, err
	}
	return &job, nil
}

// Update updates an AI job
func (r *aiJobRepository) Update(job *models.AIJob) error {
	return r.db.Save(job).Error
}
