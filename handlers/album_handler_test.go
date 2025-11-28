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

func TestAlbumHandler_ListAlbums(t *testing.T) {
	t.Parallel()

	// US4-AS4: View albums list with names, covers, and media count
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

	// Create albums with media
	album1 := testutil.CreateTestAlbum(db, map[string]interface{}{
		"owner_id": user.ID,
		"name":     "Vacation 2025",
	})
	_ = testutil.CreateTestAlbum(db, map[string]interface{}{
		"owner_id": user.ID,
		"name":     "Family Photos",
	})

	// Add media to albums
	media1 := testutil.CreateTestMedia(db, map[string]interface{}{
		"owner_id": user.ID,
	})
	media2 := testutil.CreateTestMedia(db, map[string]interface{}{
		"owner_id": user.ID,
	})

	db.Exec("INSERT INTO album_media (album_id, media_id, added_at) VALUES (?, ?, NOW())", album1.ID, media1.ID)
	db.Exec("INSERT INTO album_media (album_id, media_id, added_at) VALUES (?, ?, NOW())", album1.ID, media2.ID)

	userService := services.NewUserService(db).Build()
	sessionService := services.NewSessionService(db).Build()
	albumService := services.NewAlbumService(db).Build()

	mux := handlers.SetupAlbumRoutes(userService, sessionService, albumService)

	// Request albums list
	req := httptest.NewRequest("GET", "/albums", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.ID,
	})
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp pb.ListAlbumsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Principle VIII: Verify acceptance scenario
	// User should see:
	// - Album names ✅
	// - Cover thumbnails (TODO: implement cover_thumbnail_url)
	// - Media count per album (TODO: implement media_count)

	if len(resp.Albums) != 2 {
		t.Errorf("Expected 2 albums, got %d", len(resp.Albums))
	}

	// Verify albums are present with correct names
	albumNames := make(map[string]bool)
	for _, album := range resp.Albums {
		albumNames[album.Name] = true
		
		// Verify owner is correct
		if album.OwnerId != user.ID {
			t.Errorf("Album %s has wrong owner", album.Name)
		}
	}

	if !albumNames["Vacation 2025"] {
		t.Error("Expected 'Vacation 2025' album in list")
	}
	if !albumNames["Family Photos"] {
		t.Error("Expected 'Family Photos' album in list")
	}

	// Note: media_count and cover_thumbnail_url would be verified here
	// once those features are fully implemented in the service
}

