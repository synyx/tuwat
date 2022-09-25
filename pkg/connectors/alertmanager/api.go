package alertmanager

type Alert struct {
	Labels      map[string]string `json:"labels"`
	StartsAt    string            `json:"startsAt"`
	Annotations map[string]string `json:"annotations"`
	Status      Status            `json:"status"`
}

type Status struct {
	State      string   `json:"state"`
	SilencedBy []string `json:"silencedBy"`
}
