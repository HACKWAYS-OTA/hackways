package platform

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgconn"
)

// ErrorDetail represents specific field or parameter validation errors.
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

// ErrorBody represents the standard error payload structure.
type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// ErrorEnvelope wraps ErrorBody in a deterministic JSON root key.
type ErrorEnvelope struct {
	Error ErrorBody `json:"error"`
}

// RespondJSON marshals data to JSON and writes it with status code and headers.
func RespondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if payload == nil || status == http.StatusNoContent {
		return
	}

	_ = json.NewEncoder(w).Encode(payload)
}

// RespondError writes a standardized error envelope JSON response.
func RespondError(w http.ResponseWriter, status int, code, message string, details map[string]any) {
	envelope := ErrorEnvelope{
		Error: ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	RespondJSON(w, status, envelope)
}

// MapDatabaseError maps PostgreSQL error codes to standard HTTP status, error code, and user message.
func MapDatabaseError(err error) (int, string, string) {
	if err == nil {
		return http.StatusOK, "", ""
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return http.StatusConflict, "CONFLICT", "Resource already exists or duplicate key violated"
		case "23503": // foreign_key_violation
			return http.StatusBadRequest, "INVALID_REFERENCE", "Referenced entity does not exist or constraint violated"
		case "23514": // check_violation
			return http.StatusUnprocessableEntity, "CHECK_VIOLATION", "Constraint check violated: invalid entity data"
		case "57014": // query_canceled (e.g. statement_timeout)
			return http.StatusServiceUnavailable, "STATEMENT_TIMEOUT", "Operation timed out under high load. Please retry."
		case "55P03": // lock_not_available
			return http.StatusServiceUnavailable, "LOCK_TIMEOUT", "Could not obtain lock on resource. Please retry."
		case "40P01": // deadlock_detected
			return http.StatusServiceUnavailable, "DEADLOCK_DETECTED", "Transaction conflict detected. Please retry."
		}
	}

	return http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected database error occurred"
}

// RespondDatabaseError translates a database error and emits a formatted error response.
func RespondDatabaseError(w http.ResponseWriter, err error) {
	status, code, msg := MapDatabaseError(err)
	RespondError(w, status, code, msg, nil)
}
