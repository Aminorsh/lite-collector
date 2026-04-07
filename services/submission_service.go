package services

import (
	"fmt"

	"lite-collector/models"
	"lite-collector/repository"
)

// SubmissionWithValues is the combined view returned to callers
type SubmissionWithValues struct {
	Submission *models.Submission
	Values     map[string]interface{}
}

// SubmissionService handles submission-related operations
type SubmissionService struct {
	submissionRepo repository.SubmissionRepository
	aiJobRepo      repository.AIJobRepository
}

// NewSubmissionService creates a new SubmissionService instance
func NewSubmissionService(submissionRepo repository.SubmissionRepository, aiJobRepo repository.AIJobRepository) *SubmissionService {
	return &SubmissionService{
		submissionRepo: submissionRepo,
		aiJobRepo:      aiJobRepo,
	}
}

// CreateSubmission persists a new submission and enqueues an AI anomaly-detection job.
// TODO Phase 3: replace mock with real DB + AI job creation
func (s *SubmissionService) CreateSubmission(formID string, submitterID uint64, values map[string]interface{}) (*models.Submission, error) {
	// TODO: parse formID, validate form exists and is published, persist submission
	// and submission_values via repository, then call aiJobRepo.Create
	submission := &models.Submission{
		ID:          1,
		FormID:      1,
		SubmitterID: submitterID,
		Status:      0,
	}
	return submission, nil
}

// GetMySubmissionWithValues returns the caller's submission for a form together
// with its field values as a flat map.
// TODO Phase 3: replace mock with real repository calls
func (s *SubmissionService) GetMySubmissionWithValues(formID string, userID uint64) (*SubmissionWithValues, error) {
	// TODO: parse formID, query submissionRepo.FindByFormIDAndSubmitterID,
	// then submissionRepo.FindValuesBySubmissionID
	if formID != "1" {
		return nil, fmt.Errorf("submission not found")
	}
	return &SubmissionWithValues{
		Submission: &models.Submission{
			ID:          1,
			FormID:      1,
			SubmitterID: userID,
			Status:      1,
		},
		Values: map[string]interface{}{
			"f_001": "张三",
			"f_002": "技术部",
		},
	}, nil
}
