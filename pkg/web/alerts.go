package web

import (
	"errors"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/websocket"

	"github.com/synyx/tuwat/pkg/web/common"
)

func (h *WebHandler) alerts(w http.ResponseWriter, req *http.Request) {

	dashboardName := filepath.Base(req.URL.Path)
	dashboardName = strings.TrimPrefix(dashboardName, "/")

	aggregate := h.aggregator.Alerts(dashboardName)

	if req.Header.Get("Accept") == "text/vnd.turbo-stream.html" {
		renderer := h.partialRenderer(req, dashboardName, "alerts.gohtml")
		renderer(w, 200, webContent{Content: aggregate})
	} else {
		renderer := h.baseRenderer(req, dashboardName, "alerts.gohtml")
		renderer(w, 200, webContent{Content: aggregate})
	}
}

func (h *WebHandler) wsalerts(s *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			switch err := err.(type) {
			case error:
				if errors.Is(err, DisconnectError) {
					slog.DebugContext(s.Request().Context(), "panic serving", slog.Any("error", err))
				} else {
					slog.InfoContext(s.Request().Context(), "panic serving", slog.Any("error", err))
				}
			default:
				slog.InfoContext(s.Request().Context(), "panic serving", slog.Any("error", err))
			}
		}
		_ = s.Close()
	}()

	// get dashboard name from /ws/alerts/... directly from path, as information gets lost from the websocket upgrade.
	dashboardName := filepath.Base(s.Request().URL.Path)
	dashboardName = strings.TrimPrefix(dashboardName, "ws")
	dashboardName = strings.TrimPrefix(dashboardName, "/")

	renderer := h.wsRenderer(s, "alerts.gohtml")

	tr := trace.SpanFromContext(s.Request().Context()).SpanContext().TraceID().String()
	slog.InfoContext(s.Request().Context(), "Registering websocket connection",
		slog.String("dashboard", dashboardName))
	update := h.aggregator.Register(tr)
	defer h.aggregator.Unregister(tr)

	for {
		select {
		case _, ok := <-update:
			if !ok {
				slog.DebugContext(s.Request().Context(), "stop sending to websocket client, update channel closed")
				update = nil
			}

			slog.DebugContext(s.Request().Context(), "sending to websocket client")
			aggregate := h.aggregator.Alerts(dashboardName)
			renderer(webContent{Content: aggregate})
		case <-s.Request().Context().Done():
			slog.DebugContext(s.Request().Context(), "stop sending to websocket client, req ctx done")
			return
		}
	}
}

func (h *WebHandler) ssealerts(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			switch err := err.(type) {
			case error:
				slog.InfoContext(req.Context(), "panic serving", slog.Any("error", err))
			default:
				slog.InfoContext(req.Context(), "panic serving", slog.Any("error", err))
			}
		}
	}()

	dashboardName := common.GetField(req, 0)

	renderer, cancel := h.sseRenderer(w, req, "alerts.gohtml")
	defer cancel()

	tr := trace.SpanFromContext(req.Context()).SpanContext().TraceID().String()
	slog.InfoContext(req.Context(), "Registering sse connection")
	update := h.aggregator.Register(tr)
	defer h.aggregator.Unregister(tr)

	for {
		select {
		case _, ok := <-update:
			if !ok {
				slog.DebugContext(req.Context(), "stop sending to sse client")
				return
			}

			slog.DebugContext(req.Context(), "sending to sse client")
			aggregate := h.aggregator.Alerts(dashboardName)
			if err := renderer(webContent{Content: aggregate}); err != nil {
				slog.DebugContext(req.Context(), "stop sending to sse client", slog.Any("error", err))
				return
			}
		case <-req.Context().Done():
			slog.InfoContext(req.Context(), "stop sending to sse client")
			return
		}
	}
}
