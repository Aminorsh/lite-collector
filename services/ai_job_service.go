package services

import (
	"encoding/json"
	"errors"
	"fmt"

	"lite-collector/models"
	"lite-collector/repository"
	"lite-collector/utils"

	"gorm.io/gorm"
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
	fid := formID
	job := &models.AIJob{
		UserID:  userID,
		JobType: "generate_report",
		FormID:  &fid,
		Status:  0,
		Input:   fmt.Sprintf(`{"form_id":%d}`, formID),
	}
	if err := s.aiJobRepo.Create(job); err != nil {
		return 0, utils.ErrInternal
	}
	return job.ID, nil
}

// EnqueueFormGeneration creates a queued AI job to generate a form schema from a description.
func (s *AIJobService) EnqueueFormGeneration(userID uint64, description string) (uint64, error) {
	payload, err := json.Marshal(map[string]string{"description": description})
	if err != nil {
		return 0, utils.ErrInternal
	}
	job := &models.AIJob{
		UserID:  userID,
		JobType: "generate_form",
		Status:  0,
		Input:   string(payload),
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

// GetLatestReport returns the most recent completed report job for the given form,
// or nil without error when no completed report exists.
func (s *AIJobService) GetLatestReport(formID uint64) (*models.AIJob, error) {
	job, err := s.aiJobRepo.FindLatestCompletedByForm(formID, "generate_report")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, utils.ErrInternal
	}
	return job, nil
}

// ListPendingJobs returns the user's in-flight jobs plus any that finished in
// the last 10 minutes, so the banner can surface "recently completed" state too.
func (s *AIJobService) ListPendingJobs(userID uint64) ([]models.AIJob, error) {
	jobs, err := s.aiJobRepo.FindPendingAndRecentByUser(userID, 10)
	if err != nil {
		return nil, utils.ErrInternal
	}
	return jobs, nil
}
