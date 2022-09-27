package aggregation

import (
	"context"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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

	current       Aggregate
	connectors    []connectors.Connector
	whereTempl    *template.Template
	registrations map[any]chan<- bool
	mu            *sync.RWMutex // Protecting Registrations
	cmu           *sync.RWMutex // Protecting Configuration
	amu           *sync.RWMutex // Protecting current Aggregate
}

type result struct {
	collector string
	alerts    []connectors.Alert
	error     error
}

var (
	regCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "gonagdash_aggregator_registrations",
		Help: "Currently registered aggregation client.",
	})
)

func init() {
	prometheus.MustRegister(regCount)
}

func NewAggregator(cfg *config.Config) *Aggregator {
	return &Aggregator{
		interval:      1 * time.Minute,
		connectors:    cfg.Connectors,
		whereTempl:    cfg.WhereTemplate,
		registrations: make(map[any]chan<- bool),
		mu:            new(sync.RWMutex),
		cmu:           new(sync.RWMutex),
		amu:           new(sync.RWMutex),
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

	a.cmu.RLock()
	collectors := a.connectors
	a.cmu.RUnlock()

	for _, c := range collectors {
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

	a.cmu.RLock()
	whereTempl := a.whereTempl
	a.cmu.RUnlock()

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
			err := whereTempl.ExecuteTemplate(buf, "where", al)
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

	a.amu.Lock()
	a.current = Aggregate{
		CheckTime: time.Now(),
		Alerts:    alerts,
	}
	a.amu.Unlock()

	a.notify(ctx)
}

func (a *Aggregator) Alerts() Aggregate {
	a.amu.RLock()
	defer a.amu.RUnlock()

	return a.current
}

func (a *Aggregator) Register(handler any) <-chan bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	if r, ok := a.registrations[handler]; ok {
		close(r)
		delete(a.registrations, handler)
	}

	r := make(chan bool, 1)
	a.registrations[handler] = r
	return r
}

func (a *Aggregator) Unregister(handler any) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if r, ok := a.registrations[handler]; ok {
		close(r)
		delete(a.registrations, handler)
	}
}

func (a *Aggregator) notify(ctx context.Context) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	otelzap.Ctx(ctx).Debug("Notifying", zap.Any("count", len(a.registrations)))

	var toUnregister []any

	for thing, r := range a.registrations {
		select {
		case r <- true:
			otelzap.Ctx(ctx).Debug("Notified", zap.Any("thing", thing))
		case <-time.After(1 * time.Second):
			toUnregister = append(toUnregister, thing)
		}
	}

	for _, thing := range toUnregister {
		otelzap.Ctx(ctx).Debug("Force unregistering", zap.Any("thing", thing))
		a.Unregister(thing)
	}

	regCount.Set(float64(len(a.registrations)))
}

func (a *Aggregator) Reconfigure(cfg *config.Config) {
	a.cmu.Lock()
	defer a.cmu.Unlock()

	a.connectors = cfg.Connectors
	a.whereTempl = cfg.WhereTemplate
}
