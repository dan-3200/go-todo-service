package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"app/internal/infrastructure/logging"
)

func TestTraceHandlerAddsTraceAndSpanIDs(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(logging.NewTraceHandler(slog.NewJSONHandler(&output, nil)))

	provider := sdktrace.NewTracerProvider()
	ctx, span := provider.Tracer("test").Start(context.Background(), "operation")
	defer span.End()

	logger.InfoContext(ctx, "message")

	var record map[string]any
	require.NoError(t, json.Unmarshal(output.Bytes(), &record))
	require.NotEmpty(t, record["trace_id"])
	require.NotEmpty(t, record["span_id"])
}

func TestOTelHandlerEmitsSlogRecord(t *testing.T) {
	exporter := &logExporter{}
	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewSimpleProcessor(exporter)),
	)
	t.Cleanup(func() {
		require.NoError(t, provider.Shutdown(context.Background()))
	})

	logger := slog.New(logging.NewOTelHandler(provider.Logger("app"), slog.LevelInfo)).
		With("service", "todo-api").
		WithGroup("todo")

	logger.WarnContext(context.Background(), "todo inválido", slog.Int64("id", 42))

	require.Len(t, exporter.records, 1)
	record := exporter.records[0]
	require.Equal(t, "todo inválido", record.Body().AsString())
	require.Equal(t, "WARN", record.SeverityText())

	attributes := map[string]attribute.Value{}
	record.WalkAttributes(func(value attribute.KeyValue) bool {
		attributes[string(value.Key)] = value.Value
		return true
	})
	require.Equal(t, "todo-api", attributes["service"].AsString())
	require.Equal(t, int64(42), attributes["todo.id"].AsInt64())
}

type logExporter struct {
	records []sdklog.Record
}

func (e *logExporter) Export(_ context.Context, records []sdklog.Record) error {
	for _, record := range records {
		e.records = append(e.records, record.Clone())
	}
	return nil
}

func (*logExporter) ForceFlush(context.Context) error { return nil }

func (*logExporter) Shutdown(context.Context) error { return nil }
