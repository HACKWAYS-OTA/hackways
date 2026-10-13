package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrSessionNotFound = errors.New("session not found or expired")
	ErrInvalidToken    = errors.New("invalid session token")
)

// SessionUser encapsulates session data and authenticated user identity.
type SessionUser struct {
	SessionID     pgtype.UUID
	UserID        pgtype.UUID
	Email         string
	DisplayName   string
	AvatarURL     *string
	EmailVerified bool
	ExpiresAt     time.Time
}

// SessionManager manages server-side PostgreSQL sessions.
type SessionManager struct {
	queries *db.Queries
}

// NewSessionManager creates a SessionManager.
func NewSessionManager(queries *db.Queries) *SessionManager {
	return &SessionManager{queries: queries}
}

// GenerateToken generates a cryptographically secure 32-byte random hex token.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashToken computes a SHA-256 hash of the session token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// CreateSession generates a new session token, hashes it, and stores the session record in PostgreSQL.
func (sm *SessionManager) CreateSession(ctx context.Context, userID pgtype.UUID, userAgent, ipStr string, duration time.Duration) (string, *db.Sessions, error) {
	token, err := GenerateToken()
	if err != nil {
		return "", nil, err
	}

	tokenHash := HashToken(token)
	expiresAt := time.Now().Add(duration)

	var ipAddr *netip.Addr
	if ipStr != "" {
		if addr, err := netip.ParseAddr(ipStr); err == nil {
			ipAddr = &addr
		}
	}

	params := db.CreateSessionParams{
		UserID:    userID,
		TokenHash: tokenHash,
		UserAgent: pgtype.Text{String: userAgent, Valid: userAgent != ""},
		IpAddress: ipAddr,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	}

	session, err := sm.queries.CreateSession(ctx, params)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create session record: %w", err)
	}

	return token, &session, nil
}

// GetSession validates a plaintext token by its SHA-256 hash and retrieves user identity.
func (sm *SessionManager) GetSession(ctx context.Context, token string) (*SessionUser, error) {
	if token == "" {
		return nil, ErrInvalidToken
	}

	tokenHash := HashToken(token)
	row, err := sm.queries.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	// Update last_seen asynchronously in background or inline
	_ = sm.queries.UpdateSessionLastSeen(ctx, row.ID)

	var avatarURL *string
	if row.AvatarUrl.Valid {
		avatarURL = &row.AvatarUrl.String
	}

	return &SessionUser{
		SessionID:     row.ID,
		UserID:        row.UserID,
		Email:         row.Email,
		DisplayName:   row.DisplayName,
		AvatarURL:     avatarURL,
		EmailVerified: row.EmailVerified,
		ExpiresAt:     row.ExpiresAt.Time,
	}, nil
}

// RevokeSession revokes an active session by deleting its record by token hash.
func (sm *SessionManager) RevokeSession(ctx context.Context, token string) error {
	if token == "" {
		return ErrInvalidToken
	}
	tokenHash := HashToken(token)
	return sm.queries.DeleteSessionByTokenHash(ctx, tokenHash)
}

// RevokeAllUserSessions revokes all sessions associated with a specific user.
func (sm *SessionManager) RevokeAllUserSessions(ctx context.Context, userID pgtype.UUID) error {
	return sm.queries.DeleteSessionsByUserID(ctx, userID)
}
