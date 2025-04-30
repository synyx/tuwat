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
	Id             string         `json:"id"`
	Status         string         `json:"status"`
	Severity       string         `json:"severity"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	Control        control        `json:"control"`
	Entity         entity         `json:"entity"`
	EntitySnapshot entitySnapShot `json:"entitySnapshot"`
	Project        project        `json:"project"`
	SourceRules    []sourceRule   `json:"sourceRules"`
}

type control struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type entity struct {
	Id   string `json:"id"`
	Type string `json:"type"`
	Name string `json:"name"`
}

type entitySnapShot struct {
	Id                      string `json:"id"`
	Type                    string `json:"type"`
	Name                    string `json:"name"`
	Status                  string `json:"status"`
	CloudPlatform           string `json:"cloudPlatform"`
	Region                  string `json:"region"`
	KubernetesClusterId     string `json:"kubernetesClusterId"`
	KubernetesClusterName   string `json:"kubernetesClusterName"`
	KubernetesNamespaceName string `json:"kubernetesNamespaceName"`
}

type project struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type sourceRule struct {
	TypeName    string `json:"__typeName"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
