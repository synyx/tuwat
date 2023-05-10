package web

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/web/actuator"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Handle(appCtx context.Context, cfg *config.Config, webHandler http.Handler) {

	muxer := newTracedMuxer()

	muxer.Handle("actuator", "/actuator/health", actuator.HealthAggregator)
	muxer.Handle("actuator", "/actuator/info", actuator.NewVersionHandler())
	muxer.Handle("actuator", "/actuator/prometheus", promhttp.Handler())
	muxer.Handle("static", "/static/", http.StripPrefix("/static", newNoListingFileServer(cfg)))
	muxer.Handle("web", "/", webHandler)

	Serve(appCtx, cfg.WebAddr, muxer.handler)
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
