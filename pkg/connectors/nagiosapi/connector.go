package nagiosapi

import (
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
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
	Tag       string
	NagiosURL string
	common.HTTPConfig
}

func NewConnector(cfg *Config) *Connector {
	return &Connector{*cfg, cfg.HTTPConfig.Client()}
}

func (c *Connector) Tag() string {
	return c.config.Tag
}

func (c *Connector) Collect(ctx context.Context) ([]connectors.Alert, error) {
	content, err := c.collectHosts(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert
	for hostName, host := range content {
		if host.ProblemHasBeenAcknowledged == "1" {
			continue
		} else if host.NotificationsEnabled == "0" {
			continue
		} else if i, e := strconv.ParseInt(host.ScheduledDowntimeDepth, 10, 32); e == nil && i > 0 {
			continue
		} else if host.CurrentState != "0" {
			state, err := strconv.ParseInt(host.CurrentState, 10, 32)
			if err != nil {
				slog.ErrorContext(ctx, "Cannot parse", slog.Any("error", err))
			}
			stateChange, err := strconv.ParseInt(host.LastStateChange, 10, 64)
			if err != nil {
				slog.ErrorContext(ctx, "Cannot parse", slog.Any("error", err))
			}

			alert := connectors.Alert{
				Labels: map[string]string{
					"Hostname": hostName,
					"Source":   c.config.URL,
					"Type":     "Host",
				},
				Start:       time.Unix(stateChange, 0),
				State:       fromHostState(state),
				Description: "Host down",
				Details:     host.PluginOutput,
				Links: []html.HTML{
					html.HTML("<a href=\"" + c.config.NagiosURL + "/cgi-bin/extinfo.cgi?type=1&host=" + hostName + "\" target=\"_blank\" alt=\"Home\">🏠</a>"),
				},
			}
			alert.Silence = c.createSilencer(alert)
			alerts = append(alerts, alert)
			continue
		}

		for serviceName, service := range host.Services {
			if service.CurrentState == "0" {
				continue
			} else if service.NotificationsEnabled == "0" {
				continue
			} else if service.ProblemHasBeenAcknowledged == "1" {
				continue
			} else if i, e := strconv.ParseInt(service.ScheduledDowntimeDepth, 10, 32); e == nil && i > 0 {
				continue
			}

			state, err := strconv.ParseInt(service.CurrentState, 10, 32)
			if err != nil {
				slog.ErrorContext(ctx, "Cannot parse", slog.Any("error", err))
			}
			stateChange, err := strconv.ParseInt(service.LastStateChange, 10, 64)
			if err != nil {
				slog.ErrorContext(ctx, "Cannot parse", slog.Any("error", err))
			}

			alert := connectors.Alert{
				Labels: map[string]string{
					"Hostname": hostName,
					"Source":   c.config.URL,
					"Type":     "Service",
				},
				Start:       time.Unix(stateChange, 0),
				State:       fromServiceState(state),
				Description: serviceName,
				Details:     service.PluginOutput,
				Links: []html.HTML{
					html.HTML("<a href=\"" + c.config.NagiosURL + "/cgi-bin/extinfo.cgi?type=2&host=" + hostName + "&service=" + serviceName + "\" target=\"_blank\" alt=\"Home\">🏠</a>"),
				},
			}
			alert.Silence = c.createSilencer(alert)
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

func (c *Connector) String() string {
	return fmt.Sprintf("NagiosAPI (%s)", c.config.URL)
}

func (c *Connector) collectHosts(ctx context.Context) (map[string]host, error) {
	slog.DebugContext(ctx, "getting alerts", slog.String("url", c.config.URL+"/state"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+"/state", nil)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)

	var response response
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}
	if !response.Success {
		// TODO(jo): the `content` map is overloaded with an error string
		return nil, fmt.Errorf("API failure response")
	}

	return response.Content, nil
}

// see https://assets.nagios.com/downloads/nagioscore/docs/nagioscore/4/en/hostchecks.html
// Host State Determination
func fromHostState(state int64) connectors.State {
	switch state {
	case 0: // OK
		return connectors.OK
	case 1: // WARNING
		// assumption: 0 = Don't use aggressive host checking (default)
		// this means we should treat this as OK, however, be cautious.
		return connectors.Warning
	case 2: // CRITICAL
		fallthrough
	case 3: // UNKNOWN
		// within nagios the DOWN state can map to different states (DOWN,
		// UNREACHABLE) based on host dependencies.  As tuwat does not model
		// alert dependencies currently, both UNREACHABLE (All parents are
		// either DOWN or UNREACHABLE) and DOWN map to critical.
		return connectors.Critical
	}
	return connectors.Unknown
}

func fromServiceState(state int64) connectors.State {
	return connectors.State(state)
}
