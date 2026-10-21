package platform

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("APP_ENV", "development")
	t.Setenv("DATABASE_URL", "postgres://localhost/hackways_config_only?sslmode=disable")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("expected no error loading default config, got %v", err)
	}

	if cfg.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Port)
	}

	if cfg.AppEnv != "development" {
		t.Errorf("expected default AppEnv 'development', got %q", cfg.AppEnv)
	}

	if len(cfg.CORSAllowedOrigins) == 0 {
		t.Error("expected non-empty CORSAllowedOrigins")
	}
}

func TestLoadConfig_InvalidPort(t *testing.T) {
	os.Setenv("PORT", "invalid-port")
	defer os.Unsetenv("PORT")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error for non-numeric port, got nil")
	}
}

func TestLoadConfig_ProductionSecretValidation(t *testing.T) {
	os.Setenv("APP_ENV", "production")
	os.Setenv("SESSION_SECRET", "dev-short-secret")
	defer os.Unsetenv("APP_ENV")
	defer os.Unsetenv("SESSION_SECRET")

	_, err := LoadConfig()
	if err == nil {
		t.Fatal("expected error when running in production with dev secret, got nil")
	}
}
