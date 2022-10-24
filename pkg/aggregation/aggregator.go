package aggregation

import (
	"context"
	html "html/template"
	"regexp"
	"sort"
	"strings"
	"sync"
	text "text/template"
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
	Blocked   []BlockedAlert
}

type Alert struct {
	Id      string
	Where   string
	Tag     string
	What    string
	Details string
	When    time.Duration
	Status  string
	Links   []html.HTML
	Labels  map[string]string
	Silence connectors.SilencerFunc
}

type BlockedAlert struct {
	Alert
	Reason string
}

type Aggregator struct {
	interval time.Duration

	current       Aggregate
	connectors    []connectors.Connector
	whereTempl    *text.Template
	registrations map[any]chan<- bool
	mu            *sync.RWMutex // Protecting Registrations
	cmu           *sync.RWMutex // Protecting Configuration
	amu           *sync.RWMutex // Protecting current Aggregate
	blockRules    []config.Rule
}

type result struct {
	tag       string
	alerts    []connectors.Alert
	error     error
	connector connectors.Connector
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
		interval:   cfg.Interval,
		connectors: cfg.Connectors,
		whereTempl: cfg.WhereTemplate,
		blockRules: cfg.Filter,

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
	for _, c := range a.connectors {
		otelzap.Ctx(ctx).Debug("Adding collection", zap.String("tag", c.Tag()))
		wg.Add(1)
		go func(c connectors.Connector) {
			defer wg.Done()

			alerts, err := c.Collect(ctx)
			otelzap.Ctx(ctx).Info("Collected alerts", zap.String("tag", c.Tag()), zap.Int("count", len(alerts)), zap.Error(err))
			collect <- result{
				tag:       c.Tag(),
				alerts:    alerts,
				error:     err,
				connector: c,
			}
		}(c)
	}
	a.cmu.RUnlock()

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
	var blockedAlerts []BlockedAlert

	for _, r := range results {
		if r.error != nil {
			alert := Alert{
				Where:   "gonagdash",
				Tag:     r.tag,
				What:    "Collection Failure",
				Details: r.error.Error(),
				When:    0 * time.Second,
				Status:  connectors.Critical.String(),
			}
			alerts = append(alerts, alert)
		}

		for _, al := range r.alerts {
			labels := make(map[string]string)
			for k, v := range al.Labels {
				if v != "" {
					labels[k] = v
				}
			}

			where := labels["Hostname"]
			buf := new(strings.Builder)
			err := whereTempl.ExecuteTemplate(buf, "where", al)
			if err == nil {
				where = buf.String()
			}

			alert := Alert{
				Id:      connectors.RandomAlertId(),
				Where:   where,
				Tag:     r.tag,
				What:    al.Description,
				Details: al.Details,
				When:    time.Now().Sub(al.Start),
				Status:  al.State.String(),
				Links:   al.Links,
				Labels:  labels,
				Silence: al.Silence,
			}

			if alert.Silence != nil {
				alert.Links = append(alert.Links,
					html.HTML(`<form class="txtform" action="/alerts/`+alert.Id+`/silence" method="post"><button class="txtbtn" value="silence" type="submit">🔇</button></form>`))
			}

			if reason := a.allow(alert); reason == "" {
				alerts = append(alerts, alert)
			} else {
				blockedAlerts = append(blockedAlerts, BlockedAlert{Alert: alert, Reason: reason})
			}
		}
	}

	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].When < alerts[j].When
	})

	a.amu.Lock()
	a.current = Aggregate{
		CheckTime: time.Now(),
		Alerts:    alerts,
		Blocked:   blockedAlerts,
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
	a.blockRules = cfg.Filter
}

func (a *Aggregator) allow(alert Alert) string {
outside:
	for _, rule := range a.blockRules {
		if rule.What != nil && rule.What.MatchString(alert.What) {
			return rule.Description
		} else {
			res := make(map[string]*regexp.Regexp)
			for l, r := range rule.Labels {
				if x, ok := alert.Labels[l]; !ok {
					// if one label does not match, leave entry alone
					continue outside
				} else {
					res[x] = r
				}
			}
			for a, b := range res {
				if b.MatchString(a) {
					return rule.Description
				}
			}
		}
	}

	return ""
}

func (a *Aggregator) Silence(ctx context.Context, alertId, user string) {
	var alert Alert

	a.amu.RLock()
	for _, a := range a.current.Alerts {
		if a.Id == alertId {
			alert = a
			break
		}
	}
	a.amu.RUnlock()

	if alert.Silence != nil {
		alert.Silence(ctx, 24*time.Hour, user)
	}
}
