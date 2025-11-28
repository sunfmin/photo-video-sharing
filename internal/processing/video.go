package processing

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/opentracing/opentracing-go"

	"github.com/yourorg/photo-video-sharing/services"
)

// VideoProcessor implements video processing operations
type VideoProcessor struct {
	tracer opentracing.Tracer
}

// NewVideoProcessor creates a new video processor
func NewVideoProcessor() *VideoProcessor {
	return &VideoProcessor{
		tracer: opentracing.NoopTracer{},
	}
}

// WithTracer adds OpenTracing support
func (p *VideoProcessor) WithTracer(tracer opentracing.Tracer) *VideoProcessor {
	p.tracer = tracer
	return p
}

// GenerateVideoThumbnail extracts a frame from video at specified second
func (p *VideoProcessor) GenerateVideoThumbnail(ctx context.Context, videoPath string, atSecond int) (io.Reader, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "VideoProcessor.GenerateThumbnail")
	defer span.Finish()
	span.SetTag("video.path", videoPath)
	span.SetTag("video.at_second", atSecond)

	// Create temporary output file for thumbnail
	tmpDir := os.TempDir()
	outputPath := filepath.Join(tmpDir, fmt.Sprintf("thumbnail_%d.jpg", os.Getpid()))
	defer os.Remove(outputPath) // Clean up

	// Extract frame using FFmpeg
	// Seek to specified second and extract one frame
	err := ffmpeg.Input(videoPath, ffmpeg.KwArgs{
		"ss": fmt.Sprintf("%d", atSecond), // Seek to second
	}).
		Output(outputPath, ffmpeg.KwArgs{
			"vframes": 1,        // Extract 1 frame
			"q:v":     2,        // High quality
			"vf":      "scale=800:600:force_original_aspect_ratio=decrease", // Resize
		}).
		OverWriteOutput(). // Overwrite if exists
		Silent(true).      // Suppress FFmpeg output
		Run()

	if err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("ffmpeg extract frame: %w", err)
	}

	// Read the generated thumbnail
	data, err := os.ReadFile(outputPath)
	if err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("read thumbnail file: %w", err)
	}

	return bytes.NewReader(data), nil
}

// GenerateImageThumbnail is a no-op for VideoProcessor
func (p *VideoProcessor) GenerateImageThumbnail(ctx context.Context, reader io.Reader, size services.ThumbnailSize) (io.Reader, error) {
	return nil, fmt.Errorf("image processing not supported by VideoProcessor")
}

// ExtractEXIF is a no-op for VideoProcessor
func (p *VideoProcessor) ExtractEXIF(ctx context.Context, reader io.Reader) (map[string]interface{}, error) {
	return nil, fmt.Errorf("EXIF extraction not supported for videos")
}

// GetImageDimensions is a no-op for VideoProcessor
func (p *VideoProcessor) GetImageDimensions(ctx context.Context, reader io.Reader) (width, height int, err error) {
	return 0, 0, fmt.Errorf("dimension extraction not supported by VideoProcessor (use video metadata)")
}

