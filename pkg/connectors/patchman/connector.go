package patchman

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
)

type Connector struct {
	config Config
	client *http.Client

	osCache     map[string]*os
	archCache   map[string]*arch
	domainCache map[string]*domain
}

type Config struct {
	Tag string
	common.HTTPConfig
}

func NewConnector(cfg *Config) *Connector {
	return &Connector{
		config:      *cfg,
		client:      cfg.HTTPConfig.Client(),
		osCache:     make(map[string]*os),
		archCache:   make(map[string]*arch),
		domainCache: make(map[string]*domain),
	}
}

func (c *Connector) Tag() string {
	return c.config.Tag
}

func (c *Connector) Collect(ctx context.Context) ([]connectors.Alert, error) {
	hosts, err := c.collectHosts(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert

	for _, host := range hosts {
		if host.SecurityUpdateCount == 0 && !host.RebootRequired {
			continue
		}

		last, err := time.Parse("2006-01-02T15:04:05", host.LastReport)
		if err != nil {
			slog.ErrorContext(ctx, "Cannot parse", slog.Any("error", err))
		}

		details := fmt.Sprintf("Security Updates: %d, Updates: %d, Needs Reboot: %t",
			host.SecurityUpdateCount, host.BugfixUpdateCount, host.RebootRequired)

		os, _ := getCached(ctx, c, c.osCache, host.OSURL)
		arch, _ := getCached(ctx, c, c.archCache, host.ArchURL)
		domain, _ := getCached(ctx, c, c.domainCache, host.DomainURL)

		alert := connectors.Alert{
			Labels: map[string]string{
				"Hostname": host.Hostname,
				"Source":   c.config.URL,
				"Type":     "Host",
				"tags":     host.Tags,
				"os":       os.Name,
				"arch":     arch.Name,
				"domain":   domain.Name,
			},
			Start:       last,
			State:       connectors.Critical,
			Description: "Host Security critical",
			Details:     details,
			Links: []html.HTML{
				html.HTML("<a href=\"" + c.config.URL + "/host/" + url.QueryEscape(host.Hostname) + "/\" target=\"_blank\" alt=\"Home\">🏠</a>"),
			},
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Connector) String() string {
	return fmt.Sprintf("Patchman (%s)", c.config.URL)
}

func (c *Connector) collectHosts(ctx context.Context) ([]host, error) {
	var response []host
	next := "/api/host/"

	for next != "" {
		body, err := c.get(ctx, next)
		if err != nil {
			return nil, err
		}
		defer body.Close()

		decoder := json.NewDecoder(body)

		// read open bracket
		t, err := decoder.Token()
		if err != nil {
			slog.ErrorContext(ctx, "Cannot parse", slog.Any("error", err))
		}

		if d, ok := t.(json.Delim); ok && d == '{' {
			// Paging necessary
		pageHandler:
			for t, err := decoder.Token(); err == nil; t, err = decoder.Token() {
				if s, ok := t.(string); ok && s == "next" {
					s, err := decoder.Token()
					if s, ok := s.(string); ok && err == nil {
						u, _ := url.Parse(s)
						next = "/api/host/?" + u.RawQuery
					} else {
						next = ""
					}
				}

				if s, ok := t.(string); ok && s == "results" {
					t, err := decoder.Token()
					if d, ok := t.(json.Delim); ok && d != '[' {
						slog.ErrorContext(ctx, "Cannot parse", slog.Any("error", err))
					}
					break pageHandler
				}
			}
		} else {
			next = ""
		}

		// while the array contains values
		for decoder.More() {
			var h host
			// decode an array value (Message)
			err := decoder.Decode(&h)
			if err != nil {
				slog.ErrorContext(ctx, "Cannot parse", slog.Any("error", err))
			}

			response = append(response, h)
		}

		// read closing bracket
		if _, err := decoder.Token(); err != nil {
			slog.ErrorContext(ctx, "Cannot parse", slog.Any("error", err))
			return nil, err
		}

		slog.DebugContext(ctx, "Would pull next", slog.String("url", next))
	}

	return response, nil
}

func getCached[T any](ctx context.Context, c *Connector, cache map[string]*T, rawUrl string) (*T, error) {
	if element, ok := cache[rawUrl]; ok {
		return element, nil
	}

	split := strings.Split(rawUrl, "/")
	typ := split[len(split)-3]
	id := split[len(split)-2]

	body, err := c.get(ctx, "/api/"+typ+"/"+id+"/")
	if err != nil {
		return new(T), err
	}
	defer body.Close()

	decoder := json.NewDecoder(body)

	var element T
	if err = decoder.Decode(&element); err != nil {
		return &element, err
	}

	cache[rawUrl] = &element
	return &element, nil
}

func (c *Connector) get(ctx context.Context, endpoint string) (io.ReadCloser, error) {
	slog.DebugContext(ctx, "getting alerts", slog.String("url", c.config.URL+endpoint))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.URL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: c.config.Insecure},
	}
	client := &http.Client{Transport: tr}

	res, err := client.Do(req)
	slog.DebugContext(ctx, "patchman get", slog.String("url", req.URL.String()), slog.Any("error", err))
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}
