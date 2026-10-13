package platform

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
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

func TestMapDatabaseError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Nil error",
			err:            nil,
			expectedStatus: http.StatusOK,
			expectedCode:   "",
		},
		{
			name: "Unique violation",
			err: &pgconn.PgError{
				Code: "23505",
			},
			expectedStatus: http.StatusConflict,
			expectedCode:   "CONFLICT",
		},
		{
			name: "Foreign key violation",
			err: &pgconn.PgError{
				Code: "23503",
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_REFERENCE",
		},
		{
			name: "Check constraint violation",
			err: &pgconn.PgError{
				Code: "23514",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedCode:   "CHECK_VIOLATION",
		},
		{
			name: "Statement timeout query canceled",
			err: &pgconn.PgError{
				Code: "57014",
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   "STATEMENT_TIMEOUT",
		},
		{
			name: "Lock not available",
			err: &pgconn.PgError{
				Code: "55P03",
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   "LOCK_TIMEOUT",
		},
		{
			name: "Deadlock detected",
			err: &pgconn.PgError{
				Code: "40P01",
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedCode:   "DEADLOCK_DETECTED",
		},
		{
			name:           "Unknown generic error",
			err:            errors.New("unexpected socket crash"),
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, code, _ := MapDatabaseError(tc.err)
			if status != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, status)
			}
			if code != tc.expectedCode {
				t.Errorf("expected code %q, got %q", tc.expectedCode, code)
			}
		})
	}
}

func TestRespondDatabaseError(t *testing.T) {
	rec := httptest.NewRecorder()
	pgErr := &pgconn.PgError{Code: "23505"}

	RespondDatabaseError(rec, pgErr)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rec.Code)
	}

	var env ErrorEnvelope
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode json body: %v", err)
	}

	if env.Error.Code != "CONFLICT" {
		t.Errorf("expected code 'CONFLICT', got %q", env.Error.Code)
	}
}
