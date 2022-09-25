package web

import (
	"fmt"
	"net/http"
	"time"

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

	for {
		aggregate := h.aggregator.Alerts()

		renderer(webContent{Content: aggregate})
		time.Sleep(10 * time.Second)
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

	renderer := h.sseRenderer(w, req, "alerts.gohtml")

	ping := time.NewTicker(10 * time.Second)
	defer ping.Stop()
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case _ = <-ping.C:
			_, _ = fmt.Fprint(w, "event: ping\n\n")
			w.(http.Flusher).Flush()
		case _ = <-ticker.C:
			otelzap.Ctx(req.Context()).Info("sending to sse client")
			aggregate := h.aggregator.Alerts()

			renderer(webContent{Content: aggregate})
		case <-req.Context().Done():
			otelzap.Ctx(req.Context()).Info("stop sending to sse client")
			return
		}
	}
}
