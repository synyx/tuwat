package wizio

import "time"

type issuesResponse struct {
	Data data `json:"data"`
}
type data struct {
	IssuesV2 issuesV2 `json:"issuesV2"`
}

type issuesV2 struct {
	TotalCount int    `json:"totalCount"`
	Nodes      []node `json:"nodes"`
}

type node struct {
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
