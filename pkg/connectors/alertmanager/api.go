package alertmanager

type alert struct {
	Labels      map[string]string `json:"labels"`
	StartsAt    string            `json:"startsAt"`
	Annotations map[string]string `json:"annotations"`
	Status      status            `json:"status"`
}

type status struct {
	State      string   `json:"state"`
	SilencedBy []string `json:"silencedBy"`
}
