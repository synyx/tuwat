package web

import (
	"net/http"
)

func (h *webHandler) alerts(w http.ResponseWriter, req *http.Request) {

	aggregate := h.aggregator.Alerts()

	renderer := h.baseRenderer(req, "alerts.gohtml")

	renderer(w, 404, webContent{Content: aggregate})
}
