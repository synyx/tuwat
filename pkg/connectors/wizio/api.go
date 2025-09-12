package wizio

import "time"

type issuesResponse struct {
	Data issueData `json:"data"`
}
type issueData struct {
	IssuesV2 issuesV2 `json:"issuesV2"`
}

type issuesV2 struct {
	TotalCount int         `json:"totalCount"`
	Nodes      []issueNode `json:"nodes"`
}

type issueNode struct {
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

type threatCenterItemOrderField string

const (
	PublishedAt threatCenterItemOrderField = "PUBLISHED_AT"
	PinnedAt    threatCenterItemOrderField = "PINNED_AT"
)

type threatCenterOrder struct {
	Direction orderDirection             `json:"direction"`
	Field     threatCenterItemOrderField `json:"field"`
}

type threatCenterItem struct {
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

type threatCenterItems struct {
	TotalCount  int                `json:"totalCount"`
	ThreatNodes []threatCenterItem `json:"nodes"`
}
