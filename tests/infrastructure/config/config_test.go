package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"app/internal/infrastructure/config"
)

func TestFromEnvironmentLoadsAllSettings(t *testing.T) {
	t.Setenv("API_PORT", "9000")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("POSTGRES_USER", "user")
	t.Setenv("POSTGRES_PASSWORD", "password")
	t.Setenv("POSTGRES_HOST", "host")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_DB", "todos")
	t.Setenv("POSTGRES_MAX_OPEN_CONNS", "20")
	t.Setenv("POSTGRES_MAX_IDLE_CONNS", "5")
	t.Setenv("POSTGRES_CONN_MAX_LIFETIME", "1h")
	t.Setenv("POSTGRES_CONN_MAX_IDLE_TIME", "10m")
	t.Setenv("OTEL_ENABLED", "true")
	t.Setenv("OTEL_SERVICE_NAME", "todo-api")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "collector:4317")
	t.Setenv("OTEL_COLLECTOR_INSECURE", "false")
	t.Setenv("OTEL_SHUTDOWN_TIMEOUT", "10s")
	t.Setenv("OTEL_TRACER_RATIO", "0.25")

	settings := config.FromEnvironment()

	require.Equal(t, "9000", settings.Server.Port)
	require.Equal(t, config.AuthConfig{JWTSecret: "test-secret"}, settings.Auth)
	require.Equal(t, config.PostgresConfig{
		User:            "user",
		Password:        "password",
		Host:            "host",
		Port:            "5433",
		Database:        "todos",
		MaxOpenConns:    20,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 10 * time.Minute,
	}, settings.Postgres)
	require.Equal(t, config.TelemetryConfig{
		Enabled:           true,
		ServiceName:       "todo-api",
		CollectorAddr:     "collector:4317",
		CollectorInsecure: false,
		ShutdownTimeout:   10 * time.Second,
		TracerRatio:       0.25,
	}, settings.Telemetry)
}

func TestFromEnvironmentUsesDefaultAPIPort(t *testing.T) {
	t.Setenv("API_PORT", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("POSTGRES_MAX_OPEN_CONNS", "invalid")
	t.Setenv("POSTGRES_MAX_IDLE_CONNS", "-1")
	t.Setenv("POSTGRES_CONN_MAX_LIFETIME", "invalid")
	t.Setenv("POSTGRES_CONN_MAX_IDLE_TIME", "invalid")
	t.Setenv("OTEL_ENABLED", "invalid")
	t.Setenv("OTEL_SERVICE_NAME", "")
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_COLLECTOR_INSECURE", "invalid")
	t.Setenv("OTEL_SHUTDOWN_TIMEOUT", "invalid")
	t.Setenv("OTEL_TRACER_RATIO", "invalid")

	settings := config.FromEnvironment()

	require.Equal(t, "8200", settings.Server.Port)
	require.Empty(t, settings.Auth.JWTSecret)
	require.Equal(t, 10, settings.Postgres.MaxOpenConns)
	require.Equal(t, 2, settings.Postgres.MaxIdleConns)
	require.Equal(t, 30*time.Minute, settings.Postgres.ConnMaxLifetime)
	require.Equal(t, 5*time.Minute, settings.Postgres.ConnMaxIdleTime)
	require.False(t, settings.Telemetry.Enabled)
	require.Equal(t, "go-todo-service", settings.Telemetry.ServiceName)
	require.Equal(t, "localhost:4317", settings.Telemetry.CollectorAddr)
	require.True(t, settings.Telemetry.CollectorInsecure)
	require.Equal(t, 5*time.Second, settings.Telemetry.ShutdownTimeout)
	require.Equal(t, 1.0, settings.Telemetry.TracerRatio)
}
