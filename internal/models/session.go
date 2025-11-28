package models

import (
	"time"
)

// Session represents an active user session
type Session struct {
	ID         string    `gorm:"primaryKey;type:uuid"`
	UserID     string    `gorm:"type:uuid;not null;index"`
	ExpiresAt  time.Time `gorm:"not null;index"`
	LastUsedAt time.Time `gorm:"not null"`
	CreatedAt  time.Time
}

// TableName specifies the table name for Session model
func (Session) TableName() string {
	return "sessions"
}

