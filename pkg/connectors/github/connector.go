package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"net/http"
	"time"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type Connector struct {
	config *Config
	client *http.Client
}

type Config struct {
	common.HTTPConfig
	Tag   string
	Repos []string
}

func NewConnector(cfg *Config) *Connector {
	if cfg.URL == "" {
		cfg.URL = "https://api.github.com"
	}
	return &Connector{cfg, cfg.HTTPConfig.Client()}
}

func (c *Connector) Tag() string {
	return c.config.Tag
}

func (c *Connector) Collect(ctx context.Context) ([]connectors.Alert, error) {

	var alerts []connectors.Alert
	for _, repo := range c.config.Repos {
		issues, err := c.collectIssues(ctx, repo)
		if err != nil {
			return nil, err
		}
		for _, issue := range issues {
			last, err := time.Parse(time.RFC3339, issue.UpdatedAt)
			if err != nil {
				otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
			}

			project := repo
			descr := fmt.Sprintf("MR !%d: %s", issue.Number, issue.Title)
			details := fmt.Sprintf("Author: %s, Assigned To: %s", issue.User.Login, issue.Assignee.Login)
			typ := "Issue"
			url := issue.HTMLUrl
			draft := false
			if issue.PullRequest.URL != "" {
				typ = "PullRequest"
				url = issue.PullRequest.HTMLUrl
				draft = issue.Draft
			}

			alert := connectors.Alert{
				Labels: map[string]string{
					"Project":  project,
					"Author":   issue.User.Login,
					"Assignee": issue.Assignee.Login,
					"Source":   c.config.URL,
					"Draft":    fmt.Sprintf("%t", draft),
					"Type":     typ,
				},
				Start:       last,
				State:       connectors.Warning,
				Description: descr,
				Details:     details,
				Links: []html.HTML{
					html.HTML("<a href=\"" + url + "\" target=\"_blank\" alt=\"Home\">🏠</a>"),
				},
			}
			alerts = append(alerts, alert)
		}
	}

	return alerts, nil
}

func (c *Connector) String() string {
	return fmt.Sprintf("GitHub (%s)", c.config.URL)
}

func (c *Connector) collectIssues(ctx context.Context, repo string) ([]issue, error) {
	body, err := c.get(ctx, c.config.URL+"/repos/"+repo+"/issues")
	if err != nil {
		return nil, err
	}
	defer body.Close()

	b, _ := io.ReadAll(body)
	buf := bytes.NewBuffer(b)

	decoder := json.NewDecoder(buf)

	var response []issue
	err = decoder.Decode(&response)
	if err != nil {
		otelzap.Ctx(ctx).DPanic("Cannot parse",
			zap.String("url", c.config.URL),
			zap.String("data", buf.String()),
			zap.Error(err))
		return nil, err
	}

	return response, nil
}

func (c *Connector) get(ctx context.Context, endpoint string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	q := req.URL.Query()
	q.Add("state", "open")
	q.Add("per_page", "100")
	q.Add("page", "1")
	req.URL.RawQuery = q.Encode()
	url := req.URL.String()

	res, err := c.client.Do(req)
	if err != nil {
		otelzap.Ctx(ctx).DPanic("Cannot parse", zap.String("url", url), zap.Error(err))
		return nil, err
	}

	return res.Body, nil
}
