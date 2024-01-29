package alertmanager

type alertmanagerStatus struct {
	ClusterStatus clusterStatus      `json:"clusterStatus"`
	VersionInfo   versionInfo        `json:"versionInfo"`
	Config        alertmanagerConfig `json:"config"`
	Uptime        string             `json:"uptime"`
}

type clusterStatus struct {
	Status string `json:"status"`
}

type versionInfo struct {
	Version   string `json:"version"`
	Revision  string `json:"revision"`
	Branch    string `json:"branch"`
	BuildUser string `json:"buildUser"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
}

type alertmanagerConfig struct {
	Original string `json:"original"`
}

type gettableAlert struct {
	Annotations map[string]string `json:"annotations"`
	Fingerprint string            `json:"fingerprint"`
	StartsAt    string            `json:"startsAt"`
	UpdatedAt   string            `json:"updatedAt"`
	EndsAt      string            `json:"endsAt"`
	Status      alertStatus       `json:"status"`
}
type alertStatus struct {
	State       string   `json:"state"`
	SilencedBy  []string `json:"silencedBy"`
	InhibitedBy []string `json:"inhibitedBy"`
}
