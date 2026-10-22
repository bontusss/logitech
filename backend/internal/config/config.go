package config

import (
	"os"
	"strings"

	"backend/internal/db"
)

// Config holds application configuration loaded from environment variables.
type Config struct {
	Port          string
	DatabaseURL   string
	RedisURL      string
	JWTSecret     string
	AdminUsername string
	AdminPassword string
	AdminEmail    string
	// StaffIDPrefix is the prefix used for auto-generated staff IDs
	// (e.g. "GAL" produces GAL-001, GAL-002, ...).
	StaffIDPrefix string
}

// Load loads configuration from environment variables with sensible defaults.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = db.DefaultDSN
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "logitech-default-development-secret-key-32bytes!!"
	}

	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		adminUsername = "admin"
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "adminpassword123"
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "admin@logitech.example"
	}

	staffIDPrefix := os.Getenv("STAFF_ID_PREFIX")
	if staffIDPrefix == "" {
		staffIDPrefix = "GAL"
	}
	staffIDPrefix = strings.ToUpper(strings.TrimSpace(staffIDPrefix))

	return &Config{
		Port:          port,
		DatabaseURL:   dbURL,
		RedisURL:      redisURL,
		JWTSecret:     jwtSecret,
		AdminUsername: adminUsername,
		AdminPassword: adminPassword,
		AdminEmail:    adminEmail,
		StaffIDPrefix: staffIDPrefix,
	}
}
