package log

import (
	"fmt"
	"log"
	"os"

	"github.com/go-logr/stdr"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/synyx/tuwat/pkg/config"
)

func Initialize(cfg *config.Config) func() {
	logger, err := cfg.Logger.Build(
		zap.WithCaller(true),
		zap.AddStacktrace(zap.ErrorLevel),
	)
	if err != nil {
		fmt.Println("failed to set up logging system:", err)
		os.Exit(1)
	}

	var reversionFunctions []func()
	revert := func(reversionFunctions []func()) func() {
		return func() {
			for _, f := range reversionFunctions {
				f()
			}
		}
	}

	reversionFunctions = append(reversionFunctions, zap.ReplaceGlobals(logger))

	logrLogger := NewLogrZapBridge(logger)
	otel.SetLogger(logrLogger)
	reversionFunctions = append(reversionFunctions, func() { otel.SetLogger(stdr.New(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile))) })

	stdLogger := newStdLoggerBridge(logger)
	reversionFunctions = append(reversionFunctions, stdLogger.ReplaceGlobals())

	otelLogger := otelzap.New(
		logger,
		otelzap.WithMinLevel(zap.DebugLevel),
		otelzap.WithCaller(true),
		otelzap.WithStackTrace(false),
		otelzap.WithTraceIDField(true),
	)
	reversionFunctions = append(reversionFunctions, otelzap.ReplaceGlobals(otelLogger))

	otelLogger.Info("initialized logger", zap.String("environment", cfg.Environment))

	return revert(reversionFunctions)
}
