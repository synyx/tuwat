package graylog

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
	sourceAlerts, err := c.collectAlertEvents(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert

	for _, sourceAlert := range sourceAlerts.Events {
		var streams []string
		for _, stream := range sourceAlerts.Context.Streams {
			streams = append(streams, stream.Title)
		}

		labels := map[string]string{
			"Source": sourceAlert.Event.Source,
			"Stream": strings.Join(streams, ","),
		}
		alert := connectors.Alert{
			Labels:      labels,
			Start:       parseTime(sourceAlert.Event.TimeRangeStart),
			State:       connectors.Warning,
			Description: sourceAlert.Event.Message,
			Details:     sourceAlerts.Context.EventDefinitions[sourceAlert.Event.EventDefinitionId].Description,
			Links: []html.HTML{
				html.HTML("<a href=\"" + c.config.URL + "/alerts" + "\" target=\"_blank\" alt=\"Home\">🏠</a>"),
			},
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Connector) String() string {
	return fmt.Sprintf("Alertmanager (%s)", c.config.URL)
}

func (c *Connector) collectAlertEvents(ctx context.Context) (eventsSearchResults, error) {
	// TODO: Use pagination, however, we're unlikely to hit this limit for unresolved alerts
	body := eventsSearchParameters{
		Query:   "",
		Page:    1,
		PerPage: 25,
		Filter: eventsSearchFilter{
			Alerts: AlertsFilterOnly,
		},
		TimeRange: timeRange{
			Type:  TimeRangeRelative,
			Range: 60,
		},
	}

	res, err := c.post(ctx, "/api/events/search", body)
	if err != nil {
		return eventsSearchResults{}, err
	}
	defer res.Body.Close()

	b, _ := io.ReadAll(res.Body)
	buf := bytes.NewBuffer(b)
	decoder := json.NewDecoder(buf)

	var response eventsSearchResults
	err = decoder.Decode(&response)
	if err != nil {
		slog.ErrorContext(ctx, "Cannot parse",
			slog.String("url", c.config.URL),
			slog.Any("status", res.StatusCode),
			slog.Any("error", err))
		return eventsSearchResults{}, err
	}

	return response, nil
}

func (c *Connector) post(ctx context.Context, endpoint string, body interface{}) (*http.Response, error) {

	slog.DebugContext(ctx, "getting alerts", slog.String("url", c.config.URL+endpoint))

	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.URL+endpoint, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func parseTime(timeField string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999Z", timeField)
	if err != nil {
		return time.Time{}
	}
	return t
}