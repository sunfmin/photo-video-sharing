package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"gorm.io/gorm"

	"github.com/yourorg/photo-video-sharing/handlers"
	"github.com/yourorg/photo-video-sharing/services"
	"github.com/yourorg/photo-video-sharing/testutil"
	pb "github.com/yourorg/photo-video-sharing/api/gen/v1"
)

func TestMediaHandler_UploadPhoto(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		filename       string
		contentType    string
		fileSize       int64
		wantStatusCode int
		wantError      string
	}{
		{
			name:           "US1-AS1: Upload JPG photo under 50MB",
			filename:       "test-photo.jpg",
			contentType:    "image/jpeg",
			fileSize:       1024000, // 1MB
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "Upload PNG photo",
			filename:       "test-photo.png",
			contentType:    "image/png",
			fileSize:       2048000, // 2MB
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "File too large (over 50MB)",
			filename:       "huge-photo.jpg",
			contentType:    "image/jpeg",
			fileSize:       52428801, // 50MB + 1 byte
			wantStatusCode: http.StatusRequestEntityTooLarge,
			wantError:      "file size exceeds limit",
		},
		{
			name:           "Unsupported file type",
			filename:       "document.pdf",
			contentType:    "application/pdf",
			fileSize:       1024,
			wantStatusCode: http.StatusBadRequest,
			wantError:      "unsupported",
		},
		{
			name:           "Empty file",
			filename:       "empty.jpg",
			contentType:    "image/jpeg",
			fileSize:       0,
			wantStatusCode: http.StatusBadRequest,
			wantError:      "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			db, cleanup := testutil.SetupTestDB(t)
			defer cleanup()
			defer testutil.TruncateTables(db, "media", "users", "sessions")

			if err := services.AutoMigrate(db); err != nil {
				t.Fatalf("Failed to migrate: %v", err)
			}

			// Create test user and session
			user := testutil.CreateTestUser(db, map[string]interface{}{
				"email": "user@example.com",
			})
			session := testutil.CreateTestSession(db, user.ID, nil)

			// Create mock storage
			storage := testutil.NewMockStorage()

			// Create services with builder pattern
			userService := services.NewUserService(db).Build()
			sessionService := services.NewSessionService(db).Build()
			mediaService := services.NewMediaService(db).WithStorage(storage).Build()

			// Setup routes
			mux := handlers.SetupMediaRoutes(userService, sessionService, mediaService)

			// Create multipart form with file upload
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			
			// Add file field
			fileWriter, err := writer.CreateFormFile("file", tt.filename)
			if err != nil {
				t.Fatalf("Failed to create form file: %v", err)
			}
			
			// Write fake file content
			fakeContent := make([]byte, tt.fileSize)
			if _, err := fileWriter.Write(fakeContent); err != nil {
				t.Fatalf("Failed to write file content: %v", err)
			}
			
			writer.Close()

			// Create request
			req := httptest.NewRequest("POST", "/media", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.AddCookie(&http.Cookie{
				Name:  "session_id",
				Value: session.ID,
			})
			rec := httptest.NewRecorder()

			// Execute
			mux.ServeHTTP(rec, req)

			// Assert status code
			if rec.Code != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatusCode, rec.Code, rec.Body.String())
			}

			// For successful upload
			if tt.wantStatusCode == http.StatusCreated {
				var resp pb.UploadMediaResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Principle V: Build expected from fixtures
				// Copy only truly random fields: ID, timestamps, storage_path (contains UUID)
				expected := &pb.UploadMediaResponse{
					Media: &pb.Media{
						Id:          resp.Media.Id,          // Random UUID - from response
						OwnerId:     user.ID,                // From fixture
						Filename:    tt.filename,            // From test case
						FileType:    tt.contentType,         // From test case
						FileSize:    tt.fileSize,            // From test case
						StoragePath: resp.Media.StoragePath, // Contains UUID - from response
						UploadedAt:  resp.Media.UploadedAt,  // Timestamp - from response
						CreatedAt:   resp.Media.CreatedAt,   // Timestamp - from response
						UpdatedAt:   resp.Media.UpdatedAt,   // Timestamp - from response
					},
					Message: resp.Message, // Dynamic message about processing
				}

				// Principle V: Use cmp.Diff with protocmp.Transform
				if diff := cmp.Diff(expected, &resp, protocmp.Transform()); diff != "" {
					t.Errorf("Response mismatch (-want +got):\n%s", diff)
				}

				// Verify file was uploaded to storage
				if !storage.Exists(context.Background(), resp.Media.StoragePath) {
					t.Error("File was not uploaded to storage")
				}
			}

			// For errors
			if tt.wantError != "" {
				body := rec.Body.String()
				if !strings.Contains(body, tt.wantError) {
					t.Errorf("Expected error containing %q, got: %s", tt.wantError, body)
				}
			}
		})
	}
}

