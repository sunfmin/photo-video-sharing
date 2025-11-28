package models

import (
	"time"

	"gorm.io/datatypes"
)

// Media represents a photo or video file
type Media struct {
	ID            string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OwnerID       string         `gorm:"type:uuid;not null;index:idx_owner_uploaded,priority:1"`
	Filename      string         `gorm:"not null"`
	FileType      string         `gorm:"not null;index"`
	FileSize      int64          `gorm:"not null"`
	StoragePath   string         `gorm:"not null"`
	ThumbnailPath *string        // Nullable
	Width         *int           // Nullable
	Height        *int           // Nullable
	Duration      *int           // Nullable (video only)
	EXIFData      datatypes.JSON `gorm:"type:jsonb"` // Nullable
	UploadedAt    time.Time      `gorm:"not null;index:idx_owner_uploaded,priority:2,sort:desc"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TableName specifies the table name for Media model
func (Media) TableName() string {
	return "media"
}

