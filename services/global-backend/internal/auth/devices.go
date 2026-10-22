package auth

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/hackways/global-backend/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const deviceCookie = "__Host-hackways_device"
const adminDeviceCookie = "__Host-hackways_admin_device"
const deviceLifetime = 180 * 24 * time.Hour

var errCredentialsChanged = errors.New("credentials changed during authentication")

func deviceName(admin bool) string {
	if admin {
		return adminDeviceCookie
	}
	return deviceCookie
}
func setDeviceCookie(w http.ResponseWriter, token string, admin bool) {
	sameSite := http.SameSiteLaxMode
	if admin {
		sameSite = http.SameSiteStrictMode
	}
	http.SetCookie(w, &http.Cookie{Name: deviceName(admin), Value: token, Path: "/", Secure: true, HttpOnly: true, SameSite: sameSite, MaxAge: int(deviceLifetime.Seconds()), Expires: time.Now().Add(deviceLifetime)})
}

// The caller holds the account row lock. Recognition only deduplicates alerts;
// it never grants a session, role or authentication bypass.
func rememberDevice(ctx context.Context, q *db.Queries, user db.LockAuthenticationUserRow, r *http.Request, admin bool, locationProxies []netip.Prefix) (string, error) {
	token := GetSessionToken(r, deviceName(admin))
	decoded, err := hex.DecodeString(token)
	if err != nil || len(decoded) != 32 {
		token, err = GenerateToken()
		if err != nil {
			return "", err
		}
	}
	hash := HashToken(token)
	known, err := q.IsRecognizedDevice(ctx, db.IsRecognizedDeviceParams{UserID: user.ID, TokenHash: hash})
	if err != nil {
		return "", err
	}
	agent := []rune(r.UserAgent())
	if len(agent) > 512 {
		agent = agent[:512]
	}
	if err = q.RememberRecognizedDevice(ctx, db.RememberRecognizedDeviceParams{UserID: user.ID, TokenHash: hash, UserAgent: string(agent)}); err != nil {
		return "", err
	}
	if !user.DeviceBaselineAt.Valid {
		// Existing accounts have no historical recognition cookie. Establish their
		// first baseline without falsely calling that regular browser a new device.
		if err = q.MarkDeviceBaseline(ctx, user.ID); err != nil {
			return "", err
		}
	} else if !known && user.EmailVerified {
		details := describeDevice(string(agent))
		payload, marshalErr := json.Marshal(struct {
			Device   string    `json:"device_name"`
			Browser  string    `json:"browser_name"`
			Label    string    `json:"device_label"`
			Location string    `json:"approx_location"`
			At       time.Time `json:"signed_in_at"`
			Action   string    `json:"action_url"`
		}{details.device, details.browser, browserLabel(string(agent)), approximateLocation(r, locationProxies), time.Now().UTC(), "https://hackways.in/account"})
		if marshalErr != nil {
			return "", marshalErr
		}
		if err = enqueueAccountEmail(ctx, q, user.Email, user.DisplayName, "sign_in_alert", payload); err != nil {
			return "", err
		}
	}
	if err = q.BoundRecognizedDevices(ctx, user.ID); err != nil {
		return "", err
	}
	return token, nil
}

func browserLabel(agent string) string {
	browser := "Browser"
	os := ""
	switch {
	case strings.Contains(agent, "Edg/"):
		browser = "Microsoft Edge"
	case strings.Contains(agent, "Firefox/"):
		browser = "Firefox"
	case strings.Contains(agent, "Chrome/"):
		browser = "Chrome"
	case strings.Contains(agent, "Safari/"):
		browser = "Safari"
	}
	switch {
	case strings.Contains(agent, "Android"):
		os = "Android"
	case strings.Contains(agent, "iPhone") || strings.Contains(agent, "iPad"):
		os = "iOS"
	case strings.Contains(agent, "Windows"):
		os = "Windows"
	case strings.Contains(agent, "Macintosh"):
		os = "macOS"
	case strings.Contains(agent, "Linux"):
		os = "Linux"
	}
	if os != "" {
		return browser + " on " + os
	}
	return browser
}

