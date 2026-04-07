package models

import (
	"time"

	"gorm.io/gorm"
)

// Submission represents a form submission
type Submission struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	FormID     uint64    `gorm:"not null;index" json:"form_id"`
	SubmitterID uint64    `gorm:"not null;index" json:"submitter_id"`
	Status     int8      `gorm:"default:0" json:"status"` // 0:pending 1:normal 2:has_anomaly
	SubmittedAt time.Time `gorm:"autoCreateTime" json:"submitted_at"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for Submission
func (Submission) TableName() string {
	return "submissions"
}

// BeforeCreate hook to set timestamps
func (s *Submission) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	s.CreatedAt = now
	s.UpdatedAt = now
	return nil
}

// BeforeUpdate hook to update timestamp
func (s *Submission) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}