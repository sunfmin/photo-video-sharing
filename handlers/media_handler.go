package handlers

import (
	"net/http"

	"github.com/opentracing/opentracing-go"

	"github.com/yourorg/photo-video-sharing/services"
	pb "github.com/yourorg/photo-video-sharing/api/gen/v1"
)

// MediaHandler handles media-related HTTP requests
type MediaHandler struct {
	mediaService services.MediaService
	tracer       opentracing.Tracer
}

// NewMediaHandler creates a new MediaHandler
// mediaService can be nil for routes that don't need media functionality
func NewMediaHandler(mediaService services.MediaService) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
		tracer:       opentracing.NoopTracer{},
	}
}

// WithTracer adds OpenTracing support
func (h *MediaHandler) WithTracer(tracer opentracing.Tracer) *MediaHandler {
	h.tracer = tracer
	return h
}

// Upload handles media file upload (multipart/form-data)
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Principle XI: Create OpenTracing span for HTTP endpoint
	span := h.tracer.StartSpan("POST /media")
	defer span.Finish()
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)

	ctx := r.Context()

	// Get user ID from context (set by auth middleware)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusUnauthorized)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse multipart form (32MB max memory)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusBadRequest)
		WriteError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusBadRequest)
		WriteError(w, http.StatusBadRequest, "No file uploaded")
		return
	}
	defer file.Close()

	// Get file metadata
	filename := header.Filename
	contentType := header.Header.Get("Content-Type")
	fileSize := header.Size

	// Detect MIME type from filename if generic type
	if contentType == "application/octet-stream" || contentType == "" {
		contentType = detectMIMEFromFilename(filename)
	}

	span.SetTag("filename", filename)
	span.SetTag("content_type", contentType)
	span.SetTag("file_size", fileSize)

	// Upload via service
	resp, err := h.mediaService.Upload(ctx, &pb.UploadMediaRequest{}, file, filename, contentType, fileSize, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusCreated)
	WriteJSON(w, http.StatusCreated, resp)
}

// List handles media gallery listing
func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	// Principle XI: Create OpenTracing span
	span := h.tracer.StartSpan("GET /media")
	defer span.Finish()
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)

	ctx := r.Context()

	// Get user ID from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusUnauthorized)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Parse query parameters
	req := &pb.ListMediaRequest{}
	// TODO: Parse pagination and filter params from query string

	resp, err := h.mediaService.List(ctx, req, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	span.SetTag("result_count", len(resp.Media))
	WriteJSON(w, http.StatusOK, resp)
}

// Get handles retrieving a single media item with presigned URL
func (h *MediaHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Principle XI: Create OpenTracing span
	span := h.tracer.StartSpan("GET /media/{id}")
	defer span.Finish()
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)

	ctx := r.Context()

	// Get user ID from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusUnauthorized)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract media ID from path
	mediaID := r.PathValue("id")
	if mediaID == "" {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusBadRequest)
		WriteError(w, http.StatusBadRequest, "Missing media ID")
		return
	}

	span.SetTag("media_id", mediaID)

	resp, err := h.mediaService.Get(ctx, &pb.GetMediaRequest{Id: mediaID}, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	WriteJSON(w, http.StatusOK, resp)
}

// Delete handles media deletion
func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// Principle XI: Create OpenTracing span
	span := h.tracer.StartSpan("DELETE /media/{id}")
	defer span.Finish()
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)

	ctx := r.Context()

	// Get user ID from context
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusUnauthorized)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	// Extract media ID from path
	mediaID := r.PathValue("id")
	if mediaID == "" {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusBadRequest)
		WriteError(w, http.StatusBadRequest, "Missing media ID")
		return
	}

	span.SetTag("media_id", mediaID)

	resp, err := h.mediaService.Delete(ctx, &pb.DeleteMediaRequest{Id: mediaID}, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	WriteJSON(w, http.StatusOK, resp)
}

// ListSharedWithMe lists media shared with current user
func (h *MediaHandler) ListSharedWithMe(w http.ResponseWriter, r *http.Request, shareService services.ShareService) {
	span := h.tracer.StartSpan("GET /media/shared")
	defer span.Finish()

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	resp, err := shareService.ListSharedWithMe(ctx, &pb.ListSharedWithMeRequest{}, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	WriteJSON(w, http.StatusOK, resp)
}

// Helper function to detect MIME type from filename extension
func detectMIMEFromFilename(filename string) string {
	ext := filename[len(filename)-4:]
	if len(filename) > 4 {
		ext = filename[len(filename)-4:]
	} else {
		return "application/octet-stream"
	}

	switch ext {
	case ".jpg", ".jpeg", "jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case "heic":
		return "image/heic"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	default:
		return "application/octet-stream"
	}
}

