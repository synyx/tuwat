package alertmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
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
	Tag     string
	Cluster string
	common.HTTPConfig
}

func NewConnector(cfg *Config) *Connector {
	c := &Connector{config: *cfg, client: cfg.HTTPConfig.Client()}

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
		severity := ""
		if s, ok := sourceAlert.Labels["severity"]; ok {
			severity = s
		}

		if sourceAlert.Status.State == "suppressed" {
			continue
		} else if len(sourceAlert.Status.SilencedBy) > 0 {
			continue
		} else if severity == "none" {
			continue
		}

		state := connectors.Unknown
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
			slog.ErrorContext(ctx, "Cannot parse: Unknown state",
				slog.Any("state", sourceAlert.Status.State),
				slog.Any("severity", severity),
			)
		}

		last, err := time.Parse("2006-01-02T15:04:05Z07", sourceAlert.StartsAt)
		if err != nil {
			slog.ErrorContext(ctx, "Cannot parse", slog.Any("error", err))
		}

		var links []html.HTML
		if link, ok := sourceAlert.Annotations["runbook"]; ok {
			links = append(links, html.HTML("<a href=\""+link+"\" target=\"_blank\" alt=\"Runbook\">📖</a>"))
		}

		filterLabels := map[string]string{
			"uid": sourceAlert.Labels["uid"],
		}
		if filter, err := json.Marshal(filterLabels); err == nil {
			link := c.config.URL + "/#/alerts?filter=" + url.QueryEscape(string(filter))
			links = append(links, html.HTML("<a href=\""+link+"\" target=\"_blank\"  alt=\"Home\">🏠</a>"))
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

func (c *Connector) String() string {
	return fmt.Sprintf("Alertmanager (%s)", c.config.URL)
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
	res, err := c.get(ctx, "/api/v2/alerts")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, _ := io.ReadAll(res.Body)
	buf := bytes.NewBuffer(b)

	decoder := json.NewDecoder(buf)

	var response []alert
	err = decoder.Decode(&response)
	if err != nil {
		slog.ErrorContext(ctx, "Cannot parse",
			slog.String("url", c.config.URL),
			slog.String("data", buf.String()),
			slog.Any("status", res.StatusCode),
			slog.Any("error", err))
		return nil, err
	}

	return response, nil
}

func (c *Connector) get(ctx context.Context, endpoint string) (*http.Response, error) {

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

	return res, nil
}
