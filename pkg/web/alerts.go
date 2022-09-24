package web

import (
	"net/http"
)

func (h *webHandler) alerts(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(404)

	renderer := h.baseRenderer(req, "404.gohtml")

	renderer(w, 404, webContent{})
}
