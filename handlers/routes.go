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
	mediaHandler := NewMediaHandler(nil) // Media service not initialized in auth-only setup
	albumHandler := NewAlbumHandler(nil) // Album service not initialized in auth-only setup
	shareHandler := NewShareHandler(nil) // Share service not initialized in auth-only setup

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

	// Media routes (protected) - only if media handler is initialized
	if mediaHandler.mediaService != nil {
		authMiddleware := NewAuthMiddleware(sessionService)
		mux.Handle("POST /media", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.Upload)))
		mux.Handle("GET /media", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.List)))
		mux.Handle("GET /media/{id}", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.Get)))
		mux.Handle("DELETE /media/{id}", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.Delete)))
		// Note: /media/shared requires ShareService, added in SetupShareRoutes
	}

	// Album routes (protected) - only if album handler is initialized
	if albumHandler.albumService != nil {
		authMiddleware := NewAuthMiddleware(sessionService)
		mux.Handle("POST /albums", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.Create)))
		mux.Handle("GET /albums", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.List)))
		mux.Handle("GET /albums/{id}", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.Get)))
		mux.Handle("PUT /albums/{id}", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.Update)))
		mux.Handle("DELETE /albums/{id}", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.Delete)))
		mux.Handle("POST /albums/{id}/media", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.AddMedia)))
		mux.Handle("DELETE /albums/{id}/media", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.RemoveMedia)))
		mux.Handle("GET /albums/{id}/media", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.GetAlbumMedia)))
	}

	// Share routes (protected) - only if share handler is initialized
	if shareHandler.shareService != nil {
		authMiddleware := NewAuthMiddleware(sessionService)
		mux.Handle("POST /shares/media", authMiddleware.RequireAuth(http.HandlerFunc(shareHandler.ShareMedia)))
		mux.Handle("POST /shares/album", authMiddleware.RequireAuth(http.HandlerFunc(shareHandler.ShareAlbum)))
		mux.Handle("DELETE /shares/{id}", authMiddleware.RequireAuth(http.HandlerFunc(shareHandler.RevokeShare)))
		mux.Handle("GET /shares/media/{id}", authMiddleware.RequireAuth(http.HandlerFunc(shareHandler.ListMediaShares)))
		mux.Handle("GET /shares/album/{id}", authMiddleware.RequireAuth(http.HandlerFunc(shareHandler.ListAlbumShares)))
		mux.Handle("GET /albums/shared", authMiddleware.RequireAuth(http.HandlerFunc(shareHandler.ListSharedAlbums)))
	}

	return mux
}

// SetupMediaRoutes is a helper for tests that need media routes
// Principle IV: Shared routing configuration
func SetupMediaRoutes(
	userService services.UserService,
	sessionService services.SessionService,
	mediaService services.MediaService,
) http.Handler {
	mux := http.NewServeMux()

	// Create handlers
	userHandler := NewUserHandler(userService, sessionService)
	mediaHandler := NewMediaHandler(mediaService)
	authMiddleware := NewAuthMiddleware(sessionService)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Authentication routes (public)
	mux.HandleFunc("POST /auth/register", userHandler.Register)
	mux.HandleFunc("POST /auth/login", userHandler.Login)

	// Media routes (protected)
	mux.Handle("POST /media", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.Upload)))
	mux.Handle("GET /media", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.List)))
	mux.Handle("GET /media/{id}", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.Get)))
	mux.Handle("DELETE /media/{id}", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.Delete)))

	return mux
}

// SetupAlbumRoutes is a helper for tests that need album routes
// Principle IV: Shared routing configuration
func SetupAlbumRoutes(
	userService services.UserService,
	sessionService services.SessionService,
	albumService services.AlbumService,
) http.Handler {
	mux := http.NewServeMux()

	// Create handlers
	albumHandler := NewAlbumHandler(albumService)
	authMiddleware := NewAuthMiddleware(sessionService)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Album routes (protected)
	mux.Handle("POST /albums", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.Create)))
	mux.Handle("GET /albums", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.List)))
	mux.Handle("GET /albums/{id}", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.Get)))
	mux.Handle("PUT /albums/{id}", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.Update)))
	mux.Handle("DELETE /albums/{id}", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.Delete)))
	mux.Handle("POST /albums/{id}/media", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.AddMedia)))
	mux.Handle("DELETE /albums/{id}/media", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.RemoveMedia)))
	mux.Handle("GET /albums/{id}/media", authMiddleware.RequireAuth(http.HandlerFunc(albumHandler.GetAlbumMedia)))

	return mux
}

// SetupShareRoutes is a helper for tests that need sharing routes
// Principle IV: Shared routing configuration
func SetupShareRoutes(
	userService services.UserService,
	sessionService services.SessionService,
	mediaService services.MediaService,
	shareService services.ShareService,
) http.Handler {
	mux := http.NewServeMux()

	// Create handlers
	mediaHandler := NewMediaHandler(mediaService)
	shareHandler := NewShareHandler(shareService)
	authMiddleware := NewAuthMiddleware(sessionService)

	// Authentication routes
	userHandler := NewUserHandler(userService, sessionService)
	mux.HandleFunc("POST /auth/register", userHandler.Register)
	mux.HandleFunc("POST /auth/login", userHandler.Login)

	// Media routes (protected)
	mux.Handle("GET /media/{id}", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.Get)))
	mux.Handle("DELETE /media/{id}", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.Delete)))
	mux.Handle("GET /media/shared", authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mediaHandler.ListSharedWithMe(w, r, shareService)
	})))

	// Share routes (protected)
	mux.Handle("POST /shares/media", authMiddleware.RequireAuth(http.HandlerFunc(shareHandler.ShareMedia)))
	mux.Handle("POST /shares/album", authMiddleware.RequireAuth(http.HandlerFunc(shareHandler.ShareAlbum)))
	mux.Handle("DELETE /shares/{id}", authMiddleware.RequireAuth(http.HandlerFunc(shareHandler.RevokeShare)))
	mux.Handle("GET /shares/media/{id}", authMiddleware.RequireAuth(http.HandlerFunc(shareHandler.ListMediaShares)))

	return mux
}

// InitTracer initializes OpenTracing with NoopTracer for development
func InitTracer() opentracing.Tracer {
	return opentracing.NoopTracer{}
}

