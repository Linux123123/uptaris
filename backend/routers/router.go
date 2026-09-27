package routers

import (
	"log/slog"
	"net/http"
	"time"

	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/uptaris/uptaris/backend/internal/config"
	"github.com/uptaris/uptaris/backend/internal/handlers"
	"github.com/uptaris/uptaris/backend/internal/middleware"
	"github.com/uptaris/uptaris/backend/internal/swagger"
)

// Configure connects middleware and route groups.
func Configure(cfg config.Config, httpHandlers *handlers.Handlers, logger *slog.Logger) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	if err := engine.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		panic(err)
	}

	engine.HandleMethodNotAllowed = true
	engine.Use(middleware.RequestLogger(logger), middleware.Recovery())
	if cfg.SentryDSN != "" {
		engine.Use(sentrygin.New(sentrygin.Options{Repanic: true}))
	}

	engine.Use(middleware.Security(cfg))
	engine.NoRoute(handlers.NotFound)
	engine.NoMethod(handlers.MethodNotAllowed)
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	engine.GET("/healthz", httpHandlers.Health)
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swagger.FilesHandler()))

	v1 := engine.Group("/api/v1")
	v1.Use(middleware.RequireJSON())
	registerRoutes(v1, httpHandlers)

	return engine
}

func registerRoutes(v1 *gin.RouterGroup, api *handlers.Handlers) {
	registerAuth(v1, api)
	registerAccounts(v1, api)
	registerSystem(v1, api)
	registerServers(v1, api)
	registerMonitors(v1, api)
	registerIncidents(v1, api)
	registerUsers(v1, api)
}
