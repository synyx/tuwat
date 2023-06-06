package gitlabmr

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

type Connector struct {
	config Config
}

type Config struct {
	Tag      string
	Projects []string
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

func (c *Connector) String() string {
	return fmt.Sprintf("GitLab MRs (%s)", c.config.URL)
}

// collectMRs collects merge requests from GitLab.  It either gets all available merge requests
// from the whole instance, or only selected projects.
func (c *Connector) collectMRs(ctx context.Context) ([]mergeRequest, error) {
	if c.config.Projects == nil {
		return c.collectMRsFrom(ctx, "/api/v4/merge_requests")
	}

	var mrs []mergeRequest
	for _, id := range c.config.Projects {
		id := url.PathEscape(id)

		if m, err := c.collectMRsFrom(ctx, fmt.Sprintf("/projects/%s/merge_requests", id)); err != nil {
			return mrs, err
		} else {
			mrs = append(mrs, m...)
		}
	}

	return mrs, nil
}

// collectMRsFrom will collect GitLab merge requests.
//
// see https://docs.gitlab.com/ee/api/merge_requests.html for more information.
func (c *Connector) collectMRsFrom(ctx context.Context, from string) ([]mergeRequest, error) {
	query := map[string]string{
		"wip":      "no",
		"state":    "opened",
		"order_by": "updated_at",
		"sort":     "desc",
		"scope":    "all",
	}

	var mergeRequests []mergeRequest

	for {
		body, next, err := c.get(ctx, from, query)
		if err != nil {
			return nil, err
		}
		defer body.Close()

		b, _ := io.ReadAll(body)
		buf := bytes.NewBuffer(b)

		decoder := json.NewDecoder(buf)

		var mrs []mergeRequest
		err = decoder.Decode(&mrs)
		if err != nil {
			otelzap.Ctx(ctx).DPanic("Cannot parse",
				zap.String("url", c.config.URL),
				zap.String("data", buf.String()),
				zap.Error(err))
			return nil, err
		}
		mergeRequests = append(mergeRequests, mrs...)

		if next != "" {
			query["page"] = next
			continue
		} else {
			break
		}
	}

	return mergeRequests, nil
}

// get a single page from GitLab API.
//
// Besides the obvious body and an error it will return a number (as a string)
// to be used for pulling the next page in case of pagination.  By default, it
// will get 100 results.  The calling code is responsible for collecting more
// results.
func (c *Connector) get(ctx context.Context, endpoint string, query map[string]string) (io.ReadCloser, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+endpoint, nil)
	if err != nil {
		return nil, "", err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.config.BearerToken)

	q := req.URL.Query()
	q.Set("per_age", "100")
	q.Set("page", "1")
	for k, v := range query {
		q.Set(k, v)
	}
	req.URL.RawQuery = q.Encode()
	reqUrl := req.URL.String()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.config.Insecure},
	}
	client := &http.Client{Transport: otelhttp.NewTransport(tr)}

	res, err := client.Do(req)
	if err != nil {
		otelzap.Ctx(ctx).DPanic("Cannot parse", zap.String("url", reqUrl), zap.Error(err))
		return nil, "", err
	}

	return res.Body, res.Header.Get("x-next-page"), nil
}
