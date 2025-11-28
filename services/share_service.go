package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/yourorg/photo-video-sharing/internal/models"
	pb "github.com/yourorg/photo-video-sharing/api/gen/v1"
)

// ShareService interface defines sharing operations
type ShareService interface {
	ShareMedia(ctx context.Context, req *pb.ShareMediaRequest, userID string) (*pb.ShareMediaResponse, error)
	ShareAlbum(ctx context.Context, req *pb.ShareAlbumRequest, userID string) (*pb.ShareAlbumResponse, error)
	RevokeShare(ctx context.Context, req *pb.RevokeShareRequest, userID string) (*pb.RevokeShareResponse, error)
	ListMediaShares(ctx context.Context, req *pb.ListMediaSharesRequest, userID string) (*pb.ListMediaSharesResponse, error)
	ListSharedWithMe(ctx context.Context, req *pb.ListSharedWithMeRequest, userID string) (*pb.ListSharedWithMeResponse, error)
}

// shareService implements ShareService
type shareService struct {
	db     *gorm.DB
	tracer opentracing.Tracer
}

// shareServiceBuilder builds ShareService
type shareServiceBuilder struct {
	db     *gorm.DB
	tracer opentracing.Tracer
}

// NewShareService creates a new ShareService builder (Principle X: Builder Pattern)
func NewShareService(db *gorm.DB) *shareServiceBuilder {
	return &shareServiceBuilder{
		db:     db,
		tracer: opentracing.NoopTracer{},
	}
}

// WithTracer adds OpenTracing support
func (b *shareServiceBuilder) WithTracer(tracer opentracing.Tracer) *shareServiceBuilder {
	b.tracer = tracer
	return b
}

// Build constructs the ShareService
func (b *shareServiceBuilder) Build() ShareService {
	return &shareService{
		db:     b.db,
		tracer: b.tracer,
	}
}

// ShareMedia shares media with users by email
func (s *shareService) ShareMedia(ctx context.Context, req *pb.ShareMediaRequest, userID string) (*pb.ShareMediaResponse, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "ShareService.ShareMedia")
	defer span.Finish()
	span.SetTag("media_id", req.MediaId)
	span.SetTag("recipient_count", len(req.RecipientEmails))

	// Verify user owns the media
	var media models.Media
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", req.MediaId, userID).First(&media).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrMediaNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	var shares []*pb.Share
	var errors []string
	successCount := int32(0)

	for _, email := range req.RecipientEmails {
		// Look up recipient by email
		var recipient models.User
		if err := s.db.WithContext(ctx).Where("email = ?", email).First(&recipient).Error; err != nil {
			errors = append(errors, fmt.Sprintf("User %s not found", email))
			continue
		}

		// Check if trying to share with self
		if recipient.ID == userID {
			errors = append(errors, fmt.Sprintf("Cannot share with yourself (%s)", email))
			continue
		}

		// Check if already shared
		var existingShare models.Share
		if err := s.db.WithContext(ctx).
			Where("media_id = ? AND shared_with_user_id = ?", req.MediaId, recipient.ID).
			First(&existingShare).Error; err == nil {
			errors = append(errors, fmt.Sprintf("Already shared with %s", email))
			continue
		}

		// Create share
		share := models.Share{
			ID:               uuid.New().String(),
			MediaID:          &req.MediaId,
			OwnerID:          userID,
			SharedWithUserID: recipient.ID,
			CreatedAt:        time.Now(),
		}

		if err := s.db.WithContext(ctx).Create(&share).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Failed to share with %s", email))
			continue
		}

		// Add to response
		shares = append(shares, &pb.Share{
			Id:               share.ID,
			MediaId:          req.MediaId,
			OwnerId:          userID,
			SharedWithUserId: recipient.ID,
			SharedWithEmail:  email,
			CreatedAt:        timestamppb.New(share.CreatedAt),
		})
		successCount++
	}

	span.SetTag("success_count", successCount)
	return &pb.ShareMediaResponse{
		Shares:       shares,
		Errors:       errors,
		SuccessCount: successCount,
	}, nil
}

