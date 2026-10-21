package platform

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds runtime configuration loaded from environment variables.
type Config struct {
	BindHost            string
	AppEnv              string
	Port                int
	DatabaseURL         string
	SessionSecret       string
	CORSAllowedOrigins  []string
	DBMaxConns          int
	DBMinConns          int
	DBMaxConnLifetime   time.Duration
	DBMaxConnIdleTime   time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	IdleTimeout         time.Duration
	ShutdownTimeout     time.Duration
	GoogleClientID      string
	GoogleClientSecret  string
	GoogleRedirectURL   string
	GoogleReturnOrigins []string
}

// LoadConfig populates Config from environment variables with sensible defaults.
func LoadConfig() (*Config, error) {
	portStr := getEnv("PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		return nil, fmt.Errorf("invalid PORT environment variable %q: must be between 1 and 65535", portStr)
	}

	appEnv := getEnv("APP_ENV", "development")
	sessionSecret := os.Getenv("SESSION_SECRET")

	normalizedSessionSecret := strings.ToLower(strings.TrimSpace(sessionSecret))
	if appEnv == "production" && (len(normalizedSessionSecret) < 32 || strings.HasPrefix(normalizedSessionSecret, "dev-") || strings.HasPrefix(normalizedSessionSecret, "change-me") || strings.HasPrefix(normalizedSessionSecret, "replace-me")) {
		return nil, fmt.Errorf("SESSION_SECRET must be configured with a secure key in production")
	}

	corsOrigins := strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://127.0.0.1:3000"), ",")
	for i := range corsOrigins {
		corsOrigins[i] = strings.TrimSpace(corsOrigins[i])
		parsedOrigin, originErr := url.Parse(corsOrigins[i])
		if appEnv == "production" && (originErr != nil || parsedOrigin.Scheme != "https" || parsedOrigin.Hostname() == "" || parsedOrigin.User != nil || parsedOrigin.Path != "" || parsedOrigin.RawQuery != "" || parsedOrigin.Fragment != "") {
			return nil, fmt.Errorf("production CORS origins must be explicit HTTPS origins")
		}
	}

	dbMaxConns, _ := strconv.Atoi(getEnv("DB_MAX_CONNS", "28"))
	if dbMaxConns <= 0 || dbMaxConns > 60 {
		dbMaxConns = 28
	}

	dbMinConns, _ := strconv.Atoi(getEnv("DB_MIN_CONNS", "4"))
	if dbMinConns < 0 || dbMinConns > dbMaxConns {
		dbMinConns = min(4, dbMaxConns)
	}

	if strings.TrimSpace(os.Getenv("DATABASE_URL")) == "" {
		return nil, fmt.Errorf("DATABASE_URL must be configured")
	}
	googleClientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	googleClientSecret := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
	googleRedirectURL := strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URL"))
	if (googleClientID != "" || googleClientSecret != "" || googleRedirectURL != "") && (googleClientID == "" || googleClientSecret == "" || googleRedirectURL == "") {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, and GOOGLE_REDIRECT_URL must be configured together")
	}
	googleReturnOrigins := parseOrigins(getEnv("GOOGLE_ALLOWED_RETURN_ORIGINS", strings.Join(corsOrigins, ",")))
	for _, origin := range googleReturnOrigins {
		parsedOrigin, originErr := url.Parse(origin)
		if originErr != nil || parsedOrigin.Scheme == "" || parsedOrigin.Hostname() == "" || parsedOrigin.User != nil || parsedOrigin.Path != "" || parsedOrigin.RawQuery != "" || parsedOrigin.Fragment != "" {
			return nil, fmt.Errorf("GOOGLE_ALLOWED_RETURN_ORIGINS must contain explicit origins")
		}
		if appEnv == "production" && parsedOrigin.Scheme != "https" {
			return nil, fmt.Errorf("production GOOGLE_ALLOWED_RETURN_ORIGINS must use HTTPS")
		}
	}
	bindHost := "127.0.0.1"
	if appEnv == "production" {
		bindHost = "0.0.0.0"
	}
	bindHost = getEnv("BIND_HOST", bindHost)
	if net.ParseIP(bindHost) == nil {
		return nil, fmt.Errorf("BIND_HOST must be an IP address")
	}
	cfg := &Config{
		BindHost:            bindHost,
		AppEnv:              appEnv,
		Port:                port,
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		SessionSecret:       sessionSecret,
		CORSAllowedOrigins:  corsOrigins,
		DBMaxConns:          dbMaxConns,
		DBMinConns:          dbMinConns,
		DBMaxConnLifetime:   15 * time.Minute,
		DBMaxConnIdleTime:   5 * time.Minute,
		ReadTimeout:         10 * time.Second,
		WriteTimeout:        30 * time.Second,
		IdleTimeout:         60 * time.Second,
		ShutdownTimeout:     15 * time.Second,
		GoogleClientID:      googleClientID,
		GoogleClientSecret:  googleClientSecret,
		GoogleRedirectURL:   googleRedirectURL,
		GoogleReturnOrigins: googleReturnOrigins,
	}

	return cfg, nil
}

func parseOrigins(value string) []string {
	var origins []string
	for _, raw := range strings.Split(value, ",") {
		if origin := strings.TrimSpace(raw); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
