package platform

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	payload := map[string]string{"status": "ok"}

	RespondJSON(rec, http.StatusOK, payload)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "application/json; charset=utf-8" {
		t.Errorf("expected json content-type, got %q", contentType)
	}

	var res map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode json body: %v", err)
	}

	if res["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", res["status"])
	}
}

func TestRespondError(t *testing.T) {
	rec := httptest.NewRecorder()

	RespondError(rec, http.StatusNotFound, "RESOURCE_NOT_FOUND", "The resource was not found", map[string]any{"id": "123"})

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var env ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode error body: %v", err)
	}

	if env.Error.Code != "RESOURCE_NOT_FOUND" {
		t.Errorf("expected error code 'RESOURCE_NOT_FOUND', got %q", env.Error.Code)
	}

	if env.Error.Message != "The resource was not found" {
		t.Errorf("expected message 'The resource was not found', got %q", env.Error.Message)
	}

	if env.Error.Details["id"] != "123" {
		t.Errorf("expected details id to be '123', got %v", env.Error.Details["id"])
	}
}
