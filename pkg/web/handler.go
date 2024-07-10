package web

import (
	"context"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/web/common"
)

func Handle(appCtx context.Context, cfg *config.Config, webHandler, alertmanagerApi http.Handler) {
	muxer := newTracedMuxer()

	muxer.Handle("static", "/static/", http.StripPrefix("/static", newNoListingFileServer()))
	muxer.Handle("api", "/api/alertmanager/", http.StripPrefix("/api/alertmanager", alertmanagerApi))
	muxer.Handle("web", "/", webHandler)

	common.Serve(appCtx, cfg.WebAddr, muxer.handler)
}

type tracedMuxer struct {
	handler *http.ServeMux
}

func newTracedMuxer() tracedMuxer {
	return tracedMuxer{http.NewServeMux()}
}

func (t tracedMuxer) Handle(name, pattern string, handler http.Handler) {
	handler = otelhttp.WithRouteTag(pattern, handler)
	handler = otelhttp.NewHandler(handler, name)
	t.handler.Handle(pattern, handler)
}
