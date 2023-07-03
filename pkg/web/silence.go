package web

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/synyx/tuwat/pkg/connectors"
)

func (h *webHandler) silence(w http.ResponseWriter, req *http.Request) {

	user := "jo"
	if hdr := req.Header.Get("X-Auth-Request-User"); hdr != "" {
		user = hdr
	}

	alertId := getField(req, 0)

	h.aggregator.Silence(req.Context(), alertId, user)

	if req.Header.Get("Accept") == "text/vnd.turbo-stream.html" {
		dashboardName := getField(req, 0)
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

func (h *webHandler) silences(w http.ResponseWriter, req *http.Request) {
	renderer := h.baseRenderer(req, "silences.gohtml")

	silences := h.silencer.Silences()
	content := struct {
		ExternalId string
		Labels     string
		Silences   []connectors.Silence
	}{
		ExternalId: req.URL.Query().Get("externalId"),
		Labels:     req.URL.Query().Get("labels"),
		Silences:   silences,
	}

	renderer(w, http.StatusOK, webContent{Content: content})
}

func (h *webHandler) refreshSilence(w http.ResponseWriter, req *http.Request) {

	h.silencer.Refresh(req.Context())

	w.Header().Set("Location", "/silences")
	w.WriteHeader(http.StatusSeeOther)
}

func (h *webHandler) addSilence(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		return
	}
	externalId := req.FormValue("externalId")
	labelsString := req.FormValue("labels")
	labels := parseLabels(labelsString)

	h.silencer.SetSilence(externalId, labels)

	w.Header().Set("Location", "/silences")
	w.WriteHeader(http.StatusSeeOther)
}

func (h *webHandler) delSilence(w http.ResponseWriter, req *http.Request) {
	externalId := getField(req, 0)

	h.silencer.DeleteSilence(externalId)

	w.Header().Set("Location", "/silences")
	w.WriteHeader(http.StatusSeeOther)
}

func parseLabels(labelString string) map[string]string {
	re := regexp.MustCompile(`[,\r\n]+`)
	labelPairs := re.Split(labelString, -1)

	labels := make(map[string]string)
	for _, pair := range labelPairs {
		if kv := strings.SplitN(pair, "=", 2); len(kv) == 2 {
			labels[kv[0]] = kv[1]
		}
	}
	return labels
}
