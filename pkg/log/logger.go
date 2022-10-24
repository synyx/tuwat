package log

import (
	"fmt"
	"log"
	"os"

	"github.com/go-logr/stdr"
	"github.com/synyx/tuwat/pkg/config"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

func Initialize(cfg *config.Config) func() {
	var logger *zap.Logger
	var err error
	if cfg.Environment == "prod" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment(
			zap.WithCaller(true),
			zap.AddStacktrace(zap.ErrorLevel),
		)
	}
	if err != nil {
		fmt.Println("failed to set up logging system", err)
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

	otelLogger := otelzap.New(
		logger,
		otelzap.WithMinLevel(zap.DebugLevel),
		otelzap.WithCaller(true),
		otelzap.WithStackTrace(false),
		otelzap.WithTraceIDField(true),
	)
	reversionFunctions = append(reversionFunctions, otelzap.ReplaceGlobals(otelLogger))

	return revert(reversionFunctions)
}
