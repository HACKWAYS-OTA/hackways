package platform

import (
	"errors"
	"github.com/jackc/pgx/v5"
	"net"
	"strconv"
)

// Preview mode can run only on loopback against a separately restored local
// database. It must never enable an unreviewed service against production data.
func adminPreviewMode(environment, bindHost, databaseURL, flag string) (bool, error) {
	if flag == "" {
		return false, nil
	}
	enabled, err := strconv.ParseBool(flag)
	if err != nil {
		return false, errors.New("ADMIN_PREVIEW_ONLY must be a boolean")
	}
	if !enabled {
		return false, nil
	}
	parsed, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return false, errors.New("invalid local preview database configuration")
	}
	bindIP := net.ParseIP(bindHost)
	databaseIP := net.ParseIP(parsed.Host)
	localDatabase := parsed.Host == "localhost" || (databaseIP != nil && databaseIP.IsLoopback())
	for _, fallback := range parsed.Fallbacks {
		fallbackIP := net.ParseIP(fallback.Host)
		if fallback.Host != "localhost" && (fallbackIP == nil || !fallbackIP.IsLoopback()) {
			localDatabase = false
		}
	}
	if environment != "development" || bindIP == nil || !bindIP.IsLoopback() || !localDatabase || parsed.Database != "hackways_admin_preview" {
		return false, errors.New("admin preview requires development, a loopback listener and the local hackways_admin_preview database")
	}
	return true, nil
}
