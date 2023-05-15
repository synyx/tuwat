package web

import (
	"context"
	"net/http"

	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/web/common"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Handle(appCtx context.Context, cfg *config.Config, webHandler http.Handler) {
	muxer := newTracedMuxer()

	muxer.Handle("static", "/static/", http.StripPrefix("/static", newNoListingFileServer(cfg)))
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
