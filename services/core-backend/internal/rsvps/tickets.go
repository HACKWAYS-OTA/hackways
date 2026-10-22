package rsvps

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/skip2/go-qrcode"
)

type TicketClaims struct {
	TicketID  pgtype.UUID `json:"ticket_id"`
	EventID   pgtype.UUID `json:"event_id"`
	ExpiresAt int64       `json:"expires_at"`
}

func ticketKey() (ed25519.PrivateKey, error) {
	seed, err := base64.StdEncoding.DecodeString(os.Getenv("TICKET_SIGNING_SEED"))
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, errors.New("ticket signing seed is not configured")
	}
	return ed25519.NewKeyFromSeed(seed), nil
}
func (h *Handler) Ticket(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	var id pgtype.UUID
	if err := id.Scan(chi.URLParam(r, "id")); err != nil || !id.Valid {
		platform.RespondError(w, 400, "INVALID_ID", "Invalid registration ID", nil)
		return
	}
	ticket, err := h.queries.GetAccountTicket(r.Context(), db.GetAccountTicketParams{ID: id, UserID: user.UserID})
	if err != nil {
		respondFailure(w, err)
		return
	}
	event, err := h.queries.GetEventByID(r.Context(), ticket.EventID)
	if err != nil {
		respondFailure(w, err)
		return
	}
	if event.Status != db.EventStatusPublished {
		platform.RespondError(w, 409, "TICKET_UNAVAILABLE", "The event is not currently available", nil)
		return
	}
	key, err := ticketKey()
	if err != nil {
		platform.RespondError(w, 503, "TICKET_SIGNING_UNAVAILABLE", "Ticket signing is unavailable; your registration remains recorded", nil)
		return
	}
	claims := TicketClaims{ticket.ID, ticket.EventID, event.EndsAt.Time.Add(24 * time.Hour).Unix()}
	body, err := json.Marshal(claims)
	if err != nil {
		respondFailure(w, err)
		return
	}
	code := base64.RawURLEncoding.EncodeToString(body) + "." + base64.RawURLEncoding.EncodeToString(ed25519.Sign(key, body))
	png, err := qrcode.Encode(code, qrcode.Medium, 384)
	if err != nil {
		platform.RespondError(w, 500, "PASS_RENDER_FAILED", "Your pass could not be displayed. Please try again", nil)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	platform.RespondJSON(w, 200, struct {
		QRImage   string                 `json:"qr_image"`
		Data      db.GetAccountTicketRow `json:"data"`
		Code      string                 `json:"code"`
		PublicKey string                 `json:"public_key"`
		Algorithm string                 `json:"algorithm"`
		ShortCode string                 `json:"short_code"`
	}{"data:image/png;base64," + base64.StdEncoding.EncodeToString(png), ticket, code, base64.StdEncoding.EncodeToString(key.Public().(ed25519.PublicKey)), "Ed25519", ticket.PublicCode})
}

// VerifyTicketCode checks the signed payload; one-use admission is still a DB transaction.
func VerifyTicketCode(code string) (TicketClaims, error) {
	var claims TicketClaims
	key, err := ticketKey()
	if err != nil {
		return claims, err
	}
	parts := strings.Split(code, ".")
	if len(parts) != 2 || len(code) > 2048 {
		return claims, errors.New("invalid ticket")
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return claims, errors.New("invalid ticket")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !ed25519.Verify(key.Public().(ed25519.PublicKey), body, signature) {
		return claims, errors.New("invalid ticket")
	}
	if err = json.Unmarshal(body, &claims); err != nil || !claims.TicketID.Valid || !claims.EventID.Valid || claims.ExpiresAt <= time.Now().Unix() {
		return claims, errors.New("expired or invalid ticket")
	}
	return claims, nil
}
