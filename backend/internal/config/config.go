package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment        string
	Host               string
	Port               string
	DatabaseURL        string
	SentryDSN          string
	CORSOrigins        []string
	JWTAccessSecret    string
	RefreshTokenPepper string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	CookieSecure       bool
	CookieSameSite     string
	TrustedProxies     []string
}

func Load() Config {
	loadEnvironment()
	cfg := Config{
		Environment:        env("APP_ENV", "development"),
		Host:               os.Getenv("API_HOST"),
		Port:               env("API_PORT", "8080"),
		DatabaseURL:        env("DATABASE_URL", "postgres://uptaris:uptaris@localhost:5432/uptaris?sslmode=disable"),
		SentryDSN:          os.Getenv("SENTRY_DSN"),
		CORSOrigins:        splitCSV(env("CORS_ORIGINS", "http://localhost:5173")),
		JWTAccessSecret:    env("JWT_ACCESS_SECRET", "development-only-change-me"),
		RefreshTokenPepper: env("REFRESH_TOKEN_PEPPER", "development-only-change-me"),
		AccessTokenTTL:     duration("ACCESS_TOKEN_TTL", "15m"),
		RefreshTokenTTL:    duration("REFRESH_TOKEN_TTL", "720h"),
		CookieSecure:       env("COOKIE_SECURE", "true") == "true",
		CookieSameSite:     env("COOKIE_SAME_SITE", "lax"),
		TrustedProxies:     splitCSV(os.Getenv("TRUSTED_PROXIES")),
	}
	if err := cfg.Validate(); err != nil {
		panic(err)
	}
	return cfg
}

func duration(key, fallback string) time.Duration {
	value, err := time.ParseDuration(env(key, fallback))
	if err != nil {
		panic("invalid duration for " + key)
	}
	return value
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

// Validate rejects unsafe production authentication and browser-origin settings.
func (c Config) Validate() error {
	if c.Environment != "development" && c.Environment != "test" && c.Environment != "production" {
		return fmt.Errorf("APP_ENV must be development, test, or production")
	}
	if c.AccessTokenTTL <= 0 || c.AccessTokenTTL > time.Hour || c.RefreshTokenTTL <= c.AccessTokenTTL {
		return fmt.Errorf("access TTL must be positive and <=1h; refresh TTL must exceed access TTL")
	}
	if c.CookieSameSite != "lax" && c.CookieSameSite != "strict" && c.CookieSameSite != "none" {
		return fmt.Errorf("COOKIE_SAME_SITE must be lax, strict, or none")
	}
	if c.CookieSameSite == "none" && !c.CookieSecure {
		return fmt.Errorf("SameSite=None requires COOKIE_SECURE=true")
	}
	if len(c.CORSOrigins) == 0 {
		return fmt.Errorf("CORS_ORIGINS must contain explicit origins")
	}
	for _, origin := range c.CORSOrigins {
		u, err := url.Parse(origin)
		if err != nil || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" || (u.Scheme != "http" && u.Scheme != "https") || strings.Contains(origin, "*") {
			return fmt.Errorf("CORS_ORIGINS must contain exact HTTP(S) origins")
		}
		if c.Environment == "production" && u.Scheme != "https" {
			return fmt.Errorf("production CORS origins require HTTPS")
		}
	}
	if c.Environment == "production" {
		for _, value := range []string{c.JWTAccessSecret, c.RefreshTokenPepper} {
			if len(value) < 32 || strings.Contains(strings.ToLower(value), "change-me") || strings.Contains(strings.ToLower(value), "development") {
				return fmt.Errorf("production auth secrets must be independent random values of at least 32 bytes")
			}
		}
		if c.JWTAccessSecret == c.RefreshTokenPepper {
			return fmt.Errorf("production auth secrets must differ")
		}
		if !c.CookieSecure {
			return fmt.Errorf("production requires COOKIE_SECURE=true")
		}
	}
	return nil
}
