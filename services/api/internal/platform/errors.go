package platform

import (
	"encoding/json"
	"net/http"
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
