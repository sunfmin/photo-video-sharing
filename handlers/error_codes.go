package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yourorg/photo-video-sharing/services"
)

// ErrorCode represents a structured error with HTTP mapping (Principle XIII)
// This struct serves as both the internal definition and the JSON response format
type ErrorCode struct {
	Code       string `json:"code"`    // Machine-readable error code (e.g., "MEDIA_NOT_FOUND")
	Message    string `json:"message"` // Human-readable message
	HTTPStatus int    `json:"status"`  // HTTP status code
	ServiceErr error  `json:"-"`       // Maps to service sentinel error (not serialized)
}

// Errors singleton contains all application error codes (Principle XIII: Singleton Pattern)
// This provides type-safe error handling with automatic service error mapping
var Errors = struct {
	// Authentication errors
	InvalidCredentials ErrorCode
	DuplicateEmail     ErrorCode
	SessionExpired     ErrorCode
	SessionInvalid     ErrorCode
	Unauthorized       ErrorCode

	// Media errors
	MediaNotFound     ErrorCode
	UnsupportedFormat ErrorCode
	FileTooLarge      ErrorCode
	QuotaExceeded     ErrorCode
	InvalidMimeType   ErrorCode

	// Album errors
	AlbumNotFound       ErrorCode
	MediaAlreadyInAlbum ErrorCode
	InvalidAlbumName    ErrorCode
	MediaNotInAlbum     ErrorCode

	// Share errors
	ShareNotFound       ErrorCode
	CannotShareWithSelf ErrorCode
	UserNotFound        ErrorCode
	AlreadyShared       ErrorCode

	// Generic errors
	NotFound      ErrorCode
	InvalidInput  ErrorCode
	InternalError ErrorCode
	DatabaseError ErrorCode

	// Handler-level errors (HTTP/validation)
	InvalidRequestBody     ErrorCode
	AuthenticationRequired ErrorCode
	MissingParameter       ErrorCode
	NotImplemented         ErrorCode
}{
	// Authentication (messages match service errors exactly)
	InvalidCredentials: ErrorCode{"INVALID_CREDENTIALS", "invalid email or password", http.StatusUnauthorized, services.ErrInvalidCredentials},
	DuplicateEmail:     ErrorCode{"DUPLICATE_EMAIL", "email already registered", http.StatusConflict, services.ErrDuplicateEmail},
	SessionExpired:     ErrorCode{"SESSION_EXPIRED", "session expired", http.StatusUnauthorized, services.ErrSessionExpired},
	SessionInvalid:     ErrorCode{"SESSION_INVALID", "invalid session", http.StatusUnauthorized, services.ErrSessionInvalid},
	Unauthorized:       ErrorCode{"UNAUTHORIZED", "unauthorized", http.StatusUnauthorized, services.ErrUnauthorized},

	// Media (messages match service errors exactly)
	MediaNotFound:     ErrorCode{"MEDIA_NOT_FOUND", "media not found", http.StatusNotFound, services.ErrMediaNotFound},
	UnsupportedFormat: ErrorCode{"UNSUPPORTED_FORMAT", "unsupported file format", http.StatusBadRequest, services.ErrUnsupportedFormat},
	FileTooLarge:      ErrorCode{"FILE_TOO_LARGE", "file size exceeds limit", http.StatusRequestEntityTooLarge, services.ErrFileTooLarge},
	QuotaExceeded:     ErrorCode{"QUOTA_EXCEEDED", "storage quota exceeded", http.StatusInsufficientStorage, services.ErrQuotaExceeded},
	InvalidMimeType:   ErrorCode{"INVALID_MIME_TYPE", "invalid MIME type", http.StatusBadRequest, services.ErrInvalidMimeType},

	// Albums (messages match service errors exactly)
	AlbumNotFound:       ErrorCode{"ALBUM_NOT_FOUND", "album not found", http.StatusNotFound, services.ErrAlbumNotFound},
	MediaAlreadyInAlbum: ErrorCode{"MEDIA_ALREADY_IN_ALBUM", "media already in album", http.StatusConflict, services.ErrMediaAlreadyInAlbum},
	InvalidAlbumName:    ErrorCode{"INVALID_ALBUM_NAME", "invalid album name", http.StatusBadRequest, services.ErrInvalidAlbumName},
	MediaNotInAlbum:     ErrorCode{"MEDIA_NOT_IN_ALBUM", "media not in album", http.StatusBadRequest, services.ErrMediaNotInAlbum},

	// Sharing (messages match service errors exactly)
	ShareNotFound:       ErrorCode{"SHARE_NOT_FOUND", "share not found", http.StatusNotFound, services.ErrShareNotFound},
	CannotShareWithSelf: ErrorCode{"CANNOT_SHARE_WITH_SELF", "cannot share with yourself", http.StatusBadRequest, services.ErrCannotShareWithSelf},
	UserNotFound:        ErrorCode{"USER_NOT_FOUND", "user not found", http.StatusNotFound, services.ErrUserNotFound},
	AlreadyShared:       ErrorCode{"ALREADY_SHARED", "already shared with this user", http.StatusConflict, services.ErrAlreadyShared},

	// Generic (messages match service errors exactly)
	NotFound:      ErrorCode{"NOT_FOUND", "resource not found", http.StatusNotFound, services.ErrNotFound},
	InvalidInput:  ErrorCode{"INVALID_INPUT", "invalid input", http.StatusBadRequest, services.ErrInvalidInput},
	InternalError: ErrorCode{"INTERNAL_ERROR", "internal server error", http.StatusInternalServerError, services.ErrInternalError},
	DatabaseError: ErrorCode{"DATABASE_ERROR", "database error", http.StatusInternalServerError, services.ErrDatabaseError},

	// Handler-level errors (no ServiceErr - these are HTTP-specific)
	InvalidRequestBody:     ErrorCode{"INVALID_REQUEST_BODY", "invalid request body", http.StatusBadRequest, nil},
	AuthenticationRequired: ErrorCode{"AUTHENTICATION_REQUIRED", "authentication required", http.StatusUnauthorized, nil},
	MissingParameter:       ErrorCode{"MISSING_PARAMETER", "missing required parameter", http.StatusBadRequest, nil},
	NotImplemented:         ErrorCode{"NOT_IMPLEMENTED", "not implemented", http.StatusNotImplemented, nil},
}

