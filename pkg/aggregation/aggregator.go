package aggregation

import (
	"context"
	"sync"
	"time"

	"github.com/synyx/gonagdash/pkg/config"
	"github.com/synyx/gonagdash/pkg/connectors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
)

type Aggregate struct {
	CheckTime time.Time
	Alerts    []Alert
}

type Alert struct {
	Where  string
	Tag    string
	What   string
	When   time.Duration
	Status string
}

type Aggregator struct {
	interval time.Duration

	connectors []connectors.Connector
}

type result struct {
	collector string
	alerts    []connectors.Alert
	error     error
}

func NewAggregator(cfg *config.Config) *Aggregator {
	return &Aggregator{
		interval: 1 * time.Minute,
	}
}

func (a *Aggregator) Run(ctx context.Context) {
	ticker := time.NewTicker(a.interval)
	collect := make(chan result)
	results := make(map[string]result)

run:
	for {
		select {
		case <-ticker.C:
			var wg sync.WaitGroup

			otelzap.Ctx(ctx).Info("Collecting")

			for _, c := range a.connectors {
				go func(c connectors.Connector) {
					ctx, cancel := context.WithTimeout(ctx, a.interval/2)
					defer cancel()
					defer wg.Done()

					alerts, err := c.Collect(ctx)
					collect <- result{
						collector: c.Name(),
						alerts:    alerts,
						error:     err,
					}
				}(c)
			}
			wg.Wait()
		case result := <-collect:
			results[result.collector] = result
		case <-ctx.Done():
			break run
		}
	}
	ticker.Stop()
}

func (a *Aggregator) Collect() {

}

func (a *Aggregator) Alerts() Aggregate {
	return Aggregate{
		CheckTime: time.Now(),
		Alerts: []Alert{
			{
				Where:  "kubernetes/k8s-apps",
				Tag:    "synyx",
				What:   "MR !272",
				When:   1 * time.Minute,
				Status: "yellow",
			}, {
				Where:  "foo.synyx.coffee",
				Tag:    "prod",
				What:   "MR !272",
				When:   2 * time.Hour,
				Status: "gray",
			}, {
				Where:  "foo.contargo.net",
				Tag:    "RZ1",
				What:   "MR !272",
				When:   25 * time.Hour * 24,
				Status: "red",
			},
		},
	}
}
