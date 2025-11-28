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

// AlbumService interface defines album operations
type AlbumService interface {
	Create(ctx context.Context, req *pb.CreateAlbumRequest, userID string) (*pb.CreateAlbumResponse, error)
	Get(ctx context.Context, req *pb.GetAlbumRequest, userID string) (*pb.GetAlbumResponse, error)
	List(ctx context.Context, req *pb.ListAlbumsRequest, userID string) (*pb.ListAlbumsResponse, error)
	Update(ctx context.Context, req *pb.UpdateAlbumRequest, userID string) (*pb.UpdateAlbumResponse, error)
	Delete(ctx context.Context, req *pb.DeleteAlbumRequest, userID string) (*pb.DeleteAlbumResponse, error)
	AddMedia(ctx context.Context, req *pb.AddMediaToAlbumRequest, userID string) (*pb.AddMediaToAlbumResponse, error)
	RemoveMedia(ctx context.Context, req *pb.RemoveMediaFromAlbumRequest, userID string) (*pb.RemoveMediaFromAlbumResponse, error)
	GetAlbumMedia(ctx context.Context, req *pb.GetAlbumMediaRequest, userID string) (*pb.GetAlbumMediaResponse, error)
}

// albumService implements AlbumService
type albumService struct {
	db     *gorm.DB
	tracer opentracing.Tracer
}

// albumServiceBuilder builds AlbumService
type albumServiceBuilder struct {
	db     *gorm.DB
	tracer opentracing.Tracer
}

// NewAlbumService creates a new AlbumService builder (Principle X: Builder Pattern)
func NewAlbumService(db *gorm.DB) *albumServiceBuilder {
	return &albumServiceBuilder{
		db:     db,
		tracer: opentracing.NoopTracer{},
	}
}

// WithTracer adds OpenTracing support
func (b *albumServiceBuilder) WithTracer(tracer opentracing.Tracer) *albumServiceBuilder {
	b.tracer = tracer
	return b
}

// Build constructs the AlbumService
func (b *albumServiceBuilder) Build() AlbumService {
	return &albumService{
		db:     b.db,
		tracer: b.tracer,
	}
}

// Create creates a new album
func (s *albumService) Create(ctx context.Context, req *pb.CreateAlbumRequest, userID string) (*pb.CreateAlbumResponse, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "AlbumService.Create")
	defer span.Finish()
	span.SetTag("user_id", userID)
	span.SetTag("album_name", req.Name)

	// Validate input
	if req.Name == "" {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: album name is required", ErrInvalidInput)
	}
	if len(req.Name) > 100 {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: album name too long", ErrInvalidAlbumName)
	}

	// Create album
	album := models.Album{
		ID:        uuid.New().String(),
		OwnerID:   userID,
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if req.Description != "" {
		album.Description = &req.Description
	}

	if err := s.db.WithContext(ctx).Create(&album).Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return &pb.CreateAlbumResponse{
		Album: albumModelToProto(&album),
	}, nil
}

// Get retrieves an album by ID
func (s *albumService) Get(ctx context.Context, req *pb.GetAlbumRequest, userID string) (*pb.GetAlbumResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AlbumService.Get")
	defer span.Finish()
	span.SetTag("album_id", req.Id)
	span.SetTag("user_id", userID)

	var album models.Album
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", req.Id, userID).First(&album).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrAlbumNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return &pb.GetAlbumResponse{
		Album: albumModelToProto(&album),
	}, nil
}

// List retrieves user's albums
func (s *albumService) List(ctx context.Context, req *pb.ListAlbumsRequest, userID string) (*pb.ListAlbumsResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AlbumService.List")
	defer span.Finish()
	span.SetTag("user_id", userID)

	query := s.db.WithContext(ctx).Where("owner_id = ?", userID).Order("created_at DESC")

	// Pagination
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 50
	}
	query = query.Limit(int(pageSize))

	var albums []models.Album
	if err := query.Find(&albums).Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Get total count
	var totalCount int64
	s.db.WithContext(ctx).Model(&models.Album{}).Where("owner_id = ?", userID).Count(&totalCount)

	// Convert to protobuf
	pbAlbums := make([]*pb.Album, len(albums))
	for i, album := range albums {
		pbAlbums[i] = albumModelToProto(&album)
	}

	span.SetTag("result_count", len(pbAlbums))
	return &pb.ListAlbumsResponse{
		Albums:     pbAlbums,
		TotalCount: totalCount,
	}, nil
}

// Update updates an album
func (s *albumService) Update(ctx context.Context, req *pb.UpdateAlbumRequest, userID string) (*pb.UpdateAlbumResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AlbumService.Update")
	defer span.Finish()
	span.SetTag("album_id", req.Id)

	var album models.Album
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", req.Id, userID).First(&album).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrAlbumNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Update fields
	if req.Name != "" {
		album.Name = req.Name
	}
	if req.Description != "" {
		album.Description = &req.Description
	}
	album.UpdatedAt = time.Now()

	if err := s.db.WithContext(ctx).Save(&album).Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	return &pb.UpdateAlbumResponse{
		Album: albumModelToProto(&album),
	}, nil
}

