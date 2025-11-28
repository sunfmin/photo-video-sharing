package services

import (
	"context"
	"io"
	"time"
)

// StorageService defines the interface for object storage operations
// Principle X: Service layer interface (can be implemented by MinIO, S3, local filesystem)
type StorageService interface {
	// Upload stores a file and returns the storage path
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	
	// Download retrieves a file
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	
	// Delete removes a file
	Delete(ctx context.Context, key string) error
	
	// GetPresignedURL generates a temporary URL for direct download
	GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	
	// Exists checks if a file exists
	Exists(ctx context.Context, key string) bool
}

