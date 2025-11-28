package services

import "errors"

// Authentication errors
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrDuplicateEmail     = errors.New("email already registered")
	ErrSessionExpired     = errors.New("session expired")
	ErrSessionInvalid     = errors.New("invalid session")
	ErrUnauthorized       = errors.New("unauthorized")
)

// Media errors
var (
	ErrMediaNotFound      = errors.New("media not found")
	ErrUnsupportedFormat  = errors.New("unsupported file format")
	ErrFileTooLarge       = errors.New("file size exceeds limit")
	ErrQuotaExceeded      = errors.New("storage quota exceeded")
	ErrInvalidMimeType    = errors.New("invalid MIME type")
)

// Album errors
var (
	ErrAlbumNotFound        = errors.New("album not found")
	ErrMediaAlreadyInAlbum  = errors.New("media already in album")
	ErrInvalidAlbumName     = errors.New("invalid album name")
	ErrMediaNotInAlbum      = errors.New("media not in album")
)

// Share errors
var (
	ErrShareNotFound       = errors.New("share not found")
	ErrCannotShareWithSelf = errors.New("cannot share with yourself")
	ErrUserNotFound        = errors.New("user not found")
	ErrAlreadyShared       = errors.New("already shared with this user")
)

// Generic errors
var (
	ErrNotFound       = errors.New("resource not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrInternalError  = errors.New("internal server error")
	ErrDatabaseError  = errors.New("database error")
)

