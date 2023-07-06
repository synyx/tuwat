package redmine

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	s := &Silencer{
		config:   *cfg,
		client:   cfg.HTTPConfig.Client(),
		silences: make(map[string]labels),
	}

	if cfg.SilenceStateFile != "" {
		if err := loadSilences(cfg.SilenceStateFile, &s.silences); err != nil {
			otelzap.L().Warn("loading silences failed", zap.Error(err))
		} else {
			otelzap.L().Debug("saving silences", zap.String("file", s.config.SilenceStateFile))
		}
	}

	return s
}

func (s *Silencer) Silence(labels map[string]string, id string) error {
	s.silences[id] = labels

	return nil
}

func (s *Silencer) Silenced(labels map[string]string) connectors.Silence {
	// look for the labels inside the known silences
	for id, l := range s.silences {
		found := 0
		for k, v := range labels {
			if v2, ok := l[k]; ok && v2 == v {
				found += 1
			}
		}

		// allowing for broader silences than the handed in set of labels
		if found >= len(l) {
			// get cached silence, or pretend we're not silenced even having a matched silence
			if silence, ok := s.silenced[id]; ok {
				return silence
			} else {
				continue
			}
		}
	}

	return connectors.Silence{}
}

func (s *Silencer) String() string {
	return fmt.Sprintf("Redmine Silencer (%s)", s.config.URL)
}

func (s *Silencer) Silences() []connectors.Silence {
	var silences []connectors.Silence
	for _, silence := range s.silenced {
		silences = append(silences, silence)
	}
	return silences
}

func (s *Silencer) SetSilence(id string, labels map[string]string) {
	otelzap.L().Debug("Adding silence", zap.String("id", id), zap.Any("labels", labels), zap.Int("count", len(s.silences)))

	s.silences[id] = labels

	if s.config.SilenceStateFile != "" {
		if err := storeSilences(s.config.SilenceStateFile, s.silences); err != nil {
			otelzap.L().Warn("saving silences failed", zap.Error(err))
		} else {
			otelzap.L().Debug("saving silences", zap.String("file", s.config.SilenceStateFile))
		}
	}
}

func (s *Silencer) DeleteSilence(id string) {
	otelzap.L().Debug("Deleting silence", zap.String("id", id))
	delete(s.silences, id)

	if s.config.SilenceStateFile != "" {
		if err := storeSilences(s.config.SilenceStateFile, s.silences); err != nil {
			otelzap.L().Warn("saving silences failed", zap.Error(err))
		} else {
			otelzap.L().Debug("saving silences", zap.String("file", s.config.SilenceStateFile))
		}
	}
}

func (s *Silencer) Refresh(ctx context.Context) error {
	silenced := make(map[string]connectors.Silence)

	var silenceIds []string
	for id, _ := range s.silences {
		silenceIds = append(silenceIds, id)
	}
	issues, err := s.getIssues(ctx, silenceIds)
	if err != nil {
		return err
	}

	for k, _ := range s.silences {
		if silence, err := s.getSilence(k, issues); err != nil {
			continue
		} else {
			silenced[k] = silence
		}
	}

	otelzap.L().Debug("Refreshing silences", zap.Int("count", len(silenced)))

	s.silenced = silenced

	return nil
}

func (s *Silencer) getSilence(id string, issues []issue) (connectors.Silence, error) {
	silence := connectors.Silence{
		ExternalId: id,
		Silenced:   false,
		Labels:     s.silences[id],
	}

	redmineId, err := strconv.Atoi(id)
	if err != nil {
		return silence, err
	}

	var issue issue
	for _, i := range issues {
		if strconv.Itoa(i.Id) == id {
			issue = i
		}
	}
	if issue.Id == 0 {
		return silence, errors.New("issue not found")
	}

	silence.URL = fmt.Sprintf("%s/issues/%d", s.config.HTTPConfig.URL, redmineId)

	// Silencing if the ticket is still open and the alert is firing
	if issue.ClosedOn == "" {
		silence.Silenced = true
	}

	return silence, nil
}

func (s *Silencer) getIssues(ctx context.Context, ids []string) ([]issue, error) {
	url := fmt.Sprintf("%s/issues.json", s.config.HTTPConfig.URL)
	otelzap.Ctx(ctx).Debug("getting issues", zap.String("url", url))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Redmine-API-Key", s.config.BearerToken)
	req.Header.Set("Content-Type", "application/json")

	q := req.URL.Query()
	q.Set("limit", "100")
	q.Set("offset", "0")
	q.Set("issue_id", strings.Join(ids, ","))
	req.URL.RawQuery = q.Encode()

	res, err := s.client.Do(req)
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
