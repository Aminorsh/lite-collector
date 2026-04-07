package repository

import (
	"lite-collector/models"
)

// SubmissionRepository defines the interface for submission data access
type SubmissionRepository interface {
	Create(submission *models.Submission) error
	FindByID(id uint64) (*models.Submission, error)
	FindByFormIDAndSubmitterID(formID, submitterID uint64) (*models.Submission, error)
	Update(submission *models.Submission) error
}

// submissionRepository implements SubmissionRepository using GORM
type submissionRepository struct {
	db *gorm.DB
}

// NewSubmissionRepository creates a new submission repository instance
func NewSubmissionRepository(db *gorm.DB) SubmissionRepository {
	return &submissionRepository{db: db}
}

// Create creates a new submission
func (r *submissionRepository) Create(submission *models.Submission) error {
	return r.db.Create(submission).Error
}

// FindByID finds a submission by ID
func (r *submissionRepository) FindByID(id uint64) (*models.Submission, error) {
	var submission models.Submission
	result := r.db.First(&submission, id)
	return &submission, result.Error
}

// FindByFormIDAndSubmitterID finds a submission by form ID and submitter ID
func (r *submissionRepository) FindByFormIDAndSubmitterID(formID, submitterID uint64) (*models.Submission, error) {
	var submission models.Submission
	result := r.db.Where("form_id = ? AND submitter_id = ?", formID, submitterID).First(&submission)
	return &submission, result.Error
}

// Update updates an existing submission
func (r *submissionRepository) Update(submission *models.Submission) error {
	return r.db.Save(submission).Error
}