// Delete deletes an album
func (s *albumService) Delete(ctx context.Context, req *pb.DeleteAlbumRequest, userID string) (*pb.DeleteAlbumResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AlbumService.Delete")
	defer span.Finish()
	span.SetTag("album_id", req.Id)
	span.SetTag("delete_media", req.DeleteMedia)

	tx := s.db.WithContext(ctx).Begin()
	defer tx.Rollback()

	var album models.Album
	if err := tx.Where("id = ? AND owner_id = ?", req.Id, userID).First(&album).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrAlbumNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// If delete_media is true, delete all media in album
	if req.DeleteMedia {
		// Get media IDs in album
		var mediaIDs []string
		tx.Model(&models.AlbumMedia{}).Where("album_id = ?", req.Id).Pluck("media_id", &mediaIDs)
		
		// Delete media (would trigger cascade delete of album_media)
		if len(mediaIDs) > 0 {
			tx.Where("id IN ?", mediaIDs).Delete(&models.Media{})
		}
	}

	// Delete album (CASCADE will delete album_media entries)
	if err := tx.Delete(&album).Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	if err := tx.Commit().Error; err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: commit failed", ErrDatabaseError)
	}

	return &pb.DeleteAlbumResponse{
		Success: true,
		Message: "Album deleted successfully",
	}, nil
}

// AddMedia adds media items to an album
func (s *albumService) AddMedia(ctx context.Context, req *pb.AddMediaToAlbumRequest, userID string) (*pb.AddMediaToAlbumResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AlbumService.AddMedia")
	defer span.Finish()
	span.SetTag("album_id", req.AlbumId)
	span.SetTag("media_count", len(req.MediaIds))

	// Verify album ownership
	var album models.Album
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", req.AlbumId, userID).First(&album).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrAlbumNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	addedCount := int32(0)
	var errors []string

	for _, mediaID := range req.MediaIds {
		// Verify media ownership
		var media models.Media
		if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", mediaID, userID).First(&media).Error; err != nil {
			errors = append(errors, fmt.Sprintf("Media %s not found", mediaID))
			continue
		}

		// Add to album (ignore if already exists)
		albumMedia := models.AlbumMedia{
			AlbumID: req.AlbumId,
			MediaID: mediaID,
			AddedAt: time.Now(),
		}

		if err := s.db.WithContext(ctx).Create(&albumMedia).Error; err != nil {
			// Check if already exists (duplicate key error)
			if isDuplicateError(err) {
				errors = append(errors, fmt.Sprintf("Media %s already in album", mediaID))
			} else {
				errors = append(errors, fmt.Sprintf("Failed to add media %s", mediaID))
			}
			continue
		}

		addedCount++
	}

	return &pb.AddMediaToAlbumResponse{
		AddedCount: addedCount,
		Errors:     errors,
	}, nil
}

// RemoveMedia removes media items from an album
func (s *albumService) RemoveMedia(ctx context.Context, req *pb.RemoveMediaFromAlbumRequest, userID string) (*pb.RemoveMediaFromAlbumResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AlbumService.RemoveMedia")
	defer span.Finish()
	span.SetTag("album_id", req.AlbumId)

	// Verify album ownership
	var album models.Album
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", req.AlbumId, userID).First(&album).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrAlbumNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Remove media from album
	result := s.db.WithContext(ctx).
		Where("album_id = ? AND media_id IN ?", req.AlbumId, req.MediaIds).
		Delete(&models.AlbumMedia{})

	if result.Error != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, result.Error)
	}

	return &pb.RemoveMediaFromAlbumResponse{
		RemovedCount: int32(result.RowsAffected),
	}, nil
}

// GetAlbumMedia retrieves media items in an album
func (s *albumService) GetAlbumMedia(ctx context.Context, req *pb.GetAlbumMediaRequest, userID string) (*pb.GetAlbumMediaResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "AlbumService.GetAlbumMedia")
	defer span.Finish()
	span.SetTag("album_id", req.AlbumId)

	// Verify album ownership
	var album models.Album
	if err := s.db.WithContext(ctx).Where("id = ? AND owner_id = ?", req.AlbumId, userID).First(&album).Error; err != nil {
		span.SetTag("error", true)
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("%w", ErrAlbumNotFound)
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseError, err)
	}

	// Get media in album
	var mediaList []models.Media
	query := s.db.WithContext(ctx).
		Joins("INNER JOIN album_media ON media.id = album_media.media_id").
		Where("album_media.album_id = ?", req.AlbumId).
		Order("album_media.added_at DESC")

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
		Model(&models.AlbumMedia{}).
		Where("album_id = ?", req.AlbumId).
		Count(&totalCount)

	// Convert to protobuf
	pbMedia := make([]*pb.Media, len(mediaList))
	for i, media := range mediaList {
		pbMedia[i] = mediaModelToProto(&media)
	}

	span.SetTag("result_count", len(pbMedia))
	return &pb.GetAlbumMediaResponse{
		Media:      pbMedia,
		TotalCount: totalCount,
	}, nil
}

// Helper functions

func albumModelToProto(album *models.Album) *pb.Album {
	pbAlbum := &pb.Album{
		Id:        album.ID,
		OwnerId:   album.OwnerID,
		Name:      album.Name,
		CreatedAt: timestamppb.New(album.CreatedAt),
		UpdatedAt: timestamppb.New(album.UpdatedAt),
	}

	if album.Description != nil {
		pbAlbum.Description = *album.Description
	}

	// TODO: Add cover_thumbnail_url and media_count

	return pbAlbum
}

func isDuplicateError(err error) bool {
	// Check for PostgreSQL duplicate key error
	return err != nil && (
		contains(err.Error(), "duplicate key") ||
		contains(err.Error(), "unique constraint") ||
		contains(err.Error(), "UNIQUE constraint"))
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && stringContains(s, substr)
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

