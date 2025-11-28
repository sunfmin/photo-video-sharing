package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/yourorg/photo-video-sharing/internal/models"
	pb "github.com/yourorg/photo-video-sharing/api/gen/v1"
)

// UserService handles user-related business logic
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new UserService
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// Register creates a new user account
func (s *UserService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.User, error) {
	// Validate password complexity (letter + number)
	if !hasLetterAndNumber(req.Password) {
		return nil, fmt.Errorf("%w: password must contain at least one letter and one number", ErrInvalidInput)
	}

	// Check if email already exists
	var existingUser models.User
	if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, fmt.Errorf("%w", ErrDuplicateEmail)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
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

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return userModelToProto(&user), nil
}

// Login authenticates a user and returns user info (session created by SessionService)
func (s *UserService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.User, error) {
	// Find user by email
	var user models.User
	if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrInvalidCredentials)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("%w", ErrInvalidCredentials)
	}

	return userModelToProto(&user), nil
}

// GetCurrentUser retrieves user by ID
func (s *UserService) GetCurrentUser(ctx context.Context, userID string) (*pb.User, error) {
	var user models.User
	if err := s.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return userModelToProto(&user), nil
}

// PasswordReset initiates password reset flow (simplified - just returns success)
func (s *UserService) PasswordReset(ctx context.Context, req *pb.PasswordResetRequest) (*pb.PasswordResetResponse, error) {
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

