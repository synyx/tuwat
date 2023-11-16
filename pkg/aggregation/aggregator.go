package aggregation

import (
	"context"
	"fmt"
	html "html/template"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	text "text/template"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
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
	clock    clock.Clock
	tracer   trace.Tracer

	connectors    []connectors.Connector
	whereTempl    *text.Template
	registrations sync.Map
	cmu           *sync.RWMutex // Protecting Configuration
	amu           *sync.RWMutex // Protecting current Aggregate
	current       map[string]Aggregate
	dashboards    map[string]*config.Dashboard

	lastAccess atomic.Value

	CheckTime time.Time
}

type result struct {
	tag       string
	alerts    []connectors.Alert
	error     error
	connector connectors.Connector
}

var (
	regCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tuwat_aggregator_registrations",
		Help: "Currently registered aggregation client.",
	})
)

func init() {
	prometheus.MustRegister(regCount)
}

func NewAggregator(cfg *config.Config, clock clock.Clock) *Aggregator {

	a := &Aggregator{
		interval:   cfg.Interval,
		connectors: cfg.Connectors,
		whereTempl: cfg.WhereTemplate,
		current:    make(map[string]Aggregate),
		dashboards: cfg.Dashboards,

		registrations: sync.Map{},
		cmu:           new(sync.RWMutex),
		amu:           new(sync.RWMutex),

		clock:  clock,
		tracer: otel.Tracer("aggregator"),
	}

	a.lastAccess.Store(clock.Now())
	if a.interval == 0 {
		a.interval = 1 * time.Minute
	}

	return a
}

func (a *Aggregator) nrRegistrations() int {
	nrRegistrations := 0
	a.registrations.Range(func(key, value any) bool {
		nrRegistrations += 1
		return true
	})
	return nrRegistrations
}

func (a *Aggregator) active() bool {
	if a.nrRegistrations() > 0 {
		// As long as there are open connections, we should be active
		return true
	} else if la := a.lastAccess.Load(); la == nil {
		// On startup, we should be active
		return true
	} else if t, ok := la.(time.Time); ok && t.Before(a.clock.Now().Add(-a.interval*3)) {
		// The last access was more than 3 intervals ago, we should be inactive
		return false
	}

	// The last access was very recent, we should be active
	return true
}

func (a *Aggregator) Run(ctx context.Context) {
	ticker := a.clock.Ticker(a.interval)
	defer ticker.Stop()

	collect := make(chan result, 20)
	var results []result

	otelzap.Ctx(ctx).Info("Collecting on Start")
	go a.collect(ctx, collect)

	active := true
	for {
		select {
		case <-ticker.C:
			if active && !a.active() {
				otelzap.Ctx(ctx).Info("Deactivating collection due to inactivity")
				active = false
				continue
			} else if !a.active() {
				otelzap.Ctx(ctx).Debug("Skipping collection")
				continue
			} else if !active && a.active() {
				otelzap.Ctx(ctx).Info("Reactivating collection due to activity")
				active = true
			}

			go a.collect(ctx, collect)
		case r, ok := <-collect:
			if !ok {
				for _, dashboard := range a.dashboards {
					a.aggregate(ctx, dashboard, results)
				}
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

	startTime := time.Now()
	ctx, span := a.tracer.Start(ctx, "collection", trace.WithTimestamp(startTime))

	ctx, cancel := context.WithTimeout(ctx, a.interval/2)
	defer cancel()

	a.cmu.RLock()
	for _, c := range a.connectors {
		otelzap.Ctx(ctx).Debug("Adding collection", zap.String("tag", c.Tag()))
		wg.Add(1)
		go func(c connectors.Connector) {
			defer wg.Done()

			alerts, err := c.Collect(ctx)

			// Be graceful on errors accessing the handed-in channel
			defer func() {
				if e := recover(); e != nil {
					err = fmt.Errorf("error delivering result %w", e.(error))
				}
				otelzap.Ctx(ctx).Info("Collected alerts",
					zap.String("collector", c.String()),
					zap.String("tag", c.Tag()),
					zap.Int("count", len(alerts)),
					zap.Error(err))
			}()

			r := result{
				tag:       c.Tag(),
				alerts:    alerts,
				error:     err,
				connector: c,
			}
			select {
			case collect <- r:
			// ok
			case <-ctx.Done():
				err = fmt.Errorf("timeout delivering result %w", err)
				break
			}
		}(c)
	}
	a.cmu.RUnlock()

	wg.Wait()
	otelzap.Ctx(ctx).Debug("Collection end")
	span.End(trace.WithTimestamp(time.Now()))
	close(collect)
}

func (a *Aggregator) aggregate(ctx context.Context, dashboard *config.Dashboard, results []result) {
	otelzap.Ctx(ctx).Info("Aggregating results", zap.String("dashboard", dashboard.Name), zap.Int("count", len(results)))

	a.cmu.RLock()
	whereTempl := a.whereTempl
	a.cmu.RUnlock()

	var alerts []Alert
	var blockedAlerts []BlockedAlert

	for _, r := range results {
		if r.error != nil {
			alert := Alert{
				Where:   "tuwat",
				Tag:     r.tag,
				What:    "Collection Failure on " + r.connector.String(),
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
				When:    a.clock.Now().Sub(al.Start),
				Status:  al.State.String(),
				Links:   al.Links,
				Labels:  labels,
				Silence: al.Silence,
			}

			if alert.Silence != nil {
				alert.Links = append(alert.Links,
					html.HTML(`<form class="txtform" action="/alerts/`+alert.Id+`/silence" method="post"><button class="txtbtn" value="silence" type="submit">🔇</button></form>`))
			}

			if reason := a.allow(dashboard, alert); reason == "" {
				alerts = append(alerts, alert)
			} else {
				blockedAlerts = append(blockedAlerts, BlockedAlert{Alert: alert, Reason: reason})
			}
		}
	}

	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].When < alerts[j].When
	})

	sort.Slice(blockedAlerts, func(i, j int) bool {
		return blockedAlerts[i].When < blockedAlerts[j].When
	})

	a.amu.Lock()
	a.CheckTime = a.clock.Now()
	a.current[dashboard.Name] = Aggregate{
		CheckTime: a.CheckTime,
		Alerts:    alerts,
		Blocked:   blockedAlerts,
	}
	a.amu.Unlock()

	a.notify(ctx)
}

