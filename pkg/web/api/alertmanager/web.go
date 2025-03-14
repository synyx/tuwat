package alertmanager

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"slices"
	"time"

	"github.com/synyx/tuwat/pkg/aggregation"
	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/version"
	"github.com/synyx/tuwat/pkg/web/common"
)

type alertmanagerHandler struct {
	routes []common.Route

	aggregator *aggregation.Aggregator
	started    time.Time
	cfg        *config.Config
}

func ApiV2(cfg *config.Config, aggregator *aggregation.Aggregator) http.Handler {
	handler := &alertmanagerHandler{
		aggregator: aggregator,
		started:    time.Now(),
		cfg:        cfg,
	}
	handler.routes = []common.Route{
		common.NewRoute("GET", "/v2/status", handler.status),
		common.NewRoute("GET", "/v2/alerts", handler.alerts),
	}

	return handler
}

func (h *alertmanagerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			switch err := err.(type) {
			case error:
				slog.ErrorContext(r.Context(), "panic serving", slog.Any("error", err))
			default:
				slog.ErrorContext(r.Context(), "panic serving", slog.Any("error", err))
			}

			http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)
		}
	}()

	if ok := common.HandleRoute(h.routes, w, r); !ok {
		http.Error(w, "404 not found", http.StatusNotFound)
	}
}

func (h *alertmanagerHandler) status(w http.ResponseWriter, _ *http.Request) {
	status := alertmanagerStatus{
		ClusterStatus: clusterStatus{Status: "ready"},
		VersionInfo: versionInfo{
			Version:   version.Info.Version,
			Revision:  version.Info.Revision,
			Branch:    "unknown",
			BuildUser: "Jane Doe",
			BuildDate: version.Info.ReleaseDate,
			GoVersion: version.Info.GoVersion,
		},
		Config: alertmanagerConfig{Original: ""},
		Uptime: h.started.Format(time.RFC3339),
	}
	writer := json.NewEncoder(w)
	if err := writer.Encode(&status); err != nil {
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
	}
}

func (h *alertmanagerHandler) alerts(w http.ResponseWriter, r *http.Request) {
	var gettableAlerts []gettableAlert

	active := true
	inhibited := true
	var filters []fieldMatcher

	if val := r.URL.Query().Get("active"); val != "" {
		active = val == "true"
	}
	if val := r.URL.Query().Get("inhibited"); val != "" {
		inhibited = val == "true"
	}
	if ok := r.URL.Query().Has("filter"); ok {
		for _, val := range r.URL.Query()["filter"] {
			matchers := parseFilter(val)
			filters = append(filters, matchers...)
		}
	}

	for _, dashboard := range h.cfg.Dashboards {
		aggregate := h.aggregator.Alerts(dashboard.Name)
		for _, alert := range aggregate.Alerts {
			if alert.Status == "green" {
				continue
			}

			gettableAlerts = append(gettableAlerts, mapAlert(dashboard.Name, aggregate, alert, "active"))
		}
		for _, alertGroup := range aggregate.GroupedAlerts {
			for _, alert := range alertGroup.Alerts {
				if alert.Status == "green" {
					continue
				}

				gettableAlerts = append(gettableAlerts, mapAlert(dashboard.Name, aggregate, alert, "active"))
			}
		}
		for _, alert := range aggregate.Blocked {
			if alert.Status == "green" {
				continue
			}

			gettableAlerts = append(gettableAlerts, mapAlert(dashboard.Name, aggregate, alert.Alert, "suppressed"))
		}
	}

	if !active {
		gettableAlerts = slices.DeleteFunc(gettableAlerts, func(x gettableAlert) bool {
			return x.Status.State == "active"
		})
	}
	if !inhibited {
		gettableAlerts = slices.DeleteFunc(gettableAlerts, func(x gettableAlert) bool {
			return x.Status.State == "suppressed"
		})
	}

	if len(filters) > 0 {
		gettableAlerts = slices.DeleteFunc(gettableAlerts, func(x gettableAlert) bool {
			matches := 0
			for _, filter := range filters {
				if s, ok := x.Annotations[filter.field]; !ok {
					return true
				} else if filter.m.MatchString(s) {
					matches += 1
				}
			}

			return matches != len(filters)
		})
	}

	writer := json.NewEncoder(w)
	if err := writer.Encode(gettableAlerts); err != nil {
		http.Error(w, "500 internal server error", http.StatusInternalServerError)
	}
}

func mapAlert(dashboard string, aggregate aggregation.Aggregate, alert aggregation.Alert, state string) gettableAlert {
	labels := alert.Labels
	labels["dashboard"] = dashboard
	switch alert.Status {
	case "red":
		labels["severity"] = "critical"
	case "yellow":
		labels["severity"] = "warning"
	default:
		labels["severity"] = ""
	}

	ga := gettableAlert{
		Annotations: labels,
		Fingerprint: alert.Id,
		StartsAt:    time.Now().Add(alert.When).Format(time.RFC3339),
		UpdatedAt:   aggregate.CheckTime.Format(time.RFC3339),
		EndsAt:      time.Now().Add(1 * time.Hour).Format(time.RFC3339),
		Status: alertStatus{
			State:       state,
			SilencedBy:  nil,
			InhibitedBy: nil,
		},
	}

	return ga
}
