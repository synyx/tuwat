package actuator

import (
	"context"
	"time"

	"github.com/benbjohnson/clock"
)

const checkInterval = time.Second * 10

type HealthCheck func(ctx context.Context) (status Status, message string)

type HealthAccumulator struct {
	clock        clock.Clock
	healthChecks map[string]HealthCheck
}

func NewHealthAccumulator(clock clock.Clock) *HealthAccumulator {
	return &HealthAccumulator{
		clock:        clock,
		healthChecks: make(map[string]HealthCheck),
	}
}

func (h *HealthAccumulator) Register(component string, check HealthCheck) {
	h.healthChecks[component] = check
}

func (h *HealthAccumulator) Run(appCtx context.Context) {
	tick := h.clock.Ticker(checkInterval)
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
