package web

import (
	"net/http"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

func (h *webHandler) alerts(w http.ResponseWriter, req *http.Request) {

	renderer := h.baseRenderer(req, "alerts.gohtml")

	aggregate := h.aggregator.Alerts()

	renderer(w, 200, webContent{Content: aggregate})
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

	otelzap.Ctx(s.Request().Context()).Info("Registering websocket connection")
	update := h.aggregator.Register(h)
	defer h.aggregator.Unregister(h)

	for {
		select {
		case _, ok := <-update:
			if !ok {
				otelzap.Ctx(s.Request().Context()).Debug("stop sending to websocket client")
				return
			}

			otelzap.Ctx(s.Request().Context()).Debug("sending to websocket client")
			aggregate := h.aggregator.Alerts()
			renderer(webContent{Content: aggregate})
		case <-s.Request().Context().Done():
			otelzap.Ctx(s.Request().Context()).Debug("stop sending to websocket client")
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

	otelzap.Ctx(req.Context()).Info("Registering sse connection")
	update := h.aggregator.Register(h)
	defer h.aggregator.Unregister(h)

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
