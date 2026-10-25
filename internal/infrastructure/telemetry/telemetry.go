package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"google.golang.org/grpc/credentials/insecure"

	"app/internal/infrastructure/config"
)

const (
	metricExportInterval   = 5 * time.Second
	defaultShutdownTimeout = 5 * time.Second
	initializerTimeout     = 5 * time.Second
	defaultTracerRatio     = 1.0
)

type TelemetryConfig struct {
	ServiceName       string
	CollectorAddr     string
	CollectorInsecure bool
	ShutdownTimeout   time.Duration
	TracerRatio       float64
}

func SetupTelemetry(ctx context.Context, cfg config.TelemetryConfig) (func() error, error) {
	if strings.TrimSpace(cfg.CollectorAddr) == "" {
		return nil, nil
	}

	return SetTelemetry(ctx, TelemetryConfig{
		ServiceName:       cfg.ServiceName,
		CollectorAddr:     cfg.CollectorAddr,
		CollectorInsecure: cfg.CollectorInsecure,
		ShutdownTimeout:   cfg.ShutdownTimeoutDuration(),
		TracerRatio:       cfg.TracerRatio,
	})
}

func LogTelemetryInitialization(ctx context.Context, logger *slog.Logger, collectorAddr string, err error) {
	if logger == nil {
		return
	}

	if strings.TrimSpace(collectorAddr) == "" {
		logger.InfoContext(ctx, "OpenTelemetry desabilitado: collector não configurado")
		return
	}
	if err != nil {
		logger.WarnContext(ctx, "falha ao inicializar OpenTelemetry; continuando sem telemetria",
			slog.String("collector_addr", collectorAddr),
			slog.Any("error", err),
		)
	}
}

func SetTelemetry(ctx context.Context, settings TelemetryConfig) (func() error, error) {
	if strings.TrimSpace(settings.CollectorAddr) == "" {
		return nil, newOTelError("endereço do collector não definido", nil)
	}

	initCtx, cancelInit := context.WithTimeout(ctx, initializerTimeout)
	defer cancelInit()

	res, err := resource.New(initCtx, resource.WithAttributes(semconv.ServiceName(settings.ServiceName)))
	if err != nil {
		return nil, newOTelError("falha ao criar recurso OTel", err)
	}

	tracerProvider, err := setTracer(initCtx, settings.CollectorAddr, settings.CollectorInsecure, res, settings.TracerRatio)
	if err != nil {
		return nil, newOTelError("falha ao configurar tracer", err)
	}

	meterProvider, err := setMetric(initCtx, settings.CollectorAddr, settings.CollectorInsecure, res)
	if err != nil {
		_ = tracerProvider.Shutdown(initCtx)
		return nil, newOTelError("falha ao configurar métricas", err)
	}

	loggerProvider, err := setLogger(initCtx, settings.CollectorAddr, settings.CollectorInsecure, res)
	if err != nil {
		_ = meterProvider.Shutdown(initCtx)
		_ = tracerProvider.Shutdown(initCtx)
		return nil, newOTelError("falha ao configurar logger", err)
	}

	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	otel.SetLoggerProvider(loggerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	shutdownTimeout := settings.ShutdownTimeout
	if shutdownTimeout <= 0 {
		shutdownTimeout = defaultShutdownTimeout
	}

	cleanup := func() error {
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancelShutdown()

		if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
			return newOTelError("falha ao desligar o tracer provider", err)
		}
		if err := meterProvider.Shutdown(shutdownCtx); err != nil {
			return newOTelError("falha ao desligar o metric provider", err)
		}
		if err := loggerProvider.Shutdown(shutdownCtx); err != nil {
			return newOTelError("falha ao desligar o logger provider", err)
		}
		return nil
	}

	return cleanup, nil
}

func setTracer(ctx context.Context, collectorAddr string, collectorInsecure bool, res *resource.Resource, ratio float64) (*sdktrace.TracerProvider, error) {
	options := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(collectorAddr)}
	if collectorInsecure {
		options = append(options, otlptracegrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	exporter, err := otlptracegrpc.New(ctx, options...)
	if err != nil {
		return nil, err
	}
	if ratio <= 0 || ratio > 1 {
		ratio = defaultTracerRatio
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(ratio))),
	), nil
}

func setMetric(ctx context.Context, collectorAddr string, collectorInsecure bool, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	options := []otlpmetricgrpc.Option{otlpmetricgrpc.WithEndpoint(collectorAddr)}
	if collectorInsecure {
		options = append(options, otlpmetricgrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	exporter, err := otlpmetricgrpc.New(ctx, options...)
	if err != nil {
		return nil, err
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(metricExportInterval))),
	)
	if err := runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second)); err != nil {
		_ = meterProvider.Shutdown(ctx)
		return nil, err
	}

	return meterProvider, nil
}

func setLogger(ctx context.Context, collectorAddr string, collectorInsecure bool, res *resource.Resource) (*sdklog.LoggerProvider, error) {
	options := []otlploggrpc.Option{otlploggrpc.WithEndpoint(collectorAddr)}
	if collectorInsecure {
		options = append(options, otlploggrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	exporter, err := otlploggrpc.New(ctx, options...)
	if err != nil {
		return nil, err
	}

	return sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
	), nil
}

type otelError struct {
	message string
	err     error
}

func (e *otelError) Error() string {
	if e.err == nil {
		return e.message
	}
	return fmt.Sprintf("%s: %v", e.message, e.err)
}

func (e *otelError) Unwrap() error { return e.err }

func newOTelError(message string, err error) error {
	return &otelError{message: message, err: err}
}
