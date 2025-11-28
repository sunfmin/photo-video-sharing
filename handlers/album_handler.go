package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/opentracing/opentracing-go"

	"github.com/yourorg/photo-video-sharing/services"
	pb "github.com/yourorg/photo-video-sharing/api/gen/v1"
)

// AlbumHandler handles album-related HTTP requests
type AlbumHandler struct {
	albumService services.AlbumService
	tracer       opentracing.Tracer
}

// NewAlbumHandler creates a new AlbumHandler
func NewAlbumHandler(albumService services.AlbumService) *AlbumHandler {
	return &AlbumHandler{
		albumService: albumService,
		tracer:       opentracing.NoopTracer{},
	}
}

// WithTracer adds OpenTracing support
func (h *AlbumHandler) WithTracer(tracer opentracing.Tracer) *AlbumHandler {
	h.tracer = tracer
	return h
}

// Create handles album creation
func (h *AlbumHandler) Create(w http.ResponseWriter, r *http.Request) {
	span := h.tracer.StartSpan("POST /albums")
	defer span.Finish()
	span.SetTag("http.method", r.Method)

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req pb.CreateAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.SetTag("error", true)
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resp, err := h.albumService.Create(ctx, &req, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusCreated)
	WriteJSON(w, http.StatusCreated, resp)
}

// List handles listing albums
func (h *AlbumHandler) List(w http.ResponseWriter, r *http.Request) {
	span := h.tracer.StartSpan("GET /albums")
	defer span.Finish()

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	resp, err := h.albumService.List(ctx, &pb.ListAlbumsRequest{}, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	WriteJSON(w, http.StatusOK, resp)
}

// Get handles retrieving a single album
func (h *AlbumHandler) Get(w http.ResponseWriter, r *http.Request) {
	span := h.tracer.StartSpan("GET /albums/{id}")
	defer span.Finish()

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	albumID := r.PathValue("id")
	if albumID == "" {
		span.SetTag("error", true)
		WriteError(w, http.StatusBadRequest, "Missing album ID")
		return
	}

	resp, err := h.albumService.Get(ctx, &pb.GetAlbumRequest{Id: albumID}, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	WriteJSON(w, http.StatusOK, resp)
}

// Update handles album updates
func (h *AlbumHandler) Update(w http.ResponseWriter, r *http.Request) {
	span := h.tracer.StartSpan("PUT /albums/{id}")
	defer span.Finish()

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	albumID := r.PathValue("id")
	var reqBody struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		span.SetTag("error", true)
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req := &pb.UpdateAlbumRequest{
		Id:          albumID,
		Name:        reqBody.Name,
		Description: reqBody.Description,
	}

	resp, err := h.albumService.Update(ctx, req, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	WriteJSON(w, http.StatusOK, resp)
}

// Delete handles album deletion
func (h *AlbumHandler) Delete(w http.ResponseWriter, r *http.Request) {
	span := h.tracer.StartSpan("DELETE /albums/{id}")
	defer span.Finish()

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	albumID := r.PathValue("id")
	// TODO: Parse delete_media parameter from query string

	resp, err := h.albumService.Delete(ctx, &pb.DeleteAlbumRequest{Id: albumID, DeleteMedia: false}, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	WriteJSON(w, http.StatusOK, resp)
}

// AddMedia handles adding media to album
func (h *AlbumHandler) AddMedia(w http.ResponseWriter, r *http.Request) {
	span := h.tracer.StartSpan("POST /albums/{id}/media")
	defer span.Finish()

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	albumID := r.PathValue("id")
	var reqBody struct {
		MediaIDs []string `json:"media_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		span.SetTag("error", true)
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req := &pb.AddMediaToAlbumRequest{
		AlbumId:  albumID,
		MediaIds: reqBody.MediaIDs,
	}

	resp, err := h.albumService.AddMedia(ctx, req, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	WriteJSON(w, http.StatusOK, resp)
}

// RemoveMedia handles removing media from album
func (h *AlbumHandler) RemoveMedia(w http.ResponseWriter, r *http.Request) {
	span := h.tracer.StartSpan("DELETE /albums/{id}/media")
	defer span.Finish()

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	albumID := r.PathValue("id")
	var reqBody struct {
		MediaIDs []string `json:"media_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		span.SetTag("error", true)
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req := &pb.RemoveMediaFromAlbumRequest{
		AlbumId:  albumID,
		MediaIds: reqBody.MediaIDs,
	}

	resp, err := h.albumService.RemoveMedia(ctx, req, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	WriteJSON(w, http.StatusOK, resp)
}

// GetAlbumMedia handles retrieving album contents
func (h *AlbumHandler) GetAlbumMedia(w http.ResponseWriter, r *http.Request) {
	span := h.tracer.StartSpan("GET /albums/{id}/media")
	defer span.Finish()

	ctx := r.Context()
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		span.SetTag("error", true)
		WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	albumID := r.PathValue("id")

	resp, err := h.albumService.GetAlbumMedia(ctx, &pb.GetAlbumMediaRequest{AlbumId: albumID}, userID)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	span.SetTag("http.status_code", http.StatusOK)
	WriteJSON(w, http.StatusOK, resp)
}

