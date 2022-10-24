package gitlabmr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type Connector struct {
	config Config
}

type Config struct {
	Tag string
	connectors.HTTPConfig
}

func NewConnector(cfg Config) *Connector {
	return &Connector{cfg}
}

func (c *Connector) Tag() string {
	return c.config.Tag
}

func (c *Connector) Collect(ctx context.Context) ([]connectors.Alert, error) {
	mRs, err := c.collectMRs(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert

	for _, mr := range mRs {
		last, err := time.Parse(time.RFC3339, mr.UpdatedAt)
		if err != nil {
			otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
		}

		project := strings.SplitN(mr.References.Full, "!", 2)[0]
		descr := "MR " + mr.References.Short + ": " + mr.Title
		details := fmt.Sprintf("Author: %s, Assigned To: %s", mr.Author.Name, mr.Assignee.Name)
		alert := connectors.Alert{
			Labels: map[string]string{
				"Project":   project,
				"Milestone": mr.Milestone.Title,
				"Author":    mr.Author.Name,
				"Assignee":  mr.Assignee.Name,
				"Source":    c.config.URL,
				"Type":      "PullRequest",
			},
			Start:       last,
			State:       connectors.Warning,
			Description: descr,
			Details:     details,
			Links: []html.HTML{
				html.HTML("<a href=\"" + mr.WebUrl + "\" target=\"_blank\" alt=\"Home\">🏠</a>"),
			},
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Connector) collectMRs(ctx context.Context) ([]mergeRequest, error) {
	body, err := c.get("/api/v4/merge_requests", ctx)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	b, _ := io.ReadAll(body)
	buf := bytes.NewBuffer(b)

	decoder := json.NewDecoder(buf)

	var response []mergeRequest
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

func (c *Connector) get(endpoint string, ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.BearerToken)

	q := req.URL.Query()
	q.Add("wip", "no")
	q.Add("state", "opened")
	q.Add("order_by", "updated_at")
	q.Add("sort", "desc")
	q.Add("scope", "all")
	req.URL.RawQuery = q.Encode()
	url := req.URL.String()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		otelzap.Ctx(ctx).DPanic("Cannot parse", zap.String("url", url), zap.Error(err))
		return nil, err
	}

	return res.Body, nil
}
