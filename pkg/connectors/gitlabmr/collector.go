package gitlabmr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/synyx/gonagdash/pkg/connectors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type Collector struct {
	config Config
}

type Config struct {
	Tag string
	connectors.HTTPConfig

	TargetBranch string
}

func NewCollector(cfg Config) *Collector {
	return &Collector{cfg}
}

func (c *Collector) Tag() string {
	return c.config.Tag
}

func (c *Collector) Collect(ctx context.Context) ([]connectors.Alert, error) {
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
			},
			Start:       last,
			State:       connectors.Warning,
			Description: descr,
			Details:     details,
			Links: map[string]string{
				"MR": mr.WebUrl,
			},
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Collector) collectMRs(ctx context.Context) ([]Alert, error) {
	body, err := c.get("/merge_requests", ctx)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	b, _ := io.ReadAll(body)
	buf := bytes.NewBuffer(b)

	decoder := json.NewDecoder(buf)

	var response []Alert
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

func (c *Collector) get(endpoint string, ctx context.Context) (io.ReadCloser, error) {
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
	if c.config.TargetBranch != "" {
		q.Add("target_branch", c.config.TargetBranch)
	}
	req.URL.RawQuery = q.Encode()
	url := req.URL.String()

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		otelzap.Ctx(ctx).DPanic("Cannot parse", zap.String("url", url), zap.Error(err))
		return nil, err
	}

	return res.Body, nil
}
