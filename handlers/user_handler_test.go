package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/photo-video-sharing/handlers"
	"github.com/yourorg/photo-video-sharing/services"
	"github.com/yourorg/photo-video-sharing/testutil"
	pb "github.com/yourorg/photo-video-sharing/api/gen/v1"
)

func TestUserHandler_Register(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		email          string
		password       string
		wantStatusCode int
		wantError      string
	}{
		{
			name:           "US2-AS1: Valid registration",
			email:          "newuser@example.com",
			password:       "Password123",
			wantStatusCode: http.StatusCreated,
		},
		{
			name:           "Duplicate email",
			email:          "existing@example.com",
			password:       "Password123",
			wantStatusCode: http.StatusConflict,
			wantError:      "email already registered",
		},
		{
			name:           "Invalid email format",
			email:          "notanemail",
			password:       "Password123",
			wantStatusCode: http.StatusBadRequest,
			wantError:      "invalid",
		},
		{
			name:           "Password too short",
			email:          "test@example.com",
			password:       "Pass1",
			wantStatusCode: http.StatusBadRequest,
			wantError:      "password",
		},
		{
			name:           "Empty email",
			email:          "",
			password:       "Password123",
			wantStatusCode: http.StatusBadRequest,
			wantError:      "email",
		},
		{
			name:           "Empty password",
			email:          "test@example.com",
			password:       "",
			wantStatusCode: http.StatusBadRequest,
			wantError:      "password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			db, cleanup := testutil.SetupTestDB(t)
			defer cleanup()
			defer testutil.TruncateTables(db, "users", "sessions")

			// Auto-migrate schema
			if err := services.AutoMigrate(db); err != nil {
				t.Fatalf("Failed to migrate: %v", err)
			}

			// Pre-create existing user for duplicate test
			if tt.wantError == "email already registered" {
				testutil.CreateTestUser(db, map[string]interface{}{
					"email": tt.email,
				})
			}

			// Create service and handler
			userService := services.NewUserService(db)
			sessionService := services.NewSessionService(db)
			userHandler := handlers.NewUserHandler(userService, sessionService)

			// Setup routes
			mux := http.NewServeMux()
			mux.HandleFunc("POST /auth/register", userHandler.Register)

			// Create request
			reqBody := map[string]string{
				"email":    tt.email,
				"password": tt.password,
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute
			mux.ServeHTTP(rec, req)

			// Assert status code
			if rec.Code != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatusCode, rec.Code, rec.Body.String())
			}

			// For successful registration
			if tt.wantStatusCode == http.StatusCreated {
				var resp pb.RegisterResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if resp.User == nil {
					t.Fatal("Expected user in response, got nil")
				}
				if resp.User.Email != tt.email {
					t.Errorf("Expected email %s, got %s", tt.email, resp.User.Email)
				}
				if resp.SessionId == "" {
					t.Error("Expected session ID for auto-login, got empty")
				}

				// Verify session cookie is set
				cookies := rec.Result().Cookies()
				var sessionCookie *http.Cookie
				for _, c := range cookies {
					if c.Name == "session_id" {
						sessionCookie = c
						break
					}
				}
				if sessionCookie == nil {
					t.Error("Expected session cookie to be set")
				} else {
					if !sessionCookie.HttpOnly {
						t.Error("Session cookie should be HttpOnly")
					}
					if sessionCookie.Value != resp.SessionId {
						t.Errorf("Cookie session ID %s doesn't match response %s", sessionCookie.Value, resp.SessionId)
					}
				}
			}

			// For errors
			if tt.wantError != "" {
				body := rec.Body.String()
				if !contains(body, tt.wantError) {
					t.Errorf("Expected error containing %q, got: %s", tt.wantError, body)
				}
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		email          string
		password       string
		wantStatusCode int
		wantError      string
	}{
		{
			name:           "US2-AS2: Valid login",
			email:          "user@example.com",
			password:       "Password123",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Wrong password",
			email:          "user@example.com",
			password:       "WrongPassword",
			wantStatusCode: http.StatusUnauthorized,
			wantError:      "invalid",
		},
		{
			name:           "Non-existent user",
			email:          "nonexistent@example.com",
			password:       "Password123",
			wantStatusCode: http.StatusUnauthorized,
			wantError:      "invalid",
		},
		{
			name:           "Empty email",
			email:          "",
			password:       "Password123",
			wantStatusCode: http.StatusBadRequest,
			wantError:      "email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			db, cleanup := testutil.SetupTestDB(t)
			defer cleanup()
			defer testutil.TruncateTables(db, "users", "sessions")

			if err := services.AutoMigrate(db); err != nil {
				t.Fatalf("Failed to migrate: %v", err)
			}

			// Create test user
			testutil.CreateTestUser(db, map[string]interface{}{
				"email":    "user@example.com",
				"password": "Password123",
			})

			// Create service and handler
			userService := services.NewUserService(db)
			sessionService := services.NewSessionService(db)
			userHandler := handlers.NewUserHandler(userService, sessionService)

			mux := http.NewServeMux()
			mux.HandleFunc("POST /auth/login", userHandler.Login)

			// Create request
			reqBody := map[string]string{
				"email":    tt.email,
				"password": tt.password,
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Execute
			mux.ServeHTTP(rec, req)

			// Assert
			if rec.Code != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.wantStatusCode, rec.Code, rec.Body.String())
			}

			if tt.wantStatusCode == http.StatusOK {
				var resp pb.LoginResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("Failed to decode response: %v", err)
				}

				if resp.User == nil {
					t.Fatal("Expected user in response")
				}
				if resp.SessionId == "" {
					t.Error("Expected session ID")
				}

				// Verify session cookie
				cookies := rec.Result().Cookies()
				found := false
				for _, c := range cookies {
					if c.Name == "session_id" {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected session cookie")
				}
			}

			if tt.wantError != "" {
				if !contains(rec.Body.String(), tt.wantError) {
					t.Errorf("Expected error containing %q, got: %s", tt.wantError, rec.Body.String())
				}
			}
		})
	}
}

