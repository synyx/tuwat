package icinga2

type HostResponse struct {
	Results []HostAttrs `json:"results"`
}
type HostAttrs struct {
	Host Host `json:"attrs"`
}
type Host struct {
	DisplayName         string   `json:"display_name"`
	State               int      `json:"state"`
	LastStateChange     float64  `json:"last_state_change"`
	Acknowledgement     int      `json:"acknowledgement"`
	DowntimeDepth       int      `json:"downtime_depth"`
	EnableNotifications bool     `json:"enable_notifications"`
	Output              string   `json:"output"`
	MaxCheckAttempts    int      `json:"max_check_attempts"`
	CheckAttempt        int      `json:"check_attempt"`
	Groups              []string `json:"groups"`
	NotesUrl            string   `json:"notes_url"`
}

type ServiceResponse struct {
	Results []serviceAttrs `json:"results"`
}
type serviceAttrs struct {
	Service service `json:"attrs"`
}
type service struct {
	HostName            string      `json:"host_name"`
	Name                string      `json:"name"`
	DisplayName         string      `json:"display_name"`
	Zone                string      `json:"zone"`
	State               int         `json:"state"`
	LastStateChange     float64     `json:"last_state_change"`
	Acknowledgement     int         `json:"acknowledgement"`
	DowntimeDepth       int         `json:"downtime_depth"`
	EnableNotifications bool        `json:"enable_notifications"`
	LastCheckResult     checkResult `json:"last_check_result"`
	MaxCheckAttempts    int         `json:"max_check_attempts"`
	CheckAttempt        int         `json:"check_attempt"`
	Groups              []string    `json:"groups"`
	NotesUrl            string      `json:"notes_url"`
}

type checkResult struct {
	Output string `json:"output"`
}
