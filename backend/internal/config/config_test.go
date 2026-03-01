package config_test

import (
	"testing"

	"github.com/pkordes/rv-logbook/backend/internal/config"
	"github.com/stretchr/testify/require"
)

// TestLoad_defaults verifies that optional env vars fall back to their defaults
// when only the required DATABASE_URL is provided.
func TestLoad_defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://rvlogbook:rvlogbook@localhost:5432/rvlogbook")
	t.Setenv("PORT", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("CORS_ORIGINS", "")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "8080", cfg.Port)
	require.Equal(t, "info", cfg.LogLevel)
	require.Equal(t, "postgres://rvlogbook:rvlogbook@localhost:5432/rvlogbook", cfg.DatabaseURL)
	require.Equal(t, []string{"http://localhost:5173"}, cfg.CORSOrigins)
}

// TestLoad_overrides verifies that all values can be overridden via env vars.
func TestLoad_overrides(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@db:5432/mydb")
	t.Setenv("PORT", "9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("CORS_ORIGINS", "https://app.example.com, https://admin.example.com")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "9090", cfg.Port)
	require.Equal(t, "debug", cfg.LogLevel)
	require.Equal(t, "postgres://user:pass@db:5432/mydb", cfg.DatabaseURL)
	require.Equal(t, []string{"https://app.example.com", "https://admin.example.com"}, cfg.CORSOrigins)
}

// TestLoad_missingRequired verifies that an error is returned when DATABASE_URL
// is not set, and that the error message names the missing variable.
func TestLoad_missingRequired(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := config.Load()

	require.Error(t, err)
	require.ErrorContains(t, err, "DATABASE_URL")
}
