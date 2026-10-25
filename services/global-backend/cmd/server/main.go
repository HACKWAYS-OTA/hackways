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
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/global-backend/db"
	"github.com/hackways/global-backend/internal/admin"
	"github.com/hackways/global-backend/internal/analytics"
	"github.com/hackways/global-backend/internal/auth"
	"github.com/hackways/global-backend/internal/mcpbridge"
	"github.com/hackways/global-backend/internal/notifications"
	"github.com/hackways/global-backend/internal/platform"
	"github.com/hackways/global-backend/internal/programs"
	"github.com/hackways/global-backend/internal/rollouts"
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
		configuredSender, senderErr := notifications.NewACSEmailSender(acsEndpoint, acsKey, acsSender, os.Getenv("ACS_TRANSACTIONAL_SENDER_ADDRESS"), os.Getenv("ACS_PROMOTIONAL_SENDER_ADDRESS"))
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
	var activityDone <-chan struct{}
	if !cfg.AdminPreviewOnly {
		activityDone = auth.StartSessionActivity(workerCtx, db.New(pool), logger)
	}
	var housekeepingWG sync.WaitGroup
	runtimeCache := rollouts.New(db.New(pool))
	if err := runtimeCache.Refresh(workerCtx); err != nil {
		logger.Warn("runtime configuration unavailable; presentation features use their safe baseline")
	}
	housekeepingWG.Add(1)
	go func() { defer housekeepingWG.Done(); runtimeCache.Run(workerCtx, logger) }()

	if pool != nil && emailConfigured && !cfg.AdminPreviewOnly {
		outboxWorker = notifications.NewOutboxWorker(pool, emailSender, logger, 2*time.Second)
		if filter := strings.TrimSpace(os.Getenv("EMAIL_OUTBOX_TEMPLATE_FILTER")); filter != "" {
			var values []string
			for _, value := range strings.Split(filter, ",") {
				values = append(values, strings.TrimSpace(value))
			}
			if err := outboxWorker.SetTemplateFilter(values); err != nil {
				logger.Error("invalid outbox template filter")
				os.Exit(1)
			}
		}
		outboxWorker.Start(workerCtx)
	}

	if !cfg.AdminPreviewOnly {
		housekeepingWG.Add(1)
		go func() { defer housekeepingWG.Done(); notifications.Housekeeping(workerCtx, pool, logger) }()
	}
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
		IsProduction: cfg.AppEnv == "production",
	})
	authHandler.SetLocationProxies(cfg.AuthLocationProxyCIDRs)
	authHandler.SetRollouts(runtimeCache)
	storageHandler := storage.NewHandler(sasSigner, pool)
	programHandler := programs.New(pool)

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/", func(w http.ResponseWriter, r *http.Request) {
			platform.RespondJSON(w, http.StatusOK, map[string]string{
				"service": "Hackways Global Backend",
				"version": "v1",
			})
		})
		if cfg.AdminPreviewOnly {
			v1.Mount("/auth/admin", authHandler.AdminRoutes())
		} else {
			v1.Mount("/auth", authHandler.Routes())
			v1.Mount("/analytics", analytics.NewHandler(pool, cfg.SessionSecret, cfg.AppEnv == "development").Routes())
			v1.Mount("/programs", programHandler.Routes())
			v1.With(auth.RequireAuth(authHandler.SessionManager())).Mount("/storage", storageHandler.Routes())
		}
		v1.Mount("/admin/global/programs", programHandler.AdminRoutes())
		v1.Mount("/admin/global", admin.NewHandler(pool, runtimeCache).Routes())

	})

	if auth.MCPEnabled() && !cfg.AdminPreviewOnly {
		coreBase := os.Getenv("MCP_CORE_API_ORIGIN")
		if coreBase == "" {
			coreBase = "http://core-backend:8081"
		}
		globalBase := "http://127.0.0.1:" + fmt.Sprint(cfg.Port)
		bridge, err := mcpbridge.New(pool, logger, coreBase, globalBase)
		if err != nil {
			logger.Error("MCP configuration invalid")
			os.Exit(1)
		}
		r.Mount("/mcp", bridge)
	}
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
		if activityDone != nil {
			<-activityDone
		}

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
