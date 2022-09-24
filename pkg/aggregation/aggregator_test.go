package aggregation

import (
	"context"
	"testing"
	"time"

	"github.com/synyx/gonagdash/pkg/config"
	"github.com/synyx/gonagdash/pkg/connectors"
)

func TestAggregation(t *testing.T) {
	collect := make(chan result)
	var results []result
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	connector := &mockConnector{}
	cfg := &config.Config{Connectors: []connectors.Connector{connector}}
	a := NewAggregator(cfg)
	go a.collect(ctx, collect)
	results = append(results, <-collect)

	if len(results) != 1 {
		t.Error()
	}
}

type mockConnector struct {
}

func (m *mockConnector) Name() string {
	return "mock"
}

func (m *mockConnector) Collect(ctx context.Context) ([]connectors.Alert, error) {
	alerts := []connectors.Alert{
		{
			Tags: map[string]string{
				"Hostname": "kubernetes/k8s-apps",
			},
			Description: "Service Down",
			Start:       time.Now().Add(-1 * time.Minute),
			State:       connectors.Warning,
		}, {
			Tags: map[string]string{
				"Hostname": "nagios",
			},
			Description: "Weird",
			Start:       time.Now().Add(-2 * time.Hour),
			State:       connectors.Unknown,
		}, {
			Tags: map[string]string{
				"Hostname": "gitlab",
			},
			Description: "MR !272",
			Start:       time.Now().Add(-25 * time.Hour * 24),
			State:       connectors.Critical,
		},
	}
	return alerts, nil
}
