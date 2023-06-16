package aggregation

import (
	"context"

	"github.com/synyx/tuwat/pkg/web/actuator"
)

func NewAggregatorHealthCheck(aggregator *Aggregator) actuator.HealthCheck {
	return func(ctx context.Context) (status actuator.Status, message string) {
		if aggregator.CheckTime.IsZero() {
			return actuator.OutOfService, "STARTING"
		}
		if aggregator.active() {
			// Have requests, but no checks, stuck
			if aggregator.CheckTime.Before(aggregator.clock.Now().Add(-aggregator.interval * 3)) {
				return actuator.Down, "OLD"
			}
			return actuator.Up, "OK"
		}
		return actuator.Unknown, "INACTIVE"
	}
}
