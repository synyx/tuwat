package nagiosapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	connectors.HTTPConfig
}

func NewCollector(cfg Config) *Collector {
	return &Collector{cfg}
}

func (c *Collector) Name() string {
	return c.config.Name
}

func (c *Collector) Collect(ctx context.Context) ([]connectors.Alert, error) {
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
				otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
			}
			stateChange, err := strconv.ParseInt(host.LastStateChange, 10, 64)
			if err != nil {
				otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
			}

			alert := connectors.Alert{
				Tags: map[string]string{
					"Hostname": hostName,
				},
				Start:       time.Unix(stateChange, 0),
				State:       connectors.State(state),
				Description: "Host down",
				Details:     host.PluginOutput,
			}
			alerts = append(alerts, alert)
			continue
		}

		for serviceName, service := range host.Services {
			if service.CurrentState == "0" {
				continue
			} else if service.ProblemHasBeenAcknowledged == "1" {
				continue
			}

			state, err := strconv.ParseInt(service.CurrentState, 10, 32)
			if err != nil {
				otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
			}
			stateChange, err := strconv.ParseInt(service.LastStateChange, 10, 64)
			if err != nil {
				otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
			}

			alert := connectors.Alert{
				Tags: map[string]string{
					"Hostname": hostName,
				},
				Start:       time.Unix(stateChange, 0),
				State:       connectors.State(state),
				Description: serviceName,
				Details:     service.PluginOutput,
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

func (c *Collector) collectHosts(ctx context.Context) (map[string]Host, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+"/state", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)

	var response Response
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
