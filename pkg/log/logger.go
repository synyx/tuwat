package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/embedded"
	"go.opentelemetry.io/otel/log/global"

	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/version"
)

func Initialize(cfg *config.Config) {
	provider := newSlogProvider(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	global.SetLoggerProvider(provider)
	slog.SetDefault(otelslog.NewLogger(version.Info.Application, otelslog.WithLoggerProvider(provider)))

	slog.Info("initialized logger", slog.String("environment", cfg.Environment))
}

type SlogProvider struct {
	embedded.LoggerProvider
	handler slog.Handler
}

func newSlogProvider(handler slog.Handler) SlogProvider {
	return SlogProvider{handler: handler}
}

func (s SlogProvider) Logger(name string, options ...otellog.LoggerOption) otellog.Logger {
	return SlogLogger{
		handler: s.handler,
		name:    name,
		options: options,
	}
}

// SlogLogger is an OpenTelemetry Logger
type SlogLogger struct {
	embedded.Logger
	handler slog.Handler
	name    string
	options []otellog.LoggerOption
}

func (s SlogLogger) Emit(ctx context.Context, record otellog.Record) {
	fields := make([]slog.Attr, 0, record.AttributesLen())
	record.WalkAttributes(func(v otellog.KeyValue) bool {
		if !v.Value.Empty() {
			if v.Value.Kind() == otellog.KindString && v.Value.String() == "<nil>" {
				return true
			}
			fields = append(fields, slog.Any(v.Key, v.Value))
		}
		return true
	})

	msg := slogBody(record.Body())

	var pcs [1]uintptr
	runtime.Callers(5, pcs[:]) // skip [Callers, Emit, slogotel, ..]
	r := slog.NewRecord(record.Timestamp(), slogLevel(record.Severity()), msg, pcs[0])
	r.AddAttrs(fields...)
	_ = s.handler.Handle(ctx, r)
}

func (s SlogLogger) Enabled(context.Context, otellog.Record) bool {
	return true
}

func slogLevel(severity otellog.Severity) slog.Level {
	return slog.Level(severity - 9)
}

func slogBody(value otellog.Value) string {
	switch value.Kind() {
	case otellog.KindString:
		return value.String()
	case otellog.KindEmpty:
		return ""
	default:
		return fmt.Sprintf("%v", value)
	}
}
