package web

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/synyx/tuwat/pkg/web/common"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

func (h *webHandler) alerts(w http.ResponseWriter, req *http.Request) {

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

func (h *webHandler) wsalerts(s *websocket.Conn) {
	defer func() {
		if err := recover(); err != nil {
			switch err := err.(type) {
			case error:
				otelzap.Ctx(s.Request().Context()).Info("panic serving", zap.Error(err))
			default:
				otelzap.Ctx(s.Request().Context()).Info("panic serving", zap.Any("error", err))
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
	otelzap.Ctx(s.Request().Context()).Info("Registering websocket connection",
		zap.String("client", tr),
		zap.String("dashboard", dashboardName))
	update := h.aggregator.Register(tr)
	defer h.aggregator.Unregister(tr)

	for {
		select {
		case _, ok := <-update:
			if !ok {
				otelzap.Ctx(s.Request().Context()).Debug("stop sending to websocket client, update channel closed",
					zap.String("client", tr))
				update = nil
			}

			otelzap.Ctx(s.Request().Context()).Debug("sending to websocket client",
				zap.String("client", tr))
			aggregate := h.aggregator.Alerts(dashboardName)
			renderer(webContent{Content: aggregate})
		case <-s.Request().Context().Done():
			otelzap.Ctx(s.Request().Context()).Debug("stop sending to websocket client, req ctx done",
				zap.String("client", tr))
			return
		}
	}
}

func (h *webHandler) ssealerts(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			switch err := err.(type) {
			case error:
				otelzap.Ctx(req.Context()).Info("panic serving", zap.Error(err))
			default:
				otelzap.Ctx(req.Context()).Info("panic serving", zap.Any("error", err))
			}
		}
	}()

	dashboardName := common.GetField(req, 0)

	renderer, cancel := h.sseRenderer(w, req, "alerts.gohtml")
	defer cancel()

	tr := trace.SpanFromContext(req.Context()).SpanContext().TraceID().String()
	otelzap.Ctx(req.Context()).Info("Registering sse connection", zap.String("client", tr))
	update := h.aggregator.Register(tr)
	defer h.aggregator.Unregister(tr)

	for {
		select {
		case _, ok := <-update:
			if !ok {
				otelzap.Ctx(req.Context()).Debug("stop sending to sse client", zap.String("client", tr))
				return
			}

			otelzap.Ctx(req.Context()).Debug("sending to sse client", zap.String("client", tr))
			aggregate := h.aggregator.Alerts(dashboardName)
			if err := renderer(webContent{Content: aggregate}); err != nil {
				otelzap.Ctx(req.Context()).Debug("stop sending to sse client",
					zap.String("client", tr),
					zap.Error(err))
				return
			}
		case <-req.Context().Done():
			otelzap.Ctx(req.Context()).Info("stop sending to sse client", zap.String("client", tr))
			return
		}
	}
}
