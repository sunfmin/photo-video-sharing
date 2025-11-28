package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/yourorg/photo-video-sharing/internal/models"
	pb "github.com/yourorg/photo-video-sharing/api/gen/v1"
)

// UserService interface defines user operations
type UserService interface {
	Register(ctx context.Context, req *pb.RegisterRequest) (*pb.User, error)
	Login(ctx context.Context, req *pb.LoginRequest) (*pb.User, error)
	GetCurrentUser(ctx context.Context, userID string) (*pb.User, error)
	PasswordReset(ctx context.Context, req *pb.PasswordResetRequest) (*pb.PasswordResetResponse, error)
}

// userService implements UserService
type userService struct {
	db     *gorm.DB
	tracer opentracing.Tracer
}

// userServiceBuilder builds UserService with optional dependencies
type userServiceBuilder struct {
	db     *gorm.DB
	tracer opentracing.Tracer
}

// NewUserService creates a new UserService builder (Principle X: Builder Pattern)
func NewUserService(db *gorm.DB) *userServiceBuilder {
	return &userServiceBuilder{
		db:     db,
		tracer: opentracing.NoopTracer{}, // Default to noop
	}
}

// WithTracer adds OpenTracing support (optional)
func (b *userServiceBuilder) WithTracer(tracer opentracing.Tracer) *userServiceBuilder {
	b.tracer = tracer
	return b
}

// Build constructs the UserService
func (b *userServiceBuilder) Build() UserService {
	return &userService{
		db:     b.db,
		tracer: b.tracer,
	}
}

// Register creates a new user account
func (s *userService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.User, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserService.Register")
	defer span.Finish()
	span.SetTag("email", req.Email)

	// Validate password complexity (letter + number)
	if !hasLetterAndNumber(req.Password) {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: password must contain at least one letter and one number", ErrInvalidInput)
	}

	// Check if email already exists
	var existingUser models.User
	if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w", ErrDuplicateEmail)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: failed to hash password", ErrInternalError)
	}

	// Create user
	user := models.User{
		ID:            uuid.New().String(),
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		StorageUsed:   0,
		StorageQuota:  524288000, // 500MB default
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Principle XII: Use context-aware database operations
	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return userModelToProto(&user), nil
}

// Login authenticates a user and returns user info (session created by SessionService)
func (s *userService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.User, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserService.Login")
	defer span.Finish()
	span.SetTag("email", req.Email)

	// Find user by email
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&user).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w", ErrInvalidCredentials)
	}

	return userModelToProto(&user), nil
}

// GetCurrentUser retrieves user by ID
func (s *userService) GetCurrentUser(ctx context.Context, userID string) (*pb.User, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserService.GetCurrentUser")
	defer span.Finish()
	span.SetTag("user_id", userID)

	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return userModelToProto(&user), nil
}

// PasswordReset initiates password reset flow (simplified - just returns success)
func (s *userService) PasswordReset(ctx context.Context, req *pb.PasswordResetRequest) (*pb.PasswordResetResponse, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "UserService.PasswordReset")
	defer span.Finish()
	span.SetTag("email", req.Email)

	// Verify user exists
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Don't reveal if user exists for security
			return &pb.PasswordResetResponse{
				EmailSent: false,
				Message:   "If an account exists with this email, a password reset link has been sent.",
			}, nil
		}
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// In real implementation, would generate reset token and send email
	// For now, just return success message
	return &pb.PasswordResetResponse{
		EmailSent: false, // Not actually sending email in this implementation
		Message:   "If an account exists with this email, a password reset link has been sent.",
	}, nil
}

// Helper functions

func userModelToProto(user *models.User) *pb.User {
	return &pb.User{
		Id:           user.ID,
		Email:        user.Email,
		StorageUsed:  user.StorageUsed,
		StorageQuota: user.StorageQuota,
		CreatedAt:    timestamppb.New(user.CreatedAt),
		UpdatedAt:    timestamppb.New(user.UpdatedAt),
	}
}

func hasLetterAndNumber(s string) bool {
	hasLetter := false
	hasNumber := false
	for _, ch := range s {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
			hasLetter = true
		}
		if ch >= '0' && ch <= '9' {
			hasNumber = true
		}
	}
	return hasLetter && hasNumber
}

