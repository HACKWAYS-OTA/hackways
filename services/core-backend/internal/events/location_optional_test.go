package events

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/hackways/core-backend/db"
)

// Uses existing mandatory-CI fixtures and the explicit isolated _test database.
func TestEventLifecycle_LocationCanBeProvidedLater(t *testing.T) {
	f := ownerSetup(t)
	api := NewHandler(f.pool).Routes()
	start, end := time.Now().Add(4*time.Hour), time.Now().Add(5*time.Hour)
	for _, format := range []string{"in_person", "online", "hybrid"} {
		t.Run(format, func(t *testing.T) {
			payload := createEventRequest{Status: "published", Title: "Location later " + format,
				Description: "An isolated event without location details", Format: format,
				StartsAt: start.Format(time.RFC3339), EndsAt: end.Format(time.RFC3339), Timezone: "UTC"}
			body, _ := json.Marshal(payload)
			created := requestJSON(t, api, http.MethodPost, "/", body, f.cookie)
			if created.Code != http.StatusCreated {
				t.Fatalf("blank location blocked creation: %d %s", created.Code, created.Body.String())
			}
			var record struct {
				Data db.Events `json:"data"`
			}
			if json.Unmarshal(created.Body.Bytes(), &record) != nil || record.Data.Status != db.EventStatusPublished || record.Data.LocationName.Valid || record.Data.OnlineUrl.Valid {
				t.Fatal("event was not actually published with empty location details")
			}
			path := "/" + idString(record.Data.ID)
			// Editing a published event must not impose a missing venue or URL.
			update := updateRequest{Title: record.Data.Title, Description: "Updated location-later event", Format: format,
				StartsAt: start, EndsAt: end, Timezone: "UTC", Mode: "instant", AllowGuest: true, OfferMinutes: 10,
				Revision: record.Data.Revision, Acknowledge: true}
			body, _ = json.Marshal(update)
			updated := requestJSON(t, api, http.MethodPut, path, body, f.cookie)
			if updated.Code != http.StatusOK {
				t.Fatalf("blank location blocked editing: %d %s", updated.Code, updated.Body.String())
			}
			if json.Unmarshal(updated.Body.Bytes(), &record) != nil {
				t.Fatal("invalid event edit response")
			}
			update.Revision = record.Data.Revision
			update.OnlineURL = "http://invalid.example/meeting"
			body, _ = json.Marshal(update)
			if rejected := requestJSON(t, api, http.MethodPut, path, body, f.cookie); rejected.Code != http.StatusUnprocessableEntity {
				t.Fatalf("supplied unsafe URL accepted during editing: %d", rejected.Code)
			}
			// Draft publication uses the same optional-location rule as creation.
			payload.Status = "draft"
			body, _ = json.Marshal(payload)
			draft := requestJSON(t, api, http.MethodPost, "/", body, f.cookie)
			if draft.Code != http.StatusCreated || json.Unmarshal(draft.Body.Bytes(), &record) != nil {
				t.Fatal("draft creation failed")
			}
			body, _ = json.Marshal(transitionRequest{Status: db.EventStatusPublished, Revision: record.Data.Revision})
			published := requestJSON(t, api, http.MethodPost, "/"+idString(record.Data.ID)+"/status", body, f.cookie)
			if published.Code != http.StatusOK || json.Unmarshal(published.Body.Bytes(), &record) != nil || record.Data.Status != db.EventStatusPublished {
				t.Fatalf("blank location blocked draft publication: %d %s", published.Code, published.Body.String())
			}
			payload.Status = "published"
			payload.OnlineURL = "http://invalid.example/meeting"
			body, _ = json.Marshal(payload)
			if rejected := requestJSON(t, api, http.MethodPost, "/", body, f.cookie); rejected.Code != http.StatusBadRequest {
				t.Fatalf("supplied unsafe URL accepted during creation: %d", rejected.Code)
			}
		})
	}
}

func TestEventPublication_RequiredDetailsRemainValidated(t *testing.T) {
	f := ownerSetup(t)
	api := NewHandler(f.pool).Routes()
	valid := createEventRequest{Status: "published", Title: "Required details CI", Description: "Valid event description",
		Format: "online", StartsAt: time.Now().Add(3 * time.Hour).Format(time.RFC3339),
		EndsAt: time.Now().Add(4 * time.Hour).Format(time.RFC3339), Timezone: "UTC"}
	for _, kind := range []string{"title", "description", "past", "end", "timezone"} {
		payload := valid
		switch kind {
		case "title":
			payload.Title = " "
		case "description":
			payload.Description = " "
		case "past":
			payload.StartsAt = time.Now().Add(-time.Hour).Format(time.RFC3339)
		case "end":
			payload.EndsAt = time.Now().Add(time.Hour).Format(time.RFC3339)
		case "timezone":
			payload.Timezone = "not/a/timezone"
		}
		body, _ := json.Marshal(payload)
		if rejected := requestJSON(t, api, http.MethodPost, "/", body, f.cookie); rejected.Code != http.StatusBadRequest {
			t.Fatalf("invalid %s accepted during creation: %d", kind, rejected.Code)
		}
	}
	// Corrupt draft fixtures model legacy rows, not application-created records.
	// Publication must reject these under the existing event row lock.
	for _, kind := range []string{"title", "description", "past", "timezone", "url"} {
		payload := valid
		payload.Status = "draft"
		body, _ := json.Marshal(payload)
		created := requestJSON(t, api, http.MethodPost, "/", body, f.cookie)
		var record struct {
			Data db.Events `json:"data"`
		}
		if created.Code != http.StatusCreated || json.Unmarshal(created.Body.Bytes(), &record) != nil {
			t.Fatal("draft fixture failed")
		}
		var err error
		switch kind {
		case "title":
			_, err = f.pool.Exec(context.Background(), "UPDATE events SET title=' ' WHERE id=$1", record.Data.ID)
		case "description":
			_, err = f.pool.Exec(context.Background(), "UPDATE events SET description=' ' WHERE id=$1", record.Data.ID)
		case "past":
			_, err = f.pool.Exec(context.Background(), "UPDATE events SET starts_at=clock_timestamp()-interval '1 hour' WHERE id=$1", record.Data.ID)
		case "timezone":
			_, err = f.pool.Exec(context.Background(), "UPDATE events SET timezone='not/a/timezone' WHERE id=$1", record.Data.ID)
		case "url":
			_, err = f.pool.Exec(context.Background(), "UPDATE events SET online_url='http://invalid.example/meeting' WHERE id=$1", record.Data.ID)
		}
		if err != nil {
			t.Fatal(err)
		}
		body, _ = json.Marshal(transitionRequest{Status: db.EventStatusPublished, Revision: record.Data.Revision})
		if rejected := requestJSON(t, api, http.MethodPost, "/"+idString(record.Data.ID)+"/status", body, f.cookie); rejected.Code != http.StatusUnprocessableEntity {
			t.Fatalf("invalid %s draft was published: %d %s", kind, rejected.Code, rejected.Body.String())
		}
	}
}
