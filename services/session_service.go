package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"gorm.io/gorm"

	"github.com/yourorg/photo-video-sharing/internal/models"
)

// SessionService interface defines session operations
type SessionService interface {
	CreateSession(ctx context.Context, userID string) (string, error)
	ValidateSession(ctx context.Context, sessionID string) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error
	CleanupExpiredSessions(ctx context.Context) (int64, error)
}

// sessionService implements SessionService
type sessionService struct {
	db     *gorm.DB
	tracer opentracing.Tracer
}

// sessionServiceBuilder builds SessionService with optional dependencies
type sessionServiceBuilder struct {
	db     *gorm.DB
	tracer opentracing.Tracer
}

// NewSessionService creates a new SessionService builder (Principle X: Builder Pattern)
func NewSessionService(db *gorm.DB) *sessionServiceBuilder {
	return &sessionServiceBuilder{
		db:     db,
		tracer: opentracing.NoopTracer{}, // Default to noop
	}
}

// WithTracer adds OpenTracing support (optional)
func (b *sessionServiceBuilder) WithTracer(tracer opentracing.Tracer) *sessionServiceBuilder {
	b.tracer = tracer
	return b
}

// Build constructs the SessionService
func (b *sessionServiceBuilder) Build() SessionService {
	return &sessionService{
		db:     b.db,
		tracer: b.tracer,
	}
}

// CreateSession creates a new session for a user
func (s *sessionService) CreateSession(ctx context.Context, userID string) (string, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionService.CreateSession")
	defer span.Finish()
	span.SetTag("user_id", userID)

	sessionID := uuid.New().String()
	session := models.Session{
		ID:         sessionID,
		UserID:     userID,
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour), // 7 days
		LastUsedAt: time.Now(),
		CreatedAt:  time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(&session).Error; err != nil {
		span.SetTag("error", true)
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return sessionID, nil
}

// ValidateSession validates a session and returns the user ID
func (s *sessionService) ValidateSession(ctx context.Context, sessionID string) (string, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionService.ValidateSession")
	defer span.Finish()

	var session models.Session
	if err := s.db.WithContext(ctx).Where("id = ?", sessionID).First(&session).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return "", fmt.Errorf("%w", ErrSessionInvalid)
		}
		return "", fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		span.SetTag("error", true)
		return "", fmt.Errorf("%w", ErrSessionExpired)
	}

	// Update last used time
	session.LastUsedAt = time.Now()
	if err := s.db.WithContext(ctx).Save(&session).Error; err != nil {
		// Don't fail validation if update fails, just log
		// In production, would log this error
	}

	span.SetTag("user_id", session.UserID)
	return session.UserID, nil
}

// DeleteSession deletes a session (logout)
func (s *sessionService) DeleteSession(ctx context.Context, sessionID string) error {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionService.DeleteSession")
	defer span.Finish()

	if err := s.db.WithContext(ctx).Where("id = ?", sessionID).Delete(&models.Session{}).Error; err != nil {
		span.SetTag("error", true)
		return fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}
	return nil
}

// CleanupExpiredSessions removes all expired sessions
func (s *sessionService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "SessionService.CleanupExpiredSessions")
	defer span.Finish()

	result := s.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.Session{})
	if result.Error != nil {
		span.SetTag("error", true)
		return 0, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	span.SetTag("deleted_count", result.RowsAffected)
	return result.RowsAffected, nil
}