func TestMediaHandler_UploadVideo(t *testing.T) {
	t.Parallel()

	// US1-AS2: Upload MP4 video under 500MB
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	defer testutil.TruncateTables(db, "media", "users", "sessions")

	if err := services.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create test user and session
	user := testutil.CreateTestUser(db, map[string]interface{}{
		"email": "user@example.com",
	})
	session := testutil.CreateTestSession(db, user.ID, nil)

	// Create mock storage
	storage := testutil.NewMockStorage()

	// Create services
	userService := services.NewUserService(db).Build()
	sessionService := services.NewSessionService(db).Build()
	mediaService := services.NewMediaService(db).WithStorage(storage).Build()

	mux := handlers.SetupMediaRoutes(userService, sessionService, mediaService)

	// Create multipart form with video
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	fileWriter, _ := writer.CreateFormFile("file", "test-video.mp4")
	fakeVideo := make([]byte, 10240000) // 10MB video
	fileWriter.Write(fakeVideo)
	writer.Close()

	req := httptest.NewRequest("POST", "/media", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.ID,
	})
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp pb.UploadMediaResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Media.FileType != "video/mp4" {
		t.Errorf("Expected video/mp4, got %s", resp.Media.FileType)
	}

	// Verify message about thumbnail processing
	if !strings.Contains(resp.Message, "processing") {
		t.Error("Expected message about thumbnail processing")
	}
}

func TestMediaHandler_ListMedia(t *testing.T) {
	t.Parallel()

	// US1-AS3: View gallery with uploaded media sorted by date
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	defer testutil.TruncateTables(db, "media", "users", "sessions")

	if err := services.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create test user and session
	user := testutil.CreateTestUser(db, map[string]interface{}{
		"email": "user@example.com",
	})
	session := testutil.CreateTestSession(db, user.ID, nil)

	// Create test media items with different upload times
	media1 := testutil.CreateTestMedia(db, map[string]interface{}{
		"owner_id":    user.ID,
		"filename":    "photo1.jpg",
		"uploaded_at": time.Now().Add(-2 * time.Hour),
	})
	media2 := testutil.CreateTestMedia(db, map[string]interface{}{
		"owner_id":    user.ID,
		"filename":    "photo2.jpg",
		"uploaded_at": time.Now().Add(-1 * time.Hour),
	})
	media3 := testutil.CreateTestMedia(db, map[string]interface{}{
		"owner_id":    user.ID,
		"filename":    "video1.mp4",
		"file_type":   "video/mp4",
		"uploaded_at": time.Now(),
	})

	// Create services
	storage := testutil.NewMockStorage()
	userService := services.NewUserService(db).Build()
	sessionService := services.NewSessionService(db).Build()
	mediaService := services.NewMediaService(db).WithStorage(storage).Build()

	mux := handlers.SetupMediaRoutes(userService, sessionService, mediaService)

	// Request media list
	req := httptest.NewRequest("GET", "/media", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.ID,
	})
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp pb.ListMediaResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify we got all media items
	if len(resp.Media) != 3 {
		t.Errorf("Expected 3 media items, got %d", len(resp.Media))
	}

	// Principle V: Build expected from fixtures in sorted order (newest first)
	// The service sorts by uploaded_at DESC, so: media3 > media2 > media1
	expectedMedia := []*pb.Media{
		{
			Id:          media3.ID,
			OwnerId:     user.ID,
			Filename:    "video1.mp4",
			FileType:    "video/mp4",
			FileSize:    media3.FileSize,
			StoragePath: media3.StoragePath,
			UploadedAt:  resp.Media[0].UploadedAt, // Timestamps from response
			CreatedAt:   resp.Media[0].CreatedAt,
			UpdatedAt:   resp.Media[0].UpdatedAt,
		},
		{
			Id:          media2.ID,
			OwnerId:     user.ID,
			Filename:    "photo2.jpg",
			FileType:    media2.FileType,
			FileSize:    media2.FileSize,
			StoragePath: media2.StoragePath,
			UploadedAt:  resp.Media[1].UploadedAt,
			CreatedAt:   resp.Media[1].CreatedAt,
			UpdatedAt:   resp.Media[1].UpdatedAt,
		},
		{
			Id:          media1.ID,
			OwnerId:     user.ID,
			Filename:    "photo1.jpg",
			FileType:    media1.FileType,
			FileSize:    media1.FileSize,
			StoragePath: media1.StoragePath,
			UploadedAt:  resp.Media[2].UploadedAt,
			CreatedAt:   resp.Media[2].CreatedAt,
			UpdatedAt:   resp.Media[2].UpdatedAt,
		},
	}

	// Add optional fields from fixtures
	for i, media := range []*testutil.TestMedia{media3, media2, media1} {
		if media.Width != nil {
			expectedMedia[i].Width = int32(*media.Width)
		}
		if media.Height != nil {
			expectedMedia[i].Height = int32(*media.Height)
		}
		if media.ThumbnailPath != nil {
			expectedMedia[i].ThumbnailPath = *media.ThumbnailPath
		}
	}

	expected := &pb.ListMediaResponse{
		Media:      expectedMedia,
		TotalCount: 3, // Total from fixtures
	}

	// Principle V: Use cmp.Diff with protocmp.Transform
	if diff := cmp.Diff(expected, &resp, protocmp.Transform()); diff != "" {
		t.Errorf("Response mismatch (-want +got):\n%s", diff)
	}
}

