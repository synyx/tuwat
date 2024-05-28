package web

import (
	"context"
	"net/http"

	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/web/common"
)

func Handle(appCtx context.Context, cfg *config.Config, webHandler, alertmanagerApi http.Handler) {
	muxer := http.NewServeMux()

	muxer.Handle("/static/", http.StripPrefix("/static", newNoListingFileServer(cfg)))
	muxer.Handle("/api/alertmanager/", http.StripPrefix("/api/alertmanager", alertmanagerApi))
	muxer.Handle("/", webHandler)

	common.Serve(appCtx, cfg.WebAddr, muxer)
}
