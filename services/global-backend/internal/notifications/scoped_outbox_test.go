package notifications

import (
	"context"
	"fmt"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/testsupport"
	"github.com/jackc/pgx/v5/pgxpool"
	"testing"
	"time"
)

func TestScopedOutboxClaimLeavesHistoricalTemplatesUntouched(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)
	email := fmt.Sprintf("scoped-mail-%d@example.invalid", time.Now().UnixNano())
	defer func() { _, _ = pool.Exec(ctx, "DELETE FROM email_outbox WHERE recipient_email=$1", email) }()
	allowed, err := q.EnqueueEmailOutbox(ctx, db.EnqueueEmailOutboxParams{RecipientEmail: email, TemplateType: "admin_daily_otp", PayloadJson: []byte(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	other, err := q.EnqueueEmailOutbox(ctx, db.EnqueueEmailOutboxParams{RecipientEmail: email, TemplateType: "rsvp_confirmation", PayloadJson: []byte(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := q.ClaimEmailOutbox(ctx, db.ClaimEmailOutboxParams{BatchLimit: 2, TemplateTypes: []string{"admin_daily_otp"}})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, row := range rows {
		if row.TemplateType != "admin_daily_otp" {
			t.Fatal("scoped worker claimed a different template")
		}
		if row.ID == allowed.ID {
			found = true
		}
	}
	if !found {
		t.Fatal("scoped verification mail was not claimed")
	}
	var unchanged bool
	if err = pool.QueryRow(ctx, "SELECT status='pending' AND lease_token IS NULL AND first_sent_at IS NULL FROM email_outbox WHERE id=$1", other.ID).Scan(&unchanged); err != nil || !unchanged {
		t.Fatal("historical template state was changed")
	}
}
