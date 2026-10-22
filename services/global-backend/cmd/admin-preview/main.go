// Command admin-preview provisions only a local restored preview database.
// It never accepts a production host or a different database name.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

type runtimeSource struct {
	DatabaseURL string `json:"database_url"`
	AdminEmail  string `json:"admin_email"`
}
type postmanValue struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}
type postmanEnvironment struct {
	Name   string         `json:"name"`
	Values []postmanValue `json:"values"`
}

func main() {
	source := flag.String("runtime", "", "Protected local runtime JSON")
	output := flag.String("postman-environment", "", "Protected Postman environment output")
	flag.Parse()
	if *source == "" || *output == "" {
		fail("Provide protected runtime and Postman environment paths")
	}
	raw, err := os.ReadFile(*source)
	if err != nil {
		fail("Could not read protected local runtime source")
	}
	var settings runtimeSource
	if json.Unmarshal(raw, &settings) != nil || settings.DatabaseURL == "" || settings.AdminEmail == "" {
		fail("Invalid protected runtime source")
	}
	os.Setenv("DATABASE_URL", settings.DatabaseURL)
	os.Setenv("APP_ENV", "development")
	os.Setenv("BIND_HOST", "127.0.0.1")
	os.Setenv("ADMIN_PREVIEW_ONLY", "true")
	cfg, err := platform.LoadConfig()
	if err != nil || !cfg.AdminPreviewOnly {
		fail("Refusing non-local preview target")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		fail("Could not open local preview database")
	}
	defer pool.Close()
	tx, err := pool.Begin(ctx)
	if err != nil {
		fail("Could not begin local preview setup")
	}
	defer tx.Rollback(ctx)
	q := db.New(tx)
	user, err := q.GetUserByEmail(ctx, settings.AdminEmail)
	if err != nil {
		fail("Approved account is absent from the restored snapshot")
	}
	userID, err := q.GrantLocalPreviewAdmin(ctx, settings.AdminEmail)
	if err != nil {
		fail("Approved account must have a verified email before preview access is granted")
	}
	token, session, err := auth.NewSessionManager(q).CreateSession(ctx, userID, "local-preview-postman", "127.0.0.1", 30*time.Minute)
	if err != nil {
		fail("Could not create isolated preview session")
	}
	if q.RecordPlatformAdminSession(ctx, session.ID) != nil {
		fail("Could not mark isolated admin session")
	}
	environment := postmanEnvironment{Name: "Hackways local real-data preview", Values: []postmanValue{
		{Key: "base_url", Value: "http://localhost:3003", Type: "default", Enabled: true},
		{Key: "admin_session", Value: token, Type: "secret", Enabled: true},
	}}
	content, err := json.MarshalIndent(environment, "", "  ")
	if err != nil {
		fail("Could not encode protected Postman environment")
	}
	if filepath.Clean(filepath.Dir(*source)) != filepath.Clean(filepath.Dir(*output)) {
		fail("Postman environment must stay beside the protected source")
	}
	if os.WriteFile(*output, content, 0600) != nil {
		fail("Could not protect Postman environment")
	}
	if tx.Commit(ctx) != nil {
		fail("Could not confirm local preview access")
	}
	fmt.Println("Granted local preview access to the approved existing verified account.")
	fmt.Printf("Existing password sign-in supported: %t; production credentials unchanged.\n", strings.HasPrefix(user.PasswordHash, "$2"))
	fmt.Println("Created a protected 30-minute local Postman session; no token values emitted.")
}
func fail(message string) { fmt.Fprintln(os.Stderr, message); os.Exit(1) }
