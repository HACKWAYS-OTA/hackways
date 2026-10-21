package auth

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const (
	googleProvider         = "google"
	googleNonceCookie      = "__Host-hackways_google_nonce"
	googleNonceCookieDev   = "hackways_google_nonce"
	googleStateLifetime    = 10 * time.Minute
	googleExchangeLifetime = 2 * time.Minute
)

// GoogleOAuthConfig contains only non-secret routing/configuration values plus
// the protected client secret loaded from the runtime environment.
type GoogleOAuthConfig struct {
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	ReturnOrigins []string
	IsProduction  bool
}

func (c GoogleOAuthConfig) enabled() bool {
	return strings.TrimSpace(c.ClientID) != "" && strings.TrimSpace(c.ClientSecret) != "" && strings.TrimSpace(c.RedirectURL) != ""
}

func (c GoogleOAuthConfig) oauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  c.RedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
	}
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// GoogleStart redirects the browser to Google's consent screen.
func (h *Handler) GoogleStart(w http.ResponseWriter, r *http.Request) {
	if !h.google.enabled() {
		platform.RespondError(w, http.StatusServiceUnavailable, "GOOGLE_AUTH_UNAVAILABLE", "Google sign-in is not configured", nil)
		return
	}

	returnTo := r.URL.Query().Get("return_to")
	if returnTo == "" {
		if len(h.google.ReturnOrigins) == 0 {
			platform.RespondError(w, http.StatusServiceUnavailable, "GOOGLE_AUTH_UNAVAILABLE", "Google sign-in has no allowed return destination", nil)
			return
		}
		returnTo = h.google.ReturnOrigins[0] + "/account"
	}
	if !h.google.validReturnTo(returnTo) {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_RETURN_URL", "The sign-in destination is not allowed", nil)
		return
	}

	state, err := GenerateToken()
	if err != nil {
		platform.RespondError(w, http.StatusInternalServerError, "OAUTH_STATE_ERROR", "Unable to start Google sign-in", nil)
		return
	}
	verifier, err := GenerateToken()
	if err != nil {
		platform.RespondError(w, http.StatusInternalServerError, "OAUTH_STATE_ERROR", "Unable to start Google sign-in", nil)
		return
	}
	nonce, err := GenerateToken()
	if err != nil {
		platform.RespondError(w, http.StatusInternalServerError, "OAUTH_STATE_ERROR", "Unable to start Google sign-in", nil)
		return
	}
	if err := h.queries.CreateOAuthState(r.Context(), db.CreateOAuthStateParams{
		StateHash: HashToken(state), Verifier: verifier, ReturnTo: returnTo,
		NonceHash: HashToken(nonce), ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(googleStateLifetime), Valid: true},
	}); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	h.setGoogleNonceCookie(w, nonce)

	challengeBytes := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(challengeBytes[:])
	authURL := h.google.oauthConfig().AuthCodeURL(state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// GoogleCallback validates Google's response and creates a one-time exchange
// code. The app host exchanges that code to receive its normal host-only cookie.
func (h *Handler) GoogleCallback(w http.ResponseWriter, r *http.Request) {
	if !h.google.enabled() {
		platform.RespondError(w, http.StatusServiceUnavailable, "GOOGLE_AUTH_UNAVAILABLE", "Google sign-in is not configured", nil)
		return
	}
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if state == "" {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_OAUTH_STATE", "The Google sign-in session expired or is invalid", nil)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	payload, err := h.queries.ConsumeOAuthState(ctx, HashToken(state))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			platform.RespondError(w, http.StatusBadRequest, "INVALID_OAUTH_STATE", "The Google sign-in session expired or is invalid", nil)
			return
		}
		platform.RespondDatabaseError(w, err)
		return
	}
	if providerErr := r.URL.Query().Get("error"); providerErr != "" {
		h.redirectGoogleError(w, r, payload.ReturnTo, "Google sign-in was cancelled")
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		platform.RespondError(w, http.StatusBadRequest, "MISSING_OAUTH_CODE", "Google did not return an authorization code", nil)
		return
	}

	token, err := h.google.oauthConfig().Exchange(ctx, code, oauth2.SetAuthURLParam("code_verifier", payload.Verifier))
	if err != nil {
		platform.RespondError(w, http.StatusBadRequest, "OAUTH_EXCHANGE_FAILED", "Google sign-in could not be completed", nil)
		return
	}
	info, err := fetchGoogleUserInfo(ctx, token)
	if err != nil || !info.VerifiedEmail || info.Sub == "" || !validEmail(info.Email) {
		platform.RespondError(w, http.StatusUnauthorized, "GOOGLE_ACCOUNT_UNVERIFIED", "Google did not provide a verified email account", nil)
		return
	}

	userID, err := h.upsertGoogleIdentity(ctx, info)
	if err != nil {
		if errors.Is(err, errIdentityConflict) {
			h.redirectGoogleError(w, r, payload.ReturnTo, "This Google account is already linked to another Hackways account")
			return
		}
		platform.RespondDatabaseError(w, err)
		return
	}
	exchange, err := GenerateToken()
	if err != nil {
		platform.RespondError(w, http.StatusInternalServerError, "OAUTH_EXCHANGE_FAILED", "Unable to finish Google sign-in", nil)
		return
	}
	if err := h.queries.CreateOAuthExchangeToken(ctx, db.CreateOAuthExchangeTokenParams{
		TokenHash: HashToken(exchange),
		UserID:    userID,
		NonceHash: payload.NonceHash,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(googleExchangeLifetime), Valid: true},
	}); err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	redirectURL, _ := url.Parse(payload.ReturnTo)
	query := redirectURL.Query()
	query.Set("google_exchange", exchange)
	redirectURL.RawQuery = query.Encode()
	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

