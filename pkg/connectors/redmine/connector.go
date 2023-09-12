package redmine

import (
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type Connector struct {
	config Config
	client *http.Client
}

type Config struct {
	Tag          string
	AssignedToId string
	common.HTTPConfig
}

func NewConnector(cfg *Config) *Connector {
	// by default use the current user as reference
	if cfg.AssignedToId == "" {
		cfg.AssignedToId = "me"
	}

	return &Connector{*cfg, cfg.HTTPConfig.Client()}
}

func (c *Connector) Tag() string {
	return c.config.Tag
}

func (c *Connector) Collect(ctx context.Context) ([]connectors.Alert, error) {
	issues, err := c.collectIssues(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert
	for _, issue := range issues {
		due, err := time.Parse("2006-01-02", issue.DueDate)
		if err != nil {
			continue
		}

		state := connectors.Critical
		if issue.DueDate == time.Now().Format("2006-01-02") {
			state = connectors.Warning
		}

		alert := connectors.Alert{
			Labels: map[string]string{
				"Project":  issue.Project.Name,
				"Ticket":   fmt.Sprintf("#%d", issue.Id),
				"Source":   c.config.URL,
				"Due":      issue.DueDate,
				"Type":     "Ticket",
				"Status":   issue.Status.Name,
				"Priority": issue.Priority.Name,
				"Author":   issue.Author.Name,
				"Assigned": issue.AssignedTo.Name,
				"Subject":  issue.Subject,
				"Private":  strconv.FormatBool(issue.IsPrivate),
			},
			Start:       due,
			State:       state,
			Description: issue.Subject,
			Details:     issue.Description,
			Links: []html.HTML{
				html.HTML("<a href=\"" + c.config.URL + "/issues/" + strconv.Itoa(issue.Id) + "\" target=\"_blank\" alt=\"Home\">🏠</a>"),
			},
		}
		alerts = append(alerts, alert)
		continue

	}

	return alerts, nil
}

func (c *Connector) String() string {
	return fmt.Sprintf("Redmine (%s)", c.config.URL)
}

func (c *Connector) collectIssues(ctx context.Context) ([]issue, error) {
	otelzap.Ctx(ctx).Debug("getting issues", zap.String("url", c.config.HTTPConfig.URL+"/issues.json"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.HTTPConfig.URL+"/issues.json", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Redmine-API-Key", c.config.BearerToken)

	q := req.URL.Query()
	q.Set("limit", "100")
	q.Set("offset", "0")
	q.Set("assigned_to_id", c.config.AssignedToId)
	q.Set("status_id", "open")
	q.Set("due_date", "<="+time.Now().Format("2006-01-02"))
	req.URL.RawQuery = q.Encode()

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

	return response.Issues, nil
}
