package alertmanager

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/synyx/gonagdash/pkg/connectors"
	"golang.org/x/oauth2/clientcredentials"
)

type Collector struct {
	config Config
	oauth2 clientcredentials.Config
}

type Config struct {
	Name string
	URL  string

	// oauth2
	ClientId     string
	ClientSecret string
	TokenURL     string
}

func NewCollector(cfg Config) *Collector {
	oauth2 := clientcredentials.Config{
		ClientID:       cfg.ClientId,
		ClientSecret:   cfg.ClientSecret,
		TokenURL:       cfg.TokenURL,
		Scopes:         nil,
		EndpointParams: nil,
		AuthStyle:      0,
	}

	return &Collector{cfg, oauth2}
}

func (c *Collector) Name() string {
	return c.config.Name
}

func (c *Collector) Collect(ctx context.Context) ([]connectors.Alert, error) {
	sourceAlerts, err := c.collectAlerts(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert

	if len(sourceAlerts) == 0 {
		u, _ := url.Parse(c.config.URL)

		alert := connectors.Alert{
			Tags: map[string]string{
				"Hostname": u.Host,
			},
			Start:       time.Now(),
			State:       connectors.Critical,
			Description: "DeadMansSwitch dead",
		}
		alerts = append(alerts, alert)
	}

	for _, sourceAlert := range sourceAlerts {
		if sourceAlert.Status.State == "active" || sourceAlert.Status.State == "suppressed" {
			continue
		} else if len(sourceAlert.Status.SilencedBy) > 0 {
			continue
		}

		last, err := time.Parse("2006-01-02T15:04:05.000000", sourceAlert.StartsAt)
		if err != nil {
			panic(err)
		}

		descr := strings.Join(k8sLabels(sourceAlert.Labels, "alertname", "pod"), ":")
		alert := connectors.Alert{
			Tags: map[string]string{
				"Hostname": strings.Join(k8sLabels(sourceAlert.Labels, "cluster", "namespace"), ":"),
			},
			Start:       last,
			State:       connectors.Critical,
			Description: descr,
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func k8sLabels(haystack map[string]string, needles ...string) []string {
	var out []string
	for _, needle := range needles {
		if label, ok := haystack[needle]; ok {
			out = append(out, label)
		}
	}
	return out
}

func (c *Collector) collectAlerts(ctx context.Context) ([]Alert, error) {
	body, err := c.get("/api/v2/alerts", ctx)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	decoder := json.NewDecoder(body)

	var response []Alert
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Collector) get(endpoint string, ctx context.Context) (io.ReadCloser, error) {

	client := c.oauth2.Client(ctx)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}
