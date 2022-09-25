package patchman

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/synyx/gonagdash/pkg/connectors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
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
	var response []Host
	next := "/api/host/"

	for next != "" {
		body, err := c.get(next, ctx)
		if err != nil {
			return nil, err
		}
		defer body.Close()

		decoder := json.NewDecoder(body)

		// read open bracket
		t, err := decoder.Token()
		if err != nil {
			otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
		}

		if d, ok := t.(json.Delim); ok && d == '{' {
			// Paging necessary
		pageHandler:
			for t, err := decoder.Token(); err == nil; t, err = decoder.Token() {
				if err != nil {
					otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
				}

				if s, ok := t.(string); ok && s == "next" {
					s, err := decoder.Token()
					if s, ok := s.(string); ok && err == nil {
						u, _ := url.Parse(s)
						next = "/api/host/?" + u.RawQuery
					} else {
						next = ""
					}
				}

				if s, ok := t.(string); ok && s == "results" {
					t, err := decoder.Token()
					if d, ok := t.(json.Delim); ok && d != '[' {
						otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
					}
					break pageHandler
				}
			}
		} else {
			next = ""
		}

		// while the array contains values
		for decoder.More() {
			var h Host
			// decode an array value (Message)
			err := decoder.Decode(&h)
			if err != nil {
				otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
			}

			response = append(response, h)
		}

		// read closing bracket
		t, err = decoder.Token()
		if err != nil {
			otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
			return nil, err
		}

		otelzap.Ctx(ctx).Info("Would pull next", zap.String("url", next))
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
