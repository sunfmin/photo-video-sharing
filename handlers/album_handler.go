package handlers

import (
	"net/http"
)

// AlbumHandler handles album-related HTTP requests
type AlbumHandler struct {
	// Will be implemented in Phase 5 (US4)
}

// NewAlbumHandler creates a new AlbumHandler
func NewAlbumHandler() *AlbumHandler {
	return &AlbumHandler{}
}

// Stub methods (to be implemented in Phase 5)
func (h *AlbumHandler) Create(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *AlbumHandler) List(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *AlbumHandler) Get(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *AlbumHandler) Update(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *AlbumHandler) Delete(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *AlbumHandler) AddMedia(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *AlbumHandler) RemoveMedia(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *AlbumHandler) GetAlbumMedia(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

