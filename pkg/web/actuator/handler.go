package actuator

import (
	"context"
	"net/http"
	_ "net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/web/common"
)

var HealthAggregator *HealthActuator

func init() {
	HealthAggregator = newHealthActuator()
}

func SetHealth(check string, status Status, message string) {
	HealthAggregator.Set(check, status, message)
}

func Handle(ctx context.Context, cfg *config.Config) {
	// Use default serve mux, as the pprof /debug endpoints are registered there as well
	muxer := http.DefaultServeMux

	muxer.Handle("/actuator/health", HealthAggregator)
	muxer.Handle("/actuator/info", NewVersionHandler())
	muxer.Handle("/actuator/prometheus", promhttp.Handler())

	common.Serve(ctx, cfg.ManagementAddr, muxer)
}
