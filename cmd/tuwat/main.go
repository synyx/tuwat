package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/benbjohnson/clock"

	"github.com/synyx/tuwat/pkg/aggregation"
	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/log"
	"github.com/synyx/tuwat/pkg/version"
	"github.com/synyx/tuwat/pkg/web"
	"github.com/synyx/tuwat/pkg/web/actuator"
	"github.com/synyx/tuwat/pkg/web/api/alertmanager"
)

func main() {
	appCtx := ApplicationContext()
	cfg, err := config.NewConfiguration()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if cfg.PrintVersion {
		fmt.Println(version.Info.HumanReadable())
		return
	}

	log.Initialize(cfg)
	log.InitializeTracer(appCtx, cfg)

	clk := clock.New()
	aggregator := aggregation.NewAggregator(cfg, clk)
	webHandler := web.NewWebHandler(cfg, aggregator)
	alertmanagerApi := alertmanager.ApiV2(cfg, aggregator)

	acc := actuator.NewHealthAccumulator(clk)
	acc.Register("aggregation", aggregation.NewAggregatorHealthCheck(aggregator))

	go web.Handle(appCtx, cfg, webHandler, alertmanagerApi)
	go aggregator.Run(appCtx)
	go acc.Run(appCtx)
	go actuator.Handle(appCtx, cfg)

	reconfigure := make(chan os.Signal, 1)
	signal.Notify(reconfigure, syscall.SIGHUP)
	for {
		select {
		case <-reconfigure:
			slog.InfoContext(appCtx, "Rereading configuration")

			cfg, err := config.NewConfiguration()
			if err != nil {
				slog.ErrorContext(appCtx, "Failed to read new configuration", slog.Any("error", err))
				break
			}

			aggregator.Reconfigure(cfg)
		case <-appCtx.Done():
			slog.InfoContext(appCtx, "Exiting")
			return
		}
	}
}
