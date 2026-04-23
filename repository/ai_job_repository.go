package repository

import (
	"fmt"
	"time"

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
	// FindBySubmissionIDs returns completed detect_anomaly jobs keyed by submission ID.
	FindBySubmissionIDs(submissionIDs []uint64) ([]models.AIJob, error)
	// FindLatestCompletedByForm returns the most recent completed job of the given type for a form.
	// Returns gorm.ErrRecordNotFound when none exists.
	FindLatestCompletedByForm(formID uint64, jobType string) (*models.AIJob, error)
	// FindPendingAndRecentByUser returns a user's in-flight jobs plus any that finished
	// within the last `recentMinutes`. Used to drive the global "AI 生成中" banner.
	FindPendingAndRecentByUser(userID uint64, recentMinutes int) ([]models.AIJob, error)
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

// FindLatestCompletedByForm returns the most recent completed job of the given type for the given form.
func (r *aiJobRepository) FindLatestCompletedByForm(formID uint64, jobType string) (*models.AIJob, error) {
	var job models.AIJob
	silent := r.db.Session(&gorm.Session{Logger: r.db.Logger.LogMode(logger.Silent)})
	result := silent.
		Where("form_id = ? AND job_type = ? AND status = 2", formID, jobType).
		Order("finished_at DESC").
		First(&job)
	if result.Error != nil {
		return nil, result.Error
	}
	return &job, nil
}

// FindPendingAndRecentByUser returns in-flight and recently-finished jobs for a user.
func (r *aiJobRepository) FindPendingAndRecentByUser(userID uint64, recentMinutes int) ([]models.AIJob, error) {
	var jobs []models.AIJob
	cutoff := time.Now().Add(-time.Duration(recentMinutes) * time.Minute)
	result := r.db.Where(
		"user_id = ? AND (status IN (0,1) OR (status IN (2,3) AND finished_at IS NOT NULL AND finished_at > ?))",
		userID, cutoff,
	).Order("created_at DESC").Find(&jobs)
	if result.Error != nil {
		return nil, result.Error
	}
	return jobs, nil
}

// FindBySubmissionIDs returns completed detect_anomaly jobs matching the given submission IDs.
// The caller extracts submission_id from each job's Input JSON field.
func (r *aiJobRepository) FindBySubmissionIDs(submissionIDs []uint64) ([]models.AIJob, error) {
	if len(submissionIDs) == 0 {
		return nil, nil
	}
	// Build LIKE conditions to match {"submission_id":N} in the input field
	var jobs []models.AIJob
	tx := r.db.Where("job_type = ? AND status = 2", "detect_anomaly")
	conditions := make([]string, len(submissionIDs))
	args := make([]interface{}, len(submissionIDs))
	for i, sid := range submissionIDs {
		conditions[i] = "input LIKE ?"
		args[i] = fmt.Sprintf(`%%"submission_id":%d%%`, sid)
	}
	query := ""
	for i, c := range conditions {
		if i > 0 {
			query += " OR "
		}
		query += c
	}
	result := tx.Where(query, args...).Find(&jobs)
	return jobs, result.Error
}
