package patchman

import "time"

type host struct {
	Hostname            string `json:"hostname"`
	ReverseDNS          string `json:"reversedns"`
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

func parseTime(timeField string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999999", timeField)
	if err != nil {
		return time.Time{}
	}
	return t
}
