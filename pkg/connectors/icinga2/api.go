package icinga2

type HostResponse struct {
	Results []HostAttrs `json:"results"`
}
type HostAttrs struct {
	Host Host `json:"attrs"`
}
type Host struct {
	DisplayName         string  `json:"display_name"`
	State               int     `json:"state"`
	LastStateChange     float64 `json:"last_state_change"`
	Acknowledgement     int     `json:"acknowledgement"`
	DowntimeDepth       int     `json:"downtime_depth"`
	EnableNotifications bool    `json:"enable_notifications"`
	Output              string  `json:"output"`
	MaxCheckAttempts    int     `json:"max_check_attempts"`
	CheckAttempt        int     `json:"check_attempt"`
}

type ServiceResponse struct {
	Results []ServiceAttrs `json:"results"`
}
type ServiceAttrs struct {
	Service Service `json:"attrs"`
}
type Service struct {
	HostName            string      `json:"host_name"`
	DisplayName         string      `json:"display_name"`
	State               int         `json:"state"`
	LastStateChange     float64     `json:"last_state_change"`
	Acknowledgement     int         `json:"acknowledgement"`
	DowntimeDepth       int         `json:"downtime_depth"`
	EnableNotifications bool        `json:"enable_notifications"`
	LastCheckResult     CheckResult `json:"last_check_result"`
	MaxCheckAttempts    int         `json:"max_check_attempts"`
	CheckAttempt        int         `json:"check_attempt"`
}

type CheckResult struct {
	Output string `json:"output"`
}
