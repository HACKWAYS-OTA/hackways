// Package testsupport is used only by the repository's mandatory CI tests.
package testsupport

import (
	"net/url"
	"os"
	"strings"
	"testing"
)

// DatabaseURL prevents fixtures from falling back to the application database.
func DatabaseURL(t *testing.T) string {
	t.Helper()
	raw := os.Getenv("DATABASE_URL")
	if raw == "" {
		t.Skip("explicit isolated DATABASE_URL is required")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal("invalid isolated database configuration")
	}
	schema := parsed.Query().Get("search_path")
	isolatedSchema := strings.HasPrefix(schema, "workflow_") && strings.Contains(schema, "_ci_") && strings.IndexFunc(schema, func(c rune) bool { return !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') }) < 0
	if !strings.HasSuffix(parsed.Path, "_test") && !isolatedSchema {
		t.Fatal("refusing database-backed tests outside an isolated _test database or workflow CI schema")
	}
	return raw
}
