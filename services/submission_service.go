package services

import (
	"lite-collector/models"
	"lite-collector/repository"
)

// SubmissionService handles submission-related operations
type SubmissionService struct {
	submissionRepo repository.SubmissionRepository
}

// NewSubmissionService creates a new SubmissionService instance with dependency injection
func NewSubmissionService(submissionRepo repository.SubmissionRepository) *SubmissionService {
	return &SubmissionService{
		submissionRepo: submissionRepo,
	}
}

// CreateSubmission creates a new form submission
func (s *SubmissionService) CreateSubmission(formID string, submitterID uint64, values map[string]interface{}) (*models.Submission, error) {
	// In a real implementation, this would:
	// 1. Validate the form exists and is published
	// 2. Validate the submission values against the form schema
	// 3. Create submission record
	// 4. Create submission value records for each field
	// 5. Return the created submission

	// For now, we'll return a mock submission
	return &models.Submission{
		ID:         1,
		FormID:     1, // Would parse formID in real implementation
		SubmitterID: submitterID,
		Status:     0, // pending
	}, nil
}

// GetMySubmission gets the current user's submission for a form
func (s *SubmissionService) GetMySubmission(formID string, userID uint64) (*models.Submission, error) {
	// In a real implementation, this would query for submission by form_id and submitter_id
	// For now, we'll return a mock submission if ID is "1"
	if formID == "1" {
		return &models.Submission{
			ID:         1,
			FormID:     1,
			SubmitterID: userID,
			Status:     1, // normal
		}, nil
	}
	return nil, nil // Would return error in real implementation
}

// GetSubmissionValues gets all values for a submission
func (s *SubmissionService) GetSubmissionValues(submissionID uint64) ([]models.SubmissionValue, error) {
	// In a real implementation, this would query submission_values by submission_id
	// For now, we'll return mock values
	return []models.SubmissionValue{
		{
			ID:           1,
			SubmissionID: submissionID,
			FieldKey:     "f_001",
			Value:        "张三",
		},
		{
			ID:           2,
			SubmissionID: submissionID,
			FieldKey:     "f_002",
			Value:        "技术部",
		},
	}, nil
}