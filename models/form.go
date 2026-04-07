package models

import (
	"time"

	"gorm.io/gorm"
)

// Form represents a data collection form
type Form struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OwnerID     uint64    `gorm:"not null;index" json:"owner_id"`
	Title       string    `gorm:"size:128;not null" json:"title"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Schema      []byte    `gorm:"type:json;not null" json:"form_schema"` // JSON schema for form fields
	Status      int8      `gorm:"default:0" json:"status"`          // 0:draft 1:published 2:archived
	TemplateYear *int16   `gorm:"type:year" json:"template_year,omitempty"` // NULL if not a template
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for Form
func (Form) TableName() string {
	return "forms"
}

// BeforeCreate hook to set timestamps
func (f *Form) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	f.CreatedAt = now
	f.UpdatedAt = now
	return nil
}

// BeforeUpdate hook to update timestamp
func (f *Form) BeforeUpdate(tx *gorm.DB) error {
	f.UpdatedAt = time.Now()
	return nil
}