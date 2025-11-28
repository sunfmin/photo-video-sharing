package models

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Share represents sharing of media or album with a user
type Share struct {
	ID               string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MediaID          *string   `gorm:"type:uuid;index:idx_media_shared,priority:1"` // Nullable
	AlbumID          *string   `gorm:"type:uuid;index:idx_album_shared,priority:1"` // Nullable
	OwnerID          string    `gorm:"type:uuid;not null"`
	SharedWithUserID string    `gorm:"type:uuid;not null;index:idx_media_shared,priority:2;index:idx_album_shared,priority:2;index:idx_shared_with"`
	CreatedAt        time.Time
}

// TableName specifies the table name for Share model
func (Share) TableName() string {
	return "shares"
}

// BeforeCreate validates that exactly one of MediaID or AlbumID is set (XOR constraint)
func (s *Share) BeforeCreate(tx *gorm.DB) error {
	if (s.MediaID == nil && s.AlbumID == nil) || (s.MediaID != nil && s.AlbumID != nil) {
		return fmt.Errorf("exactly one of MediaID or AlbumID must be set")
	}
	return nil
}

