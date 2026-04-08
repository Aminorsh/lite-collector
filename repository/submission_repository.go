package repository

import (
	"lite-collector/models"

	"gorm.io/gorm"
)

// SubmissionRepository defines the interface for submission data access
type SubmissionRepository interface {
	Create(submission *models.Submission) error
	FindByID(id uint64) (*models.Submission, error)
	FindByFormIDAndSubmitterID(formID, submitterID uint64) (*models.Submission, error)
	Update(submission *models.Submission) error
	CreateValue(value *models.SubmissionValue) error
	FindValuesBySubmissionID(submissionID uint64) ([]models.SubmissionValue, error)
}

// submissionRepository implements SubmissionRepository using GORM
type submissionRepository struct {
	db *gorm.DB
}

// NewSubmissionRepository creates a new submission repository instance
func NewSubmissionRepository(db *gorm.DB) SubmissionRepository {
	return &submissionRepository{db: db}
}

func (r *submissionRepository) Create(submission *models.Submission) error {
	return r.db.Create(submission).Error
}

func (r *submissionRepository) FindByID(id uint64) (*models.Submission, error) {
	var submission models.Submission
	result := r.db.First(&submission, id)
	return &submission, result.Error
}

func (r *submissionRepository) FindByFormIDAndSubmitterID(formID, submitterID uint64) (*models.Submission, error) {
	var submission models.Submission
	result := r.db.Where("form_id = ? AND submitter_id = ?", formID, submitterID).First(&submission)
	return &submission, result.Error
}

func (r *submissionRepository) Update(submission *models.Submission) error {
	return r.db.Save(submission).Error
}

func (r *submissionRepository) CreateValue(value *models.SubmissionValue) error {
	return r.db.Create(value).Error
}

func (r *submissionRepository) FindValuesBySubmissionID(submissionID uint64) ([]models.SubmissionValue, error) {
	var values []models.SubmissionValue
	result := r.db.Where("submission_id = ?", submissionID).Find(&values)
	return values, result.Error
}
