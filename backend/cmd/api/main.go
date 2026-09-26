// @title Uptaris API
// @version 0.1.0
// @description REST API for Uptaris infrastructure monitoring.
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.bearerauth bearerauth
//
//go:generate sh -c "cd ../.. && go run github.com/swaggo/swag/v2/cmd/swag@v2.0.0-rc4 init --v3.1 --parseDependency -g cmd/api/main.go -o openapi"
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

	"github.com/getsentry/sentry-go"
	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/config"
	"github.com/uptaris/uptaris/backend/internal/database"
	"github.com/uptaris/uptaris/backend/internal/handlers"
	"github.com/uptaris/uptaris/backend/internal/inventory"
	"github.com/uptaris/uptaris/backend/internal/observability"
	"github.com/uptaris/uptaris/backend/internal/users"
	_ "github.com/uptaris/uptaris/backend/openapi"
	"github.com/uptaris/uptaris/backend/routers"
)

func main() { os.Exit(run()) }

func run() int {
	cfg := config.Load()
	logger := observability.NewLogger(cfg.Environment, os.Stdout, nil)

	if cfg.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:            cfg.SentryDSN,
			Environment:    cfg.Environment,
			SendDefaultPII: false,
			BeforeSend:     observability.ScrubEvent,
		}); err != nil {
			logger.Error("initialize sentry", "error", err)
			return 1
		}
		defer sentry.Flush(2 * time.Second)
	}

	if cfg.SentryDSN != "" {
		logger = observability.NewLogger(cfg.Environment, os.Stdout, sentry.CurrentHub())
	}
	slog.SetDefault(logger)
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		logger.Error("open database", "error_type", fmt.Sprintf("%T", err))
		return 1
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("get database connection")
		return 1
	}
	defer sqlDB.Close()
	httpHandlers := handlers.New(auth.New(db, cfg), inventory.New(db), users.New(db), cfg, sqlDB.PingContext)
	router := routers.Configure(cfg, httpHandlers, logger)
	server := &http.Server{
		Addr:              net.JoinHostPort(cfg.Host, cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		logger.Info("API listening", "address", server.Addr, "environment", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-stop:
	case err := <-serveErr:
		logger.Error("serve API", "error", err)
		return 1
	}
	defer signal.Stop(stop)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("shutdown API", "error", err)
	}
	return 0
}
