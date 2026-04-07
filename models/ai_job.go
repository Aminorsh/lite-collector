package models

import (
	"time"

	"gorm.io/gorm"
)

// AIJob represents an asynchronous AI job
type AIJob struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"not null;index" json:"user_id"`
	JobType   string    `gorm:"size:32;not null;index" json:"type"` // generate_form | generate_report | detect_anomaly
	Status    int8      `gorm:"default:0" json:"status"`            // 0:queued 1:processing 2:done 3:failed
	Input     string    `gorm:"type:text" json:"input,omitempty"`   // Input data for the AI job
	Output    string    `gorm:"type:text" json:"output,omitempty"`  // Output/result from the AI job
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	FinishedAt *time.Time `gorm:"type:datetime" json:"finished_at,omitempty"` // When job finished
}

// TableName specifies the table name for AIJob
func (AIJob) TableName() string {
	return "ai_jobs"
}

// BeforeCreate hook to set timestamps
func (aj *AIJob) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	aj.CreatedAt = now
	return nil
}

// BeforeUpdate hook to update timestamp
func (aj *AIJob) BeforeUpdate(tx *gorm.DB) error {
	now := time.Now()
	aj.FinishedAt = &now
	return nil
}