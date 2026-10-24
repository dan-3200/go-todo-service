package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"app/internal/infrastructure/config"
)

func TestFromEnvironmentLoadsAllSettings(t *testing.T) {
	t.Setenv("API_PORT", "9000")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD", "password")
	t.Setenv("POSTGRES_HOST", "host")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_DB", "todos")

	settings := config.FromEnvironment()

	require.Equal(t, "9000", settings.Server.Port)
	require.Equal(t, config.PostgresConfig{
		User:     "user",
		Password: "password",
		Host:     "host",
		Port:     "5433",
		Database: "todos",
	}, settings.Postgres)
}

func TestFromEnvironmentUsesDefaultAPIPort(t *testing.T) {
	t.Setenv("API_PORT", "")

	settings := config.FromEnvironment()

	require.Equal(t, "8200", settings.Server.Port)
}
