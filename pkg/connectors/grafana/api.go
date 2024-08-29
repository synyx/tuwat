package grafana

type alertingRulesResult struct {
	Status string            `json:"status"`
	Data   alertingRulesData `json:"data"`
}

type alertingRulesData struct {
	Groups []alertingRulesGroup `json:"groups"`
}

type alertingRulesGroup struct {
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
	Labels      map[string]string `json:"labels"`
	Alerts      []alert           `json:"alerts"`
	Type        string            `json:"type"`
}

type alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	State       string            `json:"state"`
	ActiveAt    string            `json:"activeAt"`
	Value       int               `json:"value"`
}

type alertingState = string

const (
	alertingStateFiring   = "firing"
	alertingStateAlerting = "alerting"
	alertingStateInactive = "inactive"
	alertingStateNoData   = "nodata"
	alertingStateNormal   = "Normal"
	alertingStatePending  = "pending"
	alertingStateError    = "error"
)
