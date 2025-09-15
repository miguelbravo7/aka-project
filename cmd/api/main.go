package main

import (
	"context"
	"net/http"
	"time"

	"aka-project/internal/api"
	"aka-project/internal/config"
	"aka-project/internal/db"
	"aka-project/internal/helper"
	"aka-project/internal/logger"
	internal_middleware "aka-project/internal/middleware"
	"aka-project/internal/repository"
	"aka-project/internal/telemetry"

	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	otlpmetricgrpc "go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
)

func main() {
	logger.New()
	log.Info().Msg("Application starting...")
	cfg := config.Load()
	ctx := context.Background()

	// Telemetry
	// Create OTLP metric exporter
	metricExporter, err := otlpmetricgrpc.New(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create OTLP metric exporter")
	}

	// Create MeterProvider with OTLP exporter
	mp := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter)),
	)

	tele, err := telemetry.NewTelemetryWithMeterProvider("aka-project", mp)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init telemetry")
	}
	defer func() {
		// Flush metrics before shutdown
		if err := mp.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("failed to flush metrics")
		}
		tele.Shutdown(ctx)
	}()

	// Postgres
	pool, err := pgxpool.New(ctx, cfg.DBUrl)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect db")
	}
	q := db.New(pool)

	// Redis
	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})

	// Middleware
	mw, err := internal_middleware.NewMiddleware(redisClient, cfg.RateLimitSpec, cfg.APIKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create middleware")
	}

	// Repository + handlers
	characterRepo := repository.NewCharacterRepo(q, helper.FetchPage)
	characterHandler, err := api.NewCharacterHandler(characterRepo, tele.Meter)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create character handler")
	}
	healthHandler := &api.HealthHandler{DB: pool, Redis: redisClient}

	// Router
	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.RealIP, internal_middleware.Logger)

	r.Get("/healthcheck", healthHandler.Check)

	r.Group(func(r chi.Router) {
		r.Use(mw.RequireAPIKey)
		r.Use(mw.RateLimit)

		r.Get("/characters", otelhttp.NewHandler(
			http.HandlerFunc(characterHandler.GetCharacters),
			"CreateCharacters",
		).ServeHTTP)
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Info().Msgf("server listening on port %s", cfg.Port)

	// Start the server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	// Listen for OS signals to perform a graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit // Block until a signal is received

	log.Info().Msg("Shutting down server...")

	// Create a context with a timeout for the shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("server shutdown failed")
	}

	log.Info().Msg("Server stopped.")
}