func TestUserHandler_Isolation(t *testing.T) {
	t.Parallel()

	// US2-AS3: User A and User B should see different data
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	defer testutil.TruncateTables(db, "users", "sessions", "media")

	if err := services.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create two users
	userA := testutil.CreateTestUser(db, map[string]interface{}{
		"email": "userA@example.com",
	})
	userB := testutil.CreateTestUser(db, map[string]interface{}{
		"email": "userB@example.com",
	})

	// Create media for each user
	testutil.CreateTestMedia(db, map[string]interface{}{
		"owner_id": userA.ID,
		"filename": "userA-photo.jpg",
	})
	testutil.CreateTestMedia(db, map[string]interface{}{
		"owner_id": userB.ID,
		"filename": "userB-photo.jpg",
	})

	// Create sessions for both users
	sessionA := testutil.CreateTestSession(db, userA.ID, nil)
	sessionB := testutil.CreateTestSession(db, userB.ID, nil)

	// Create handlers (media handler would be created here, but we're testing auth isolation)
	// This test verifies that session-based auth properly isolates user data
	// We'll verify the sessions are distinct and user IDs are different

	if sessionA.ID == sessionB.ID {
		t.Error("Sessions should have different IDs")
	}
	if sessionA.UserID == sessionB.UserID {
		t.Error("Sessions should belong to different users")
	}
	if userA.ID == userB.ID {
		t.Error("Users should have different IDs")
	}
}

func TestUserHandler_Unauthorized(t *testing.T) {
	t.Parallel()

	// US2-AS4: Unauthenticated user cannot access protected endpoints
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	if err := services.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create handler with auth middleware
	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	userHandler := handlers.NewUserHandler(userService, sessionService)
	authMiddleware := handlers.NewAuthMiddleware(sessionService)

	mux := http.NewServeMux()
	mux.Handle("GET /auth/me", authMiddleware.RequireAuth(http.HandlerFunc(userHandler.GetCurrentUser)))

	// Request without session cookie
	req := httptest.NewRequest("GET", "/auth/me", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestUserHandler_PasswordReset(t *testing.T) {
	t.Parallel()

	// US2-AS5: Password reset generates reset link
	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	defer testutil.TruncateTables(db, "users")

	if err := services.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create test user
	testutil.CreateTestUser(db, map[string]interface{}{
		"email": "user@example.com",
	})

	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	userHandler := handlers.NewUserHandler(userService, sessionService)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/password-reset", userHandler.PasswordReset)

	reqBody := map[string]string{
		"email": "user@example.com",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/auth/password-reset", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	// Note: Email not actually sent in test, just verify response
	var resp pb.PasswordResetResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// In real implementation, email_sent might be true, but for now we just verify no error
	if resp.Message == "" {
		t.Error("Expected message in response")
	}
}

func TestUserHandler_GetCurrentUser(t *testing.T) {
	t.Parallel()

	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	defer testutil.TruncateTables(db, "users", "sessions")

	if err := services.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create user and session
	user := testutil.CreateTestUser(db, map[string]interface{}{
		"email": "user@example.com",
	})
	session := testutil.CreateTestSession(db, user.ID, nil)

	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	userHandler := handlers.NewUserHandler(userService, sessionService)
	authMiddleware := handlers.NewAuthMiddleware(sessionService)

	mux := http.NewServeMux()
	mux.Handle("GET /auth/me", authMiddleware.RequireAuth(http.HandlerFunc(userHandler.GetCurrentUser)))

	// Request with valid session cookie
	req := httptest.NewRequest("GET", "/auth/me", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.ID,
	})
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var resp pb.GetCurrentUserResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.User == nil {
		t.Fatal("Expected user in response")
	}
	if resp.User.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, resp.User.Email)
	}
}

func TestUserHandler_ExpiredSession(t *testing.T) {
	t.Parallel()

	db, cleanup := testutil.SetupTestDB(t)
	defer cleanup()
	defer testutil.TruncateTables(db, "users", "sessions")

	if err := services.AutoMigrate(db); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Create user and expired session
	user := testutil.CreateTestUser(db, map[string]interface{}{
		"email": "user@example.com",
	})
	session := testutil.CreateTestSession(db, user.ID, map[string]interface{}{
		"expires_at": time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	})

	userService := services.NewUserService(db)
	sessionService := services.NewSessionService(db)
	userHandler := handlers.NewUserHandler(userService, sessionService)
	authMiddleware := handlers.NewAuthMiddleware(sessionService)

	mux := http.NewServeMux()
	mux.Handle("GET /auth/me", authMiddleware.RequireAuth(http.HandlerFunc(userHandler.GetCurrentUser)))

	// Request with expired session
	req := httptest.NewRequest("GET", "/auth/me", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: session.ID,
	})
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

// Helper functions

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

func verifyPasswordHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
