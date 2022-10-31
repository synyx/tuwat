package actuator

import (
	"encoding/json"
	"github.com/synyx/tuwat/pkg/version"
	"net/http"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type versionHandler struct {
	versionInfo *version.VersionInfo
}

func NewVersionHandler() *versionHandler {
	return &versionHandler{versionInfo: &version.Info}
}

func (v *versionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)

	if err := encoder.Encode(v.versionInfo); err != nil {
		otelzap.Ctx(r.Context()).Debug("error serving info", zap.Error(err))
	}
}
