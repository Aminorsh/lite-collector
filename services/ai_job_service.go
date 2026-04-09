package services

import (
	"fmt"

	"lite-collector/models"
	"lite-collector/repository"
	"lite-collector/utils"
)

// AIJobService handles AI job-related operations
type AIJobService struct {
	aiJobRepo repository.AIJobRepository
}

// NewAIJobService creates a new AIJobService instance
func NewAIJobService(aiJobRepo repository.AIJobRepository) *AIJobService {
	return &AIJobService{aiJobRepo: aiJobRepo}
}

// EnqueueAnomalyDetection creates a queued AI job record for the given submission.
// Actual processing happens in Phase 3 when the async worker is implemented.
func (s *AIJobService) EnqueueAnomalyDetection(userID, submissionID uint64) error {
	job := &models.AIJob{
		UserID:  userID,
		JobType: "detect_anomaly",
		Status:  0, // queued
		Input:   fmt.Sprintf(`{"submission_id":%d}`, submissionID),
	}
	if err := s.aiJobRepo.Create(job); err != nil {
		return utils.ErrInternal
	}
	return nil
}

// EnqueueReport creates a queued AI job to generate a summary report for a form.
func (s *AIJobService) EnqueueReport(userID, formID uint64) (uint64, error) {
	job := &models.AIJob{
		UserID:  userID,
		JobType: "generate_report",
		Status:  0,
		Input:   fmt.Sprintf(`{"form_id":%d}`, formID),
	}
	if err := s.aiJobRepo.Create(job); err != nil {
		return 0, utils.ErrInternal
	}
	return job.ID, nil
}

// GetJobStatus returns the current status of an AI job.
func (s *AIJobService) GetJobStatus(jobID uint64) (*models.AIJob, error) {
	job, err := s.aiJobRepo.FindByID(jobID)
	if err != nil {
		return nil, utils.ErrNotFound
	}
	return job, nil
}
