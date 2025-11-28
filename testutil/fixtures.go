package testutil

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateTestUser creates a test user with default or custom values
func CreateTestUser(db *gorm.DB, overrides map[string]interface{}) *TestUser {
	user := &TestUser{
		ID:            uuid.New().String(),
		Email:         "test@example.com",
		PasswordHash:  mustHashPassword("Password123"),
		StorageUsed:   0,
		StorageQuota:  524288000, // 500MB
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Apply overrides
	if email, ok := overrides["email"].(string); ok {
		user.Email = email
	}
	if password, ok := overrides["password"].(string); ok {
		user.PasswordHash = mustHashPassword(password)
	}
	if quota, ok := overrides["storage_quota"].(int64); ok {
		user.StorageQuota = quota
	}
	if used, ok := overrides["storage_used"].(int64); ok {
		user.StorageUsed = used
	}

	if err := db.Create(user).Error; err != nil {
		panic("failed to create test user: " + err.Error())
	}

	return user
}

// CreateTestSession creates a test session for a user
func CreateTestSession(db *gorm.DB, userID string, overrides map[string]interface{}) *TestSession {
	session := &TestSession{
		ID:         uuid.New().String(),
		UserID:     userID,
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
		LastUsedAt: time.Now(),
		CreatedAt:  time.Now(),
	}

	// Apply overrides
	if expiresAt, ok := overrides["expires_at"].(time.Time); ok {
		session.ExpiresAt = expiresAt
	}

	if err := db.Create(session).Error; err != nil {
		panic("failed to create test session: " + err.Error())
	}

	return session
}

// CreateTestMedia creates a test media item
func CreateTestMedia(db *gorm.DB, overrides map[string]interface{}) *TestMedia {
	media := &TestMedia{
		ID:            uuid.New().String(),
		OwnerID:       "", // Must be provided
		Filename:      "test-photo.jpg",
		FileType:      "image/jpeg",
		FileSize:      1024000, // 1MB
		StoragePath:   "/test/test-photo.jpg",
		ThumbnailPath: stringPtr("/test/test-photo_thumb.jpg"),
		Width:         intPtr(1920),
		Height:        intPtr(1080),
		UploadedAt:    time.Now(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Apply overrides
	if ownerID, ok := overrides["owner_id"].(string); ok {
		media.OwnerID = ownerID
	}
	if filename, ok := overrides["filename"].(string); ok {
		media.Filename = filename
	}
	if fileType, ok := overrides["file_type"].(string); ok {
		media.FileType = fileType
	}
	if fileSize, ok := overrides["file_size"].(int64); ok {
		media.FileSize = fileSize
	}
	if storagePath, ok := overrides["storage_path"].(string); ok {
		media.StoragePath = storagePath
	}

	if media.OwnerID == "" {
		panic("owner_id is required for CreateTestMedia")
	}

	if err := db.Create(media).Error; err != nil {
		panic("failed to create test media: " + err.Error())
	}

	return media
}

// CreateTestAlbum creates a test album
func CreateTestAlbum(db *gorm.DB, overrides map[string]interface{}) *TestAlbum {
	album := &TestAlbum{
		ID:          uuid.New().String(),
		OwnerID:     "", // Must be provided
		Name:        "Test Album",
		Description: stringPtr("A test album"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Apply overrides
	if ownerID, ok := overrides["owner_id"].(string); ok {
		album.OwnerID = ownerID
	}
	if name, ok := overrides["name"].(string); ok {
		album.Name = name
	}
	if desc, ok := overrides["description"].(*string); ok {
		album.Description = desc
	}

	if album.OwnerID == "" {
		panic("owner_id is required for CreateTestAlbum")
	}

	if err := db.Create(album).Error; err != nil {
		panic("failed to create test album: " + err.Error())
	}

	return album
}

// CreateTestShare creates a test share
func CreateTestShare(db *gorm.DB, overrides map[string]interface{}) *TestShare {
	share := &TestShare{
		ID:               uuid.New().String(),
		OwnerID:          "", // Must be provided
		SharedWithUserID: "", // Must be provided
		CreatedAt:        time.Now(),
	}

	// Apply overrides
	if ownerID, ok := overrides["owner_id"].(string); ok {
		share.OwnerID = ownerID
	}
	if sharedWith, ok := overrides["shared_with_user_id"].(string); ok {
		share.SharedWithUserID = sharedWith
	}
	if mediaID, ok := overrides["media_id"].(string); ok {
		share.MediaID = stringPtr(mediaID)
	}
	if albumID, ok := overrides["album_id"].(string); ok {
		share.AlbumID = stringPtr(albumID)
	}

	if share.OwnerID == "" || share.SharedWithUserID == "" {
		panic("owner_id and shared_with_user_id are required for CreateTestShare")
	}

	if err := db.Create(share).Error; err != nil {
		panic("failed to create test share: " + err.Error())
	}

	return share
}

// Helper functions

func mustHashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic("failed to hash password: " + err.Error())
	}
	return string(hash)
}

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

// Test model structs (mirror internal/models but with public fields for tests)

type TestUser struct {
	ID            string    `gorm:"primaryKey;type:uuid"`
	Email         string    `gorm:"uniqueIndex;not null"`
	PasswordHash  string    `gorm:"not null"`
	StorageUsed   int64     `gorm:"default:0"`
	StorageQuota  int64     `gorm:"default:524288000"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (TestUser) TableName() string {
	return "users"
}

type TestSession struct {
	ID         string    `gorm:"primaryKey;type:uuid"`
	UserID     string    `gorm:"type:uuid;not null;index"`
	ExpiresAt  time.Time `gorm:"not null;index"`
	LastUsedAt time.Time `gorm:"not null"`
	CreatedAt  time.Time
}

func (TestSession) TableName() string {
	return "sessions"
}

type TestMedia struct {
	ID            string    `gorm:"primaryKey;type:uuid"`
	OwnerID       string    `gorm:"type:uuid;not null;index"`
	Filename      string    `gorm:"not null"`
	FileType      string    `gorm:"not null;index"`
	FileSize      int64     `gorm:"not null"`
	StoragePath   string    `gorm:"not null"`
	ThumbnailPath *string
	Width         *int
	Height        *int
	Duration      *int
	EXIFData      []byte    `gorm:"type:jsonb"`
	UploadedAt    time.Time `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (TestMedia) TableName() string {
	return "media"
}

type TestAlbum struct {
	ID          string    `gorm:"primaryKey;type:uuid"`
	OwnerID     string    `gorm:"type:uuid;not null;index"`
	Name        string    `gorm:"not null;size:100"`
	Description *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (TestAlbum) TableName() string {
	return "albums"
}

type TestShare struct {
	ID               string    `gorm:"primaryKey;type:uuid"`
	MediaID          *string   `gorm:"type:uuid"`
	AlbumID          *string   `gorm:"type:uuid"`
	OwnerID          string    `gorm:"type:uuid;not null"`
	SharedWithUserID string    `gorm:"type:uuid;not null"`
	CreatedAt        time.Time
}

func (TestShare) TableName() string {
	return "shares"
}

