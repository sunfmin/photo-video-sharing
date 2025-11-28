package models

import (
	"time"
)

// Album represents a collection of media items
type Album struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OwnerID     string    `gorm:"type:uuid;not null;index:idx_owner_created,priority:1"`
	Name        string    `gorm:"not null;size:100"`
	Description *string   // Nullable
	CreatedAt   time.Time `gorm:"index:idx_owner_created,priority:2,sort:desc"`
	UpdatedAt   time.Time
}

// TableName specifies the table name for Album model
func (Album) TableName() string {
	return "albums"
}

// AlbumMedia is the junction table for many-to-many relationship between albums and media
type AlbumMedia struct {
	AlbumID string    `gorm:"primaryKey;type:uuid"`
	MediaID string    `gorm:"primaryKey;type:uuid;index"`
	AddedAt time.Time `gorm:"autoCreateTime"`
}

// TableName specifies the table name for AlbumMedia model
func (AlbumMedia) TableName() string {
	return "album_media"
}

