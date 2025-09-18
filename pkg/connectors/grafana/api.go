package grafana

// https://raw.githubusercontent.com/grafana/grafana/main/pkg/services/ngalert/api/tooling/post.json
// https://prometheus.io/docs/prometheus/latest/querying/api/#rules

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
	State       alertingRuleState `json:"state"`
	Name        string            `json:"name"`
	ActiveAt    string            `json:"activeAt"`
	Annotations map[string]string `json:"annotations"`
	Labels      map[string]string `json:"labels,omitempty"`
	Alerts      []alert           `json:"alerts,omitempty"`
	Type        string            `json:"type"`
}

type alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	State       alertingState     `json:"state"`
	ActiveAt    string            `json:"activeAt"`
	Value       string            `json:"value"`
}

type alertingRuleState = string

const (
	alertingStateFiring   alertingRuleState = "firing"
	alertingStatePending  alertingRuleState = "pending"
	alertingStateInactive alertingRuleState = "inactive"
)

type alertingState = string

const (
	alertingStateAlerting alertingState = "alerting"
	alertingStateNoData   alertingState = "nodata"
	alertingStateNormal   alertingState = "normal"
	alertingStateError    alertingState = "error"
)