type googleExchangeRequest struct {
	Code string `json:"code"`
}

// GoogleExchange consumes a one-time callback code and sets the existing
// host-only Hackways session cookie on the app that initiated sign-in.
func (h *Handler) GoogleExchange(w http.ResponseWriter, r *http.Request) {
	var req googleExchangeRequest
	if err := platform.DecodeJSON(r, &req); err != nil || strings.TrimSpace(req.Code) == "" {
		platform.RespondError(w, http.StatusBadRequest, "INVALID_EXCHANGE", "Invalid Google sign-in exchange", nil)
		return
	}
	nonceCookie, err := r.Cookie(h.googleNonceCookieName())
	if err != nil || nonceCookie.Value == "" {
		platform.RespondError(w, http.StatusUnauthorized, "INVALID_EXCHANGE", "The Google sign-in exchange is not bound to this browser", nil)
		return
	}
	userID, err := h.queries.ConsumeOAuthExchangeToken(r.Context(), db.ConsumeOAuthExchangeTokenParams{TokenHash: HashToken(strings.TrimSpace(req.Code)), NonceHash: HashToken(nonceCookie.Value)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			platform.RespondError(w, http.StatusUnauthorized, "INVALID_EXCHANGE", "The Google sign-in exchange expired or was already used", nil)
			return
		}
		platform.RespondDatabaseError(w, err)
		return
	}
	h.clearGoogleNonceCookie(w)
	token, _, err := h.sm.CreateSession(r.Context(), userID, r.UserAgent(), r.RemoteAddr, 7*24*time.Hour)
	if err != nil {
		platform.RespondError(w, http.StatusInternalServerError, "SESSION_ERROR", "Failed to establish session", nil)
		return
	}
	SetSessionCookie(w, DefaultSessionCookie, token, 7*24*time.Hour, h.isProduction)
	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		platform.RespondDatabaseError(w, err)
		return
	}
	platform.RespondJSON(w, http.StatusOK, map[string]interface{}{"user": UserResponse{
		ID: formatUUID(user.ID), Email: user.Email, DisplayName: user.DisplayName,
		AvatarURL: nullableText(user.AvatarUrl), EmailVerified: user.EmailVerified,
	}})
}

var errIdentityConflict = errors.New("google identity conflict")

