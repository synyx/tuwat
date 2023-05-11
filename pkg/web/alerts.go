package web

import (
	"net/http"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

func (h *webHandler) alerts(w http.ResponseWriter, req *http.Request) {

	aggregate := h.aggregator.Alerts()

	if req.Header.Get("Accept") == "text/vnd.turbo-stream.html" {
		renderer := h.partialRenderer(req, "alerts.gohtml")
		renderer(w, 200, webContent{Content: aggregate})
	} else {
		renderer := h.baseRenderer(req, "alerts.gohtml")
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

	renderer := h.wsRenderer(s, "alerts.gohtml")

	tr := trace.SpanFromContext(s.Request().Context()).SpanContext().TraceID().String()
	otelzap.Ctx(s.Request().Context()).Info("Registering websocket connection", zap.String("id", tr))
	update := h.aggregator.Register(tr)
	defer h.aggregator.Unregister(tr)

	for {
		select {
		case _, ok := <-update:
			if !ok {
				otelzap.Ctx(s.Request().Context()).Debug("stop sending to websocket client, update channel closed")
				update = nil
			}

			otelzap.Ctx(s.Request().Context()).Debug("sending to websocket client")
			aggregate := h.aggregator.Alerts()
			renderer(webContent{Content: aggregate})
		case <-s.Request().Context().Done():
			otelzap.Ctx(s.Request().Context()).Debug("stop sending to websocket client, req ctx done")
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

	renderer, cancel := h.sseRenderer(w, req, "alerts.gohtml")
	defer cancel()

	tr := trace.SpanFromContext(req.Context()).SpanContext().TraceID().String()
	otelzap.Ctx(req.Context()).Info("Registering sse connection", zap.String("id", tr))
	update := h.aggregator.Register(tr)
	defer h.aggregator.Unregister(tr)

	for {
		select {
		case _, ok := <-update:
			if !ok {
				otelzap.Ctx(req.Context()).Debug("stop sending to sse client")
				return
			}

			otelzap.Ctx(req.Context()).Debug("sending to sse client")
			aggregate := h.aggregator.Alerts()
			if err := renderer(webContent{Content: aggregate}); err != nil {
				otelzap.Ctx(req.Context()).Debug("stop sending to sse client", zap.Error(err))
				return
			}
		case <-req.Context().Done():
			otelzap.Ctx(req.Context()).Info("stop sending to sse client")
			return
		}
	}
}
