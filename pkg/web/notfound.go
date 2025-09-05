package web

import (
	"net/http"
)

func (h *WebHandler) notFound(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(404)

	renderer := h.baseRenderer(req, "_base.html", "404.gohtml")

	renderer(w, 404, webContent{})
}
