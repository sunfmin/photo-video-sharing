package models

import (
	"time"
)

// User represents a registered account
type User struct {
	ID            string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email         string    `gorm:"uniqueIndex;not null"`
	PasswordHash  string    `gorm:"not null"`
	StorageUsed   int64     `gorm:"default:0"`
	StorageQuota  int64     `gorm:"default:524288000"` // 500MB
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

