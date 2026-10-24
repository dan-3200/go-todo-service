package logging

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"
)

// TraceHandler decorates another slog handler with the active OpenTelemetry
// trace and span identifiers, when the log call carries a traced context.
type TraceHandler struct {
	next slog.Handler
}

func NewTraceHandler(next slog.Handler) *TraceHandler {
	return &TraceHandler{next: next}
}

func (h *TraceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *TraceHandler) Handle(ctx context.Context, record slog.Record) error {
	spanContext := trace.SpanContextFromContext(ctx)
	if spanContext.IsValid() {
		record.AddAttrs(
			slog.String("trace_id", spanContext.TraceID().String()),
			slog.String("span_id", spanContext.SpanID().String()),
		)
	}

	return h.next.Handle(ctx, record)
}

func (h *TraceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewTraceHandler(h.next.WithAttrs(attrs))
}

func (h *TraceHandler) WithGroup(name string) slog.Handler {
	return NewTraceHandler(h.next.WithGroup(name))
}
