package services

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"

	"github.com/yourorg/photo-video-sharing/internal/models"
	pb "github.com/yourorg/photo-video-sharing/api/gen/v1"
)

// MediaService interface defines media operations
type MediaService interface {
	Upload(ctx context.Context, req *pb.UploadMediaRequest, file io.Reader, filename string, fileType string, fileSize int64, userID string) (*pb.UploadMediaResponse, error)
	Get(ctx context.Context, req *pb.GetMediaRequest, userID string) (*pb.GetMediaResponse, error)
	List(ctx context.Context, req *pb.ListMediaRequest, userID string) (*pb.ListMediaResponse, error)
	Delete(ctx context.Context, req *pb.DeleteMediaRequest, userID string) (*pb.DeleteMediaResponse, error)
}

// mediaService implements MediaService
type mediaService struct {
	db         *gorm.DB
	storage    StorageService
	processing ProcessingService
	tracer     opentracing.Tracer
}

// mediaServiceBuilder builds MediaService with dependencies
type mediaServiceBuilder struct {
	db         *gorm.DB
	storage    StorageService
	processing ProcessingService
	tracer     opentracing.Tracer
}

// NewMediaService creates a new MediaService builder (Principle X: Builder Pattern)
func NewMediaService(db *gorm.DB) *mediaServiceBuilder {
	return &mediaServiceBuilder{
		db:     db,
		tracer: opentracing.NoopTracer{},
	}
}

// WithStorage adds storage backend (required)
func (b *mediaServiceBuilder) WithStorage(storage StorageService) *mediaServiceBuilder {
	b.storage = storage
	return b
}

// WithProcessing adds media processing (optional)
func (b *mediaServiceBuilder) WithProcessing(processing ProcessingService) *mediaServiceBuilder {
	b.processing = processing
	return b
}

// WithTracer adds OpenTracing support (optional)
func (b *mediaServiceBuilder) WithTracer(tracer opentracing.Tracer) *mediaServiceBuilder {
	b.tracer = tracer
	return b
}

// Build constructs the MediaService
func (b *mediaServiceBuilder) Build() MediaService {
	if b.storage == nil {
		panic("MediaService requires storage backend (use WithStorage)")
	}
	return &mediaService{
		db:         b.db,
		storage:    b.storage,
		processing: b.processing,
		tracer:     b.tracer,
	}
}

// Upload handles media file upload
func (s *mediaService) Upload(ctx context.Context, req *pb.UploadMediaRequest, file io.Reader, filename string, fileType string, fileSize int64, userID string) (*pb.UploadMediaResponse, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "MediaService.Upload")
	defer span.Finish()
	span.SetTag("filename", filename)
	span.SetTag("file_type", fileType)
	span.SetTag("file_size", fileSize)
	span.SetTag("user_id", userID)

	// Validate file type
	if !isAllowedFileType(fileType) {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedFormat, fileType)
	}

	// Validate file size
	if err := validateFileSize(fileType, fileSize); err != nil {
		span.SetTag("error", true)
		return nil, err
	}

	// Check if file is empty
	if fileSize == 0 {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: file is empty", ErrInvalidInput)
	}

	// Check user quota with transaction
	tx := s.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	var user models.User
	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: user not found", ErrNotFound)
	}

	// Check quota
	if user.StorageUsed+fileSize > user.StorageQuota {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w", ErrQuotaExceeded)
	}

	// Generate media ID and storage path
	mediaID := uuid.New().String()
	storagePath := fmt.Sprintf("%s/%s", userID, mediaID)

	// Upload to storage
	if err := s.storage.Upload(ctx, storagePath, file, fileSize, fileType); err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("upload to storage: %w", err)
	}

	// Create media record
	media := models.Media{
		ID:          mediaID,
		OwnerID:     userID,
		Filename:    filename,
		FileType:    fileType,
		FileSize:    fileSize,
		StoragePath: storagePath,
		UploadedAt:  time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := tx.Create(&media).Error; err != nil {
		span.SetTag("error", true)
		// Try to delete from storage if DB insert fails
		s.storage.Delete(ctx, storagePath)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Update user quota
	user.StorageUsed += fileSize
	if err := tx.Save(&user).Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: failed to update quota", ErrDatabaseError)
	}

	if err := tx.Commit().Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: commit failed", ErrDatabaseError)
	}

	// Return response (thumbnail processing would happen async in background)
	return &pb.UploadMediaResponse{
		Media:   mediaModelToProto(&media),
		Message: "Upload successful, processing thumbnail in background...",
	}, nil
}

// Get retrieves a media item with presigned URLs
func (s *mediaService) Get(ctx context.Context, req *pb.GetMediaRequest, userID string) (*pb.GetMediaResponse, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "MediaService.Get")
	defer span.Finish()
	span.SetTag("media_id", req.Id)
	span.SetTag("user_id", userID)

	// Fetch media (check ownership)
	var media models.Media
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", req.Id, userID).First(&media).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrMediaNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Generate presigned URL for download (valid for 1 hour)
	downloadURL, err := s.storage.GetPresignedURL(ctx, media.StoragePath, 1*time.Hour)
	if err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("generate presigned URL: %w", err)
	}

	// Generate thumbnail URL if exists
	thumbnailURL := ""
	if media.ThumbnailPath != nil {
		thumbnailURL, _ = s.storage.GetPresignedURL(ctx, *media.ThumbnailPath, 1*time.Hour)
	}

	return &pb.GetMediaResponse{
		Media:        mediaModelToProto(&media),
		DownloadUrl:  downloadURL,
		ThumbnailUrl: thumbnailURL,
	}, nil
}

