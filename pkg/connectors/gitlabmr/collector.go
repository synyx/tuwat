package gitlabmr

import (
	"context"
	"encoding/json"
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
	Name string
	connectors.HTTPConfig

	TargetBranch string
}

func NewCollector(cfg Config) *Collector {
	return &Collector{cfg}
}

func (c *Collector) Name() string {
	return c.config.Name
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
		alert := connectors.Alert{
			Tags: map[string]string{
				"Hostname": project,
			},
			Start:       last,
			State:       connectors.Warning,
			Description: descr,
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

	decoder := json.NewDecoder(body)

	var response []Alert
	err = decoder.Decode(&response)
	if err != nil {
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

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}
