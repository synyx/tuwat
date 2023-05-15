package web

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/web/actuator"
	"github.com/synyx/tuwat/pkg/web/actuator/pprofhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func Handle(appCtx context.Context, cfg *config.Config, webHandler http.Handler) {
	http.Handle()
	muxer := newTracedMuxer()

	muxer.Handle("actuator", "/actuator/health", actuator.HealthAggregator)
	muxer.Handle("actuator", "/actuator/info", actuator.NewVersionHandler())
	muxer.Handle("actuator", "/actuator/pprof/cmdline", http.HandlerFunc(pprofhttp.CmdlineHandler))
	muxer.Handle("actuator", "/actuator/pprof/profile", http.HandlerFunc(pprofhttp.ProfileHandler))
	muxer.Handle("actuator", "/actuator/pprof/symbol", http.HandlerFunc(pprofhttp.SymbolHandler))
	muxer.Handle("actuator", "/actuator/pprof/trace", http.HandlerFunc(pprofhttp.TraceHandler))
	muxer.Handle("actuator", "/actuator/pprof/", http.HandlerFunc(pprofhttp.IndexHandler))
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
