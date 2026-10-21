package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler serves HTTP endpoints for authentication and session management.
type Handler struct {
	pool         *pgxpool.Pool
	queries      *db.Queries
	sm           *SessionManager
	isProduction bool
	google       GoogleOAuthConfig
}

// NewHandler constructs an Auth Handler.
func NewHandler(pool *pgxpool.Pool, isProduction bool, googleConfig ...GoogleOAuthConfig) *Handler {
	queries := db.New(pool)
	var google GoogleOAuthConfig
	if len(googleConfig) > 0 {
		google = googleConfig[0]
	}
	return &Handler{
		pool:         pool,
		queries:      queries,
		sm:           NewSessionManager(queries),
		isProduction: isProduction,
		google:       google,
	}
}

// SessionManager returns the underlying session manager.
func (h *Handler) SessionManager() *SessionManager {
	return h.sm
}

// Routes mounts authentication endpoints.
func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Post("/verify", h.Verify)
	r.Get("/google/start", h.GoogleStart)
	r.Get("/google/callback", h.GoogleCallback)
	r.Post("/google/exchange", h.GoogleExchange)

	// Protected routes
	r.Group(func(protected chi.Router) {
		protected.Use(RequireAuth(h.sm))
		protected.Get("/me", h.Me)
		protected.Get("/sessions", h.Sessions)
		protected.Delete("/sessions", h.RevokeAllSessions)
		protected.Delete("/sessions/{sessionID}", h.RevokeSession)
		protected.Get("/notifications", h.Notifications)
		protected.Post("/notifications/{notificationID}/retry", h.RetryNotification)
		protected.Post("/email-verification/request", h.RequestEmailVerification)
		protected.Post("/email-verification/confirm", h.ConfirmEmailVerification)
	})

	return r
}

