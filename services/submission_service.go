package services

import (
	"fmt"

	"lite-collector/models"
	"lite-collector/repository"
	"lite-collector/utils"
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

// CreateSubmission persists a new submission with its field values, then enqueues
// an AI anomaly-detection job. All three writes are best-effort: if the job enqueue
// fails we log and continue — the submission itself is not rolled back.
func (s *SubmissionService) CreateSubmission(formID string, submitterID uint64, values map[string]interface{}) (*models.Submission, error) {
	id, err := parseFormID(formID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	// Persist the submission record
	submission := &models.Submission{
		FormID:      id,
		SubmitterID: submitterID,
		Status:      0, // pending — AI review not done yet
	}
	if err := s.submissionRepo.Create(submission); err != nil {
		return nil, utils.ErrSubmissionCreateFail
	}

	// Persist each field value as an EAV row (values are coerced to string)
	for key, val := range values {
		sv := &models.SubmissionValue{
			SubmissionID: submission.ID,
			FieldKey:     key,
			Value:        fmt.Sprintf("%v", val),
		}
		if err := s.submissionRepo.CreateValue(sv); err != nil {
			// Non-fatal: submission exists, partial values are better than nothing.
			// Phase 3 can add proper transaction rollback if needed.
			continue
		}
	}

	// Enqueue AI anomaly detection — fire and forget, failure is non-fatal
	_ = s.aiJobRepo.Create(&models.AIJob{
		UserID:  submitterID,
		JobType: "detect_anomaly",
		Status:  0, // queued
		Input:   fmt.Sprintf(`{"submission_id":%d}`, submission.ID),
	})

	return submission, nil
}

// GetMySubmissionWithValues returns the caller's submission for a form together
// with its field values as a flat map.
func (s *SubmissionService) GetMySubmissionWithValues(formID string, userID uint64) (*SubmissionWithValues, error) {
	id, err := parseFormID(formID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}

	submission, err := s.submissionRepo.FindByFormIDAndSubmitterID(id, userID)
	if err != nil {
		return nil, utils.ErrSubmissionNotFound
	}

	rows, err := s.submissionRepo.FindValuesBySubmissionID(submission.ID)
	if err != nil {
		return nil, utils.ErrInternal
	}

	values := make(map[string]interface{}, len(rows))
	for _, row := range rows {
		values[row.FieldKey] = row.Value
	}

	return &SubmissionWithValues{
		Submission: submission,
		Values:     values,
	}, nil
}

// GetSubmissionsByFormID returns all submissions for a form.
// Ownership of the form must be verified by the caller before invoking this.
func (s *SubmissionService) GetSubmissionsByFormID(formID string) ([]models.Submission, error) {
	id, err := parseFormID(formID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}
	submissions, err := s.submissionRepo.FindByFormID(id)
	if err != nil {
		return nil, utils.ErrInternal
	}
	return submissions, nil
}

// GetSubmissionByIDWithValues returns a single submission with its field values.
// Ownership of the parent form must be verified by the caller before invoking this.
func (s *SubmissionService) GetSubmissionByIDWithValues(submissionID string) (*SubmissionWithValues, error) {
	id, err := parseFormID(submissionID)
	if err != nil {
		return nil, utils.ErrBadRequest
	}
	submission, err := s.submissionRepo.FindByID(id)
	if err != nil {
		return nil, utils.ErrSubmissionNotFound
	}
	rows, err := s.submissionRepo.FindValuesBySubmissionID(submission.ID)
	if err != nil {
		return nil, utils.ErrInternal
	}
	values := make(map[string]interface{}, len(rows))
	for _, row := range rows {
		values[row.FieldKey] = row.Value
	}
	return &SubmissionWithValues{Submission: submission, Values: values}, nil
}

func parseFormID(s string) (uint64, error) {
	var id uint64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}
