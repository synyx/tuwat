package log

import (
	"context"
	"io"
	"log/slog"
	"net/url"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/version"
)

func InitializeTracer(appCtx context.Context, cfg *config.Config) trace.Tracer {
	var tp *tracesdk.TracerProvider

	if cfg.OtelUrl != "" {
		tp = otelHttpTracer(appCtx, cfg)
	} else {
		tp = noopTracer()
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.TraceContext{},
		propagation.Baggage{}),
	)

	go tracerShutdown(appCtx, tp)

	return tp.Tracer("")
}

func stdoutTracer(cfg *config.Config) (tp *tracesdk.TracerProvider) {
	exporter, err := stdout.New()
	if err != nil {
		slog.Error("creating new stdout tracer", slog.Any("error", err))
		os.Exit(1)
	}

	return tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithSyncer(exporter),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(version.Info.Application),
			semconv.ServiceVersionKey.String(version.Info.Version),
			attribute.String("environment", cfg.Environment),
			attribute.String("instance", cfg.Instance),
		)),
	)
}

func noopTracer() *tracesdk.TracerProvider {
	exporter, err := stdout.New(stdout.WithWriter(io.Discard))
	if err != nil {
		slog.Error("creating noop tracer", slog.Any("error", err))
		os.Exit(1)
	}

	return tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.NeverSample()),
		tracesdk.WithBatcher(exporter),
	)
}

func otelHttpTracer(ctx context.Context, cfg *config.Config) *tracesdk.TracerProvider {
	u, err := url.Parse(cfg.OtelUrl)
	if err != nil {
		slog.Error("creating OTLP trace exporter", slog.Any("error", err))
		os.Exit(1)
	}

	options := []otlptracehttp.Option{otlptracehttp.WithEndpoint(u.Host)}
	if u.Scheme == "http" {
		options = append(options, otlptracehttp.WithInsecure())
	}
	if u.Path != "" && u.Path != "/" {
		options = append(options, otlptracehttp.WithURLPath(u.Path))
	}
	headers := map[string]string{
		"User-Agent": version.Info.Application + "/" + version.Info.Version,
	}
	options = append(options, otlptracehttp.WithHeaders(headers))

	exporter, err := otlptracehttp.New(ctx, options...)
	if err != nil {
		slog.Error("creating OTLP trace exporter", slog.Any("error", err))
		return noopTracer()
	}

	return tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exporter),
		// Record information about this application in a Resource.
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(version.Info.Application),
			semconv.ServiceVersionKey.String(version.Info.Version),
			attribute.String("environment", cfg.Environment),
			attribute.String("instance", cfg.Instance),
		)),
	)
}

func tracerShutdown(appCtx context.Context, tp *tracesdk.TracerProvider) {
	<-appCtx.Done()

	// Do not make the application hang when it is shutdown.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	if err := tp.Shutdown(ctx); err != nil {
		slog.Error("shutting down tracer", slog.Any("error", err))
		os.Exit(1)
	}
}
