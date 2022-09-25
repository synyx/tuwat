package patchman

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/synyx/gonagdash/pkg/connectors"
)

type Collector struct {
	config Config
}

type Config struct {
	Name string
	URL  string
}

func NewCollector(cfg Config) *Collector {
	return &Collector{cfg}
}

func (c *Collector) Name() string {
	return c.config.Name
}

func (c *Collector) Collect(ctx context.Context) ([]connectors.Alert, error) {
	hosts, err := c.collectHosts(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert

	for _, host := range hosts {
		if host.SecurityUpdateCount == 0 && host.RebootRequired == false {
			continue
		}

		last, err := time.Parse("2006-01-02T15:04:05", host.LastReport)
		if err != nil {
			otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
		}

		alert := connectors.Alert{
			Tags: map[string]string{
				"Hostname": host.Hostname,
			},
			Start:       last,
			State:       connectors.Critical,
			Description: "Host Security critical",
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Collector) collectHosts(ctx context.Context) ([]Host, error) {
	body, err := c.get("/api/host/", ctx)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	decoder := json.NewDecoder(body)

	var response []Host
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Collector) get(endpoint string, ctx context.Context) (io.ReadCloser, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}
