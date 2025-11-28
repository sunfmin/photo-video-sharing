package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/opentracing/opentracing-go"
)

// LocalStorage implements StorageService using local filesystem
// Used for development and testing
type LocalStorage struct {
	basePath string
	tracer   opentracing.Tracer
}

// NewLocalStorage creates a new local filesystem storage backend
func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Ensure base directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("create base directory: %w", err)
	}

	return &LocalStorage{
		basePath: basePath,
		tracer:   opentracing.NoopTracer{},
	}, nil
}

// WithTracer adds OpenTracing support
func (s *LocalStorage) WithTracer(tracer opentracing.Tracer) *LocalStorage {
	s.tracer = tracer
	return s
}

// Upload stores a file on local filesystem
func (s *LocalStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "LocalStorage.Upload")
	defer span.Finish()
	span.SetTag("storage.key", key)
	span.SetTag("storage.size", size)

	// Build full path
	fullPath := filepath.Join(s.basePath, key)

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		span.SetTag("error", true)
		return fmt.Errorf("create parent directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		span.SetTag("error", true)
		return fmt.Errorf("create file: %w", err)
	}
	defer file.Close()

	// Copy data
	if _, err := io.Copy(file, reader); err != nil {
		span.SetTag("error", true)
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// Download retrieves a file from local filesystem
func (s *LocalStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "LocalStorage.Download")
	defer span.Finish()
	span.SetTag("storage.key", key)

	fullPath := filepath.Join(s.basePath, key)
	file, err := os.Open(fullPath)
	if err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("open file: %w", err)
	}

	return file, nil
}

// Delete removes a file from local filesystem
func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "LocalStorage.Delete")
	defer span.Finish()
	span.SetTag("storage.key", key)

	fullPath := filepath.Join(s.basePath, key)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		span.SetTag("error", true)
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

// GetPresignedURL generates a file:// URL for local filesystem
// Note: This is a simplified implementation for development
func (s *LocalStorage) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "LocalStorage.GetPresignedURL")
	defer span.Finish()
	span.SetTag("storage.key", key)

	// For local storage, return a file:// URL
	// In production with MinIO, this would be a real presigned URL
	fullPath := filepath.Join(s.basePath, key)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		span.SetTag("error", true)
		return "", fmt.Errorf("get absolute path: %w", err)
	}

	// Return file URL (for development only)
	return fmt.Sprintf("file://%s", absPath), nil
}

// Exists checks if a file exists on local filesystem
func (s *LocalStorage) Exists(ctx context.Context, key string) bool {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "LocalStorage.Exists")
	defer span.Finish()
	span.SetTag("storage.key", key)

	fullPath := filepath.Join(s.basePath, key)
	_, err := os.Stat(fullPath)
	return err == nil
}

