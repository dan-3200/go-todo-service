package logging

import (
	"log/slog"
	"os"

	"go.opentelemetry.io/otel"
)

// New creates the application's structured logger. Logs are written as JSON
// to stdout and exported through the globally configured OTel logger provider.
func New(serviceName string) *slog.Logger {
	handler := NewMultiHandler(
		NewTraceHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
		NewOTelHandler(otel.Logger("app"), slog.LevelInfo),
	)

	return slog.New(handler).With(slog.String("service", serviceName))
}
