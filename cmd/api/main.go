package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/yourorg/photo-video-sharing/handlers"
	"github.com/yourorg/photo-video-sharing/internal/config"
	"github.com/yourorg/photo-video-sharing/internal/storage"
	"github.com/yourorg/photo-video-sharing/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting Photo Video Sharing API server...")
	log.Printf("Environment: %s", cfg.Environment)

	// Connect to database
	db, err := connectDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("✓ Connected to PostgreSQL")

	// Run migrations
	if err := services.AutoMigrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("✓ Database migrations complete")

	// Initialize storage
	stor, err := initializeStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	log.Println("✓ Storage backend initialized")

	// Initialize services with builder pattern (Principle X)
	tracer := handlers.InitTracer()
	
	userService := services.NewUserService(db).
		WithTracer(tracer).
		Build()
	
	sessionService := services.NewSessionService(db).
		WithTracer(tracer).
		Build()
	
	mediaService := services.NewMediaService(db).
		WithStorage(stor).
		WithTracer(tracer).
		Build()
	
	albumService := services.NewAlbumService(db).
		WithTracer(tracer).
		Build()

	// Setup HTTP routes (Principle IV: Shared routing configuration)
	mux := setupAllRoutes(userService, sessionService, mediaService, albumService)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server listening on http://localhost:%s", cfg.Port)
		log.Println("📸 Photo Video Sharing API is ready!")
		log.Println("\nEndpoints:")
		log.Println("  POST   /auth/register       - Register new user")
		log.Println("  POST   /auth/login          - Login")
		log.Println("  GET    /auth/me             - Get current user")
		log.Println("  POST   /media               - Upload photo/video")
		log.Println("  GET    /media               - List media gallery")
		log.Println("  GET    /media/{id}          - Get media with presigned URL")
		log.Println("  DELETE /media/{id}          - Delete media")
		log.Println("  POST   /albums              - Create album")
		log.Println("  GET    /albums              - List albums")
		log.Println("  POST   /albums/{id}/media   - Add media to album")
		log.Println("  GET    /albums/{id}/media   - Get album contents")
		log.Println("  GET    /health              - Health check")
		log.Println()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Shutting down server...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("✓ Server exited gracefully")
}

func connectDatabase(cfg *config.Config) (*gorm.DB, error) {
	// Configure logger based on environment
	logLevel := logger.Silent
	if cfg.Environment == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get database connection: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func initializeStorage(cfg *config.Config) (services.StorageService, error) {
	// Use MinIO for production, local storage for development
	if cfg.Environment == "production" {
		minioStorage, err := storage.NewMinIOStorage(storage.MinIOConfig{
			Endpoint:  cfg.MinIOEndpoint,
			AccessKey: cfg.MinIOAccessKey,
			SecretKey: cfg.MinIOSecretKey,
			Bucket:    cfg.MinIOBucket,
			UseSSL:    cfg.MinIOUseSSL,
		})
		if err != nil {
			return nil, fmt.Errorf("initialize MinIO storage: %w", err)
		}
		return minioStorage, nil
	}

	// Development: use local filesystem
	localStorage, err := storage.NewLocalStorage("./local_storage")
	if err != nil {
		return nil, fmt.Errorf("initialize local storage: %w", err)
	}
	return localStorage, nil
}

func setupAllRoutes(
	userService services.UserService,
	sessionService services.SessionService,
	mediaService services.MediaService,
	albumService services.AlbumService,
) http.Handler {
	mux := http.NewServeMux()

	// Create handlers
	userHandler := handlers.NewUserHandler(userService, sessionService)
	mediaHandler := handlers.NewMediaHandler(mediaService)
	albumHandler := handlers.NewAlbumHandler(albumService)
	authMiddleware := handlers.NewAuthMiddleware(sessionService)

	// Health check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		handlers.WriteJSON(w, http.StatusOK, map[string]string{
			"status": "healthy",
			"service": "photo-video-sharing",
		})
	})

	// Authentication routes (public)
	mux.HandleFunc("POST /auth/register", userHandler.Register)
	mux.HandleFunc("POST /auth/login", userHandler.Login)
	mux.HandleFunc("POST /auth/logout", userHandler.Logout)
	mux.Handle("GET /auth/me", authMiddleware.RequireAuth(http.HandlerFunc(userHandler.GetCurrentUser)))
	mux.HandleFunc("POST /auth/password-reset", userHandler.PasswordReset)

	// Media routes (protected)
	mux.Handle("POST /media", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.Upload)))
	mux.Handle("GET /media", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.List)))
	mux.Handle("GET /media/{id}", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.Get)))
	mux.Handle("DELETE /media/{id}", authMiddleware.RequireAuth(http.HandlerFunc(mediaHandler.Delete)))

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

