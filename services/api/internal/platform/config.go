package platform

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	AppEnv             string
	Port               int
	BaseURL            string
	DatabaseURL        string
	SessionSecret      string
	SessionName        string
	SessionMaxAge      time.Duration
	CORSAllowedOrigins []string
	StorageBackend     string
	LocalStorageDir    string
	DBMaxConns         int
	DBMinConns         int
	DBMaxConnLifetime  time.Duration
	DBMaxConnIdleTime  time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	IdleTimeout        time.Duration
	ShutdownTimeout    time.Duration
}

// LoadConfig populates Config from environment variables with sensible defaults.
func LoadConfig() (*Config, error) {
	portStr := getEnv("PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid PORT environment variable %q: must be between 1 and 65535", portStr)
	}

	appEnv := getEnv("APP_ENV", "development")
	sessionSecret := getEnv("SESSION_SECRET", "dev-insecure-secret-key-32-bytes-long!")

	if appEnv == "production" && (sessionSecret == "" || strings.HasPrefix(sessionSecret, "dev-")) {
		return nil, fmt.Errorf("SESSION_SECRET must be configured with a secure key in production")
	}

	corsOrigins := strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000"), ",")
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
	}

	dbMaxConns, _ := strconv.Atoi(getEnv("DB_MAX_CONNS", "28"))
	if dbMaxConns <= 0 || dbMaxConns > 60 {
		dbMaxConns = 28
	}

	dbMinConns, _ := strconv.Atoi(getEnv("DB_MIN_CONNS", "4"))
	if dbMinConns <= 0 || dbMinConns > dbMaxConns {
		dbMinConns = 4
	}

	cfg := &Config{
		AppEnv:             appEnv,
		Port:               port,
		BaseURL:            getEnv("BASE_URL", fmt.Sprintf("http://localhost:%d", port)),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://hackways_user:hackways_password@localhost:5432/hackways_dev?sslmode=disable"),
		SessionSecret:      sessionSecret,
		SessionName:        getEnv("SESSION_NAME", "hackways_session"),
		SessionMaxAge:      30 * 24 * time.Hour,
		CORSAllowedOrigins: corsOrigins,
		StorageBackend:     getEnv("STORAGE_BACKEND", "local"),
		LocalStorageDir:    getEnv("LOCAL_STORAGE_DIR", "./uploads"),
		DBMaxConns:         dbMaxConns,
		DBMinConns:         dbMinConns,
		DBMaxConnLifetime:  15 * time.Minute,
		DBMaxConnIdleTime:  5 * time.Minute,
		ReadTimeout:        10 * time.Second,
		WriteTimeout:       30 * time.Second,
		IdleTimeout:        60 * time.Second,
		ShutdownTimeout:    15 * time.Second,
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
