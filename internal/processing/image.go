package processing

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"

	"github.com/disintegration/imaging"
	"github.com/opentracing/opentracing-go"
	"github.com/rwcarlsen/goexif/exif"

	"github.com/yourorg/photo-video-sharing/services"
)

// ImageProcessor implements image processing operations
type ImageProcessor struct {
	tracer opentracing.Tracer
}

// NewImageProcessor creates a new image processor
func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		tracer: opentracing.NoopTracer{},
	}
}

// WithTracer adds OpenTracing support
func (p *ImageProcessor) WithTracer(tracer opentracing.Tracer) *ImageProcessor {
	p.tracer = tracer
	return p
}

// GenerateImageThumbnail creates a thumbnail from an image
func (p *ImageProcessor) GenerateImageThumbnail(ctx context.Context, reader io.Reader, size services.ThumbnailSize) (io.Reader, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "ImageProcessor.GenerateThumbnail")
	defer span.Finish()
	span.SetTag("thumbnail.size", int(size))

	// Read image
	data, err := io.ReadAll(reader)
	if err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("read image data: %w", err)
	}

	// Decode image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("decode image: %w", err)
	}

	// Resize to thumbnail size (maintain aspect ratio)
	thumbnail := imaging.Fit(img, int(size), int(size), imaging.Lanczos)

	// Encode back to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumbnail, &jpeg.Options{Quality: 85}); err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	return bytes.NewReader(buf.Bytes()), nil
}

// ExtractEXIF extracts EXIF metadata from an image
func (p *ImageProcessor) ExtractEXIF(ctx context.Context, reader io.Reader) (map[string]interface{}, error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "ImageProcessor.ExtractEXIF")
	defer span.Finish()

	// Read all data for EXIF extraction
	data, err := io.ReadAll(reader)
	if err != nil {
		span.SetTag("error", true)
		return nil, fmt.Errorf("read image data: %w", err)
	}

	// Decode EXIF
	x, err := exif.Decode(bytes.NewReader(data))
	if err != nil {
		// EXIF not present or invalid - not an error, just return empty map
		return make(map[string]interface{}), nil
	}

	// Extract useful EXIF fields
	result := make(map[string]interface{})

	// DateTime
	if dt, err := x.DateTime(); err == nil {
		result["date_time"] = dt.Format("2006-01-02 15:04:05")
	}

	// Camera make and model
	if make, err := x.Get(exif.Make); err == nil {
		if makeStr, err := make.StringVal(); err == nil {
			result["camera_make"] = makeStr
		}
	}
	if model, err := x.Get(exif.Model); err == nil {
		if modelStr, err := model.StringVal(); err == nil {
			result["camera_model"] = modelStr
		}
	}

	// GPS coordinates
	if lat, long, err := x.LatLong(); err == nil {
		result["gps_latitude"] = lat
		result["gps_longitude"] = long
	}

	// Orientation
	if orient, err := x.Get(exif.Orientation); err == nil {
		if orientInt, err := orient.Int(0); err == nil {
			result["orientation"] = orientInt
		}
	}

	// ISO
	if iso, err := x.Get(exif.ISOSpeedRatings); err == nil {
		if isoInt, err := iso.Int(0); err == nil {
			result["iso"] = isoInt
		}
	}

	// Exposure time
	if exposure, err := x.Get(exif.ExposureTime); err == nil {
		if num, denom, err := exposure.Rat2(0); err == nil {
			result["exposure_time"] = fmt.Sprintf("%d/%d", num, denom)
		}
	}

	// F-number
	if fnumber, err := x.Get(exif.FNumber); err == nil {
		if num, denom, err := fnumber.Rat2(0); err == nil {
			result["f_number"] = float64(num) / float64(denom)
		}
	}

	return result, nil
}

// GetImageDimensions returns width and height of an image
func (p *ImageProcessor) GetImageDimensions(ctx context.Context, reader io.Reader) (width, height int, err error) {
	// Principle XI: Create OpenTracing span
	span, ctx := opentracing.StartSpanFromContext(ctx, "ImageProcessor.GetDimensions")
	defer span.Finish()

	// Decode image config (faster than full decode)
	data, err := io.ReadAll(reader)
	if err != nil {
		span.SetTag("error", true)
		return 0, 0, fmt.Errorf("read image data: %w", err)
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		span.SetTag("error", true)
		return 0, 0, fmt.Errorf("decode image config: %w", err)
	}

	span.SetTag("image.width", config.Width)
	span.SetTag("image.height", config.Height)
	return config.Width, config.Height, nil
}

// GenerateVideoThumbnail is a no-op for ImageProcessor (implemented by VideoProcessor)
func (p *ImageProcessor) GenerateVideoThumbnail(ctx context.Context, videoPath string, atSecond int) (io.Reader, error) {
	return nil, fmt.Errorf("video processing not supported by ImageProcessor")
}

