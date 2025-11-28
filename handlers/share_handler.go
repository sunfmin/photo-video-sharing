package handlers

import (
	"net/http"
)

// ShareHandler handles share-related HTTP requests
type ShareHandler struct {
	// Will be implemented in Phase 6 (US3)
}

// NewShareHandler creates a new ShareHandler
func NewShareHandler() *ShareHandler {
	return &ShareHandler{}
}

// Stub methods (to be implemented in Phase 6)
func (h *ShareHandler) ShareMedia(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *ShareHandler) ShareAlbum(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *ShareHandler) RevokeShare(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *ShareHandler) ListMediaShares(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *ShareHandler) ListAlbumShares(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

func (h *ShareHandler) ListSharedAlbums(w http.ResponseWriter, r *http.Request) {
	WriteError(w, http.StatusNotImplemented, "Not implemented yet")
}

