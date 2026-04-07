package models

import (
	"time"

	"gorm.io/gorm"
)

// SubmissionValue represents a single field value in a submission (EAV model)
type SubmissionValue struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SubmissionID uint64    `gorm:"not null;index" json:"submission_id"`
	FieldKey     string    `gorm:"size:64;not null;index" json:"field_key"` // matches key in form schema
	Value        string    `gorm:"type:text" json:"value"`                  // all values stored as string
	IsAnomaly    bool      `gorm:"default:false" json:"is_anomaly"`
	AnomalyReason string   `gorm:"size:255" json:"anomaly_reason,omitempty"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for SubmissionValue
func (SubmissionValue) TableName() string {
	return "submission_values"
}

// BeforeCreate hook to set timestamps
func (sv *SubmissionValue) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	sv.CreatedAt = now
	sv.UpdatedAt = now
	return nil
}

// BeforeUpdate hook to update timestamp
func (sv *SubmissionValue) BeforeUpdate(tx *gorm.DB) error {
	sv.UpdatedAt = time.Now()
	return nil
}