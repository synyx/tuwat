package patchman

type host struct {
	Hostname            string `json:"hostname"`
	LastReport          string `json:"lastreport"`
	RebootRequired      bool   `json:"reboot_required"`
	BugfixUpdateCount   int    `json:"bugfix_update_count"`
	SecurityUpdateCount int    `json:"security_update_count"`
	UpdatedAt           string `json:"updated_at"`
	Tags                string `json:"tags"`
	OSURL               string `json:"os"`
	ArchURL             string `json:"arch"`
	DomainURL           string `json:"domain"`
}

type os struct {
	Name    string `json:"name"`
	OSGroup string `json:"osgroup"`
}

type arch struct {
	Name string `json:"name"`
}

type domain struct {
	Name string `json:"name"`
}
