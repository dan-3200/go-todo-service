package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Postgres  PostgresConfig
	Telemetry TelemetryConfig
}

type ServerConfig struct {
	Port string
}

type PostgresConfig struct {
	User            string
	Password        string
	Host            string
	Port            string
	Database        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type TelemetryConfig struct {
	Enabled           bool
	ServiceName       string
	CollectorAddr     string
	CollectorInsecure bool
	ShutdownTimeout   time.Duration
	TracerRatio       float64
}

func (c TelemetryConfig) ShutdownTimeoutDuration() time.Duration {
	if c.ShutdownTimeout <= 0 {
		return 5 * time.Second
	}
	return c.ShutdownTimeout
}

func Load() (Config, error) {
	err := godotenv.Load()
	return FromEnvironment(), err
}

func FromEnvironment() Config {
	return Config{
		Server: ServerConfig{
			Port: environmentOrDefault("API_PORT", "8200"),
		},
		Postgres: PostgresConfig{
			User:            os.Getenv("POSTGRES_USER"),
			Password:        os.Getenv("POSTGRES_PASSWORD"),
			Host:            os.Getenv("POSTGRES_HOST"),
			Port:            os.Getenv("POSTGRES_PORT"),
			Database:        os.Getenv("POSTGRES_DB"),
			MaxOpenConns:    environmentPositiveInt("POSTGRES_MAX_OPEN_CONNS", 10),
			MaxIdleConns:    environmentNonNegativeInt("POSTGRES_MAX_IDLE_CONNS", 2),
			ConnMaxLifetime: environmentDuration("POSTGRES_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: environmentDuration("POSTGRES_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		Telemetry: TelemetryConfig{
			Enabled:           environmentBool("OTEL_ENABLED", false),
			ServiceName:       environmentOrDefault("OTEL_SERVICE_NAME", "go-todo-service"),
			CollectorAddr:     environmentOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
			CollectorInsecure: environmentBool("OTEL_COLLECTOR_INSECURE", true),
			ShutdownTimeout:   environmentDuration("OTEL_SHUTDOWN_TIMEOUT", 5*time.Second),
			TracerRatio:       environmentFloat("OTEL_TRACER_RATIO", 1.0),
		},
	}
}

func environmentBool(key string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

func environmentDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func environmentFloat(key string, fallback float64) float64 {
	value, err := strconv.ParseFloat(os.Getenv(key), 64)
	if err != nil || value <= 0 || value > 1 {
		return fallback
	}
	return value
}

func environmentPositiveInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func environmentNonNegativeInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func environmentOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
