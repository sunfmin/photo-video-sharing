package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/yourorg/photo-video-sharing/internal/models"
)

// SessionService handles session management
type SessionService struct {
	db *gorm.DB
}

// NewSessionService creates a new SessionService
func NewSessionService(db *gorm.DB) *SessionService {
	return &SessionService{db: db}
}

// CreateSession creates a new session for a user
func (s *SessionService) CreateSession(ctx context.Context, userID string) (string, error) {
	sessionID := uuid.New().String()
	session := models.Session{
		ID:         sessionID,
		UserID:     userID,
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour), // 7 days
		LastUsedAt: time.Now(),
		CreatedAt:  time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(&session).Error; err != nil {
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return sessionID, nil
}

// ValidateSession validates a session and returns the user ID
func (s *SessionService) ValidateSession(ctx context.Context, sessionID string) (string, error) {
	var session models.Session
	if err := s.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("%w", ErrSessionInvalid)
		}
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		return "", fmt.Errorf("%w", ErrSessionExpired)
	}

	// Update last used time
	session.LastUsedAt = time.Now()
	if err := s.db.WithContext(ctx).Save(&session).Error; err != nil {
		// Don't fail validation if update fails, just log
		// In production, would log this error
	}

	return session.UserID, nil
}

// DeleteSession deletes a session (logout)
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	if err := s.db.WithContext(ctx).Where("id = ?", sessionID).Delete(&models.Session{}).Error; err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return nil
}

// CleanupExpiredSessions removes all expired sessions
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	result := s.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.Session{})
	if result.Error != nil {
		return 0, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}
	return result.RowsAffected, nil
}