// ShareAlbum shares an album with users by email
func (s *shareService) ShareAlbum(ctx context.Context, req *pb.ShareAlbumRequest, userID string) (*pb.ShareAlbumResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ShareService.ShareAlbum")
	defer span.Finish()
	span.SetTag("album_id", req.AlbumId)

	// Verify user owns the album
	var album models.Album
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", req.AlbumId, userID).First(&album).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrAlbumNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	var shares []*pb.Share
	var errors []string
	successCount := int32(0)

	for _, email := range req.RecipientEmails {
		var recipient models.User
		if err := s.db.WithContext(ctx).Where("email = ?", email).First(&recipient).Error; err != nil {
			errors = append(errors, fmt.Sprintf("User %s not found", email))
			continue
		}

		if recipient.ID == userID {
			errors = append(errors, fmt.Sprintf("Cannot share with yourself (%s)", email))
			continue
		}

		// Check if already shared
		var existingShare models.Share
		if err := s.db.WithContext(ctx).
			Where("album_id = ? AND shared_with_user_id = ?", req.AlbumId, recipient.ID).
			First(&existingShare).Error; err == nil {
			errors = append(errors, fmt.Sprintf("Already shared with %s", email))
			continue
		}

		// Create share
		share := models.Share{
			ID:               uuid.New().String(),
			AlbumID:          &req.AlbumId,
			OwnerID:          userID,
			SharedWithUserID: recipient.ID,
			CreatedAt:        time.Now(),
		}

		if err := s.db.WithContext(ctx).Create(&share).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Failed to share with %s", email))
			continue
		}

		shares = append(shares, &pb.Share{
			Id:               share.ID,
			AlbumId:          req.AlbumId,
			OwnerId:          userID,
			SharedWithUserId: recipient.ID,
			SharedWithEmail:  email,
			CreatedAt:        timestamppb.New(share.CreatedAt),
		})
		successCount++
	}

	return &pb.ShareAlbumResponse{
		Shares:       shares,
		Errors:       errors,
		SuccessCount: successCount,
	}, nil
}

// RevokeShare removes sharing access
func (s *shareService) RevokeShare(ctx context.Context, req *pb.RevokeShareRequest, userID string) (*pb.RevokeShareResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ShareService.RevokeShare")
	defer span.Finish()
	span.SetTag("share_id", req.ShareId)

	// Verify user owns the share
	var share models.Share
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", req.ShareId, userID).First(&share).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrShareNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Delete share
	if err := s.db.WithContext(ctx).Delete(&share).Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return &pb.RevokeShareResponse{
		Success: true,
		Message: "Share revoked successfully",
	}, nil
}

// ListMediaShares lists who a media is shared with
func (s *shareService) ListMediaShares(ctx context.Context, req *pb.ListMediaSharesRequest, userID string) (*pb.ListMediaSharesResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ShareService.ListMediaShares")
	defer span.Finish()
	span.SetTag("media_id", req.MediaId)

	// Verify ownership
	var media models.Media
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", req.MediaId, userID).First(&media).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrMediaNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Get shares
	var shares []models.Share
	if err := s.db.WithContext(ctx).Where("media_id = ?", req.MediaId).Find(&shares).Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Convert to protobuf
	pbShares := make([]*pb.Share, len(shares))
	for i, share := range shares {
		// Get recipient email
		var recipient models.User
		s.db.WithContext(ctx).Where("id = ?", share.SharedWithUserID).First(&recipient)

		pbShares[i] = &pb.Share{
			Id:               share.ID,
			MediaId:          *share.MediaID,
			OwnerId:          share.OwnerID,
			SharedWithUserId: share.SharedWithUserID,
			SharedWithEmail:  recipient.Email,
			CreatedAt:        timestamppb.New(share.CreatedAt),
		}
	}

	return &pb.ListMediaSharesResponse{
		Shares: pbShares,
	}, nil
}

// ListSharedWithMe lists media shared with current user
func (s *shareService) ListSharedWithMe(ctx context.Context, req *pb.ListSharedWithMeRequest, userID string) (*pb.ListSharedWithMeResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "ShareService.ListSharedWithMe")
	defer span.Finish()
	span.SetTag("user_id", userID)

	// Get all media shared with this user
	var mediaList []models.Media
	query := s.db.WithContext(ctx).
		Joins("INNER JOIN shares ON shares.media_id = media.id").
		Where("shares.shared_with_user_id = ?", userID).
		Order("shares.created_at DESC")

	// Pagination
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 50
	}
	query = query.Limit(int(pageSize))

	if err := query.Find(&mediaList).Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Get total count
	var totalCount int64
	s.db.WithContext(ctx).
		Model(&models.Share{}).
		Where("shared_with_user_id = ? AND media_id IS NOT NULL", userID).
		Count(&totalCount)

	// Convert to protobuf
	pbMedia := make([]*pb.Media, len(mediaList))
	for i, media := range mediaList {
		pbMedia[i] = mediaModelToProto(&media)
	}

	span.SetTag("result_count", len(pbMedia))
	return &pb.ListSharedWithMeResponse{
		Media:      pbMedia,
		TotalCount: totalCount,
	}, nil
}

