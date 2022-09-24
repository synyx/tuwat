package actuator

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/uptrace/opentelemetry-go-extra/otelzap"
	"go.uber.org/zap"
)

type Status string

const (
	Down         Status = "DOWN"
	Up           Status = "UP"
	OutOfService Status = "OUT_OF_SERVICE"
	Unknown      Status = "UNKNOWN"
)

var statusPriority = [...]Status{Up, Unknown, OutOfService, Down}

type HealthActuator struct {
	status HealthStatus

	mutex sync.RWMutex
}

type StatusComponent struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
}

type HealthStatus struct {
	Status Status `json:"status"`

	Components map[string]StatusComponent `json:"components"`
}

func newHealthActuator() *HealthActuator {
	return &HealthActuator{
		status: HealthStatus{
			Status:     Unknown,
			Components: make(map[string]StatusComponent),
		},
	}
}

func (h *HealthActuator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	h.mutex.RLock()
	str, err := json.Marshal(h.status)
	status := h.status.Status
	h.mutex.RUnlock()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		otelzap.Ctx(r.Context()).Error("error marshalling health status", zap.Error(err))
		if _, err := io.WriteString(w, "{\"status\":\"DOWN\"}"); err != nil {
			otelzap.Ctx(r.Context()).Debug("error serving health", zap.Error(err))
		}
	}

	switch status {
	case OutOfService:
		fallthrough
	case Down:
		w.WriteHeader(http.StatusServiceUnavailable)
	case Unknown:
		fallthrough
	case Up:
		w.WriteHeader(http.StatusOK)
	}

	if _, err := w.Write(str); err != nil {
		otelzap.Ctx(r.Context()).Debug("error serving health", zap.Error(err))
	}
}

func (h *HealthActuator) Status() Status {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.status.Status
}

func (h *HealthActuator) Set(check string, status Status, message string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.status.Components[check] = StatusComponent{status, message}

	h.bubbleUpStatus()
}

func (h *HealthActuator) bubbleUpStatus() {
	newStatus := statusPriority[0]
	for _, component := range h.status.Components {
		if priorityOf(component.Status) > priorityOf(newStatus) {
			newStatus = component.Status
		}
	}
	h.status.Status = newStatus
}

func priorityOf(element Status) int {
	for k, v := range statusPriority {
		if element == v {
			return k
		}
	}
	return -1
}