func TestMediaHandler_GetMedia(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		scenario       string
		setupMedia     func(db *gorm.DB, userID string) *testutil.TestMedia
		wantStatusCode int
	}{
		{
			name:     "US1-AS4: Get photo with presigned URL",
			scenario: "User clicks photo to view full quality",
			setupMedia: func(db *gorm.DB, userID string) *testutil.TestMedia {
				return testutil.CreateTestMedia(db, map[string]interface{}{
					"owner_id":  userID,
					"filename":  "vacation.jpg",
					"file_type": "image/jpeg",
				})
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:     "US1-AS5: Get video with presigned URL",
			scenario: "User clicks video for playback",
			setupMedia: func(db *gorm.DB, userID string) *testutil.TestMedia {
				return testutil.CreateTestMedia(db, map[string]interface{}{
					"owner_id":  userID,
					"filename":  "birthday.mp4",
					"file_type": "video/mp4",
				})
			},
			wantStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			db, cleanup := testutil.SetupTestDB(t)
			defer cleanup()
			defer testutil.TruncateTables(db, "media", "users", "sessions")

			if err := services.AutoMigrate(db); err != nil {
				t.Fatalf("Failed to migrate: %v", err)
			}

			user := testutil.CreateTestUser(db, map[string]interface{}{
				"email": "user@example.com",
			})
			session := testutil.CreateTestSession(db, user.ID, nil)

			media := tt.setupMedia(db, user.ID)

			// Create services
			storage := testutil.NewMockStorage()
			userService := services.NewUserService(db).Build()
			sessionService := services.NewSessionService(db).Build()
			mediaService := services.NewMediaService(db).WithStorage(storage).Build()

			mux := handlers.SetupMediaRoutes(userService, sessionService, mediaService)

			// Request specific media
			req := httptest.NewRequest("GET", "/media/"+media.ID, nil)
			req.AddCookie(&http.Cookie{
				Name:  "session_id",
				Value: session.ID,
			})
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatusCode, rec.Code, rec.Body.String())
			}

			if rec.Code != http.StatusOK {
				return // Skip further checks if failed
			}

			var resp pb.GetMediaResponse
			if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			// Principle V: Build expected from fixtures
			expected := &pb.GetMediaResponse{
				Media: &pb.Media{
					Id:          media.ID,                   // From fixture
					OwnerId:     user.ID,                    // From fixture
					Filename:    media.Filename,             // From fixture
					FileType:    media.FileType,             // From fixture
					FileSize:    media.FileSize,             // From fixture
					StoragePath: media.StoragePath,          // From fixture
					UploadedAt:  resp.Media.UploadedAt,      // Timestamp - from response
					CreatedAt:   resp.Media.CreatedAt,       // Timestamp - from response
					UpdatedAt:   resp.Media.UpdatedAt,       // Timestamp - from response
				},
				DownloadUrl:  resp.DownloadUrl,  // Dynamic presigned URL
				ThumbnailUrl: resp.ThumbnailUrl, // Dynamic presigned URL
			}

			// Add optional fields from fixture
			if media.Width != nil {
				expected.Media.Width = int32(*media.Width)
			}
			if media.Height != nil {
				expected.Media.Height = int32(*media.Height)
			}
			if media.ThumbnailPath != nil {
				expected.Media.ThumbnailPath = *media.ThumbnailPath
			}

			// Principle V: Use cmp.Diff with protocmp.Transform
			if diff := cmp.Diff(expected, &resp, protocmp.Transform()); diff != "" {
				t.Errorf("Response mismatch (-want +got):\n%s", diff)
			}

			// Verify presigned URLs are present
			if resp.DownloadUrl == "" {
				t.Error("Expected presigned download URL")
			}
			if media.ThumbnailPath != nil && resp.ThumbnailUrl == "" {
				t.Error("Expected presigned thumbnail URL when thumbnail exists")
			}
		})
	}
}

