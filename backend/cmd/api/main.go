// @title Uptaris API
// @version 0.1.0
// @description REST API for Uptaris infrastructure monitoring.
// @BasePath /api/v1
// @schemes http https
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	_ "github.com/uptaris/uptaris/backend/docs"
	"github.com/uptaris/uptaris/backend/internal/config"
	"github.com/uptaris/uptaris/backend/internal/database"
	apihttp "github.com/uptaris/uptaris/backend/internal/http"
)

func main() {
	cfg := config.Load()
	logger := newLogger(cfg.Environment)

	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{Dsn: cfg.SentryDSN, Environment: cfg.Environment}); err != nil {
			logger.Error("initialize sentry", "error", err)
			os.Exit(1)
		}
		defer sentry.Flush(2 * time.Second)
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", "error", err)
		os.Exit(1)
	}

	router := apihttp.NewRouter(cfg, db, logger)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("API listening", "address", server.Addr, "environment", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			sentry.CaptureException(err)
			logger.Error("serve API", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown API", "error", err)
	}
}

func newLogger(environment string) *slog.Logger {
	level := slog.LevelInfo
	if environment == "development" {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
