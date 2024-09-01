package alertmanager

type alert struct {
	Labels      map[string]string `json:"labels"`
	StartsAt    string            `json:"startsAt"`
	Annotations map[string]string `json:"annotations"`
	Status      status            `json:"status"`
	Receivers   []receiver        `json:"receivers,omitempty"`
}

type status struct {
	State      state    `json:"state"`
	SilencedBy []string `json:"silencedBy"`
}

type receiver struct {
	Name string `json:"name"`
}

type state = string

const (
	stateSuppressed  state = "suppressed"
	stateUnprocessed state = "unprocessed"
	stateActive      state = "active"
)

const (
	severityWarning  = "warning"
	severityError    = "error"
	severityCritical = "critical"
	severityNone     = "none"
)
