package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/opentracing/opentracing-go"

	"github.com/yourorg/photo-video-sharing/services"
	pb "github.com/yourorg/photo-video-sharing/api/gen/v1"
)

// ShareHandler handles share-related HTTP requests
type ShareHandler struct {
	shareService services.ShareService
	tracer       opentracing.Tracer
}

// NewShareHandler creates a new ShareHandler
func NewShareHandler(shareService services.ShareService) *ShareHandler {
	return &ShareHandler{
		shareService: shareService,
		tracer:       opentracing.NoopTracer{},
	}
}

// WithTracer adds OpenTracing support
func (h *ShareHandler) WithTracer(tracer opentracing.Tracer) *ShareHandler {
	h.tracer = tracer
	return h
}

// ShareMedia handles sharing media with users
func (h *ShareHandler) ShareMedia(w http.ResponseWriter, r *http.Request) {
	span := h.tracer.StartSpan("POST /shares/media")
	defer span.Finish()
	span.SetTag("http.method", r.Method)

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var reqBody struct {
		MediaID         string   `json:"media_id"`
		RecipientEmails []string `json:"recipient_emails"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		span.SetTag("error", true)
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req := &pb.ShareMediaRequest{
		MediaId:         reqBody.MediaID,
		RecipientEmails: reqBody.RecipientEmails,
	}

	resp, err := h.shareService.ShareMedia(ctx, req, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusCreated)
	WriteJSON(w, http.StatusCreated, resp)
}

// ShareAlbum handles sharing albums with users
func (h *ShareHandler) ShareAlbum(w http.ResponseWriter, r *http.Request) {
	span := h.tracer.StartSpan("POST /shares/album")
	defer span.Finish()

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var reqBody struct {
		AlbumID         string   `json:"album_id"`
		RecipientEmails []string `json:"recipient_emails"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		span.SetTag("error", true)
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req := &pb.ShareAlbumRequest{
		AlbumId:         reqBody.AlbumID,
		RecipientEmails: reqBody.RecipientEmails,
	}

	resp, err := h.shareService.ShareAlbum(ctx, req, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusCreated)
	WriteJSON(w, http.StatusCreated, resp)
}

// RevokeShare handles revoking a share
func (h *ShareHandler) RevokeShare(w http.ResponseWriter, r *http.Request) {
	span := h.tracer.StartSpan("DELETE /shares/{id}")
	defer span.Finish()

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	shareID := r.PathValue("id")
	if shareID == "" {
		span.SetTag("error", true)
		WriteError(w, http.StatusBadRequest, "Missing share ID")
		return
	}

	resp, err := h.shareService.RevokeShare(ctx, &pb.RevokeShareRequest{ShareId: shareID}, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	WriteJSON(w, http.StatusOK, resp)
}

// ListMediaShares lists who a media is shared with
func (h *ShareHandler) ListMediaShares(w http.ResponseWriter, r *http.Request) {
	span := h.tracer.StartSpan("GET /shares/media/{id}")
	defer span.Finish()

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	mediaID := r.PathValue("id")
	resp, err := h.shareService.ListMediaShares(ctx, &pb.ListMediaSharesRequest{MediaId: mediaID}, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	WriteJSON(w, http.StatusOK, resp)
}

// ListAlbumShares lists who an album is shared with (stub)
func (h *ShareHandler) ListAlbumShares(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Album share listing - TODO")
}

// ListSharedAlbums lists albums shared with current user (stub)
func (h *ShareHandler) ListSharedAlbums(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Shared album listing - TODO")
}

