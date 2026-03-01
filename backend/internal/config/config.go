// Package config loads and validates application configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration values for the API server.
// Values are populated by Load from environment variables.
type Config struct {
	// Port is the TCP port the HTTP server listens on. Defaults to "8080".
	Port string

	// DatabaseURL is the Postgres connection string. Required.
	DatabaseURL string

	// LogLevel controls the minimum log level. Defaults to "info".
	// Valid values: debug, info, warn, error.
	LogLevel string

	// CORSOrigins is the list of allowed cross-origin request origins.
	// Defaults to ["http://localhost:5173"] (Vite dev server).
	// Set CORS_ORIGINS to a comma-separated list to override.
	CORSOrigins []string

	// MaxBodyBytes is the maximum request body size in bytes.
	// Requests with a larger body are rejected with 413 before reaching a handler.
	// Defaults to 1 MiB. Set MAX_BODY_BYTES to override.
	MaxBodyBytes int64
}

// Load reads configuration from environment variables and returns a Config.
// Returns an error listing any required variables that are not set.
func Load() (Config, error) {
	cfg := Config{
		Port:         getEnv("PORT", "8080"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		CORSOrigins:  splitCSV(getEnv("CORS_ORIGINS", "http://localhost:5173")),
		MaxBodyBytes: getEnvInt64("MAX_BODY_BYTES", 1<<20),
	}

	var missing []string

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("required environment variables not set: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

// getEnv returns the value of the environment variable named by key,
// or fallback if the variable is not set or is empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// splitCSV splits a comma-separated string into a trimmed slice, ignoring empty entries.
func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// getEnvInt64 returns the env var named by key parsed as int64,
// or fallback if the variable is not set, empty, or not a valid integer.
func getEnvInt64(key string, fallback int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}
