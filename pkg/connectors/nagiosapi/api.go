package nagiosapi

type Response struct {
	Content map[string]Host `json:"content"`
	Success bool            `json:"success"`
}

type Host struct {
	NotificationsEnabled       string             `json:"notifications_enabled"`
	CurrentState               string             `json:"current_state"`
	PluginOutput               string             `json:"plugin_output"`
	LastCheck                  string             `json:"last_check"`
	LastStateChange            string             `json:"last_state_change"`
	ProblemHasBeenAcknowledged string             `json:"problem_has_been_acknowledged"`
	CurrentAttempt             string             `json:"current_attempt"`
	MaxAttempts                string             `json:"max_attempts"`
	ScheduledDowntimeDepth     string             `json:"scheduled_downtime_depth"`
	Services                   map[string]Service `json:"services"`
}

type Service struct {
	CurrentState               string `json:"current_state"`
	PluginOutput               string `json:"plugin_output"`
	LastCheck                  string `json:"last_check"`
	LastStateChange            string `json:"last_state_change"`
	ProblemHasBeenAcknowledged string `json:"problem_has_been_acknowledged"`
	CurrentAttempt             string `json:"current_attempt"`
	MaxAttempts                string `json:"max_attempts"`
	NotificationsEnabled       string `json:"notifications_enabled"`
}
