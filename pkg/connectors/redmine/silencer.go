package redmine

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type labels map[string]string

type Silencer struct {
	config Config
	client *http.Client

	silences map[string]labels
	silenced map[string]connectors.Silence
}

func NewSilencer(cfg *Config) *Silencer {
	return &Silencer{
		config:   *cfg,
		client:   cfg.HTTPConfig.Client(),
		silences: make(map[string]labels),
	}
}

func (s *Silencer) Silence(labels map[string]string, externalId string) error {
	s.silences[externalId] = labels

	return nil
}

func (s *Silencer) Silenced(labels map[string]string) connectors.Silence {
	// look for the labels inside the known silences
	for externalId, l := range s.silences {
		found := 0
		for k, v := range labels {
			if v2, ok := l[k]; ok && v2 == v {
				found += 1
			}
		}

		// allowing for broader silences than the handed in set of labels
		if found >= len(l) {
			// get cached silence, or pretend we're not silenced even having a matched silence
			if silence, ok := s.silenced[externalId]; ok {
				return silence
			} else {
				return connectors.Silence{}
			}
		}
	}

	return connectors.Silence{}
}

func (s *Silencer) String() string {
	return fmt.Sprintf("Redmine Silencer (%s)", s.config.URL)
}

func (s *Silencer) RefreshCache(ctx context.Context) {
	silenced := make(map[string]connectors.Silence)

	for k, _ := range s.silences {
		if silence, err := s.getSilence(ctx, k); err != nil {
			silenced[k] = silence
		}
	}

	s.silenced = silenced
}

func (s *Silencer) getSilence(ctx context.Context, externalId string) (connectors.Silence, error) {
	id, err := strconv.Atoi(externalId)
	if err != nil {
		return connectors.Silence{}, err
	}

	issue, err := s.getIssue(ctx, id)
	if err != nil {
		return connectors.Silence{}, err
	}

	silenced := false

	// Silencing if the ticket is still open and the alert is firing
	if issue.ClosedOn == "" {
		silenced = true
	}

	return connectors.Silence{
		ExternalId: externalId,
		URL:        fmt.Sprintf("%s/issues/%d", s.config.HTTPConfig.URL, id),
		Silenced:   silenced,
	}, nil
}

func (s *Silencer) getIssue(ctx context.Context, externalId int) (*issue, error) {
	url := fmt.Sprintf("%s/issues/%d.json", s.config.HTTPConfig.URL, externalId)
	otelzap.Ctx(ctx).Debug("getting issues", zap.String("url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("assigned_to", "me")
	req.Header.Set("status_id", "open")
	req.Header.Set("due_date", "<="+time.Now().Format("2006-01-02"))
	req.Header.Set("X-Redmine-API-Key", s.config.BearerToken)

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)

	var response issue
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
