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
	"net/url"
	"slices"
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
	Tag       string
	Cluster   string
	TimeRange int
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
	sourceAlertPages, err := c.collectAlertEvents(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert
	var seenEventDefinitions []string

	for _, sourceAlertPage := range sourceAlertPages {
		for _, sourceAlert := range sourceAlertPage.Events {
			eventAggregationId := eventToAggregationId(sourceAlert)
			if slices.Contains(seenEventDefinitions, eventAggregationId) {
				continue
			}
			seenEventDefinitions = append(seenEventDefinitions, eventAggregationId)

			var streams []string
			for _, stream := range sourceAlertPage.Context.Streams {
				streams = append(streams, stream.Title)
			}

			hostname, _ := url.Parse(c.config.URL)
			labels := map[string]string{
				"Source":    sourceAlert.Event.Source,
				"Stream":    strings.Join(streams, ","),
				"Priority":  priorityToLabel(sourceAlert.Event.Priority),
				"EventType": alertToLabel(sourceAlert.Event.Alert),
				"Hostname":  hostname.Hostname(),
			}
			alert := connectors.Alert{
				Labels:      labels,
				Start:       parseTime(sourceAlert.Event.TimeStamp),
				State:       priorityToState(sourceAlert.Event.Priority),
				Description: sourceAlert.Event.Message,
				Details:     sourceAlertPage.Context.EventDefinitions[sourceAlert.Event.EventDefinitionId].Description,
				Links: []html.HTML{
					html.HTML("<a href=\"" + c.config.URL + "/alerts" + "\" target=\"_blank\" alt=\"Home\">🏠</a>"),
				},
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

func (c *Connector) String() string {
	return fmt.Sprintf("Graylog (%s)", c.config.URL)
}

func (c *Connector) collectAlertEvents(ctx context.Context) ([]eventsSearchResults, error) {
	timeRangeSeconds := c.config.TimeRange
	if timeRangeSeconds == 0 {
		timeRangeSeconds = 600
	}

	page := 1
	var responsePage eventsSearchResults
	var result []eventsSearchResults
	for ok := true; ok; ok = len(responsePage.Events) > 0 {
		var err error
		responsePage, err = c.collectAlertEventsPage(ctx, page, timeRangeSeconds)
		if err != nil {
			return []eventsSearchResults{}, err
		}
		if len(responsePage.Events) > 0 {
			result = append(result, responsePage)
		}
		page++
	}

	return result, nil
}

func (c *Connector) collectAlertEventsPage(ctx context.Context, page int, timeRangeSeconds int) (eventsSearchResults, error) {
	body := eventsSearchParameters{
		Query:   "",
		Page:    page,
		PerPage: 100,
		TimeRange: timeRange{
			Type:  TimeRangeRelative,
			Range: timeRangeSeconds,
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
	req.Header.Set("X-Requested-By", "tuwat")

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

func priorityToLabel(priority int) string {
	switch priority {
	case priorityLow:
		return "Low"
	case priorityNormal:
		return "Normal"
	case priorityHigh:
		return "High"
	default:
		return "Unknown"
	}
}

func priorityToState(priority int) connectors.State {
	switch priority {
	case priorityLow:
		return connectors.Warning
	case priorityNormal:
		return connectors.Warning
	case priorityHigh:
		return connectors.Critical
	default:
		return connectors.Unknown
	}
}

func alertToLabel(isAlert bool) string {
	if isAlert {
		return "Alert"
	}
	return "Event"
}

func eventToAggregationId(sourceAlert eventsSearchResult) string {
	event := sourceAlert.Event
	id := event.EventDefinitionId
	var fieldValues []string
	for _, value := range event.GroupByFields {
		fieldValues = append(fieldValues, value)
	}
	slices.Sort(fieldValues)
	id += strings.Join(fieldValues, "")
	return id
}
