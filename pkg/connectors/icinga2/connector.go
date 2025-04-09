package icinga2

import (
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
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
	ignoredHosts := make(map[string]bool)

	for _, host := range hosts {
		host := host.Host
		ignoredHosts[host.DisplayName] = false

		if host.Acknowledgement > 0 {
			ignoredHosts[host.DisplayName] = true
			continue
		} else if !host.EnableNotifications {
			ignoredHosts[host.DisplayName] = true
			continue
		} else if host.DowntimeDepth > 0 {
			ignoredHosts[host.DisplayName] = true
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
			State:       fromHostState(host.State),
			Description: "Host down",
			Details:     host.Output,
			Links: []html.HTML{
				html.HTML("<a href=\"" + c.config.DashboardURL + "/dashboard#!/monitoring/host/show?host=" + host.DisplayName + "\" target=\"_blank\" alt=\"Home\">🏠</a>"),
			},
		}
		alert.Silence = c.createSilencer(alert)
		alerts = append(alerts, alert)
	}

	for _, service := range services {
		service := service.Service
		if ignore, ok := ignoredHosts[service.HostName]; ok && ignore {
			continue
		} else if service.Acknowledgement > 0 {
			continue
		} else if !service.EnableNotifications {
			continue
		} else if service.DowntimeDepth > 0 {
			continue
		} else if service.State == 0 {
			continue
		}

		var hostgroups []string
		if host, ok := hosts[service.HostName]; ok {
			hostgroups = host.Host.Groups
		}

		var links []html.HTML
		if service.NotesUrl != "" {
			links = append(links, html.HTML("<a href=\""+service.NotesUrl+"\" target=\"_blank\" alt=\"Runbook\">📖</a>"))
		}
		links = append(links, html.HTML("<a href=\""+c.config.DashboardURL+"/dashboard#!/monitoring/host/show?host="+service.HostName+"&service="+service.Name+"\" target=\"_blank\" alt=\"Home\">🏠</a>"))

		sec, dec := math.Modf(service.LastStateChange)
		alert := connectors.Alert{
			Labels: map[string]string{
				"Hostname":   service.HostName,
				"Zone":       service.Zone,
				"Source":     c.config.URL,
				"groups":     strings.Join(service.Groups, ","),
				"hostgroups": strings.Join(hostgroups, ","),
				"Type":       "Service",
			},
			Start:       time.Unix(int64(sec), int64(dec*(1e9))),
			State:       fromServiceState(service.State),
			Description: service.DisplayName,
			Details:     service.LastCheckResult.Output,
			Links:       links,
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

func (c *Connector) collectHosts(ctx context.Context) (map[string]HostAttrs, error) {
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

	results := make(map[string]HostAttrs)
	for _, host := range response.Results {
		results[host.Host.DisplayName] = host
	}

	return results, nil
}

func (c *Connector) get(endpoint string, ctx context.Context) (io.ReadCloser, error) {
	slog.DebugContext(ctx, "getting alerts", slog.String("url", c.config.URL+endpoint))

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

// see: https://icinga.com/docs/icinga-2/latest/doc/03-monitoring-basics/#check-result-state-mapping
func fromHostState(state int) connectors.State {
	switch state {
	case 0: // OK
		fallthrough
	case 1: // WARNING
		// both OK and WARNING mean that the host generally is considered UP.
		// However, what the API delivers and what the check result codes are
		// seem to differ.
		return connectors.Critical
	case 2: // CRITICAL
		fallthrough
	case 3: // UNKNOWN
		// bot CRITICAL and UNKNOWN are considered DOWN for hosts by icinga2.
		return connectors.Critical
	}
	return connectors.Unknown
}

func fromServiceState(state int) connectors.State {
	return connectors.State(state)
}
