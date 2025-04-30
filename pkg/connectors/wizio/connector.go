package wizio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
	"io"
	"log/slog"
	"net/http"
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
		descr := "Issue: " + node.SourceRules[0].Name
		alert := connectors.Alert{
			Labels: map[string]string{
				"Entity":     node.EntitySnapshot.Name,
				"EntityType": node.EntitySnapshot.Type,
				"Status":     node.Status,
				"Severity":   node.Severity,
				"Source":     c.config.URL,
				"Cluster":    node.EntitySnapshot.KubernetesClusterName,
				"Namespace":  node.EntitySnapshot.KubernetesNamespaceName,
			},
			Start:       node.CreatedAt,
			State:       mapState(node.Severity),
			Description: descr,
			Details:     node.SourceRules[0].Description,
			//Links: []html.HTML{
			//	html.HTML("<a href=\"" + mr.WebUrl + "\" target=\"_blank\" alt=\"Home\">🏠</a>"),
			//},
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
	// TODO
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
						control {
							id 
							name
						}
						createdAt 
						updatedAt
						dueAt
						project {
							id
							name
							slug
							businessUnit
							riskProfile {
								businessImpact
							}
						}
						status
						severity
						entity {  
							id  
							name  
							type  
						}
						entitySnapshot {
							id
							type
							name
							status 
							cloudPlatform
							region
							kubernetesClusterId
							kubernetesClusterName
							kubernetesNamespaceName
						}
						note
						serviceTickets {
							externalId
							name
							url
						}

					  	sourceRules {
							__typename
							... on Control {
								id
								name
								description
								resolutionRecommendation
								risks
								securitySubCategories {
									title
									category {
										name
										framework {
											name
										}
									}
								}
							}
							... on CloudEventRule {
								id
								name
								description
								sourceType
								type
								risks
								securitySubCategories {
									title
									category {
										name
										framework {
											name
										}
									}
								}
							}
							... on CloudConfigurationRule {
								 id
								 name
								 description
								 remediationInstructions
								 serviceType
								 risks
								 securitySubCategories {
									title
									category {
										name
										framework {
											name
										}
									}
								 }
							}
						  }
					}
					pageInfo {
						hasNextPage
						endCursor
					}
				}
			 }
    `

	numberOfIssues := c.config.NumberOfIssues
	if numberOfIssues == 0 {
		numberOfIssues = 10
	}

	query := struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query: graphqlQuery,
		Variables: map[string]interface{}{
			"first": numberOfIssues,
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
