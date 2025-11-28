package handlers

import (
	"net/http"
)

// MediaHandler handles media-related HTTP requests
type MediaHandler struct {
	// Will be implemented in Phase 4 (US1)
}

// NewMediaHandler creates a new MediaHandler
func NewMediaHandler() *MediaHandler {
	return &MediaHandler{}
}

// Stub methods (to be implemented in Phase 4)
func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *MediaHandler) Get(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *MediaHandler) ListSharedWithMe(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

