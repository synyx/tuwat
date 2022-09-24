package actuator

import (
	"context"
	"time"
)

const checkInterval = time.Second * 10

type HealthCheck func(ctx context.Context) (status Status, message string)

type healthAccumulator struct {
	healthChecks map[string]HealthCheck
}

func NewHealthAccumulator() *healthAccumulator {
	return &healthAccumulator{
		healthChecks: make(map[string]HealthCheck),
	}
}

func (h *healthAccumulator) Register(component string, check HealthCheck) {
	h.healthChecks[component] = check
}

func (h *healthAccumulator) Run(appCtx context.Context) {
	tick := time.NewTicker(checkInterval)
	defer tick.Stop()

checker:
	for {
		ctx, cancel := context.WithTimeout(appCtx, checkInterval/2)

		for component, check := range h.healthChecks {
			status, message := check(ctx)
			SetHealth(component, status, message)
		}

		select {
		case <-appCtx.Done():
			cancel()
			break checker
		case <-tick.C:
			cancel()
			continue
		}
	}
}
