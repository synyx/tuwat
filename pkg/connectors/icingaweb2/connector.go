package icingaweb2

import (
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
)

type Connector struct {
	config Config
	client *http.Client
}

type Config struct {
	Tag string
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
		ignoredHosts[host.DisplayName] = false

		state, _ := strconv.Atoi(host.State)
		if ack, _ := strconv.ParseBool(host.Acknowledgement); ack {
			ignoredHosts[host.DisplayName] = true
			continue
		} else if notification, _ := strconv.ParseBool(host.EnableNotifications); !notification {
			ignoredHosts[host.DisplayName] = true
			continue
		} else if downtime, _ := strconv.ParseBool(host.Downtime); downtime {
			ignoredHosts[host.DisplayName] = true
			continue
		} else if state == 0 {
			continue
		}

		var links []html.HTML
		links = append(links, html.HTML("<a href=\""+c.config.URL+"/dashboard#!/monitoring/host/show?host="+host.DisplayName+"\" target=\"_blank\" alt=\"Home\">🏠</a>"))

		unix, _ := strconv.ParseInt(host.LastStateChange, 10, 64)
		alert := connectors.Alert{
			Labels: map[string]string{
				"Hostname": host.DisplayName,
				"Source":   c.config.URL,
				"Type":     "Host",
			},
			Start:       time.Unix(unix, 0),
			State:       fromHostState(state),
			Description: "Host down",
			Details:     host.Output,
			Links:       links,
		}
		alerts = append(alerts, alert)
	}

	for _, service := range services {
		state, _ := strconv.Atoi(service.State)
		if ignore, ok := ignoredHosts[service.HostName]; ok && ignore {
			continue
		} else if ack, _ := strconv.ParseBool(service.Acknowledgement); ack {
			continue
		} else if notification, _ := strconv.ParseBool(service.EnableNotifications); !notification {
			continue
		} else if downtime, _ := strconv.ParseBool(service.Downtime); downtime {
			continue
		} else if state == 0 {
			continue
		}

		var links []html.HTML
		links = append(links, html.HTML("<a href=\""+c.config.URL+"/dashboard#!/monitoring/host/show?host="+service.HostName+"&service="+service.Name+"\" target=\"_blank\" alt=\"Home\">🏠</a>"))

		unix, _ := strconv.ParseInt(service.LastStateChange, 10, 64)
		alert := connectors.Alert{
			Labels: map[string]string{
				"Hostname": service.HostName,
				"Source":   c.config.URL,
				"Type":     "Service",
			},
			Start:       time.Unix(unix, 0),
			State:       fromServiceState(state),
			Description: service.DisplayName,
			Details:     service.LastCheckResult,
			Links:       links,
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Connector) String() string {
	return fmt.Sprintf("IcingaWeb2 (%s)", c.config.URL)
}

func (c *Connector) collectServices(ctx context.Context) ([]service, error) {
	body, err := c.get("/monitoring/list/services", ctx)
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

	return response, nil
}

func (c *Connector) collectHosts(ctx context.Context) (map[string]host, error) {
	body, err := c.get("/monitoring/list/hosts", ctx)
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

	results := make(map[string]host)
	for _, host := range response {
		results[host.DisplayName] = host
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
	defer res.Body.Close()

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
