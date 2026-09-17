package icingaweb2

type HostResponse []host

type host struct {
	DisplayName         string `json:"host_display_name"`
	State               string `json:"host_state"`
	LastStateChange     string `json:"host_last_state_change"`
	Acknowledgement     string `json:"host_acknowledged"`
	Downtime            string `json:"host_in_downtime"`
	EnableNotifications string `json:"host_notifications_enabled"`
	Output              string `json:"host_output"`
	CheckAttempt        string `json:"host_attempt"`
}

type ServiceResponse []service

type service struct {
	HostName            string `json:"host_display_name"`
	Name                string `json:"service_display_name"`
	DisplayName         string `json:"service_description"`
	State               string `json:"service_state"`
	LastStateChange     string `json:"service_last_state_change"`
	Acknowledgement     string `json:"service_acknowledged"`
	Downtime            string `json:"service_in_downtime"`
	EnableNotifications string `json:"service_notifications_enabled"`
	LastCheckResult     string `json:"service_output"`
	CheckAttempt        string `json:"service_attempt"`
}
