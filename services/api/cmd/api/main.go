package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hackways/api/internal/events"
	"github.com/hackways/api/internal/platform"
	"github.com/hackways/api/internal/rsvps"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Initialize structured logging
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

	logger.Info("starting Hackways API server",
		"environment", cfg.AppEnv,
		"port", cfg.Port,
	)

	// 3. Initialize PostgreSQL Connection Pool
	var pool *pgxpool.Pool
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to parse database configuration", "error", err)
		os.Exit(1)
	}
	poolConfig.MaxConns = int32(cfg.DBMaxConns)
	poolConfig.MinConns = int32(cfg.DBMinConns)
	poolConfig.MaxConnLifetime = cfg.DBMaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.DBMaxConnIdleTime

	initCtx, initCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer initCancel()

	pool, err = pgxpool.NewWithConfig(initCtx, poolConfig)
	if err != nil {
		logger.Warn("could not establish immediate connection to database; starting in degraded mode", "error", err)
	} else {
		logger.Info("connected to PostgreSQL",
			"max_conns", cfg.DBMaxConns,
			"min_conns", cfg.DBMinConns,
		)
	}

	// 4. Initialize HTTP router
	r := chi.NewRouter()

	// 5. Attach standard middleware stack
	r.Use(platform.RequestID)
	r.Use(platform.Logger(logger))
	r.Use(platform.Recoverer(logger))
	r.Use(platform.CORS(cfg.CORSAllowedOrigins))

	// 6. Mount health and platform diagnostic routes
	r.Get("/health", platform.HealthHandler(cfg.AppEnv, pool))

	eventsHandler := events.NewHandler(pool)
	rsvpsHandler := rsvps.NewHandler(pool)

	// 7. Mount API v1 route tree
	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/", func(w http.ResponseWriter, r *http.Request) {
			platform.RespondJSON(w, http.StatusOK, map[string]string{
				"name":    "Hackways REST API",
				"version": "v1",
				"status":  "operational",
			})
		})

		v1.Mount("/events", eventsHandler.Routes())
		v1.Mount("/rsvps", rsvpsHandler.Routes())
	})

	// 8. Configure HTTP server with production timeouts
	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// 9. Graceful shutdown handler
	shutdownErr := make(chan error, 1)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		logger.Info("shutting down server gracefully...")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		if pool != nil {
			logger.Info("closing PostgreSQL connection pool...")
			pool.Close()
		}

		shutdownErr <- srv.Shutdown(ctx)
	}()

	// 10. Start listening
	logger.Info("HTTP listener ready", "addr", serverAddr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server listener error", "error", err)
		os.Exit(1)
	}

	// 11. Wait for shutdown completion
	if err := <-shutdownErr; err != nil {
		logger.Error("error during graceful shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped cleanly")
}
