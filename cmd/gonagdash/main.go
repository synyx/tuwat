package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/synyx/gonagdash/pkg/aggregation"
	"github.com/synyx/gonagdash/pkg/buildinfo"
	"github.com/synyx/gonagdash/pkg/config"
	"github.com/synyx/gonagdash/pkg/log"
	"github.com/synyx/gonagdash/pkg/web"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

func main() {
	appCtx := ApplicationContext()
	cfg, err := config.NewConfiguration()
	if err != nil {
		panic(err)
	}

	if cfg.PrintVersion {
		fmt.Print(buildinfo.Version)
		if buildinfo.GitSHA != "" {
			fmt.Printf(" (%s)", buildinfo.GitSHA)
		}
		fmt.Println()
		return
	}

	log.Initialize(cfg)
	log.InitializeTracer(appCtx, cfg)

	aggregator := aggregation.NewAggregator(cfg)
	webHandler := web.WebHandler(cfg, aggregator)

	go web.Handle(appCtx, cfg, webHandler)
	go aggregator.Run(appCtx)

	reconfigure := make(chan os.Signal, 1)
	signal.Notify(reconfigure, syscall.SIGHUP)
	for {
		select {
		case <-reconfigure:
			otelzap.Ctx(appCtx).Info("Rereading configuration")

			cfg, err := config.NewConfiguration()
			if err != nil {
				otelzap.Ctx(appCtx).Error("Failed to read new configuration", zap.Error(err))
			}

			aggregator.Reconfigure(cfg)
		case <-appCtx.Done():
			otelzap.Ctx(appCtx).Info("Exiting")
			return
		}
	}
}
