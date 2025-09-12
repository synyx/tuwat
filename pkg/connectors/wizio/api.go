package wizio

import "time"

type issuesResponse struct {
	Data issuesData `json:"data"`
}
type issuesData struct {
	IssuesV2 issuesV2 `json:"issuesV2"`
}

type issuesV2 struct {
	TotalCount int     `json:"totalCount"`
	Nodes      []issue `json:"nodes"`
}

type issue struct {
	Id             string          `json:"id"`
	Status         string          `json:"status"`
	Severity       string          `json:"severity"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	EntitySnapshot entitySnapShot  `json:"entitySnapshot"`
	ServiceTickets []serviceTicket `json:"serviceTickets"`
	SourceRules    []sourceRule    `json:"sourceRules"`
	Projects       []project       `json:"projects"`
}

type entitySnapShot struct {
	Id                      string            `json:"id"`
	Type                    string            `json:"type"`
	Name                    string            `json:"name"`
	Status                  string            `json:"status"`
	KubernetesClusterName   string            `json:"kubernetesClusterName"`
	KubernetesNamespaceName string            `json:"kubernetesNamespaceName"`
	Tags                    map[string]string `json:"tags"`
}

type serviceTicket struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type sourceRule struct {
	TypeName    string `json:"__typeName"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type project struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type orderDirection string

const (
	Ascending  orderDirection = "ASC"
	Descending orderDirection = "DESC"
)

type issueOrderField string

const (
	Id                  issueOrderField = "ID"
	Severity            issueOrderField = "SEVERITY"
	CreatedAt           issueOrderField = "CREATED_AT"
	ResolvedAt          issueOrderField = "RESOLVED_AT"
	StatusChangedAt     issueOrderField = "STATUS_CHANGED_AT"
	ThreatLastGroupedAt issueOrderField = "THREAT_LAST_GROUPED_AT"
)

type issueOrder struct {
	Direction orderDirection  `json:"direction"`
	Field     issueOrderField `json:"field"`
}

type issueStatus string

const (
	Open       issueStatus = "OPEN"
	InProgress issueStatus = "IN_PROGRESS"
	Resolved   issueStatus = "RESOLVED"
	Rejected   issueStatus = "REJECTED"
)

type severity string

const (
	Informational severity = "INFORMATIONAL"
	Low           severity = "LOW"
	Medium        severity = "MEDIUM"
	High          severity = "HIGH"
	Critical      severity = "CRITICAL"
)

type issueFilters struct {
	Status   []issueStatus `json:"status"`
	Severity []severity    `json:"severity"`
}
