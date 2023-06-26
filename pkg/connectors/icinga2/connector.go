package icinga2

import (
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type Connector struct {
	config Config
	client *http.Client
}

type Config struct {
	Tag          string
	DashboardURL string
	common.HTTPConfig
}

func NewConnector(cfg *Config) *Connector {
	return &Connector{*cfg, cfg.HTTPConfig.Client()}
}

func (c *Connector) Tag() string {
	return c.config.Tag
}

func (c *Connector) Collect(ctx context.Context) ([]connectors.Alert, error) {
	hosts, err := c.collectHosts(ctx)
	if err != nil {
		return nil, err
	}
	services, err := c.collectServices(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert
	problemHosts := make(map[string]bool)

	for _, host := range hosts {
		host := host.Host
		if host.Acknowledgement > 0 {
			continue
		} else if !host.EnableNotifications {
			continue
		} else if host.DowntimeDepth > 0 {
			continue
		} else if host.State == 0 {
			continue
		}

		sec, dec := math.Modf(host.LastStateChange)
		alert := connectors.Alert{
			Labels: map[string]string{
				"Hostname": host.DisplayName,
				"Source":   c.config.URL,
				"groups":   strings.Join(host.Groups, ","),
				"Type":     "Host",
			},
			Start:       time.Unix(int64(sec), int64(dec*(1e9))),
			State:       connectors.State(host.State),
			Description: "Host down",
			Details:     host.Output,
			Links: []html.HTML{
				html.HTML("<a href=\"" + c.config.DashboardURL + "/dashboard#!/monitoring/host/show?host=" + host.DisplayName + "\" target=\"_blank\" alt=\"Home\">🏠</a>"),
			},
		}
		alert.Silence = c.createSilencer(alert)
		alerts = append(alerts, alert)
		problemHosts[host.DisplayName] = true
	}

	for _, service := range services {
		service := service.Service
		if x, ok := problemHosts[service.HostName]; ok && x {
			continue
		} else if !service.EnableNotifications {
			continue
		} else if service.DowntimeDepth > 0 {
			continue
		} else if service.State == 0 {
			continue
		}

		sec, dec := math.Modf(service.LastStateChange)
		alert := connectors.Alert{
			Labels: map[string]string{
				"Hostname": service.HostName,
				"Zone":     service.Zone,
				"Source":   c.config.URL,
				"groups":   strings.Join(service.Groups, ","),
				"Type":     "Service",
			},
			Start:       time.Unix(int64(sec), int64(dec*(1e9))),
			State:       connectors.State(service.State),
			Description: service.DisplayName,
			Details:     service.LastCheckResult.Output,
			Links: []html.HTML{
				html.HTML("<a href=\"" + c.config.DashboardURL + "/dashboard#!/monitoring/host/show?host=" + service.HostName + "&service=" + service.Name + "\" target=\"_blank\" alt=\"Home\">🏠</a>"),
			},
		}
		alert.Silence = c.createSilencer(alert)
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Connector) String() string {
	return fmt.Sprintf("Icinga2 (%s)", c.config.URL)
}

func (c *Connector) collectServices(ctx context.Context) ([]serviceAttrs, error) {
	body, err := c.get("/v1/objects/services", ctx)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	decoder := json.NewDecoder(body)

	var response ServiceResponse
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}

	return response.Results, nil
}

func (c *Connector) collectHosts(ctx context.Context) ([]HostAttrs, error) {
	body, err := c.get("/v1/objects/hosts", ctx)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	decoder := json.NewDecoder(body)

	var response HostResponse
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}

	return response.Results, nil
}

func (c *Connector) get(endpoint string, ctx context.Context) (io.ReadCloser, error) {
	otelzap.Ctx(ctx).Debug("getting alerts", zap.String("url", c.config.URL+endpoint))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		return res.Body, nil
	}

	if ct := res.Header.Get("Content-Type"); ct == "application/json" {
		e := struct {
			Error  int    `json:"error"`
			Status string `json:"status"`
		}{}
		decoder := json.NewDecoder(res.Body)

		err = decoder.Decode(&e)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get: %s", e.Status)
	}

	return nil, fmt.Errorf("failed to get, unknown status code: %d", res.StatusCode)
}
