package icinga2

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"time"

	"github.com/synyx/gonagdash/pkg/connectors"
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
			},
			Start:       time.Unix(int64(sec), int64(dec*(1e9))),
			State:       connectors.State(host.State),
			Description: "Host down",
			Details:     host.Output,
		}
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
			},
			Start:       time.Unix(int64(sec), int64(dec*(1e9))),
			State:       connectors.State(service.State),
			Description: service.DisplayName,
			Details:     service.LastCheckResult.Output,
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Collector) collectServices(ctx context.Context) ([]ServiceAttrs, error) {
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

func (c *Collector) collectHosts(ctx context.Context) ([]HostAttrs, error) {
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

func (c *Collector) get(endpoint string, ctx context.Context) (io.ReadCloser, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.config.Insecure},
	}
	client := &http.Client{Transport: tr}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}
