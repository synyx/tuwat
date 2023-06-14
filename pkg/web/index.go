package web

import (
	"net/http"

	"github.com/synyx/tuwat/pkg/config"
)

type indexContent struct {
	Dashboards map[string]*config.Dashboard
}

func (h *webHandler) index(w http.ResponseWriter, req *http.Request) {
	// if there is only a single dashboard, simply show that one
	if len(h.dashboards) == 1 {
		h.alerts(w, req)
		return
	}

	// otherwise render the list of dashboards
	renderer := h.baseRenderer(req, "index.gohtml")
	renderer(w, 200, webContent{Content: indexContent{h.dashboards}})
}
