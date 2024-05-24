package log

import (
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/log/noop"

	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/version"
)

func Initialize(cfg *config.Config) {
	provider := noop.NewLoggerProvider()
	slog.SetDefault(otelslog.NewLogger(version.Info.Application, otelslog.WithLoggerProvider(provider)))

	slog.Info("initialized logger", slog.String("environment", cfg.Environment))
}
