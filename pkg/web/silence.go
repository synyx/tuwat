package web

import "net/http"

func (h *webHandler) silence(w http.ResponseWriter, req *http.Request) {

	user := "jo"
	if hdr := req.Header.Get("X-Auth-Request-User"); hdr != "" {
		user = hdr
	}

	alertId := getField(req, 0)

	h.aggregator.Silence(req.Context(), alertId, user)

	if req.Header.Get("Accept") == "text/vnd.turbo-stream.html" {
		renderer := h.partialRenderer(req, "alerts.gohtml")
		aggregate := h.aggregator.Alerts()
		renderer(w, 200, webContent{Content: aggregate})
	} else if req.ProtoAtLeast(1, 1) {
		w.Header().Set("Location", "/")
		w.WriteHeader(303)
	} else {
		w.Header().Set("Location", "/")
		w.WriteHeader(302)
	}
}
