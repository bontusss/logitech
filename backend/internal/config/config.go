package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the runtime configuration for the API. Values come from the
// environment; cmd/server loads backend/.env in local development.
type Config struct {
	Env           string
	Port          string
	AppURL        string
	DatabaseURL   string
	RedisURL      string
	JWTSecret     string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	CORSOrigins   []string
	CookieSecure  bool
	ExposeDevKeys bool
	UploadDir     string
	APIPublicURL  string
	AppDomain     string
}

const devJWTSecret = "dev-only-jwt-secret-change-before-production"

func Load() (*Config, error) {
	env := getenv("APP_ENV", "development")

	cfg := &Config{
		Env:          env,
		Port:         getenv("PORT", "8080"),
		AppURL:       getenv("APP_URL", "http://localhost:3000"),
		DatabaseURL:  getenv("DATABASE_URL", "postgres://app:app@localhost:5432/appdb?sslmode=disable"),
		RedisURL:     getenv("REDIS_URL", "redis://localhost:6380/0"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		AccessTTL:    durationEnv("ACCESS_TOKEN_TTL", 15*time.Minute),
		RefreshTTL:   durationEnv("REFRESH_TOKEN_TTL", 30*24*time.Hour),
		UploadDir:    getenv("UPLOAD_DIR", "uploads"),
		APIPublicURL: getenv("API_PUBLIC_URL", "http://localhost:8080"),
		AppDomain:    getenv("APP_DOMAIN", "localhost"),
	}

	if cfg.JWTSecret == "" {
		if env == "production" {
			return nil, fmt.Errorf("JWT_SECRET is required in production")
		}
		cfg.JWTSecret = devJWTSecret
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	origins := []string{cfg.AppURL, "http://localhost:3000", "http://127.0.0.1:3000"}
	if raw := os.Getenv("CORS_ORIGINS"); raw != "" {
		origins = splitAndTrim(raw)
	}
	cfg.CORSOrigins = origins

	secure, _ := strconv.ParseBool(getenv("COOKIE_SECURE", "false"))
	cfg.CookieSecure = secure || env == "production"
	// Verification and reset tokens are returned only outside production so
	// the flow can be exercised before a mail provider is connected.
	cfg.ExposeDevKeys = env != "production"

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func durationEnv(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func splitAndTrim(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
