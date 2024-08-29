package grafana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"log/slog"
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
	sourceAlertGroups, err := c.collectAlerts(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert

	for _, sourceAlertGroup := range sourceAlertGroups {
		rule := sourceAlertGroup.Rules[0]
		sourceAlert := rule.Alerts[0]

		state := grafanaStateToState(sourceAlert.State)
		if state == connectors.OK {
			continue
		}

		labels := map[string]string{
			"Folder":    sourceAlert.Labels["grafana_folder"],
			"Alertname": sourceAlert.Labels["alertname"],
			"Contacts":  sourceAlert.Labels["__contacts__"],
		}

		alert := connectors.Alert{
			Labels:      labels,
			Start:       parseTime(sourceAlert.ActiveAt),
			State:       state,
			Description: rule.Name,
			Details:     rule.Annotations["message"],
			Links: []html.HTML{
				html.HTML("<a href=\"" + c.config.URL + "/alerting/grafana/" + rule.Labels["rule_uid"] + "/view?tab=instances" + "\" target=\"_blank\" alt=\"Alert\">🏠</a>"),
				html.HTML("<a href=\"" + c.config.URL + "/d/" + rule.Annotations["__dashboardUid__"] + "\" target=\"_blank\" alt=\"Dashboard\">🏠</a>"),
			},
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func grafanaStateToState(state string) connectors.State {
	switch strings.ToLower(state) {
	case alertingStateAlerting:
		return connectors.Critical
	default:
		return connectors.OK
	}
}

func (c *Connector) String() string {
	return fmt.Sprintf("Grafana (%s)", c.config.URL)
}

func (c *Connector) collectAlerts(ctx context.Context) ([]alertingRulesGroup, error) {
	res, err := c.get(ctx, "/api/prometheus/grafana/api/v1/rules")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, _ := io.ReadAll(res.Body)
	buf := bytes.NewBuffer(b)

	decoder := json.NewDecoder(buf)

	var response alertingRulesResult
	err = decoder.Decode(&response)
	if err != nil {
		slog.ErrorContext(ctx, "Cannot parse",
			slog.String("url", c.config.URL),
			slog.String("data", buf.String()),
			slog.Any("status", res.StatusCode),
			slog.Any("error", err))
		return nil, err
	}

	return response.Data.Groups, nil
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
func parseTime(timeField string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999-07:00", timeField)
	if err != nil {
		return time.Time{}
	}
	return t
}
