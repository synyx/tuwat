package orderview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"net/http"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"

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

		state := connectors.Warning
		switch sourceAlert.State {
		case 2:
			state = connectors.Critical
		case 1:
			state = connectors.Warning
		default:
			state = connectors.Unknown
		}

		alert := connectors.Alert{
			Labels: map[string]string{
				"Hostname": sourceAlert.Owner,
			},
			Start:       time.Unix(sourceAlert.Since, 0),
			State:       state,
			Description: sourceAlert.Message,
			Links:       []html.HTML{},
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Connector) String() string {
	return fmt.Sprintf("Orderview (%s)", c.config.URL)
}

func (c *Connector) collectAlerts(ctx context.Context) ([]ticket, error) {
	res, err := c.get(ctx, "/api/ticket.php")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, _ := io.ReadAll(res.Body)
	buf := bytes.NewBuffer(b)

	decoder := json.NewDecoder(buf)

	var response []ticket
	err = decoder.Decode(&response)
	if err != nil {
		otelzap.Ctx(ctx).DPanic("Cannot parse",
			zap.String("url", c.config.URL),
			zap.String("data", buf.String()),
			zap.Any("status", res.StatusCode),
			zap.Error(err))
		return nil, err
	}

	return response, nil
}

func (c *Connector) get(ctx context.Context, endpoint string) (*http.Response, error) {

	otelzap.Ctx(ctx).Debug("getting tickets", zap.String("url", c.config.URL+endpoint))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
