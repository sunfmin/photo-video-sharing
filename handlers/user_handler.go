package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/opentracing/opentracing-go"
	
	"github.com/yourorg/photo-video-sharing/services"
	pb "github.com/yourorg/photo-video-sharing/api/gen/v1"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService    services.UserService
	sessionService services.SessionService
	tracer         opentracing.Tracer
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService services.UserService, sessionService services.SessionService) *UserHandler {
	return &UserHandler{
		userService:    userService,
		sessionService: sessionService,
		tracer:         opentracing.NoopTracer{}, // Default to noop
	}
}

// WithTracer adds OpenTracing support
func (h *UserHandler) WithTracer(tracer opentracing.Tracer) *UserHandler {
	h.tracer = tracer
	return h
}

// Register handles user registration
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Principle XI: Create OpenTracing span for HTTP endpoint
	span := h.tracer.StartSpan("POST /auth/register")
	defer span.Finish()
	span.SetTag("http.method", r.Method)
	span.SetTag("http.url", r.URL.Path)
	
	ctx := r.Context()

	var req pb.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusBadRequest)
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusBadRequest)
		WriteError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Password == "" {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusBadRequest)
		WriteError(w, http.StatusBadRequest, "password is required")
		return
	}
	if len(req.Password) < 8 {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusBadRequest)
		WriteError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	// Basic email validation
	if !contains(req.Email, "@") {
		span.SetTag("error", true)
		span.SetTag("http.status_code", http.StatusBadRequest)
		WriteError(w, http.StatusBadRequest, "invalid email format")
		return
	}

	// Register user
	user, err := h.userService.Register(ctx, &req)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	// Create session for auto-login
	sessionID, err := h.sessionService.CreateSession(ctx, user.Id)
	if err != nil {
		span.SetTag("error", true)
		HandleServiceError(w, err)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	// Send response
	resp := &pb.RegisterResponse{
		User:      user,
		SessionId: sessionID,
	}
	span.SetTag("http.status_code", http.StatusCreated)
	WriteJSON(w, http.StatusCreated, resp)
}

// Login handles user login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req pb.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate input
	if req.Email == "" {
		WriteError(w, http.StatusBadRequest, "email is required")
		return
	}
	if req.Password == "" {
		WriteError(w, http.StatusBadRequest, "password is required")
		return
	}

	// Authenticate user
	user, err := h.userService.Login(r.Context(), &req)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	// Create session
	sessionID, err := h.sessionService.CreateSession(r.Context(), user.Id)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	// Send response
	resp := &pb.LoginResponse{
		User:      user,
		SessionId: sessionID,
	}
	WriteJSON(w, http.StatusOK, resp)
}

// Logout handles user logout
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Get session from cookie
	cookie, err := r.Cookie("session_id")
	if err != nil {
		WriteError(w, http.StatusBadRequest, "No session found")
		return
	}

	// Delete session
	if err := h.sessionService.DeleteSession(r.Context(), cookie.Value); err != nil {
		HandleServiceError(w, err)
		return
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // Delete cookie
	})

	resp := &pb.LogoutResponse{
		Success: true,
	}
	WriteJSON(w, http.StatusOK, resp)
}

// GetCurrentUser returns the current authenticated user
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// User ID is set by auth middleware in context
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	user, err := h.userService.GetCurrentUser(r.Context(), userID)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	resp := &pb.GetCurrentUserResponse{
		User: user,
	}
	WriteJSON(w, http.StatusOK, resp)
}

// PasswordReset handles password reset requests
func (h *UserHandler) PasswordReset(w http.ResponseWriter, r *http.Request) {
	var req pb.PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" {
		WriteError(w, http.StatusBadRequest, "email is required")
		return
	}

	resp, err := h.userService.PasswordReset(r.Context(), &req)
	if err != nil {
		HandleServiceError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, resp)
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && jsonContains(s, substr)
}

func jsonContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

