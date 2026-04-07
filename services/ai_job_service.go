package services

import (
	"lite-collector/models"
	"lite-collector/repository"
)

// AIJobService handles AI job-related operations
type AIJobService struct {
	aiJobRepo repository.AIJobRepository
}

// NewAIJobService creates a new AIJobService instance
func NewAIJobService(aiJobRepo repository.AIJobRepository) *AIJobService {
	return &AIJobService{aiJobRepo: aiJobRepo}
}

// EnqueueAnomalyDetection creates an anomaly detection job for a submission.
// TODO Phase 3: persist job to DB and dispatch to async worker
func (s *AIJobService) EnqueueAnomalyDetection(submissionID uint64) {
	// placeholder — real implementation will call s.aiJobRepo.Create(...)
}

// GetJobStatus returns the current status of an AI job.
// TODO Phase 3: replace mock with s.aiJobRepo.FindByID(jobID)
func (s *AIJobService) GetJobStatus(jobID uint64) (*models.AIJob, error) {
	return &models.AIJob{
		ID:      jobID,
		UserID:  1,
		JobType: "detect_anomaly",
		Status:  2,
		Output:  `{"anomalies": []}`,
	}, nil
}
