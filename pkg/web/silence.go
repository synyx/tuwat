package web

import "net/http"

func (h *webHandler) silence(w http.ResponseWriter, req *http.Request) {

	user := "jo"
	if hdr := req.Header.Get("X-Auth-Request-User"); hdr != "" {
		user = hdr
	}

	alertId := getField(req, 0)

	h.aggregator.Silence(req.Context(), alertId, user)

	renderer := h.partialRenderer(req, "alerts.gohtml")

	aggregate := h.aggregator.Alerts()

	renderer(w, 200, webContent{Content: aggregate})
}
