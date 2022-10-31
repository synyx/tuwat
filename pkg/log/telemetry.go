package log

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/synyx/tuwat/pkg/config"
	propagation2 "github.com/synyx/tuwat/pkg/log/propagation"
	"github.com/synyx/tuwat/pkg/version"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

func InitializeTracer(appCtx context.Context, cfg *config.Config) trace.Tracer {
	var tp *tracesdk.TracerProvider

	if cfg.JaegerUrl != "" {
		tp = jaegerTracer(cfg)
	} else {
		tp = noopTracer()
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation2.SleuthTraceContext{},
		propagation2.SleuthCurrentTraceContext{},
		propagation2.B3TraceContext{},
		propagation2.XB3TraceContext{},
		propagation.TraceContext{},
		propagation.TraceContext{},
		propagation.Baggage{}),
	)

	go tracerShutdown(appCtx, tp)

	return tp.Tracer("")
}

func stdoutTracer(cfg *config.Config) (tp *tracesdk.TracerProvider) {
	exporter, err := stdout.New(stdout.WithoutTimestamps())
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	return tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exporter),
	)
}

func jaegerTracer(cfg *config.Config) *tracesdk.TracerProvider {
	// Create the Jaeger exporter
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(cfg.JaegerUrl)))
	if err != nil {
		log.Fatal(err)
	}

	return tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
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
		log.Fatal(err)
	}
}
