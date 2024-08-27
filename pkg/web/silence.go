package web

import (
	"net/http"

	"github.com/synyx/tuwat/pkg/web/common"
)

func (h *WebHandler) silence(w http.ResponseWriter, req *http.Request) {

	user := "jo"
	if hdr := req.Header.Get("X-Auth-Request-User"); hdr != "" {
		user = hdr
	}

	alertId := common.GetField(req, 0)

	h.aggregator.Silence(req.Context(), alertId, user)

	if req.Header.Get("Accept") == "text/vnd.turbo-stream.html" {
		dashboardName := common.GetField(req, 0)
		renderer := h.partialRenderer(req, "alerts.gohtml")
		aggregate := h.aggregator.Alerts(dashboardName)
		renderer(w, 200, webContent{Content: aggregate})
	} else if req.ProtoAtLeast(1, 1) {
		w.Header().Set("Location", "/")
		w.WriteHeader(303)
	} else {
		w.Header().Set("Location", "/")
		w.WriteHeader(302)
	}
}
