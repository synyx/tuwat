package wizio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	html "html/template"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
)

type Connector struct {
	config Config
	client *http.Client
}

type Config struct {
	Tag            string
	StatusFilter   []string
	SeverityFilter []string
	common.HTTPConfig
	NumberOfIssues int
}

func NewConnector(cfg *Config) *Connector {
	// wiz.io requires the audience to be set. We only oerwrite it if Oauth2Creds are present and the audience is empty
	if cfg.OAuth2Creds.EndpointParams == nil {
		cfg.OAuth2Creds.EndpointParams = map[string][]string{}
	}
	if cfg.OAuth2Creds.EndpointParams["audience"] == nil {
		cfg.OAuth2Creds.EndpointParams["audience"] = []string{"wiz-api"}
	}

	if cfg.NumberOfIssues == 0 {
		cfg.NumberOfIssues = 10
	}
	return &Connector{*cfg, cfg.HTTPConfig.Client()}
}

func (c *Connector) Tag() string {
	return c.config.Tag
}

func (c *Connector) Collect(ctx context.Context) ([]connectors.Alert, error) {
	issueResponse, err := c.collectIssues(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert

	for _, node := range issueResponse.Data.IssuesV2.Nodes {
		description := node.SourceRules[0].Name
		namespace := node.EntitySnapshot.KubernetesNamespaceName
		if namespace == "" {
			if n, ok := node.EntitySnapshot.Tags["kustomize.toolkit.fluxcd.io/namespace"]; ok {
				namespace = n
			}
		}

		var projects []string
		for _, p := range node.Projects {
			projects = append(projects, p.Name)
		}
		slices.Sort(projects)

		labels := map[string]string{}
		maps.Copy(labels, node.EntitySnapshot.Tags)
		maps.Copy(labels, map[string]string{
			"IssueId":    node.Id,
			"Entity":     node.EntitySnapshot.Name,
			"EntityType": node.EntitySnapshot.Type,
			"Status":     node.Status,
			"Severity":   node.Severity,
			"Source":     c.config.URL,
			"Cluster":    node.EntitySnapshot.KubernetesClusterName,
			"Namespace":  namespace,
			"Type":       "Issue",
			"Projects":   strings.Join(projects, ","),
		})

		if namespace == "" {
			labels["Hostname"] = node.EntitySnapshot.Name
		}

		// This weburl includes the state filter for issues shown in the background filter
		webUrl := "https://app.wiz.io/issues#%7E%28filters%7E%28status%7E%28equals%7E%28%7E%27OPEN%7E%27IN_PROGRESS%29%29%29%7Eissue%7E%27" + node.Id + "%29"

		links := []html.HTML{
			html.HTML("<a href=\"" + webUrl + "\" target=\"_blank\" alt=\"Home\">🏠</a>"),
		}

		for _, ticket := range node.ServiceTickets {
			links = append(links, html.HTML("<a href=\""+ticket.Url+"\" target=\"_blank\" alt=\""+ticket.Name+"\">🎫</a>"))
		}

		alert := connectors.Alert{
			Labels:      labels,
			Start:       node.CreatedAt,
			State:       mapState(node.Severity),
			Description: description,
			Details:     node.SourceRules[0].Description,
			Links:       links,
		}
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Connector) String() string {
	return fmt.Sprintf("Wiz.io Issues (%s)", c.config.URL)
}

// collectIssues builds requests issues from wiz.io via graphQL query
//
// see https://win.wiz.io/reference/quickstart for a quick reference
// see https://win.wiz.io/reference/issues-query for the issues api
// see https://docs.wiz.io/docs/how-actions-and-automation-rules-work#issues
//
//	for an overview of possible fields, as the template fields seem to be equivalent to the graphql fields
func (c *Connector) collectIssues(ctx context.Context) (*issuesResponse, error) {
	graphqlQuery := `
		query IssuesTable(
			$filterBy: IssueFilters $first: Int $after: String $orderBy: IssueOrder
			) {
				issuesV2(
					filterBy: $filterBy
					first: $first
					after: $after
					orderBy: $orderBy
				) { 
					nodes {
						id
						createdAt 
						updatedAt
						projects {
							id
							name
						}
						status
						severity
						entitySnapshot {
							id
							type
							name
							status
							kubernetesClusterName
							kubernetesNamespaceName
							tags
						}
						notes {
							text
						    createdAt
						    user {
							    id
							    name
						    }
						}
						serviceTickets {
							name
							url
						}
					  	sourceRules {
							__typename
							id
							name
							description
						  }
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			 }
    `

	query := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query: graphqlQuery,
		Variables: map[string]interface{}{
			"first": c.config.NumberOfIssues,
			"filterBy": map[string]interface{}{
				"status":   c.config.StatusFilter,
				"severity": c.config.SeverityFilter,
			},
			"orderBy": map[string]interface{}{
				"direction": "DESC",
				"field":     "SEVERITY",
			},
		},
	}

	from := "/graphql"

	marshal, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	body, err := c.get(ctx, from, string(marshal))
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(body)

	var issues issuesResponse
	err = decoder.Decode(&issues)
	if err != nil {
		slog.ErrorContext(ctx, "Cannot parse",
			slog.String("url", c.config.URL+from),
			slog.Any("error", err),
		)
		return nil, err
	}

	return &issues, nil
}

func (c *Connector) collectThreats(ctx context.Context) (*issuesResponse, error) {
	graphqlQuery := `
		query ThreatTable(
			$filterBy: ThreatCenterItemFilters
			$first: Int
			$after: String
			$orderBy: ThreatCenterItemOrder
			) {
				threatCenterItems(
					filterBy: $filterBy
					first: $first
					after: $after
					orderBy: $orderBy
				) { 
					nodes {
						id
						createdAt 
						updatedAt
						projects {
							id
							name
						}
						status
						severity
						entitySnapshot {
							id
							type
							name
							status
							kubernetesClusterName
							kubernetesNamespaceName
							tags
						}
						notes {
							text
						    createdAt
						    user {
							    id
							    name
						    }
						}
						serviceTickets {
							name
							url
						}
					  	sourceRules {
							__typename
							id
							name
							description
						  }
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			 }
    `

	query := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query: graphqlQuery,
		Variables: map[string]interface{}{
			"first": c.config.NumberOfIssues,
			"filterBy": map[string]interface{}{
				"status":   c.config.StatusFilter,
				"severity": c.config.SeverityFilter,
			},
			"orderBy": map[string]interface{}{
				"direction": "DESC",
				"field":     "SEVERITY",
			},
		},
	}

	from := "/graphql"

	marshal, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	body, err := c.get(ctx, from, string(marshal))
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(body)

	var issues issuesResponse
	err = decoder.Decode(&issues)
	if err != nil {
		slog.ErrorContext(ctx, "Cannot parse",
			slog.String("url", c.config.URL+from),
			slog.Any("error", err),
		)
		return nil, err
	}

	return &issues, nil
}

func (c *Connector) get(ctx context.Context, endpoint string, query string) (io.ReadCloser, error) {
	slog.DebugContext(ctx, "getting alerts", slog.String("url", c.config.URL+endpoint))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.URL+endpoint, bytes.NewBuffer([]byte(query)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func mapState(severity string) connectors.State {
	switch severity {
	case "CRITICAL":
		return connectors.Critical
	case "HIGH":
		return connectors.Warning
	case "MEDIUM":
		return connectors.Warning
	case "LOW":
		return connectors.Warning
	case "INFORMATION":
		return connectors.Warning
	default:
		return connectors.Unknown
	}
}
