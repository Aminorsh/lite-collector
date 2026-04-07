package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a WeChat user
type User struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	OpenID    string    `gorm:"size:64;uniqueIndex;not null" json:"openid"`
	Nickname  string    `gorm:"size:64" json:"nickname,omitempty"`
	AvatarURL string    `gorm:"size:255" json:"avatar_url,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to set timestamps
func (u *User) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	return nil
}

// BeforeUpdate hook to update timestamp
func (u *User) BeforeUpdate(tx *gorm.DB) error {
	u.UpdatedAt = time.Now()
	return nil
}