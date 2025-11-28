package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/opentracing/opentracing-go"
)

// MinIOStorage implements StorageService using MinIO
type MinIOStorage struct {
	client *minio.Client
	bucket string
	tracer opentracing.Tracer
}

// MinIOConfig holds MinIO configuration
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

// NewMinIOStorage creates a new MinIO storage backend
func NewMinIOStorage(config MinIOConfig) (*MinIOStorage, error) {
	// Initialize MinIO client
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &MinIOStorage{
		client: client,
		bucket: config.Bucket,
		tracer: opentracing.NoopTracer{},
	}, nil
}

// WithTracer adds OpenTracing support
func (s *MinIOStorage) WithTracer(tracer opentracing.Tracer) *MinIOStorage {
	s.tracer = tracer
	return s
}

// Upload stores a file in MinIO
func (s *MinIOStorage) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "MinIOStorage.Upload")
	defer span.Finish()
	span.SetTag("storage.key", key)
	span.SetTag("storage.size", size)

	// Ensure bucket exists
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		span.SetTag("error", true)
		return fmt.Errorf("check bucket existence: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			span.SetTag("error", true)
			return fmt.Errorf("create bucket: %w", err)
		}
	}

	// Upload file
	_, err = s.client.PutObject(ctx, s.bucket, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		span.SetTag("error", true)
		return fmt.Errorf("upload to MinIO: %w", err)
	}

	return nil
}

// Download retrieves a file from MinIO
func (s *MinIOStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "MinIOStorage.Download")
	defer span.Finish()
	span.SetTag("storage.key", key)

	object, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("download from MinIO: %w", err)
	}

	return object, nil
}

// Delete removes a file from MinIO
func (s *MinIOStorage) Delete(ctx context.Context, key string) error {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "MinIOStorage.Delete")
	defer span.Finish()
	span.SetTag("storage.key", key)

	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		span.SetTag("error", true)
		return fmt.Errorf("delete from MinIO: %w", err)
	}

	return nil
}

// GetPresignedURL generates a temporary URL for direct download
func (s *MinIOStorage) GetPresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "MinIOStorage.GetPresignedURL")
	defer span.Finish()
	span.SetTag("storage.key", key)
	span.SetTag("storage.expiry", expiry.String())

	url, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, nil)
	if err != nil {
		span.SetTag("error", true)
		return "", fmt.Errorf("generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// Exists checks if a file exists in MinIO
func (s *MinIOStorage) Exists(ctx context.Context, key string) bool {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "MinIOStorage.Exists")
	defer span.Finish()
	span.SetTag("storage.key", key)

	_, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	return err == nil
}

