package aggregation

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/synyx/tuwat/pkg/clock"
	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/log"
)

func TestAggregation(t *testing.T) {
	collect := make(chan result)
	var results []result
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	connector := &mockConnector{}
	cfg := &config.Config{Connectors: []connectors.Connector{connector}}
	a := NewAggregator(cfg, clock.NewClock())
	go a.collect(ctx, collect)

	select {
	case c := <-collect:
		results = append(results, c)
	case <-ctx.Done():
		t.Error("timeout waiting for results")
	}

	if len(results) != 1 {
		t.Error()
	}
}

func TestSkippingAggregation(t *testing.T) {
	collect := make(chan result)
	// close channel to produce a panic on test failure (should not even attempt to collect)
	close(collect)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	clk := newMockClock(time.Now())

	connector := &mockConnector{}
	cfg := &config.Config{Connectors: []connectors.Connector{connector}, Interval: 10 * time.Second}

	log.Initialize(cfg)

	a := NewAggregator(cfg, clk)

	// Use the aggregator, thus mark as accessed
	a.Alerts()

	// Test that the collection would happen, if we are inside a reasonable time since the last access
	// This test assumes that the code _will_ panic (as the channel handed-in is closed).
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		a.collect(ctx, collect)
	}()

	clk.Progress(cfg.Interval * 5)

	// after enough time, the collection should now run to completion without panicking due to the collection
	// channel already being closed.
	a.collect(ctx, collect)
}

type mockClock struct {
	now    time.Time
	timers []*struct {
		t  time.Time
		ch chan<- time.Time
	}
}

func newMockClock(t time.Time) *mockClock {
	return &mockClock{
		now: t,
		timers: make([]*struct {
			t  time.Time
			ch chan<- time.Time
		}, 0),
	}
}

func (c *mockClock) Progress(d time.Duration) {
	i := 0
	then := c.now.Add(d)
	for _, s := range c.timers {
		if s.t.Before(then) {
			s.ch <- s.t
			i++
		} else {
			break
		}
	}
	c.timers = c.timers[i:]
	c.now = then
}

func (c *mockClock) After(d time.Duration) <-chan time.Time {

	ch := make(chan time.Time)
	c.insertTimer(c.now.Add(d), ch)

	return ch
}

func (c *mockClock) insertTimer(t time.Time, ch chan<- time.Time) {
	i := sort.Search(len(c.timers), func(i int) bool {
		return c.timers[i].t.After(t)
	})
	c.timers = append(c.timers, nil)
	copy(c.timers[i+1:], c.timers[i:])

	c.timers[i] = &struct {
		t  time.Time
		ch chan<- time.Time
	}{
		t,
		ch,
	}
}

// NOTE: NewTicker ticks only once
func (c *mockClock) NewTicker(d time.Duration) *time.Ticker {
	ch := make(chan time.Time)
	c.insertTimer(c.now.Add(d), ch)

	return &time.Ticker{C: ch}
}

func (c *mockClock) Now() time.Time {
	return c.now
}

type mockConnector struct {
}

func (m *mockConnector) Tag() string {
	return "mock"
}

func (m *mockConnector) Collect(ctx context.Context) ([]connectors.Alert, error) {
	alerts := []connectors.Alert{
		{
			Labels: map[string]string{
				"Hostname": "kubernetes/k8s-apps",
			},
			Description: "Service Down",
			Start:       time.Now().Add(-1 * time.Minute),
			State:       connectors.Warning,
		}, {
			Labels: map[string]string{
				"Hostname": "nagios",
			},
			Description: "Weird",
			Start:       time.Now().Add(-2 * time.Hour),
			State:       connectors.Unknown,
		}, {
			Labels: map[string]string{
				"Hostname": "gitlab",
			},
			Description: "MR !272",
			Start:       time.Now().Add(-25 * time.Hour * 24),
			State:       connectors.Critical,
		},
	}
	return alerts, nil
}
