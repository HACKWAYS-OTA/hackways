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
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/core-backend/internal/admin"
	"github.com/hackways/core-backend/internal/events"
	"github.com/hackways/core-backend/internal/platform"
	"github.com/hackways/core-backend/internal/rsvps"
	"github.com/hackways/core-backend/internal/spaces"
	"github.com/hackways/core-backend/internal/workflows"
	"github.com/jackc/pgx/v5/pgxpool"
	"sync"
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

	// Default port to 8081 for Core Backend if not explicitly overridden
	if cfg.Port == 8080 {
		cfg.Port = 8081
	}
	// Enforce core-backend hard cap of 28 connections as per AGENTS.md Rule 2.5
	if cfg.DBMaxConns > 28 || cfg.DBMaxConns == 25 {
		cfg.DBMaxConns = 28
	}

	logger.Info("starting Hackways Core Backend Engine (Events & RSVP Capacity)",
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

	// 4. Initialize HTTP router and middleware
	r := chi.NewRouter()
	r.Use(platform.RequestID)
	r.Use(platform.Logger(logger))
	r.Use(platform.Recoverer(logger))
	r.Use(platform.RequestProtection(32))
	r.Use(platform.CORS(cfg.CORSAllowedOrigins))

	// 5. Mount health check
	r.Get("/health", platform.HealthHandler(cfg.AppEnv, pool))

	// 6. Mount domain handlers
	eventsHandler := events.NewHandler(pool)
	rsvpsHandler := rsvps.NewHandler(pool)

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Mount("/admin/core", admin.NewHandler(pool).Routes())
		v1.Get("/", func(w http.ResponseWriter, r *http.Request) {
			platform.RespondJSON(w, http.StatusOK, map[string]string{
				"service": "Hackways Core Backend",
				"version": "v1",
			})
		})
		if !cfg.AdminPreviewOnly {
			v1.Mount("/events", eventsHandler.Routes())

			v1.Mount("/rsvps", rsvpsHandler.Routes())
			v1.Mount("/spaces", spaces.NewHandler(pool).Routes())
		}
	})

	maintenanceCtx, stopMaintenance := context.WithCancel(context.Background())
	var maintenanceWG sync.WaitGroup
	if !cfg.AdminPreviewOnly {
		maintenanceWG.Add(1)
		go func() { defer maintenanceWG.Done(); workflows.New(pool).RunMaintenance(maintenanceCtx, logger) }()
	}
	defer stopMaintenance()
	// 7. HTTP Server with production timeouts
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

	// 8. Graceful shutdown handler
	shutdownErr := make(chan error, 1)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("shutting down Core Backend gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		shutdownResult := srv.Shutdown(ctx)
		stopMaintenance()
		maintenanceWG.Wait()

		if pool != nil {
			logger.Info("closing PostgreSQL connection pool...")
			pool.Close()
		}

		shutdownErr <- shutdownResult
	}()

	logger.Info("Core Backend listener ready", "addr", serverAddr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("Core Backend server error", "error", err)
		os.Exit(1)
	}

	if err := <-shutdownErr; err != nil {
		logger.Error("error during graceful shutdown", "error", err)
		os.Exit(1)
	}
	logger.Info("Core Backend stopped cleanly")
}