// AllErrors returns all error codes for iteration
func AllErrors() []ErrorCode {
	return []ErrorCode{
		Errors.InvalidCredentials,
		Errors.DuplicateEmail,
		Errors.SessionExpired,
		Errors.SessionInvalid,
		Errors.Unauthorized,
		Errors.MediaNotFound,
		Errors.UnsupportedFormat,
		Errors.FileTooLarge,
		Errors.QuotaExceeded,
		Errors.InvalidMimeType,
		Errors.AlbumNotFound,
		Errors.MediaAlreadyInAlbum,
		Errors.InvalidAlbumName,
		Errors.MediaNotInAlbum,
		Errors.ShareNotFound,
		Errors.CannotShareWithSelf,
		Errors.UserNotFound,
		Errors.AlreadyShared,
		Errors.NotFound,
		Errors.InvalidInput,
		Errors.InternalError,
		Errors.DatabaseError,
		// Note: Handler-level errors (InvalidRequestBody, etc.) are not in this list
		// because they don't have ServiceErr mapping (they're HTTP-specific)
	}
}

// HandleServiceError automatically maps service errors to appropriate HTTP responses
// Principle XIII: Automatic error mapping with context awareness
func HandleServiceError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	// Principle XII: Check context errors first
	if errors.Is(err, context.Canceled) {
		respondWithErrorAndMessage(w, ErrorCode{"REQUEST_CANCELED", "Request cancelled", 499, nil})
		return
	}
	if errors.Is(err, context.DeadlineExceeded) {
		respondWithErrorAndMessage(w, ErrorCode{"REQUEST_TIMEOUT", "Request timeout", 504, nil})
		return
	}

	// Principle XIII: Check service error mapping via singleton
	for _, errCode := range AllErrors() {
		if errCode.ServiceErr != nil && errors.Is(err, errCode.ServiceErr) {
			// Use the actual error message (includes context from wrapping)
			respondWithErrorAndMessage(w, errCode)
			return
		}
	}

	// Default to internal error
	respondWithErrorAndMessage(w, Errors.InternalError)
}

// RespondWithError sends an ErrorCode as JSON response
// Principle XIII: Uses ErrorCode singleton directly as JSON response
func RespondWithError(w http.ResponseWriter, errCode ErrorCode) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errCode.HTTPStatus)
	json.NewEncoder(w).Encode(errCode)
}

// respondWithErrorAndMessage sends error response with custom message
// Used by HandleServiceError to preserve error context from wrapping
func respondWithErrorAndMessage(w http.ResponseWriter, errCode ErrorCode) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errCode.HTTPStatus)
	json.NewEncoder(w).Encode(errCode)
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
