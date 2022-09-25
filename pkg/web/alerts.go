package web

import (
	"net/http"
	"sort"
	"time"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
)

func (h *webHandler) alerts(w http.ResponseWriter, req *http.Request) {

	aggregate := h.aggregator.Alerts()

	sort.Slice(aggregate.Alerts, func(i, j int) bool {
		return aggregate.Alerts[i].When < aggregate.Alerts[j].When
	})

	renderer := h.baseRenderer(req, "alerts.gohtml")

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
			_ = s.Close()
		}
	}()
	renderer := h.wsRenderer(s, "alerts.gohtml")

	for {
		aggregate := h.aggregator.Alerts()

		sort.Slice(aggregate.Alerts, func(i, j int) bool {
			return aggregate.Alerts[i].When < aggregate.Alerts[j].When
		})

		renderer(webContent{Content: aggregate})
		time.Sleep(10 * time.Second)
	}
}
