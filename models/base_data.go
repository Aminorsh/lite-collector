package models

import (
	"time"

	"gorm.io/gorm"
)

// BaseData represents reference data for prefilling forms
type BaseData struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	FormID    uint64    `gorm:"not null;index" json:"form_id"`
	RowKey    string    `gorm:"size:64;not null;index" json:"row_key"` // lookup key (e.g. employee ID)
	Data      []byte    `gorm:"type:json;not null" json:"data"`        // prefill values for this record
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for BaseData
func (BaseData) TableName() string {
	return "base_data"
}

// BeforeCreate hook to set timestamps
func (bd *BaseData) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	bd.CreatedAt = now
	bd.UpdatedAt = now
	return nil
}

// BeforeUpdate hook to update timestamp
func (bd *BaseData) BeforeUpdate(tx *gorm.DB) error {
	bd.UpdatedAt = time.Now()
	return nil
}