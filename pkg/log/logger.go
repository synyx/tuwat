package log

import (
	"log/slog"
	"os"

	"github.com/synyx/tuwat/pkg/config"
)

func Initialize(cfg *config.Config) {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})
	slog.SetDefault(slog.New(handler))

	slog.Info("initialized logger", slog.String("environment", cfg.Environment))
}
