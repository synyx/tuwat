package actuator

import (
	"encoding/json"
	"net/http"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"

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
		otelzap.Ctx(r.Context()).Debug("error serving info", zap.Error(err))
	}
}
