package graylog

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
)

type Connector struct {
	config Config
	client *http.Client
}

type Config struct {
	Tag     string
	Cluster string
	common.HTTPConfig
}

func NewConnector(cfg *Config) *Connector {
	c := &Connector{config: *cfg, client: cfg.HTTPConfig.Client()}

	return c
}

func (c *Connector) Tag() string {
	return c.config.Tag
}

func (c *Connector) Collect(ctx context.Context) ([]connectors.Alert, error) {
	sourceAlerts, err := c.collectAlerts(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert

	for _, sourceAlert := range sourceAlerts {

		alert := connectors.Alert{
			Labels:      nil,
			Start:       parseTime(sourceAlert.TriggeredAt),
			State:       0,
			Description: "",
			Details:     "",
			Links:       nil,
			Silence:     nil,
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Connector) String() string {
	return fmt.Sprintf("Alertmanager (%s)", c.config.URL)
}

func (c *Connector) collectAlerts(ctx context.Context) ([]alert, error) {
	res, err := c.get(ctx, "/api/streams/alerts/paginated")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)

	var response alertResult
	err = decoder.Decode(&response)
	if err != nil {
		slog.ErrorContext(ctx, "Cannot parse",
			slog.String("url", c.config.URL),
			slog.Any("status", res.StatusCode),
			slog.Any("error", err))
		return nil, err
	}

	return response.Alerts, nil
}

func (c *Connector) get(ctx context.Context, endpoint string) (*http.Response, error) {

	slog.DebugContext(ctx, "getting alerts", slog.String("url", c.config.URL+endpoint))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	// TODO: Use pagination, however, we're unlikely to hit this limit for unresolved alerts
	q := req.URL.Query()
	q.Set("skip", "0")
	q.Set("limit", "300")
	q.Set("state", "unresolved")
	req.URL.RawQuery = q.Encode()

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func parseTime(timeField string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999Z", timeField)
	if err != nil {
		return time.Time{}
	}
	return t
}
