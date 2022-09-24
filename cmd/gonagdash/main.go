package main

import (
	"github.com/synyx/gonagdash/pkg/config"
	"github.com/synyx/gonagdash/pkg/web"
)

func main() {
	appCtx := ApplicationContext()
	cfg := config.NewConfiguration()

	webHandler := web.WebHandler(cfg)

	go web.Handle(appCtx, cfg, webHandler)
	<-appCtx.Done()
}
