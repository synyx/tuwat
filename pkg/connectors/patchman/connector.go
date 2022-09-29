package patchman

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/synyx/gonagdash/pkg/connectors"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type Connector struct {
	config Config

	osCache     map[string]*OS
	archCache   map[string]*Arch
	domainCache map[string]*Domain
}

type Config struct {
	Tag string
	connectors.HTTPConfig
}

func NewConnector(cfg Config) *Connector {
	return &Connector{
		config:      cfg,
		osCache:     make(map[string]*OS),
		archCache:   make(map[string]*Arch),
		domainCache: make(map[string]*Domain),
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
		if host.SecurityUpdateCount == 0 && host.RebootRequired == false {
			continue
		}

		last, err := time.Parse("2006-01-02T15:04:05", host.LastReport)
		if err != nil {
			otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
		}

		details := fmt.Sprintf("Security Updates: %d, Updates: %d, Needs Reboot: %t",
			host.SecurityUpdateCount, host.BugfixUpdateCount, host.RebootRequired)

		os, err := getCached(ctx, c, c.osCache, host.OSURL)
		arch, err := getCached(ctx, c, c.archCache, host.ArchURL)
		domain, err := getCached(ctx, c, c.domainCache, host.DomainURL)

		alert := connectors.Alert{
			Labels: map[string]string{
				"Hostname": host.Hostname,
				"Source":   c.config.URL,
				"tags":     host.Tags,
				"os":       os.Name,
				"arch":     arch.Name,
				"domain":   domain.Name,
			},
			Start:       last,
			State:       connectors.Critical,
			Description: "Host Security critical",
			Details:     details,
			Links: map[string]string{
				"🏠": c.config.URL + "/host/" + host.Hostname + "/",
			},
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Connector) collectHosts(ctx context.Context) ([]Host, error) {
	var response []Host
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
			otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
		}

		if d, ok := t.(json.Delim); ok && d == '{' {
			// Paging necessary
		pageHandler:
			for t, err := decoder.Token(); err == nil; t, err = decoder.Token() {
				if err != nil {
					otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
				}

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
						otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
					}
					break pageHandler
				}
			}
		} else {
			next = ""
		}

		// while the array contains values
		for decoder.More() {
			var h Host
			// decode an array value (Message)
			err := decoder.Decode(&h)
			if err != nil {
				otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
			}

			response = append(response, h)
		}

		// read closing bracket
		t, err = decoder.Token()
		if err != nil {
			otelzap.Ctx(ctx).DPanic("Cannot parse", zap.Error(err))
			return nil, err
		}

		otelzap.Ctx(ctx).Info("Would pull next", zap.String("url", next))
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
	otelzap.Ctx(ctx).Debug("patchman get", zap.String("url", req.URL.String()), zap.Error(err))
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}
