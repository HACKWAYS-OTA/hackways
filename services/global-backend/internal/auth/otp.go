package auth

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/hackways/global-backend/db"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type otpPayload struct {
	Code string `json:"verification_code"`
	Hash string `json:"verification_hash"`
}

func newEmailCode() (string, string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", "", err
	}
	code := fmt.Sprintf("%06d", n.Int64())
	hash, err := bcrypt.GenerateFromPassword([]byte(code), DefaultBcryptCost)
	return code, string(hash), err
}

func validEmailCode(code string) bool {
	if len(code) != 6 {
		return false
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func matchesEmailCode(code, hash string) bool {
	if strings.HasPrefix(hash, "$2") {
		return validEmailCode(code) && bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) == nil
	}
	// Previously issued long verification tokens remain usable until expiry.
	return len(code) == 64 && HashToken(code) == hash
}

func enqueueAccountEmail(ctx context.Context, q *db.Queries, email, name, template string, payload []byte) error {
	_, err := q.EnqueueEmailOutbox(ctx, db.EnqueueEmailOutboxParams{RecipientEmail: email, RecipientName: pgtype.Text{String: name, Valid: name != ""}, TemplateType: template, PayloadJson: payload})
	return err
}

func queueEmailVerification(ctx context.Context, q *db.Queries, id pgtype.UUID, email, name string) error {
	code, hash, err := newEmailCode()
	if err != nil {
		return err
	}
	if err = q.CreateEmailVerification(ctx, db.CreateEmailVerificationParams{UserID: id, Email: email, TokenHash: hash}); err != nil {
		return err
	}
	payload, err := json.Marshal(otpPayload{Code: code, Hash: hash})
	if err != nil {
		return err
	}
	return enqueueAccountEmail(ctx, q, email, name, "email_verification", payload)
}
