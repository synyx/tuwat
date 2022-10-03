package alertmanager

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/synyx/gonagdash/pkg/connectors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"golang.org/x/oauth2/clientcredentials"
)

type Connector struct {
	config Config
	oauth2 clientcredentials.Config
}

type Config struct {
	Tag     string
	Cluster string
	connectors.HTTPConfig
}

func NewConnector(cfg Config) *Connector {
	c := &Connector{config: cfg}

	if cfg.ClientId != "" {
		c.oauth2 = clientcredentials.Config{
			ClientID:       cfg.ClientId,
			ClientSecret:   cfg.ClientSecret,
			TokenURL:       cfg.TokenURL,
			Scopes:         nil,
			EndpointParams: nil,
			AuthStyle:      0,
		}
	}

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

	if len(sourceAlerts) == 0 {
		u, _ := url.Parse(c.config.URL)

		alert := connectors.Alert{
			Labels: map[string]string{
				"Hostname": u.Host,
			},
			Start:       time.Now(),
			State:       connectors.Critical,
			Description: "DeadMansSwitch dead",
		}
		alerts = append(alerts, alert)
	}

	for _, sourceAlert := range sourceAlerts {
		severity, _ := sourceAlert.Labels["severity"]

		if sourceAlert.Status.State == "suppressed" {
			continue
		} else if len(sourceAlert.Status.SilencedBy) > 0 {
			continue
		} else if severity == "none" {
			continue
		}

		state := connectors.Critical
		if sourceAlert.Status.State == "unprocessed" {
			state = connectors.Unknown
		} else if sourceAlert.Status.State == "active" && severity == "" {
			state = connectors.Warning
		} else if sourceAlert.Status.State == "active" && severity == "warning" {
			state = connectors.Warning
		} else if sourceAlert.Status.State == "active" && severity == "critical" {
			state = connectors.Critical
		} else if sourceAlert.Status.State == "active" && severity == "error" {
			state = connectors.Critical
		} else {
			otelzap.Ctx(ctx).DPanic("Cannot parse: Unknown state",
				zap.Any("state", sourceAlert.Status.State),
				zap.Any("severity", severity),
			)
		}

		last, err := time.Parse("2006-01-02T15:04:05Z07", sourceAlert.StartsAt)
		if err != nil {
			otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
		}

		links := make(map[string]string)
		if rb, ok := sourceAlert.Annotations["runbook"]; ok {
			links["📖"] = rb
		}

		filterLabels := map[string]string{
			"uid": sourceAlert.Labels["uid"],
		}
		if filter, err := json.Marshal(filterLabels); err == nil {
			links["🏠"] = c.config.URL + "/#/alerts?filter=" + url.QueryEscape(string(filter))
		}

		descr := sourceAlert.Labels["alertname"]
		details := strings.Join(k8sLabels(sourceAlert.Annotations, "summary", "description"), "\n")
		namespace := strings.Join(k8sLabels(sourceAlert.Labels, "namespace"), ":")

		r := regexp.MustCompile(`in namespace\W+([a-zA-Z-0-9_]+)`)
		if s := r.FindAllStringSubmatch(details, 1); len(s) > 0 {
			namespace = s[0][1]
		}

		tags := map[string]string{
			"Cluster":   c.config.Cluster,
			"Namespace": namespace,
			"Source":    c.config.URL,
		}
		for k, v := range sourceAlert.Labels {
			tags[k] = v
		}

		alert := connectors.Alert{
			Labels:      tags,
			Start:       last,
			State:       state,
			Description: descr,
			Details:     details,
			Links:       links,
		}
		alert.Silence = c.createSilencer(alert)
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

func (c *Connector) collectAlerts(ctx context.Context) ([]alert, error) {
	body, err := c.get("/api/v2/alerts", ctx)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	decoder := json.NewDecoder(body)

	var response []alert
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Connector) get(endpoint string, ctx context.Context) (io.ReadCloser, error) {

	var client *http.Client
	if c.config.ClientId != "" {
		client = c.oauth2.Client(ctx)
	} else {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: c.config.Insecure},
		}
		client = &http.Client{Transport: tr}
	}

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