func (a *Aggregator) Alerts(dashboardName string) Aggregate {
	a.lastAccess.Store(a.clock.Now())

	a.amu.RLock()
	defer a.amu.RUnlock()

	if db, ok := a.current[dashboardName]; ok {
		return db
	} else if len(a.current) == 1 && dashboardName == "" {
		// use the only current dashboard as default
		for _, db := range a.current {
			return db
		}
	}

	return Aggregate{}
}

func (a *Aggregator) Register(handler string) <-chan bool {

	if r, ok := a.registrations.LoadAndDelete(handler); ok {
		close(r.(chan bool))
	}

	r := make(chan bool, 1)
	a.registrations.Store(handler, r)

	regCount.Set(float64(a.nrRegistrations()))

	return r
}

func (a *Aggregator) Unregister(handler string) {
	if r, ok := a.registrations.LoadAndDelete(handler); ok {
		close(r.(chan bool))
	}

	regCount.Set(float64(a.nrRegistrations()))
}

func (a *Aggregator) notify(ctx context.Context) {
	otelzap.Ctx(ctx).Debug("Notifying", zap.Any("count", a.nrRegistrations()))

	var toUnregister []string

	a.registrations.Range(func(key, value any) bool {
		r := value.(chan bool)
		select {
		case r <- true:
			otelzap.Ctx(ctx).Debug("Notified", zap.Any("client", key))

			a.lastAccess.Store(a.clock.Now())
		default:
			toUnregister = append(toUnregister, key.(string))
		}
		return true
	})

	for _, thing := range toUnregister {
		otelzap.Ctx(ctx).Debug("Force unregistering", zap.Any("client", thing))
		a.Unregister(thing)
	}
}

func (a *Aggregator) Reconfigure(cfg *config.Config) {
	a.cmu.Lock()
	defer a.cmu.Unlock()

	a.connectors = cfg.Connectors
	a.whereTempl = cfg.WhereTemplate
	a.dashboards = cfg.Dashboards
}

// allow will match rules against the ruleset.
func (a *Aggregator) allow(dashboard *config.Dashboard, alert Alert) string {
	reason := a.matchAlertWithReason(dashboard, alert)

	switch dashboard.Mode {
	case config.Including:
		// Revert logic when the dashboard configuration is in `including` mode.
		if reason == "" {
			return "Unmatched"
		} else {
			return ""
		}
	case config.Excluding:
		return reason
	}
	panic("unknown mode: " + dashboard.Mode.String())
}

// matchAlertWithReason will match anything which does match against any of the
// configured rules.
func (a *Aggregator) matchAlertWithReason(dashboard *config.Dashboard, alert Alert) string {
nextRule:
	for _, rule := range dashboard.Filter {
		matchers := make(map[string]config.RuleMatcher)

		// if it's a rule working on top level concepts:
		if rule.What != nil && rule.What.MatchString(alert.What) {
			// `what` contains a description what is being alerted and should be a
			// human understandable description.  The rule simply matches against
			// that.
			matchers[alert.What] = rule.What
		} else if rule.When != nil {
			// `when` is a duration, which is converted to seconds.  The rule simply matches against
			// that.
			seconds := strconv.FormatFloat(alert.When.Seconds(), 'f', 0, 64)
			matchers[seconds] = rule.When
		}

		// Test if any of the labels are applicable to the given alert
		for l, r := range rule.Labels {
			if x, ok := alert.Labels[l]; !ok {
				// if the label does not exist on the alert, it cannot match
				// thus skip this rule and try the next one
				continue nextRule
			} else {
				matchers[x] = r
			}
		}

		// If all the applicable matchers return a match, this rule matches,
		// meaning the rules are combined via `AND`.
		matchCount := 0
		for val, matcher := range matchers {
			if matcher.MatchString(val) {
				matchCount++
			}
		}
		if matchCount == len(matchers) {
			return rule.Description
		}
	}

	return ""
}

func (a *Aggregator) Silence(ctx context.Context, alertId, user string) {
	var alert Alert

	a.amu.RLock()
all:
	for _, dashboard := range a.dashboards {
		for _, a := range a.current[dashboard.Name].Alerts {
			if a.Id == alertId {
				alert = a
				break all
			}
		}
	}
	a.amu.RUnlock()

	if alert.Silence != nil {
		if err := alert.Silence(ctx, 24*time.Hour, user); err != nil {
			otelzap.Ctx(ctx).Info("error silencing", zap.Error(err))
		}
	}
}
