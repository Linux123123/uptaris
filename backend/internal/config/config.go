package config

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"slices"
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
	OAuthProviders     map[string]OAuthProviderConfig
	OAuthTokenKey      []byte
	TwoFactorKey       []byte
	FrontendURL        string
	WebAuthnRPID       string
}

type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
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
		OAuthProviders:     loadOAuthProviders(),
		OAuthTokenKey:      loadHexKey("OAUTH_TOKEN_ENCRYPTION_KEY"),
		TwoFactorKey:       loadHexKey("TWO_FACTOR_ENCRYPTION_KEY"),
		FrontendURL:        strings.TrimRight(env("FRONTEND_URL", "http://localhost:5173"), "/"),
		WebAuthnRPID:       os.Getenv("WEBAUTHN_RP_ID"),
	}

	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	return cfg
}

func loadHexKey(name string) []byte {
	key, err := hex.DecodeString(os.Getenv(name))
	if err != nil {
		panic(name + " must be hex encoded")
	}

	return key
}

func loadOAuthProviders() map[string]OAuthProviderConfig {
	ids := splitCSV(os.Getenv("OAUTH_PROVIDERS"))
	providers := make(map[string]OAuthProviderConfig, len(ids))
	for _, id := range ids {
		prefix := "OAUTH_" + strings.ToUpper(strings.ReplaceAll(id, "-", "_")) + "_"
		provider := OAuthProviderConfig{
			ClientID:     os.Getenv(prefix + "CLIENT_ID"),
			ClientSecret: os.Getenv(prefix + "CLIENT_SECRET"),
			RedirectURL:  os.Getenv(prefix + "REDIRECT_URL"),
		}

		providers[id] = provider
	}

	return providers
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
		if err != nil || u.Host == "" || u.User != nil ||
			u.RawQuery != "" || u.Fragment != "" || u.Path != "" ||
			(u.Scheme != "http" && u.Scheme != "https") || strings.Contains(origin, "*") {
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

	for id, provider := range c.OAuthProviders {
		if !validOAuthProviderID(id) {
			return fmt.Errorf("OAuth provider IDs must be lowercase URL-safe names")
		}

		if provider.ClientID == "" || provider.ClientSecret == "" || provider.RedirectURL == "" {
			return fmt.Errorf("OAuth provider %q requires client ID, secret, and redirect URL", id)
		}

		callback, err := url.Parse(provider.RedirectURL)
		if err != nil || callback.Host == "" || callback.User != nil ||
			callback.RawQuery != "" || callback.Fragment != "" ||
			(callback.Scheme != "http" && callback.Scheme != "https") {
			return fmt.Errorf("OAuth provider %q redirect URL must be absolute HTTP(S)", id)
		}

		if c.Environment == "production" && callback.Scheme != "https" {
			return fmt.Errorf("production OAuth redirect URLs require HTTPS")
		}
	}

	if len(c.OAuthProviders) > 0 && len(c.OAuthTokenKey) != 32 {
		return fmt.Errorf("OAUTH_TOKEN_ENCRYPTION_KEY must be 32 bytes encoded as 64 hex characters when OAuth providers are configured")
	}

	if len(c.TwoFactorKey) != 0 && len(c.TwoFactorKey) != 32 {
		return fmt.Errorf("TWO_FACTOR_ENCRYPTION_KEY must be 32 bytes encoded as 64 hex characters")
	}

	frontend, err := url.Parse(c.FrontendURL)
	if err != nil || frontend.Host == "" || frontend.User != nil ||
		frontend.RawQuery != "" || frontend.Fragment != "" || frontend.Path != "" ||
		(frontend.Scheme != "http" && frontend.Scheme != "https") {
		return fmt.Errorf("FRONTEND_URL must be an absolute HTTP(S) origin")
	}

	if c.Environment == "production" && frontend.Scheme != "https" {
		return fmt.Errorf("production FRONTEND_URL requires HTTPS")
	}

	if c.Environment == "production" && c.WebAuthnRPID == "" {
		return fmt.Errorf("production requires WEBAUTHN_RP_ID")
	}

	rpID := c.EffectiveWebAuthnRPID()
	host := strings.ToLower(frontend.Hostname())
	if rpID == "" || strings.ContainsAny(rpID, ":/*") ||
		(host != rpID && !strings.HasSuffix(host, "."+rpID)) {
		return fmt.Errorf("WEBAUTHN_RP_ID must be FRONTEND_URL hostname or a parent domain")
	}

	if !slices.Contains(c.CORSOrigins, c.FrontendURL) {
		return fmt.Errorf("FRONTEND_URL must appear in CORS_ORIGINS")
	}

	return nil
}

func (c Config) EffectiveWebAuthnRPID() string {
	if c.WebAuthnRPID != "" {
		return c.WebAuthnRPID
	}

	frontend, err := url.Parse(c.FrontendURL)
	if err != nil {
		return ""
	}

	return strings.ToLower(frontend.Hostname())
}

func validOAuthProviderID(id string) bool {
	if id == "" {
		return false
	}

	for _, char := range id {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' && char != '_' {
			return false
		}
	}

	return true
}
