package actuator

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/synyx/tuwat/pkg/version"
)

type InfoHandler struct {
	versionInfo *version.VersionInfo
}

func NewInfoHandler() *InfoHandler {
	return &InfoHandler{versionInfo: &version.Info}
}

func (v *InfoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)

	if err := encoder.Encode(v.versionInfo); err != nil {
		slog.DebugContext(r.Context(), "error serving info", slog.Any("error", err))
	}
}