func (h *Handler) upsertGoogleIdentity(ctx context.Context, info googleUserInfo) (pgtype.UUID, error) {
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return pgtype.UUID{}, err
	}
	defer tx.Rollback(ctx)
	q := h.queries.WithTx(tx)
	linked, err := q.GetUserByGoogleSubject(ctx, info.Sub)
	if err == nil {
		if !strings.EqualFold(linked.Email, info.Email) {
			return pgtype.UUID{}, errIdentityConflict
		}
		if err := tx.Commit(ctx); err != nil {
			return pgtype.UUID{}, err
		}
		return linked.ID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return pgtype.UUID{}, err
	}
	passwordMarker, err := GenerateToken()
	if err != nil {
		return pgtype.UUID{}, err
	}
	name := strings.TrimSpace(info.Name)
	if name == "" {
		name = strings.Split(info.Email, "@")[0]
	}
	avatar := pgtype.Text{String: strings.TrimSpace(info.Picture), Valid: strings.TrimSpace(info.Picture) != ""}
	user, err := q.GetOrCreateOAuthUser(ctx, db.GetOrCreateOAuthUserParams{
		Email: info.Email, PasswordHash: "google:" + passwordMarker, DisplayName: name, AvatarUrl: avatar,
	})
	if err != nil {
		return pgtype.UUID{}, err
	}
	if err := q.MarkUserEmailVerified(ctx, user.ID); err != nil {
		return pgtype.UUID{}, err
	}
	if err := q.LinkUserIdentity(ctx, db.LinkUserIdentityParams{UserID: user.ID, Provider: googleProvider, Subject: info.Sub}); err != nil {
		return pgtype.UUID{}, err
	}
	linked, err = q.GetUserByGoogleSubject(ctx, info.Sub)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgtype.UUID{}, errIdentityConflict
		}
		return pgtype.UUID{}, err
	}
	if !strings.EqualFold(linked.Email, info.Email) {
		return pgtype.UUID{}, errIdentityConflict
	}
	if err := tx.Commit(ctx); err != nil {
		return pgtype.UUID{}, err
	}
	return linked.ID, nil
}

func fetchGoogleUserInfo(ctx context.Context, token *oauth2.Token) (googleUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return googleUserInfo{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return googleUserInfo{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return googleUserInfo{}, fmt.Errorf("google userinfo returned %s", resp.Status)
	}
	var info googleUserInfo
	if err := json.NewDecoder(io.LimitReader(resp.Body, 32*1024)).Decode(&info); err != nil {
		return googleUserInfo{}, err
	}
	info.Email = strings.ToLower(strings.TrimSpace(info.Email))
	return info, nil
}

func validEmail(value string) bool {
	return strings.Contains(value, "@") && len(value) <= 320 && !strings.ContainsAny(value, "\r\n")
}

func (h *Handler) googleNonceCookieName() string {
	if h.isProduction {
		return googleNonceCookie
	}
	return googleNonceCookieDev
}

func (h *Handler) setGoogleNonceCookie(w http.ResponseWriter, nonce string) {
	http.SetCookie(w, &http.Cookie{Name: h.googleNonceCookieName(), Value: nonce, Path: "/", HttpOnly: true, Secure: h.isProduction, SameSite: http.SameSiteLaxMode, MaxAge: int(googleStateLifetime.Seconds()), Expires: time.Now().Add(googleStateLifetime)})
}

func (h *Handler) clearGoogleNonceCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: h.googleNonceCookieName(), Value: "", Path: "/", HttpOnly: true, Secure: h.isProduction, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(0, 0)})
}

func (c GoogleOAuthConfig) validReturnTo(value string) bool {
	u, err := url.Parse(value)
	if err != nil || u.Scheme == "" || u.Hostname() == "" || u.User != nil || u.Fragment != "" {
		return false
	}
	for _, origin := range c.ReturnOrigins {
		o, err := url.Parse(origin)
		if err != nil {
			continue
		}
		if strings.EqualFold(o.Scheme, u.Scheme) && strings.EqualFold(o.Host, u.Host) {
			return true
		}
	}
	return false
}

func (h *Handler) redirectGoogleError(w http.ResponseWriter, r *http.Request, returnTo, message string) {
	u, err := url.Parse(returnTo)
	if err != nil || !h.google.validReturnTo(returnTo) {
		platform.RespondError(w, http.StatusBadRequest, "OAUTH_CANCELLED", message, nil)
		return
	}
	q := u.Query()
	q.Set("google_error", message)
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

func nullableText(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
