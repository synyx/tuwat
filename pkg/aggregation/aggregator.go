package aggregation

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/synyx/gonagdash/pkg/config"
	"github.com/synyx/gonagdash/pkg/connectors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type Aggregate struct {
	CheckTime time.Time
	Alerts    []Alert
}

type Alert struct {
	Where   string
	Tag     string
	What    string
	Details string
	When    time.Duration
	Status  string
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
		interval:   1 * time.Minute,
		connectors: cfg.Connectors,
	}
}

func (a *Aggregator) Run(ctx context.Context) {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	collect := make(chan result, 20)
	var results []result

	otelzap.Ctx(ctx).Info("Collecting on Start")
	go a.collect(ctx, collect)

	for {
		select {
		case <-ticker.C:
			otelzap.Ctx(ctx).Info("Collecting")
			go a.collect(ctx, collect)
		case r, ok := <-collect:
			if !ok {
				a.aggregate(ctx, results)
				results = nil
				collect = make(chan result, 20)
			} else if ok {
				otelzap.Ctx(ctx).Info("Appending")
				results = append(results, r)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (a *Aggregator) collect(ctx context.Context, collect chan<- result) {
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(ctx, a.interval/2)
	defer cancel()

	for _, c := range a.connectors {
		otelzap.Ctx(ctx).Info("Adding collection", zap.String("collector", c.Name()))
		wg.Add(1)
		go func(c connectors.Connector) {
			defer wg.Done()

			alerts, err := c.Collect(ctx)
			otelzap.Ctx(ctx).Info("Collected alerts", zap.String("collector", c.Name()), zap.Int("count", len(alerts)), zap.Error(err))
			collect <- result{
				collector: c.Name(),
				alerts:    alerts,
				error:     err,
			}
		}(c)
	}
	otelzap.Ctx(ctx).Info("Waiting for collection end")
	wg.Wait()
	otelzap.Ctx(ctx).Info("Collection end")
	close(collect)
}

func (a *Aggregator) aggregate(ctx context.Context, results []result) {
	otelzap.Ctx(ctx).Info("Aggregating results", zap.Int("count", len(results)))

	var alerts []Alert
	for _, r := range results {
		for _, a := range r.alerts {
			alert := Alert{
				Where:   a.Tags["Hostname"],
				Tag:     r.collector,
				What:    a.Description,
				Details: a.Details,
				When:    time.Now().Sub(a.Start),
				Status:  a.State.String(),
			}
			alerts = append(alerts, alert)
		}
	}

	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].When < alerts[j].When
	})

	a.current = Aggregate{
		CheckTime: time.Now(),
		Alerts:    alerts,
	}
}

func (a *Aggregator) Alerts() Aggregate {
	return a.current
}
