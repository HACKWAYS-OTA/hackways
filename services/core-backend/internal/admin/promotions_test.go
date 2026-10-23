package admin

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hackways/core-backend/db"
	"github.com/hackways/core-backend/internal/auth"
	"github.com/hackways/core-backend/internal/events"
	"github.com/hackways/core-backend/internal/testsupport"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestDiscoverPromotionAdministration(t *testing.T) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, testsupport.DatabaseURL(t))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	q := db.New(pool)
	actor, err := q.CreateUser(ctx, db.CreateUserParams{Email: fmt.Sprintf("promotions-%d@example.invalid", time.Now().UnixNano()), PasswordHash: "isolated-ci-only", DisplayName: "Promotion CI"})
	if err != nil {
		t.Fatal(err)
	}
	old, err := q.DiscoverPromotionSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var saved []struct {
		Slot       int16
		ID         pgtype.UUID
		Start, End pgtype.Timestamptz
	}
	rows, err := pool.Query(ctx, "SELECT slot,event_id,starts_at,ends_at FROM discover_promotions")
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var row struct {
			Slot       int16
			ID         pgtype.UUID
			Start, End pgtype.Timestamptz
		}
		if err = rows.Scan(&row.Slot, &row.ID, &row.Start, &row.End); err != nil {
			t.Fatal(err)
		}
		saved = append(saved, row)
	}
	rows.Close()
	var fixtures []pgtype.UUID
	defer func() {
		scope := sha256.Sum256([]byte(fmt.Sprintf("admin:%x:PUT:/discover-promotions", actor.ID.Bytes)))
		_, _ = pool.Exec(ctx, "DELETE FROM idempotency_records WHERE scope_hash=$1", fmt.Sprintf("%x", scope))
		_, _ = pool.Exec(ctx, "DELETE FROM discover_promotions")
		for _, row := range saved {
			_, _ = pool.Exec(ctx, "INSERT INTO discover_promotions(slot,event_id,starts_at,ends_at) VALUES($1,$2,$3,$4)", row.Slot, row.ID, row.Start, row.End)
		}
		_, _ = pool.Exec(ctx, "UPDATE discover_promotion_settings SET revision=$1,updated_at=$2,updated_by=NULL", old.Revision, old.UpdatedAt)
		_, _ = pool.Exec(ctx, "DELETE FROM audit_log WHERE actor_user_id=$1", actor.ID)
		for _, id := range fixtures {
			_, _ = pool.Exec(ctx, "DELETE FROM events WHERE id=$1", id)
		}
		_, _ = pool.Exec(ctx, "DELETE FROM organizers WHERE user_id=$1", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM sessions WHERE user_id=$1", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM platform_administrators WHERE user_id=$1", actor.ID)
		_, _ = pool.Exec(ctx, "DELETE FROM users WHERE id=$1", actor.ID)
	}()
	_, err = pool.Exec(ctx, "UPDATE users SET email_verified=true WHERE id=$1", actor.ID)
	if err != nil {
		t.Fatal(err)
	}
	sm := auth.NewSessionManager(q)
	token, session, err := sm.CreateSession(ctx, actor.ID, "CI", "", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	router := NewHandler(pool).Routes()
	call := func(method, body, key, cookie string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/discover-promotions", strings.NewReader(body))
		r.Header.Set("Origin", "https://admin.hackways.in")
		r.Header.Set("Idempotency-Key", key)
		if cookie != "" {
			r.AddCookie(&http.Cookie{Name: cookie, Value: token})
		}
		w := httptest.NewRecorder()
		router.ServeHTTP(w, r)
		return w
	}
	if got := call("GET", "", "", auth.DefaultSessionCookie).Code; got != 401 {
		t.Fatalf("ordinary session got %d", got)
	}
	if got := call("GET", "", "", auth.AdminSessionCookie).Code; got != 403 {
		t.Fatalf("ungranted session got %d", got)
	}
	_, err = pool.Exec(ctx, "INSERT INTO platform_administrators(user_id) VALUES($1)", actor.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err = q.RecordPlatformAdminSession(ctx, session.ID); err != nil {
		t.Fatal(err)
	}
	org, err := q.CreateOrganizer(ctx, db.CreateOrganizerParams{UserID: actor.ID, DisplayName: actor.DisplayName})
	if err != nil {
		t.Fatal(err)
	}
	var selections []promotionSelection
	for i := 0; i < 6; i++ {
		event, e := q.CreateEventRecord(ctx, db.CreateEventRecordParams{OrganizerID: org.ID, Status: db.EventStatusPublished, Format: db.EventFormatInPerson, Title: fmt.Sprint("Promo CI ", i), Description: "Mandatory isolated CI", StartsAt: pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true}, EndsAt: pgtype.Timestamptz{Time: time.Now().Add(25 * time.Hour), Valid: true}, Timezone: "UTC", LocationName: pgtype.Text{String: "CI", Valid: true}})
		if e != nil {
			t.Fatal(e)
		}
		fixtures = append(fixtures, event.ID)
		selections = append(selections, promotionSelection{event.AdminReference, time.Now().Add(-time.Minute), time.Now().Add(time.Hour)})
	}
	payload := func(revision int64, items []promotionSelection) string {
		encoded, _ := json.Marshal(struct {
			Revision   int64                `json:"revision"`
			Selections []promotionSelection `json:"selections"`
		}{revision, items})
		return string(encoded)
	}
	check := func(w *httptest.ResponseRecorder, status int) {
		t.Helper()
		if w.Code != status {
			t.Fatalf("expected %d got %d: %s", status, w.Code, w.Body.String())
		}
	}
	check(call("PUT", payload(old.Revision, selections), "promotion-six-events-key", auth.AdminSessionCookie), 422)
	body := payload(old.Revision, selections[:5])
	check(call("PUT", body, "promotion-five-events-key", auth.AdminSessionCookie), 200)
	replay := call("PUT", body, "promotion-five-events-key", auth.AdminSessionCookie)
	check(replay, 200)
	if replay.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatal("retry was not replayed")
	}
	check(call("PUT", payload(old.Revision, nil), "promotion-stale-version-key", auth.AdminSessionCookie), 409)
	public := func() int {
		w := httptest.NewRecorder()
		events.NewHandler(pool).Routes().ServeHTTP(w, httptest.NewRequest("GET", "/promotions", nil))
		check(w, 200)
		var data struct {
			Data []json.RawMessage `json:"data"`
		}
		if err = json.Unmarshal(w.Body.Bytes(), &data); err != nil {
			t.Fatal(err)
		}
		return len(data.Data)
	}
	if public() != 5 {
		t.Fatal("active promotion feed did not contain five events")
	}
	if _, err = pool.Exec(ctx, "INSERT INTO discover_promotions(slot,event_id,starts_at,ends_at) VALUES(6,$1,clock_timestamp(),clock_timestamp()+interval '1 hour')", fixtures[5]); err == nil {
		t.Fatal("database allowed a sixth promotion slot")
	}
	_, err = pool.Exec(ctx, "UPDATE events SET visibility='private' WHERE id=$1", fixtures[0])
	if err != nil {
		t.Fatal(err)
	}
	if public() != 4 {
		t.Fatal("private event leaked into feed")
	}
	check(call("PUT", payload(old.Revision+1, selections[:1]), "promotion-private-event-key", auth.AdminSessionCookie), 422)
	future := selections[1]
	future.StartsAt = time.Now().Add(2 * time.Hour)
	future.EndsAt = time.Now().Add(3 * time.Hour)
	check(call("PUT", payload(old.Revision+1, []promotionSelection{future}), "promotion-scheduled-event-key", auth.AdminSessionCookie), 200)
	atStart, e := q.ActiveDiscoverPromotions(ctx, pgtype.Timestamptz{Time: future.StartsAt, Valid: true})
	if e != nil || len(atStart) != 1 {
		t.Fatal("schedule start must be inclusive")
	}
	atEnd, e := q.ActiveDiscoverPromotions(ctx, pgtype.Timestamptz{Time: future.EndsAt, Valid: true})
	if e != nil || len(atEnd) != 0 {
		t.Fatal("schedule end must be exclusive")
	}
	if public() != 0 {
		t.Fatal("future promotion leaked before schedule")
	}
	check(call("PUT", payload(old.Revision+2, []promotionSelection{}), "promotion-clear-events-key", auth.AdminSessionCookie), 200)
	if public() != 0 {
		t.Fatal("cleared banner fell back to catalogue")
	}
	results := make(chan int, 2)
	for i := 0; i < 2; i++ {
		go func(i int) {
			results <- call("PUT", payload(old.Revision+3, selections[1:3]), fmt.Sprintf("promotion-concurrent-%d-key", i), auth.AdminSessionCookie).Code
		}(i)
	}
	a, b := <-results, <-results
	if !(a == 200 && b == 409 || a == 409 && b == 200) {
		t.Fatalf("concurrent revisions returned %d and %d", a, b)
	}
}
