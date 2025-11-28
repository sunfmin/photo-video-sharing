package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yourorg/photo-video-sharing/services"
)

// ErrorResponse represents an error response sent to clients
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// ErrorMapping maps service errors to HTTP status codes
var ErrorMapping = map[error]int{
	// Authentication errors
	services.ErrInvalidCredentials: http.StatusUnauthorized,
	services.ErrDuplicateEmail:     http.StatusConflict,
	services.ErrSessionExpired:     http.StatusUnauthorized,
	services.ErrSessionInvalid:     http.StatusUnauthorized,
	services.ErrUnauthorized:       http.StatusUnauthorized,

	// Media errors
	services.ErrMediaNotFound:      http.StatusNotFound,
	services.ErrUnsupportedFormat:  http.StatusBadRequest,
	services.ErrFileTooLarge:       http.StatusRequestEntityTooLarge,
	services.ErrQuotaExceeded:      http.StatusInsufficientStorage,
	services.ErrInvalidMimeType:    http.StatusBadRequest,

	// Album errors
	services.ErrAlbumNotFound:        http.StatusNotFound,
	services.ErrMediaAlreadyInAlbum:  http.StatusConflict,
	services.ErrInvalidAlbumName:     http.StatusBadRequest,
	services.ErrMediaNotInAlbum:      http.StatusBadRequest,

	// Share errors
	services.ErrShareNotFound:       http.StatusNotFound,
	services.ErrCannotShareWithSelf: http.StatusBadRequest,
	services.ErrUserNotFound:        http.StatusNotFound,
	services.ErrAlreadyShared:       http.StatusConflict,

	// Generic errors
	services.ErrNotFound:      http.StatusNotFound,
	services.ErrInvalidInput:  http.StatusBadRequest,
	services.ErrInternalError: http.StatusInternalServerError,
	services.ErrDatabaseError: http.StatusInternalServerError,
}

// HandleServiceError automatically maps service errors to appropriate HTTP responses
func HandleServiceError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	// Find the matching error by unwrapping
	statusCode := http.StatusInternalServerError
	for sentinelErr, code := range ErrorMapping {
		if errors.Is(err, sentinelErr) {
			statusCode = code
			break
		}
	}

	// Create error response
	resp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: err.Error(),
		Code:    statusCode,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

// WriteJSON writes a JSON response
func WriteJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If encoding fails, try to send a plain error
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// WriteError writes an error response
func WriteError(w http.ResponseWriter, statusCode int, message string) {
	resp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}
	WriteJSON(w, statusCode, resp)
}

