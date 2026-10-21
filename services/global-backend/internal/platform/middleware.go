package platform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type contextKey string

const RequestIDHeader = "X-Request-ID"
const RequestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if len(id) < 8 || len(id) > 128 || strings.IndexFunc(id, func(c rune) bool {
			return !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_')
		}) >= 0 {
			id = generateID()
		}
		w.Header().Set(RequestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), RequestIDKey, id)))
	})
}
func GetRequestID(ctx context.Context) string { id, _ := ctx.Value(RequestIDKey).(string); return id }
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			defer func() {
				logger.Info("http request completed", "request_id", GetRequestID(r.Context()), "method", r.Method, "route", r.URL.Path, "status", rec.Status(), "duration_ms", time.Since(start).Milliseconds(), "bytes", rec.BytesWritten())
			}()
			next.ServeHTTP(rec, r)
		})
	}
}
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("unhandled panic", "request_id", GetRequestID(r.Context()), "stack", string(debug.Stack()))
					RespondError(w, 500, "INTERNAL_ERROR", "An unexpected server error occurred", nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// CORS also enforces origins on unsafe browser requests; CORS headers alone do not prevent CSRF.
func CORS(origins []string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(origins))
	for _, origin := range origins {
		allowed[origin] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Origin")
			origin := r.Header.Get("Origin")
			unsafe := r.Method != http.MethodGet && r.Method != http.MethodHead && r.Method != http.MethodOptions
			if origin != "" && !allowed[origin] && !allowed["*"] {
				RespondError(w, 403, "ORIGIN_DENIED", "This origin is not allowed", nil)
				return
			}
			if unsafe && origin == "" && r.Header.Get("Sec-Fetch-Site") == "cross-site" {
				RespondError(w, 403, "ORIGIN_DENIED", "Cross-site requests are not allowed", nil)
				return
			}
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-Request-ID, Idempotency-Key, If-Match")
				w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, Retry-After, ETag, Idempotency-Replayed")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(204)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequestProtection bounds request size, execution time, and simultaneous work.
// The channel is admission control, never authoritative application state.
func RequestProtection(maxConcurrent int) func(http.Handler) http.Handler {
	sem := make(chan struct{}, maxConcurrent)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Cache-Control", "no-store")
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			default:
				w.Header().Set("Retry-After", "1")
				RespondError(w, 503, "OVERLOADED", "Service is busy; retry shortly", nil)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
			if r.ContentLength > 1<<20 {
				RespondError(w, 413, "BODY_TOO_LARGE", "Request exceeds 1 MB", nil)
				return
			}
			if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
				if r.ContentLength != 0 {
					media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
					if err != nil || media != "application/json" {
						RespondError(w, 415, "JSON_REQUIRED", "Use application/json", nil)
						return
					}
				}
			}
			ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// DecodeJSON rejects ambiguous payloads, unknown fields, and trailing documents.
func DecodeJSON[T any](r *http.Request, dst *T) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("expected exactly one JSON object")
	}
	return nil
}
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "req-" + time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(b)
}
