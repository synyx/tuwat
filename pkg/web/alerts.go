package web

import (
	"net/http"
	"sort"
)

func (h *webHandler) alerts(w http.ResponseWriter, req *http.Request) {

	aggregate := h.aggregator.Alerts()

	sort.Slice(aggregate.Alerts, func(i, j int) bool {
		return aggregate.Alerts[i].When < aggregate.Alerts[j].When
	})

	renderer := h.baseRenderer(req, "alerts.gohtml")

	renderer(w, 404, webContent{Content: aggregate})
}
