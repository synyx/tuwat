package patchman

type Host struct {
	Hostname            string `json:"hostname"`
	LastReport          string `json:"lastreport"`
	RebootRequired      bool   `json:"reboot_required"`
	BugfixUpdateCount   int    `json:"bugfix_update_count"`
	SecurityUpdateCount int    `json:"security_update_count"`
	UpdatedAt           string `json:"updated_at"`
	Tags                string `json:"tags"`
}
