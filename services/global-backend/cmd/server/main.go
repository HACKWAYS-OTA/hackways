package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/notifications"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/hackways/global-backend/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// 2. Load configuration
	cfg, err := platform.LoadConfig()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Default port to 8082 for Global Backend if not explicitly overridden
	if cfg.Port == 8080 {
		cfg.Port = 8082
	}
	// Enforce global-backend hard cap of 22 connections as per AGENTS.md Rule 2.5
	if cfg.DBMaxConns > 22 || cfg.DBMaxConns == 25 {
		cfg.DBMaxConns = 22
	}

	logger.Info("starting Hackways Global Backend Engine (Auth, Storage & ACS Email)",
		"environment", cfg.AppEnv,
		"port", cfg.Port,
		"max_db_conns", cfg.DBMaxConns,
	)

	// 3. Initialize PostgreSQL Connection Pool
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to parse database configuration; check DATABASE_URL without logging its value")
		os.Exit(1)
	}
	poolConfig.MaxConns = int32(cfg.DBMaxConns)
	poolConfig.MinConns = int32(min(cfg.DBMinConns, cfg.DBMaxConns))
	poolConfig.MaxConnLifetime = cfg.DBMaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.DBMaxConnIdleTime

	initCtx, initCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer initCancel()

	pool, err := pgxpool.NewWithConfig(initCtx, poolConfig)
	if err != nil {
		logger.Error("could not initialize database pool", "error", err)
		os.Exit(1)
	} else {
		if pingErr := pool.Ping(initCtx); pingErr != nil {
			logger.Warn("database is not ready; health checks will return unavailable")
		}
		logger.Info("initialized PostgreSQL pool",
			"max_conns", cfg.DBMaxConns,
			"min_conns", cfg.DBMinConns,
		)
	}

	// 4. Initialize storage SAS signer
	sasConfig := storage.AzureSASConfig{
		AccountName: os.Getenv("AZURE_STORAGE_ACCOUNT"),
		AccountKey:  os.Getenv("AZURE_STORAGE_ACCOUNT_KEY"),
		Container:   "media-uploads",
		CDNDomain:   os.Getenv("AZURE_CDN_DOMAIN"),
	}
	if sasConfig.AccountName == "" || sasConfig.AccountKey == "" {
		logger.Warn("Azure Blob Storage is not configured; upload intents are unavailable")
	}
	sasSigner := storage.NewAzureBlobSASSigner(sasConfig)

	// 5. Initialize transactional email outbox worker
	var emailSender notifications.EmailSender = notifications.UnavailableEmailSender{}
	emailConfigured := false
	acsEndpoint, acsKey, acsSender := notifications.EmailConfiguration()
	if acsEndpoint != "" || acsKey != "" || acsSender != "" {
		configuredSender, senderErr := notifications.NewACSEmailSender(acsEndpoint, acsKey, acsSender)
		if senderErr != nil {
			logger.Error("invalid ACS email configuration", "error", senderErr)
			os.Exit(1)
		}
		emailSender = configuredSender
		emailConfigured = true
	} else {
		logger.Warn("ACS email is not configured; notification delivery is paused without consuming retries")
	}

	var outboxWorker *notifications.OutboxWorker
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	if pool != nil && emailConfigured {
		outboxWorker = notifications.NewOutboxWorker(pool, emailSender, logger, 2*time.Second)
		outboxWorker.Start(workerCtx)
	}

	var housekeepingWG sync.WaitGroup
	housekeepingWG.Add(1)
	go func() { defer housekeepingWG.Done(); notifications.Housekeeping(workerCtx, pool, logger) }()
	// 6. Initialize HTTP router and middleware
	r := chi.NewRouter()
	r.Use(platform.RequestID)
	r.Use(platform.Logger(logger))
	r.Use(platform.Recoverer(logger))
	r.Use(platform.RequestProtection(24))
	r.Use(platform.CORS(cfg.CORSAllowedOrigins))

	// 7. Mount health check
	r.Get("/health", platform.HealthHandler(cfg.AppEnv, pool))

	// 8. Mount domain handlers
	authHandler := auth.NewHandler(pool, cfg.AppEnv == "production", auth.GoogleOAuthConfig{
		ClientID: cfg.GoogleClientID, ClientSecret: cfg.GoogleClientSecret,
		RedirectURL: cfg.GoogleRedirectURL, ReturnOrigins: cfg.GoogleReturnOrigins,
		SessionSecret: cfg.SessionSecret, IsProduction: cfg.AppEnv == "production",
	})
	storageHandler := storage.NewHandler(sasSigner, pool)

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/", func(w http.ResponseWriter, r *http.Request) {
			platform.RespondJSON(w, http.StatusOK, map[string]string{
				"service": "Hackways Global Backend",
				"version": "v1",
			})
		})
		v1.Mount("/auth", authHandler.Routes())
		v1.With(auth.RequireAuth(authHandler.SessionManager())).Mount("/storage", storageHandler.Routes())
	})

	// 9. HTTP Server with production timeouts
	serverAddr := net.JoinHostPort(cfg.BindHost, fmt.Sprint(cfg.Port))
	srv := &http.Server{
		Addr:              serverAddr,
		Handler:           r,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    16 << 10,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
	}

	// 10. Graceful shutdown handler
	shutdownErr := make(chan error, 1)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("shutting down Global Backend gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		shutdownResult := srv.Shutdown(ctx)
		workerCancel()
		housekeepingWG.Wait()

		if outboxWorker != nil {
			logger.Info("stopping transactional email outbox worker...")
			workerCancel()
			outboxWorker.Stop()
		}

		if pool != nil {
			logger.Info("closing PostgreSQL connection pool...")
			pool.Close()
		}

		shutdownErr <- shutdownResult
	}()

	logger.Info("Global Backend listener ready", "addr", serverAddr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Global Backend server error", "error", err)
		os.Exit(1)
	}

	if err := <-shutdownErr; err != nil {
		logger.Error("error during graceful shutdown", "error", err)
		os.Exit(1)
	}
	logger.Info("Global Backend stopped cleanly")
}