func (h *Handler) establishSession(w http.ResponseWriter, r *http.Request, id pgtype.UUID, expectedPassword string, admin bool, proof ...db.ConsumeOAuthExchangeTokenParams) error {
	tx, err := h.pool.Begin(r.Context())
	if err != nil {
		return err
	}
	defer tx.Rollback(r.Context())
	q := h.queries.WithTx(tx)
	if err = q.SetStorageWriteTimeout(r.Context()); err != nil {
		return err
	}
	user, err := q.LockAuthenticationUser(r.Context(), id)
	if err != nil {
		return err
	}
	if expectedPassword != "" && user.PasswordHash != expectedPassword {
		return errCredentialsChanged
	}
	if len(proof) > 0 {
		exchanged, e := q.ConsumeOAuthExchangeToken(r.Context(), proof[0])
		if e != nil {
			if errors.Is(e, pgx.ErrNoRows) {
				return errCredentialsChanged
			}
			return e
		}
		if exchanged != id {
			return errCredentialsChanged
		}
	}
	if admin {
		allowed, e := q.IsPlatformAdministrator(r.Context(), id)
		if e != nil {
			return e
		}
		if !allowed {
			return errAdminAccess
		}
	}
	duration := 7 * 24 * time.Hour
	name := DefaultSessionCookie
	if admin {
		duration = 30 * time.Minute
		name = AdminSessionCookie
	}
	token, session, err := NewSessionManager(q).CreateSession(r.Context(), id, r.UserAgent(), r.RemoteAddr, duration)
	if err != nil {
		return err
	}
	if admin {
		if err = q.RecordPlatformAdminSession(r.Context(), session.ID); err != nil {
			return err
		}
	}
	device, err := rememberDevice(r.Context(), q, user, r, admin, h.locationProxies)
	if err != nil {
		return err
	}
	if err = tx.Commit(r.Context()); err != nil {
		return err
	}
	SetSessionCookie(w, name, token, duration, h.isProduction)
	setDeviceCookie(w, device, admin)
	return nil
}

var errAdminAccess = errors.New("admin access required")

// Browser APIs do not disclose a computer's hostname. Device names are coarse
// user-agent categories; location is optional trusted ingress metadata, not GPS.
type deviceDescription struct{ device, browser string }

func describeDevice(agent string) deviceDescription {
	result := deviceDescription{device: "Unidentified device", browser: "Unidentified browser"}
	switch {
	case strings.Contains(agent, "iPad"):
		result.device = "iPad"
	case strings.Contains(agent, "iPhone"):
		result.device = "iPhone"
	case strings.Contains(agent, "Android"):
		result.device = "Android device"
	case strings.Contains(agent, "Windows"):
		result.device = "Windows PC"
	case strings.Contains(agent, "Macintosh"):
		result.device = "Mac"
	case strings.Contains(agent, "Linux"):
		result.device = "Linux computer"
	}
	for _, marker := range []struct{ key, name string }{{"Edg/", "Microsoft Edge"}, {"EdgiOS/", "Microsoft Edge"}, {"EdgA/", "Microsoft Edge"}, {"OPR/", "Opera"}, {"FxiOS/", "Firefox"}, {"Firefox/", "Firefox"}, {"CriOS/", "Chrome"}, {"Chrome/", "Chrome"}, {"Version/", "Safari"}} {
		if at := strings.Index(agent, marker.key); at >= 0 {
			version := strings.TrimLeft(agent[at+len(marker.key):], " ")
			end := 0
			for end < len(version) && (version[end] >= '0' && version[end] <= '9' || version[end] == '.') {
				end++
			}
			result.browser = marker.name
			if end > 0 && end <= 32 {
				result.browser += " " + version[:end]
			}
			break
		}
	}
	return result
}

func (h *Handler) SetLocationProxies(prefixes []netip.Prefix) {
	h.locationProxies = append([]netip.Prefix(nil), prefixes...)
}
func approximateLocation(r *http.Request, proxies []netip.Prefix) string {
	peer, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return "Unavailable"
	}
	trusted := false
	for _, prefix := range proxies {
		if prefix.Contains(peer.Addr().Unmap()) {
			trusted = true
			break
		}
	}
	if !trusted {
		return "Unavailable"
	}
	clean := func(raw string) string {
		value, err := url.QueryUnescape(raw)
		if err != nil || len(value) > 200 {
			return ""
		}
		for _, c := range value {
			if !unicode.IsLetter(c) && !unicode.IsNumber(c) && !strings.ContainsRune(" .,-'", c) {
				return ""
			}
		}
		return strings.TrimSpace(value)
	}
	country := clean(r.Header.Get("X-Hackways-Country"))
	if len(country) != 2 || country == "XX" || country == "T1" {
		return "Unavailable"
	}
	parts := []string{}
	for _, header := range []string{"X-Hackways-City", "X-Hackways-Region"} {
		if value := clean(r.Header.Get(header)); value != "" && (len(parts) == 0 || parts[len(parts)-1] != value) {
			parts = append(parts, value)
		}
	}
	parts = append(parts, country)
	return strings.Join(parts, ", ")
}
