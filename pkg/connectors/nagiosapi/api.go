package nagiosapi

type response struct {
	Content map[string]host `json:"content"`
	Success bool            `json:"success"`
}

type host struct {
	NotificationsEnabled       string             `json:"notifications_enabled"`
	CurrentState               string             `json:"current_state"`
	PluginOutput               string             `json:"plugin_output"`
	LastCheck                  string             `json:"last_check"`
	LastStateChange            string             `json:"last_state_change"`
	ProblemHasBeenAcknowledged string             `json:"problem_has_been_acknowledged"`
	CurrentAttempt             string             `json:"current_attempt"`
	MaxAttempts                string             `json:"max_attempts"`
	ScheduledDowntimeDepth     string             `json:"scheduled_downtime_depth"`
	Services                   map[string]service `json:"services"`
}

type service struct {
	CurrentState               string `json:"current_state"`
	PluginOutput               string `json:"plugin_output"`
	LastCheck                  string `json:"last_check"`
	LastStateChange            string `json:"last_state_change"`
	ProblemHasBeenAcknowledged string `json:"problem_has_been_acknowledged"`
	CurrentAttempt             string `json:"current_attempt"`
	MaxAttempts                string `json:"max_attempts"`
	NotificationsEnabled       string `json:"notifications_enabled"`
}
