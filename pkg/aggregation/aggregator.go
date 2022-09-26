package aggregation

import (
	"context"
	"sort"
	"strings"
	"sync"
	"text/template"
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
	Links   map[string]string
}

type Aggregator struct {
	interval time.Duration

	current    Aggregate
	connectors []connectors.Connector
	whereTempl *template.Template
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
		whereTempl: cfg.WhereTemplate,
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
			go a.collect(ctx, collect)
		case r, ok := <-collect:
			if !ok {
				a.aggregate(ctx, results)
				results = nil
				collect = make(chan result, 20)
			} else if ok {
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
		otelzap.Ctx(ctx).Debug("Adding collection", zap.String("collector", c.Tag()))
		wg.Add(1)
		go func(c connectors.Connector) {
			defer wg.Done()

			alerts, err := c.Collect(ctx)
			otelzap.Ctx(ctx).Info("Collected alerts", zap.String("collector", c.Tag()), zap.Int("count", len(alerts)), zap.Error(err))
			collect <- result{
				collector: c.Tag(),
				alerts:    alerts,
				error:     err,
			}
		}(c)
	}
	wg.Wait()
	otelzap.Ctx(ctx).Debug("Collection end")
	close(collect)
}

func (a *Aggregator) aggregate(ctx context.Context, results []result) {
	otelzap.Ctx(ctx).Info("Aggregating results", zap.Int("count", len(results)))

	var alerts []Alert
	for _, r := range results {
		if r.error != nil {
			alert := Alert{
				Where:   "gonagdash",
				Tag:     r.collector,
				What:    "Collection Failure",
				Details: r.error.Error(),
				When:    0 * time.Second,
				Status:  connectors.Critical.String(),
			}
			alerts = append(alerts, alert)
		}

		for _, al := range r.alerts {
			where := al.Labels["Hostname"]
			buf := new(strings.Builder)
			err := a.whereTempl.ExecuteTemplate(buf, "where", al)
			if err == nil {
				where = buf.String()
			}

			alert := Alert{
				Where:   where,
				Tag:     r.collector,
				What:    al.Description,
				Details: al.Details,
				When:    time.Now().Sub(al.Start),
				Status:  al.State.String(),
				Links:   al.Links,
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