type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type UserResponse struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	DisplayName   string     `json:"display_name"`
	AvatarURL     *string    `json:"avatar_url"`
	EmailVerified bool       `json:"email_verified"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
}

// Register creates a new user account and starts a session.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := platform.DecodeJSON(r, &req); err != nil {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON body", nil)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.DisplayName = strings.TrimSpace(req.DisplayName)

	if req.Email == "" {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email is required", nil)
		return
	}
	if address, err := mail.ParseAddress(req.Email); err != nil || address.Address != req.Email {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid email format", nil)
		return
	}
	if len(req.Password) < 8 {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Password must be at least 8 characters", nil)
		return
	}
	if req.DisplayName == "" {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Display name is required", nil)
		return
	}

	if !h.allowIdentityAttempt(w, r, req.Email) {
		return
	}
	if len(req.Password) > 72 || len(req.Email) > 320 || len([]rune(req.DisplayName)) > 100 {
		platform.RespondError(w, 400, "INVALID_DETAILS", "Account fields exceed their limits", nil)
		return
	}
	hash, err := HashPassword(req.Password)
	if err != nil {
		platform.RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to secure password", nil)
		return
	}

	user, err := h.queries.CreateUser(r.Context(), db.CreateUserParams{
		Email:        req.Email,
		PasswordHash: hash,
		DisplayName:  req.DisplayName,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			platform.RespondError(w, http.StatusConflict, "EMAIL_EXISTS", "Email is already registered", nil)
			return
		}
		platform.RespondDatabaseError(w, err)
		return
	}

	// Create session
	token, _, err := h.sm.CreateSession(r.Context(), user.ID, r.UserAgent(), r.RemoteAddr, 7*24*time.Hour)
	if err != nil {
		platform.RespondError(w, http.StatusInternalServerError, "SESSION_ERROR", "Failed to establish session", nil)
		return
	}

	SetSessionCookie(w, DefaultSessionCookie, token, 7*24*time.Hour, h.isProduction)

	userIDStr := formatUUID(user.ID)
	createdAt := user.CreatedAt.Time
	var avatarURL *string
	if user.AvatarUrl.Valid {
		avatarURL = &user.AvatarUrl.String
	}

	platform.RespondJSON(w, http.StatusCreated, map[string]interface{}{
		"user": UserResponse{
			ID:            userIDStr,
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			AvatarURL:     avatarURL,
			EmailVerified: user.EmailVerified,
			CreatedAt:     &createdAt,
		},
	})
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates credentials and establishes an active session.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := platform.DecodeJSON(r, &req); err != nil {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Malformed JSON body", nil)
		return
	}

	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	if req.Email == "" || req.Password == "" {
		platform.RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Email and password are required", nil)
		return
	}

	if !h.allowIdentityAttempt(w, r, req.Email) {
		return
	}
	if len(req.Password) > 72 || len(req.Email) > 320 {
		platform.RespondError(w, 400, "INVALID_DETAILS", "Account fields exceed their limits", nil)
		return
	}
	user, err := h.queries.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			platform.RespondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password", nil)
			return
		}
		platform.RespondDatabaseError(w, err)
		return
	}

	if !CheckPasswordHash(req.Password, user.PasswordHash) {
		platform.RespondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password", nil)
		return
	}

	token, _, err := h.sm.CreateSession(r.Context(), user.ID, r.UserAgent(), r.RemoteAddr, 7*24*time.Hour)
	if err != nil {
		platform.RespondError(w, http.StatusInternalServerError, "SESSION_ERROR", "Failed to create session", nil)
		return
	}

	SetSessionCookie(w, DefaultSessionCookie, token, 7*24*time.Hour, h.isProduction)

	userIDStr := formatUUID(user.ID)
	createdAt := user.CreatedAt.Time
	var avatarURL *string
	if user.AvatarUrl.Valid {
		avatarURL = &user.AvatarUrl.String
	}

	platform.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"user": UserResponse{
			ID:            userIDStr,
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			AvatarURL:     avatarURL,
			EmailVerified: user.EmailVerified,
			CreatedAt:     &createdAt,
		},
	})
}

// Logout invalidates the active session and clears the session cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := GetSessionToken(r, DefaultSessionCookie)
	if token == "" {
		token = GetSessionToken(r, AdminSessionCookie)
	}

	if token != "" {
		if err := h.sm.RevokeSession(r.Context(), token); err != nil {
			platform.RespondDatabaseError(w, err)
			return
		}
	}

	ClearSessionCookie(w, DefaultSessionCookie)
	ClearSessionCookie(w, AdminSessionCookie)
	platform.RespondJSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// Me returns the authenticated user from session context.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok || user == nil {
		platform.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Not authenticated", nil)
		return
	}

	platform.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"user": UserResponse{
			ID:            formatUUID(user.UserID),
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			AvatarURL:     user.AvatarURL,
			EmailVerified: user.EmailVerified,
		},
	})
}

type VerifyRequest struct {
	Token string `json:"token"`
}

// Verify validates a session token (via cookie or request body) and returns the authenticated user.
func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	token := GetSessionToken(r, DefaultSessionCookie)
	if token == "" {
		token = GetSessionToken(r, AdminSessionCookie)
	}

	if token == "" && r.Body != nil {
		var req VerifyRequest
		if err := platform.DecodeJSON(r, &req); err == nil && req.Token != "" {
			token = req.Token
		}
	}

	if token == "" {
		platform.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "No session token provided", nil)
		return
	}

	user, err := h.sm.GetSession(r.Context(), token)
	if err != nil {
		if !errors.Is(err, ErrSessionNotFound) && !errors.Is(err, ErrInvalidToken) {
			platform.RespondDatabaseError(w, err)
			return
		}
		platform.RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Session expired or invalid", nil)
		return
	}

	platform.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"valid": true,
		"user": UserResponse{
			ID:            formatUUID(user.UserID),
			Email:         user.Email,
			DisplayName:   user.DisplayName,
			AvatarURL:     user.AvatarURL,
			EmailVerified: user.EmailVerified,
		},
	})
}

func formatUUID(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	src := id.Bytes
	return string([]byte{
		hexByte(src[0] >> 4), hexByte(src[0] & 0xf),
		hexByte(src[1] >> 4), hexByte(src[1] & 0xf),
		hexByte(src[2] >> 4), hexByte(src[2] & 0xf),
		hexByte(src[3] >> 4), hexByte(src[3] & 0xf),
		'-',
		hexByte(src[4] >> 4), hexByte(src[4] & 0xf),
		hexByte(src[5] >> 4), hexByte(src[5] & 0xf),
		'-',
		hexByte(src[6] >> 4), hexByte(src[6] & 0xf),
		hexByte(src[7] >> 4), hexByte(src[7] & 0xf),
		'-',
		hexByte(src[8] >> 4), hexByte(src[8] & 0xf),
		hexByte(src[9] >> 4), hexByte(src[9] & 0xf),
		'-',
		hexByte(src[10] >> 4), hexByte(src[10] & 0xf),
		hexByte(src[11] >> 4), hexByte(src[11] & 0xf),
		hexByte(src[12] >> 4), hexByte(src[12] & 0xf),
		hexByte(src[13] >> 4), hexByte(src[13] & 0xf),
		hexByte(src[14] >> 4), hexByte(src[14] & 0xf),
		hexByte(src[15] >> 4), hexByte(src[15] & 0xf),
	})
}

func hexByte(b byte) byte {
	if b < 10 {
		return '0' + b
	}
	return 'a' + (b - 10)
}

// Nginx enforces per-IP limits; this persisted per-identifier limit also bounds
// distributed credential guessing without trusting forwarded client headers.
func (h *Handler) allowIdentityAttempt(w http.ResponseWriter, r *http.Request, email string) bool {
	digest := sha256.Sum256([]byte("identity:" + strings.ToLower(strings.TrimSpace(email))))
	attempts, err := h.queries.CountAuthAttempt(r.Context(), hex.EncodeToString(digest[:]))
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return false
	}
	if attempts > 10 {
		w.Header().Set("Retry-After", "60")
		platform.RespondError(w, 429, "AUTH_RATE_LIMITED", "Too many attempts; retry in a minute", nil)
		return false
	}
	return true
}
