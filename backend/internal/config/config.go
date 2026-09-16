package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Port        string
	DatabaseURL string
	SentryDSN   string
	CORSOrigins []string
}

func Load() Config {
	loadEnvironment()
	return Config{
		Environment: env("APP_ENV", "development"),
		Port:        env("API_PORT", "8080"),
		DatabaseURL: env("DATABASE_URL", "postgres://uptaris:uptaris@localhost:5432/uptaris?sslmode=disable"),
		SentryDSN:   os.Getenv("SENTRY_DSN"),
		CORSOrigins: splitCSV(env("CORS_ORIGINS", "http://localhost:5173")),
	}
}

func loadEnvironment() {
	// Support `go run` from backend/ and binary execution from repository root.
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if origin := strings.TrimSpace(part); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
