package actuator

import (
	"encoding/json"
	"net/http"

	"github.com/synyx/tuwat/pkg/buildinfo"
	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type HealthHandler struct {
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

type gitCommitInfo struct {
	Id string `json:"id,omitempty"`
}
type gitInfo struct {
	Commit gitCommitInfo `json:"commit,omitempty"`
}
type appInfo struct {
	Version string `json:"version"`
	Name    string `json:"name"`
}
type info struct {
	App appInfo `json:"app"`
	Git gitInfo `json:"git,omitempty"`
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	info := info{
		App: appInfo{
			Version: buildinfo.Version,
			Name:    buildinfo.Service,
		},
	}
	if buildinfo.GitSHA != "" {
		info.Git = gitInfo{gitCommitInfo{Id: buildinfo.GitSHA}}
	}

	encoder := json.NewEncoder(w)

	if err := encoder.Encode(info); err != nil {
		otelzap.Ctx(r.Context()).Debug("error serving info", zap.Error(err))
	}
}
