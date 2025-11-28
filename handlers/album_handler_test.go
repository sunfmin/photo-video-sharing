package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/yourorg/photo-video-sharing/handlers"
	"github.com/yourorg/photo-video-sharing/services"
	"github.com/yourorg/photo-video-sharing/testutil"
	pb "github.com/yourorg/photo-video-sharing/api/gen/v1"
)

func TestAlbumHandler_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		albumName      string
		description    string
		wantStatusCode int
		wantError      string
	}{
		{
			name:           "US4-AS1: Create album with valid name",
			albumName:      "Vacation 2025",
			description:    "Our summer vacation photos",
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "Create album without description",
			albumName:      "Family",
			description:    "",
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "Empty album name",
			albumName:      "",
			description:    "Test",
			wantStatusCode: http.StatusBadRequest,
			wantError:      "name",
		},
		{
			name:           "Album name too long",
			albumName:      strings.Repeat("A", 101), // 101 chars
			description:    "Test",
			wantStatusCode: http.StatusBadRequest,
			wantError:      "name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			db, cleanup := testutil.SetupTestDB(t)
			defer cleanup()
			defer testutil.TruncateTables(db, "albums", "users", "sessions")

			if err := services.AutoMigrate(db); err != nil {
				t.Fatalf("Failed to migrate: %v", err)
			}

			// Create test user and session
			user := testutil.CreateTestUser(db, map[string]interface{}{
				"email": "user@example.com",
			})
			session := testutil.CreateTestSession(db, user.ID, nil)

			// Create services
			userService := services.NewUserService(db).Build()
			sessionService := services.NewSessionService(db).Build()
			albumService := services.NewAlbumService(db).Build()

			mux := handlers.SetupAlbumRoutes(userService, sessionService, albumService)

			// Create request
			reqBody := map[string]string{
				"name":        tt.albumName,
				"description": tt.description,
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/albums", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.AddCookie(&http.Cookie{
				Name:  "session_id",
				Value: session.ID,
			})
			rec := httptest.NewRecorder()

			// Execute
			mux.ServeHTTP(rec, req)

			// Assert
			if rec.Code != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatusCode, rec.Code, rec.Body.String())
			}

			if tt.wantStatusCode == http.StatusCreated {
				var resp pb.CreateAlbumResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				// Principle V: Build expected from fixtures, copy only random fields
				expected := &pb.CreateAlbumResponse{
					Album: &pb.Album{
						Id:          resp.Album.Id,          // Random UUID - from response
						OwnerId:     user.ID,                // From fixture
						Name:        tt.albumName,           // From test case
						Description: tt.description,         // From test case
						MediaCount:  0,                      // No media added yet
						CreatedAt:   resp.Album.CreatedAt,   // Timestamp - from response
						UpdatedAt:   resp.Album.UpdatedAt,   // Timestamp - from response
					},
				}

				// Principle V: Use cmp.Diff with protocmp.Transform
				if diff := cmp.Diff(expected, &resp, protocmp.Transform()); diff != "" {
					t.Errorf("Response mismatch (-want +got):\n%s", diff)
				}
			}

			if tt.wantError != "" {
				if !strings.Contains(rec.Body.String(), tt.wantError) {
					t.Errorf("Expected error containing %q, got: %s", tt.wantError, rec.Body.String())
				}
			}
		})
	}
}

func TestAlbumHandler_AddMedia(t *testing.T) {
	t.Parallel()

	// US4-AS2: Add media to album
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	defer testutil.TruncateTables(db, "album_media", "albums", "media", "users", "sessions")

	if err := services.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create test user
	user := testutil.CreateTestUser(db, map[string]interface{}{
		"email": "user@example.com",
	})
	session := testutil.CreateTestSession(db, user.ID, nil)

	// Create album and media
	album := testutil.CreateTestAlbum(db, map[string]interface{}{
		"owner_id": user.ID,
		"name":     "My Album",
	})
	media := testutil.CreateTestMedia(db, map[string]interface{}{
		"owner_id": user.ID,
		"filename": "photo.jpg",
	})

	// Create services
	userService := services.NewUserService(db).Build()
	sessionService := services.NewSessionService(db).Build()
	albumService := services.NewAlbumService(db).Build()

	mux := handlers.SetupAlbumRoutes(userService, sessionService, albumService)

	// Add media to album
	reqBody := map[string]interface{}{
		"media_ids": []string{media.ID},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/albums/"+album.ID+"/media", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.ID,
	})
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp pb.AddMediaToAlbumResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Principle V: Build expected from fixtures
	expected := &pb.AddMediaToAlbumResponse{
		AddedCount: 1,      // 1 media added
		Errors:     []string{}, // No errors expected
	}

	// Principle V: Use cmp.Diff with protocmp.Transform
	if diff := cmp.Diff(expected, &resp, protocmp.Transform()); diff != "" {
		t.Errorf("Response mismatch (-want +got):\n%s", diff)
	}

	// Verify media appears in album
	reqGet := httptest.NewRequest("GET", "/albums/"+album.ID+"/media", nil)
	reqGet.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.ID,
	})
	recGet := httptest.NewRecorder()

	mux.ServeHTTP(recGet, reqGet)

	if recGet.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recGet.Code)
	}

	var albumMedia pb.GetAlbumMediaResponse
	json.NewDecoder(recGet.Body).Decode(&albumMedia)

	if len(albumMedia.Media) != 1 {
		t.Errorf("Expected 1 media in album, got %d", len(albumMedia.Media))
	}
	if len(albumMedia.Media) > 0 && albumMedia.Media[0].Id != media.ID {
		t.Error("Wrong media in album")
	}
}

func TestAlbumHandler_RemoveMedia(t *testing.T) {
	t.Parallel()

	// US4-AS3: Remove media from album (media still in gallery)
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	defer testutil.TruncateTables(db, "album_media", "albums", "media", "users", "sessions")

	if err := services.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	user := testutil.CreateTestUser(db, map[string]interface{}{
		"email": "user@example.com",
	})
	session := testutil.CreateTestSession(db, user.ID, nil)

	album := testutil.CreateTestAlbum(db, map[string]interface{}{
		"owner_id": user.ID,
	})
	media := testutil.CreateTestMedia(db, map[string]interface{}{
		"owner_id": user.ID,
	})

	// Add media to album first
	db.Exec("INSERT INTO album_media (album_id, media_id, added_at) VALUES (?, ?, ?)",
		album.ID, media.ID, context.Background())

	userService := services.NewUserService(db).Build()
	sessionService := services.NewSessionService(db).Build()
	albumService := services.NewAlbumService(db).Build()

	mux := handlers.SetupAlbumRoutes(userService, sessionService, albumService)

	// Remove media from album
	reqBody := map[string]interface{}{
		"media_ids": []string{media.ID},
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("DELETE", "/albums/"+album.ID+"/media", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.ID,
	})
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Verify media removed from album
	var count int64
	db.Model(&testutil.TestMedia{}).
		Joins("INNER JOIN album_media ON media.id = album_media.media_id").
		Where("album_media.album_id = ?", album.ID).
		Count(&count)

	if count != 0 {
		t.Error("Media should be removed from album")
	}

	// Verify media still exists in gallery
	var mediaStillExists testutil.TestMedia
	if err := db.Where("id = ?", media.ID).First(&mediaStillExists).Error; err != nil {
		t.Error("Media should still exist in gallery after removal from album")
	}
}

