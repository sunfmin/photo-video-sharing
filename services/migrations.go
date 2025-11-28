package services

import (
	"gorm.io/gorm"

	"github.com/yourorg/photo-video-sharing/internal/models"
)

// AutoMigrate runs database migrations for all models
// This function is exported so external applications can use it
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Media{},
		&models.Album{},
		&models.AlbumMedia{},
		&models.Share{},
	)
}

