package log

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"runtime"

	"github.com/go-slog/otelslog"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/embedded"
	otelglobal "go.opentelemetry.io/otel/log/global"

	"github.com/synyx/tuwat/pkg/config"
)

var cleanFileRE = regexp.MustCompile("(.*/|^)(pkg|cmd)(.*)")

func Initialize(cfg *config.Config) {
	handler := otelslog.NewHandler(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.SourceKey:
				source, _ := a.Value.Any().(*slog.Source)
				if source != nil {
					source.File = cleanFileRE.ReplaceAllString(source.File, "$2$3")
				}
			case "error":
				if a.Value.String() == "<nil>" {
					return slog.Attr{}
				}
			}
			return a
		},
	}))
	provider := newSlogProvider(handler)
	otelglobal.SetLoggerProvider(provider)
	slog.SetDefault(slog.New(handler))

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
		if v.Value.Kind() == otellog.KindEmpty {
			return true
		}
		fields = append(fields, slog.Any(v.Key, v.Value))
		return true
	})

	msg := slogBody(record.Body())

	var pcs [1]uintptr
	runtime.Callers(5, pcs[:]) // skip [Callers, Emit, slogotel, ..]
	r := slog.NewRecord(record.Timestamp(), slogLevel(record.Severity()), msg, pcs[0])
	r.AddAttrs(fields...)
	_ = s.handler.Handle(ctx, r)
}

func (s SlogLogger) Enabled(_ context.Context, _ otellog.EnabledParameters) bool {
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
