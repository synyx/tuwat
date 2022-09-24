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

	current    Aggregate
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
	var results []result
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

			a.aggregate(results)
			results = nil
		case result := <-collect:
			results = append(results, result)
		case <-ctx.Done():
			break run
		}
	}
	ticker.Stop()
}

func (a *Aggregator) aggregate(results []result) {
	var alerts []Alert
	for _, r := range results {
		for _, a := range r.alerts {
			alert := Alert{
				Where:  a.Tags["Hostname"],
				Tag:    r.collector,
				What:   a.Description,
				When:   time.Now().Sub(a.Start),
				Status: a.State.String(),
			}
			alerts = append(alerts, alert)
		}
	}

	a.current = Aggregate{
		CheckTime: time.Now(),
		Alerts:    alerts,
	}
}

func (a *Aggregator) Alerts() Aggregate {
	return a.current
}
