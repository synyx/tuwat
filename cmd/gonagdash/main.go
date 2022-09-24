package main

import (
	"fmt"

	"github.com/synyx/gonagdash/pkg/aggregation"
	"github.com/synyx/gonagdash/pkg/buildinfo"
	"github.com/synyx/gonagdash/pkg/config"
	"github.com/synyx/gonagdash/pkg/log"
	"github.com/synyx/gonagdash/pkg/web"
)

func main() {
	appCtx := ApplicationContext()
	cfg := config.NewConfiguration()

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
	<-appCtx.Done()
}