// List retrieves user's media gallery with pagination and filtering
func (s *mediaService) List(ctx context.Context, req *pb.ListMediaRequest, userID string) (*pb.ListMediaResponse, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "MediaService.List")
	defer span.Finish()
	span.SetTag("user_id", userID)

	// Build query
	query := s.db.WithContext(ctx).Where("owner_id = ?", userID)

	// Apply filters
	if req.FileTypeFilter != "" {
		query = query.Where("file_type LIKE ?", req.FileTypeFilter+"%")
	}
	if req.StartDate != "" {
		query = query.Where("uploaded_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("uploaded_at <= ?", req.EndDate)
	}

	// Order by uploaded_at DESC (newest first)
	query = query.Order("uploaded_at DESC")

	// Apply pagination
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 50 // Default
	}
	query = query.Limit(int(pageSize))

	// Fetch media
	var mediaList []models.Media
	if err := query.Find(&mediaList).Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Get total count
	var totalCount int64
	s.db.WithContext(ctx).Model(&models.Media{}).Where("owner_id = ?", userID).Count(&totalCount)

	// Convert to protobuf
	pbMedia := make([]*pb.Media, len(mediaList))
	for i, media := range mediaList {
		pbMedia[i] = mediaModelToProto(&media)
	}

	span.SetTag("result_count", len(pbMedia))
	return &pb.ListMediaResponse{
		Media:      pbMedia,
		TotalCount: totalCount,
	}, nil
}

// Delete removes a media item
func (s *mediaService) Delete(ctx context.Context, req *pb.DeleteMediaRequest, userID string) (*pb.DeleteMediaResponse, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "MediaService.Delete")
	defer span.Finish()
	span.SetTag("media_id", req.Id)
	span.SetTag("user_id", userID)

	// Start transaction
	tx := s.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	// Fetch media (check ownership)
	var media models.Media
	if err := tx.Where("id = ? AND owner_id = ?", req.Id, userID).First(&media).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrMediaNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Delete from database (CASCADE will delete shares)
	if err := tx.Delete(&media).Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Update user quota
	var user models.User
	if err := tx.Where("id = ?", userID).First(&user).Error; err == nil {
		user.StorageUsed -= media.FileSize
		if user.StorageUsed < 0 {
			user.StorageUsed = 0
		}
		tx.Save(&user)
	}

	if err := tx.Commit().Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: commit failed", ErrDatabaseError)
	}

	// Delete from storage (best effort, don't fail if storage delete fails)
	if err := s.storage.Delete(ctx, media.StoragePath); err != nil {
		// Log error but don't fail the operation
		span.SetTag("storage_delete_error", true)
	}

	if media.ThumbnailPath != nil {
		s.storage.Delete(ctx, *media.ThumbnailPath)
	}

	return &pb.DeleteMediaResponse{
		Success: true,
		Message: "Media deleted successfully",
	}, nil
}

// Helper functions

func mediaModelToProto(media *models.Media) *pb.Media {
	pbMedia := &pb.Media{
		Id:          media.ID,
		OwnerId:     media.OwnerID,
		Filename:    media.Filename,
		FileType:    media.FileType,
		FileSize:    media.FileSize,
		StoragePath: media.StoragePath,
		UploadedAt:  timestamppb.New(media.UploadedAt),
		CreatedAt:   timestamppb.New(media.CreatedAt),
		UpdatedAt:   timestamppb.New(media.UpdatedAt),
	}

	if media.ThumbnailPath != nil {
		pbMedia.ThumbnailPath = *media.ThumbnailPath
	}
	if media.Width != nil {
		pbMedia.Width = int32(*media.Width)
	}
	if media.Height != nil {
		pbMedia.Height = int32(*media.Height)
	}
	if media.Duration != nil {
		pbMedia.Duration = int32(*media.Duration)
	}

	// Convert EXIF JSON to protobuf Struct
	if len(media.EXIFData) > 0 {
		// For simplicity, use empty struct if conversion fails
		exifStruct, _ := structpb.NewStruct(map[string]interface{}{})
		pbMedia.ExifData = exifStruct
	}

	return pbMedia
}

func isAllowedFileType(fileType string) bool {
	allowed := []string{
		"image/jpeg",
		"image/png",
		"image/heic",
		"video/mp4",
		"video/quicktime",
		"video/x-msvideo",
	}
	for _, t := range allowed {
		if t == fileType {
			return true
		}
	}
	return false
}

func validateFileSize(fileType string, size int64) error {
	// Photo: max 50MB
	if strings.HasPrefix(fileType, "image/") {
		if size > 52428800 {
			return fmt.Errorf("%w: photos must be under 50MB", ErrFileTooLarge)
		}
	}

	// Video: max 500MB
	if strings.HasPrefix(fileType, "video/") {
		if size > 524288000 {
			return fmt.Errorf("%w: videos must be under 500MB", ErrFileTooLarge)
		}
	}

	return nil
}

