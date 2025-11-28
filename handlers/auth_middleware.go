package handlers

import (
	"context"
	"net/http"

	"github.com/yourorg/photo-video-sharing/services"
)

// AuthMiddleware handles authentication for protected routes
type AuthMiddleware struct {
	sessionService *services.SessionService
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(sessionService *services.SessionService) *AuthMiddleware {
	return &AuthMiddleware{
		sessionService: sessionService,
	}
}

// RequireAuth is middleware that requires authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session cookie
		cookie, err := r.Cookie("session_id")
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Validate session
		userID, err := m.sessionService.ValidateSession(r.Context(), cookie.Value)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "Invalid or expired session")
			return
		}

		// Add user ID to context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

