package services

import (
	"lite-collector/models"
)

// AIJobService handles AI job-related operations
type AIJobService struct{}

// NewAIJobService creates a new AIJobService instance
func NewAIJobService() *AIJobService {
	return &AIJobService{}
}

// EnqueueAnomalyDetection creates an anomaly detection job for a submission
func (s *AIJobService) EnqueueAnomalyDetection(submissionID uint64) {
	// In a real implementation, this would:
	// 1. Create an AIJob record with type "detect_anomaly"
	// 2. Submit the job to a queue/worker for processing
	// 3. The worker would process the submission with AI and update the job
	//
	// For now, we'll just log that we're enqueuing the job
	// In a real app, you'd use a proper job queue like RabbitMQ, Redis Queue, etc.
}

// GetJobStatus gets the status of an AI job
func (s *AIJobService) GetJobStatus(jobID uint64) (*models.AIJob, error) {
	// In a real implementation, this would query the database
	// For now, we'll return a mock job
	return &models.AIJob{
		ID:      jobID,
		UserID:  1,
		JobType: "detect_anomaly",
		Status:  2, // done
		Output:  `{"anomalies": []}`,
	}, nil
}