func TestMediaHandler_QuotaExceeded(t *testing.T) {
	t.Parallel()

	// Edge case: User exceeds storage quota
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	defer testutil.TruncateTables(db, "media", "users", "sessions")

	if err := services.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create user with nearly full quota
	user := testutil.CreateTestUser(db, map[string]interface{}{
		"email":         "user@example.com",
		"storage_used":  int64(524000000), // 499.5MB used
		"storage_quota": int64(524288000), // 500MB total
	})
	session := testutil.CreateTestSession(db, user.ID, nil)

	storage := testutil.NewMockStorage()
	userService := services.NewUserService(db).Build()
	sessionService := services.NewSessionService(db).Build()
	mediaService := services.NewMediaService(db).WithStorage(storage).Build()

	mux := handlers.SetupMediaRoutes(userService, sessionService, mediaService)

	// Try to upload 1MB file (would exceed quota)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fileWriter, _ := writer.CreateFormFile("file", "photo.jpg")
	fileWriter.Write(make([]byte, 1024000)) // 1MB
	writer.Close()

	req := httptest.NewRequest("POST", "/media", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.ID,
	})
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInsufficientStorage {
		t.Errorf("Expected status 507, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "quota") {
		t.Error("Expected quota exceeded error")
	}
}

func TestMediaHandler_Unauthorized(t *testing.T) {
	t.Parallel()

	// Edge case: User tries to access another user's media
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	defer testutil.TruncateTables(db, "media", "users", "sessions")

	if err := services.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create two users
	user1 := testutil.CreateTestUser(db, map[string]interface{}{
		"email": "user1@example.com",
	})
	user2 := testutil.CreateTestUser(db, map[string]interface{}{
		"email": "user2@example.com",
	})
	session1 := testutil.CreateTestSession(db, user1.ID, nil)

	// User2's media
	media2 := testutil.CreateTestMedia(db, map[string]interface{}{
		"owner_id": user2.ID,
		"filename": "private-photo.jpg",
	})

	storage := testutil.NewMockStorage()
	userService := services.NewUserService(db).Build()
	sessionService := services.NewSessionService(db).Build()
	mediaService := services.NewMediaService(db).WithStorage(storage).Build()

	mux := handlers.SetupMediaRoutes(userService, sessionService, mediaService)

	// User1 tries to access User2's media
	req := httptest.NewRequest("GET", "/media/"+media2.ID, nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session1.ID,
	})
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound && rec.Code != http.StatusForbidden {
		t.Errorf("Expected status 404 or 403, got %d", rec.Code)
	}
}

