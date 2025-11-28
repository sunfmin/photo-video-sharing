package handlers

import (
	"net/http"

	"github.com/opentracing/opentracing-go"
	
	"github.com/yourorg/photo-video-sharing/services"
)

// SetupRoutes configures all application routes
// This function is used by both production and tests to ensure consistent routing
// Principle IV: Shared routing configuration ensures test accuracy
func SetupRoutes(
	userService services.UserService,
	sessionService services.SessionService,
) http.Handler {
	mux := http.NewServeMux()

	// Create handlers
	userHandler := NewUserHandler(userService, sessionService)
	mediaHandler := NewMediaHandler()
	albumHandler := NewAlbumHandler()
	shareHandler := NewShareHandler()

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Authentication routes (public)
	mux.HandleFunc("POST /auth/register", userHandler.Register)
	mux.HandleFunc("POST /auth/login", userHandler.Login)
	mux.HandleFunc("POST /auth/logout", userHandler.Logout)
	mux.HandleFunc("GET /auth/me", userHandler.GetCurrentUser)
	mux.HandleFunc("POST /auth/password-reset", userHandler.PasswordReset)

	// Media routes (protected)
	mux.HandleFunc("POST /media", mediaHandler.Upload)
	mux.HandleFunc("GET /media", mediaHandler.List)
	mux.HandleFunc("GET /media/{id}", mediaHandler.Get)
	mux.HandleFunc("DELETE /media/{id}", mediaHandler.Delete)
	mux.HandleFunc("GET /media/shared", mediaHandler.ListSharedWithMe)

	// Album routes (protected)
	mux.HandleFunc("POST /albums", albumHandler.Create)
	mux.HandleFunc("GET /albums", albumHandler.List)
	mux.HandleFunc("GET /albums/{id}", albumHandler.Get)
	mux.HandleFunc("PUT /albums/{id}", albumHandler.Update)
	mux.HandleFunc("DELETE /albums/{id}", albumHandler.Delete)
	mux.HandleFunc("POST /albums/{id}/media", albumHandler.AddMedia)
	mux.HandleFunc("DELETE /albums/{id}/media", albumHandler.RemoveMedia)
	mux.HandleFunc("GET /albums/{id}/media", albumHandler.GetAlbumMedia)

	// Share routes (protected)
	mux.HandleFunc("POST /shares/media", shareHandler.ShareMedia)
	mux.HandleFunc("POST /shares/album", shareHandler.ShareAlbum)
	mux.HandleFunc("DELETE /shares/{id}", shareHandler.RevokeShare)
	mux.HandleFunc("GET /shares/media/{id}", shareHandler.ListMediaShares)
	mux.HandleFunc("GET /shares/album/{id}", shareHandler.ListAlbumShares)
	mux.HandleFunc("GET /albums/shared", shareHandler.ListSharedAlbums)

	return mux
}

// InitTracer initializes OpenTracing with NoopTracer for development
func InitTracer() opentracing.Tracer {
	return opentracing.NoopTracer{}
}

