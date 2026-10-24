package logging

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/trace"
)

// New creates the application's structured logger. The handler can be replaced
// without changing callers that use slog.Logger.
func New(serviceName string) *slog.Logger {
	handler := NewMultiHandler(
		NewTraceHandler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})),
		NewOTelHandler(otel.Logger("app"), slog.LevelInfo),
	)

	return slog.New(handler).With(slog.String("service", serviceName))
}

// MultiHandler writes a log record to each enabled handler.
type MultiHandler struct {
	handlers []slog.Handler
}

func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return &MultiHandler{handlers: handlers}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, record slog.Record) error {
	var errs []error
	for _, handler := range h.handlers {
		if !handler.Enabled(ctx, record.Level) {
			continue
		}
		if err := handler.Handle(ctx, record); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for index, handler := range h.handlers {
		handlers[index] = handler.WithAttrs(attrs)
	}
	return NewMultiHandler(handlers...)
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for index, handler := range h.handlers {
		handlers[index] = handler.WithGroup(name)
	}
	return NewMultiHandler(handlers...)
}

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

// OTelHandler converts slog records to OpenTelemetry log records. The
// OpenTelemetry SDK associates the active trace/span from ctx during export.
type OTelHandler struct {
	logger otellog.Logger
	level  slog.Level
	attrs  []scopedAttr
	groups []string
}

type scopedAttr struct {
	attr   slog.Attr
	groups []string
}

func NewOTelHandler(logger otellog.Logger, level slog.Level) *OTelHandler {
	return &OTelHandler{logger: logger, level: level}
}

func (h *OTelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	if level < h.level {
		return false
	}
	return h.logger.Enabled(ctx, otellog.EnabledParameters{Severity: toOTelSeverity(level)})
}

func (h *OTelHandler) Handle(ctx context.Context, record slog.Record) error {
	otelRecord := otellog.Record{}
	otelRecord.SetTimestamp(record.Time)
	otelRecord.SetSeverity(toOTelSeverity(record.Level))
	otelRecord.SetSeverityText(record.Level.String())
	otelRecord.SetBody(attribute.StringValue(record.Message))
	otelRecord.AddAttributes(toOTelScopedAttributes(h.attrs)...)

	record.Attrs(func(attr slog.Attr) bool {
		otelRecord.AddAttributes(toOTelAttributes([]slog.Attr{attr}, h.groups)...)
		return true
	})

	h.logger.Emit(ctx, otelRecord)
	return nil
}

func (h *OTelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	combined := make([]scopedAttr, 0, len(h.attrs)+len(attrs))
	combined = append(combined, h.attrs...)
	for _, attr := range attrs {
		combined = append(combined, scopedAttr{
			attr:   attr,
			groups: append([]string{}, h.groups...),
		})
	}

	return &OTelHandler{
		logger: h.logger,
		level:  h.level,
		attrs:  combined,
		groups: h.groups,
	}
}

func (h *OTelHandler) WithGroup(name string) slog.Handler {
	groups := append([]string{}, h.groups...)
	if name != "" {
		groups = append(groups, name)
	}

	return &OTelHandler{
		logger: h.logger,
		level:  h.level,
		attrs:  h.attrs,
		groups: groups,
	}
}

func toOTelSeverity(level slog.Level) otellog.Severity {
	switch {
	case level < slog.LevelInfo:
		return otellog.SeverityDebug
	case level < slog.LevelWarn:
		return otellog.SeverityInfo
	case level < slog.LevelError:
		return otellog.SeverityWarn
	default:
		return otellog.SeverityError
	}
}

func toOTelAttributes(attrs []slog.Attr, groups []string) []attribute.KeyValue {
	values := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		appendOTelAttribute(&values, strings.Join(groups, "."), attr)
	}
	return values
}

func toOTelScopedAttributes(attrs []scopedAttr) []attribute.KeyValue {
	values := make([]attribute.KeyValue, 0, len(attrs))
	for _, scoped := range attrs {
		appendOTelAttribute(&values, strings.Join(scoped.groups, "."), scoped.attr)
	}
	return values
}

func appendOTelAttribute(values *[]attribute.KeyValue, prefix string, attr slog.Attr) {
	value := attr.Value.Resolve()
	key := joinAttributeKey(prefix, attr.Key)

	if value.Kind() == slog.KindGroup {
		for _, child := range value.Group() {
			appendOTelAttribute(values, key, child)
		}
		return
	}
	if key == "" {
		return
	}

	switch value.Kind() {
	case slog.KindBool:
		*values = append(*values, attribute.Bool(key, value.Bool()))
	case slog.KindInt64:
		*values = append(*values, attribute.Int64(key, value.Int64()))
	case slog.KindUint64:
		*values = append(*values, attribute.String(key, fmt.Sprint(value.Uint64())))
	case slog.KindFloat64:
		*values = append(*values, attribute.Float64(key, value.Float64()))
	case slog.KindString:
		*values = append(*values, attribute.String(key, value.String()))
	case slog.KindTime:
		*values = append(*values, attribute.String(key, value.Time().UTC().Format(time.RFC3339Nano)))
	case slog.KindDuration:
		*values = append(*values, attribute.String(key, value.Duration().String()))
	default:
		*values = append(*values, attribute.String(key, fmt.Sprint(value.Any())))
	}
}

func joinAttributeKey(prefix, key string) string {
	if prefix == "" {
		return key
	}
	if key == "" {
		return prefix
	}
	return prefix + "." + key
}
