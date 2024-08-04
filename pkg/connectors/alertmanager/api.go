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

const (
	SilenceExpired = "expired"
	SilenceActive  = "active"
	SilencePending = "pending"
)

type silenceStatus struct {
	State string `json:"state"`
}

type matcher struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	IsRegex bool   `json:"isRegex"`
	IsEqual bool   `json:"isEqual"`
}

type silence struct {
	Id        string        `json:"id"`
	Status    silenceStatus `json:"status"`
	Matchers  []matcher     `json:"matchers"`
	StartsAt  string        `json:"startsAt"`
	EndsAt    string        `json:"endsAt"`
	CreatedBy string        `json:"createdBy"`
	Comment   string        `json:"comment"`
}
