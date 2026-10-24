package logging

import "context"

type telemetrySuppressedKey struct{}

// WithoutTelemetry marks a context whose logs must not be exported to
// OpenTelemetry. It is intended for high-frequency operational probes.
func WithoutTelemetry(ctx context.Context) context.Context {
	return context.WithValue(ctx, telemetrySuppressedKey{}, true)
}

func telemetrySuppressed(ctx context.Context) bool {
	suppressed, _ := ctx.Value(telemetrySuppressedKey{}).(bool)
	return suppressed
}
