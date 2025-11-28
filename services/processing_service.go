package services

import (
	"context"
	"io"
)

// ProcessingService defines the interface for media processing operations
// Principle X: Service layer interface (image/video processing)
type ProcessingService interface {
	// GenerateImageThumbnail creates a thumbnail from an image
	GenerateImageThumbnail(ctx context.Context, reader io.Reader, size ThumbnailSize) (io.Reader, error)
	
	// ExtractEXIF extracts EXIF metadata from an image
	ExtractEXIF(ctx context.Context, reader io.Reader) (map[string]interface{}, error)
	
	// GenerateVideoThumbnail extracts a frame from a video as thumbnail
	GenerateVideoThumbnail(ctx context.Context, videoPath string, atSecond int) (io.Reader, error)
	
	// GetImageDimensions returns width and height of an image
	GetImageDimensions(ctx context.Context, reader io.Reader) (width, height int, err error)
}

// ThumbnailSize defines standard thumbnail sizes
type ThumbnailSize int

const (
	ThumbnailSmall  ThumbnailSize = 150  // 150x150 for list view
	ThumbnailMedium ThumbnailSize = 800  // 800x600 for preview
	ThumbnailLarge  ThumbnailSize = 1920 // 1920x1080 for lightbox
)

