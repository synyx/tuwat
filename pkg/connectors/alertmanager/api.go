package alertmanager

type Alert struct {
	Labels      map[string]string `json:"labels"`
	Severity    string            `json:"severity"`
	StartsAt    string            `json:"startsAt"`
	Annotations map[string]string `json:"annotations"`
	Status      Status            `json:"status"`
}

type Status struct {
	State      string   `json:"state"`
	SilencedBy []string `json:"silencedBy"`
}
