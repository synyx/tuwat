package grafana

// https://raw.githubusercontent.com/grafana/grafana/main/pkg/services/ngalert/api/tooling/post.json

type ruleResponse struct {
	Status string        `json:"status"`
	Data   ruleDiscovery `json:"data,omitempty"`
}

type ruleDiscovery struct {
	Groups []ruleGroup `json:"groups"`
}

type ruleGroup struct {
	Name  string         `json:"name"`
	File  string         `json:"file"`
	Rules []alertingRule `json:"rules"`
}

type alertingRule struct {
	State       alertingState     `json:"state"`
	Name        string            `json:"name"`
	ActiveAt    string            `json:"activeAt"`
	Health      string            `json:"health"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels,omitempty"`
	Alerts      []alert           `json:"alerts,omitempty"`
	Type        string            `json:"type"`
}

type alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	State       string            `json:"state"`
	ActiveAt    string            `json:"activeAt"`
	Value       string            `json:"value"`
}

type alertingState = string

const (
	alertingStatePending  = "pending"
	alertingStateFiring   = "firing"
	alertingStateInactive = "inactive"
)

const (
	alertingStateAlerting = "alerting"
	alertingStateNoData   = "nodata"
	alertingStateNormal   = "normal"
	alertingStateError    = "error"
)
