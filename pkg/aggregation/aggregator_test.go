package aggregation

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"

	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/log"
)

func TestAggregation(t *testing.T) {
	filter := config.Rule{
		Description: "Ignore MRs",
		Labels: map[string]config.RuleMatcher{
			"Hostname": config.ParseRuleMatcher("~= gitlab"),
		},
	}

	a := aggregator(config.Excluding, filter)
	aggregation := aggregate(a, t)
	if len(aggregation.Alerts) != 2 {
		t.Error("invalid shown", aggregation)
	}
}

func TestWhen(t *testing.T) {
	filter := config.Rule{
		Description: "Non-Escalated",
		When:        config.ParseRuleMatcher("< 86400"), // < 2d
		What:        config.ParseRuleMatcher(": Update"),
		Labels: map[string]config.RuleMatcher{
			"Type": config.ParseRuleMatcher("PullRequest"),
		},
	}

	a := aggregator(config.Excluding, filter)
	aggregation := aggregate(a, t)
	if len(aggregation.Blocked) != 1 {
		t.Error("invalid blocked", aggregation.Blocked)
	}
	if len(aggregation.Alerts) != 2 {
		t.Error("invalid shown", aggregation.Alerts)
	}
}

func aggregator(mode config.DashboardMode, filters ...config.Rule) *Aggregator {
	cfg, _ := config.NewConfiguration()
	log.Initialize(cfg)

	connector := &mockConnector{
		clock: clock.NewMock(),
	}
	cfg.Connectors = []connectors.Connector{connector}
	cfg.Dashboards = map[string]*config.Dashboard{
		"Home": {
			Name:   "Home",
			Mode:   mode,
			Filter: filters,
		},
	}

	return NewAggregator(cfg, connector.clock)
}

func aggregate(a *Aggregator, t *testing.T) Aggregate {
	collect := make(chan result)
	var results []result
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go a.collect(ctx, collect)

	select {
	case c := <-collect:
		results = append(results, c)
	case <-ctx.Done():
		t.Error("timeout waiting for results")
	}

	if len(results) != 1 {
		t.Error("Have", results)
	}
	if len(results[0].alerts) != 3 {
		t.Fatal("Make sure nr == number in mock Collect()", results)
	}

	a.aggregate(ctx, nil, a.dashboards["Home"], results)

	return a.current["Home"]
}

type mockConnector struct {
	clock clock.Clock
}

func (m *mockConnector) String() string {
	return "mock"
}

func (m *mockConnector) Tag() string {
	return "mock"
}

func (m *mockConnector) Collect(_ context.Context) ([]connectors.Alert, error) {
	alerts := []connectors.Alert{
		{
			Labels: map[string]string{
				"Hostname": "nagios",
				"Type":     "PullRequest",
			},
			Description: "MR !1: X: Update foo",
			Start:       m.clock.Now().Add(-3 * 24 * time.Hour),
			State:       connectors.Warning,
		}, {
			Labels: map[string]string{
				"Hostname": "nagios",
				"Type":     "PullRequest",
			},
			Description: "MR !2: Y: Update bar",
			Start:       m.clock.Now().Add(-2 * time.Hour),
			State:       connectors.Unknown,
		}, {
			Labels: map[string]string{
				"Hostname": "gitlab",
			},
			Description: "MR !272",
			Start:       m.clock.Now().Add(-25 * time.Hour * 24),
			State:       connectors.Critical,
		},
	}
	return alerts, nil
}
