package actuator

import (
	"encoding/json"
	"net/http"

	"github.com/synyx/tuwat/pkg/version"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type VersionHandler struct {
	versionInfo *version.VersionInfo
}

func NewVersionHandler() *VersionHandler {
	return &VersionHandler{versionInfo: &version.Info}
}

func (v *VersionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)

	if err := encoder.Encode(v.versionInfo); err != nil {
		otelzap.Ctx(r.Context()).Debug("error serving info", zap.Error(err))
	}
}
