package grafana

import (
	"context"
	"fmt"
	html "html/template"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/alertmanager"
	"github.com/synyx/tuwat/pkg/connectors/common"
)

type Connector struct {
	config Config
	ac     *alertmanager.Connector
}

type Config struct {
	Tag     string
	Cluster string
	common.HTTPConfig
}

func NewConnector(cfg *Config) *Connector {
	alertmanagerConfig := &alertmanager.Config{
		Tag:        cfg.Tag,
		Cluster:    cfg.Cluster,
		HTTPConfig: cfg.HTTPConfig,
	}
	alertmanagerConfig.URL += "/api/alertmanager/grafana"

	c := &Connector{config: *cfg, ac: alertmanager.NewConnector(alertmanagerConfig)}

	return c
}

func (c *Connector) Tag() string {
	return c.config.Tag
}

func (c *Connector) Collect(ctx context.Context) ([]connectors.Alert, error) {
	sourceAlerts, err := c.ac.Collect(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []connectors.Alert

	for _, alert := range sourceAlerts {
		alert.Description = alert.Labels["rulename"]
		alert.Details = alert.Labels["message"]
		labels := map[string]string{
			"Hostname": alert.Labels["grafana_folder"],
			"Contacts": alert.Labels["__contacts__"],
		}
		for k, v := range labels {
			alert.Labels[k] = v
		}

		alert.Links = []html.HTML{
			html.HTML("<a href=\"" + c.config.URL + "/alerting/grafana/" + alert.Labels["rule_uid"] + "/view?tab=instances" + "\" target=\"_blank\" alt=\"Alert\">🏠</a>"),
			html.HTML("<a href=\"" + c.config.URL + "/d/" + alert.Labels["__dashboardUid__"] + "\" target=\"_blank\" alt=\"Dashboard\">🏠</a>"),
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (c *Connector) String() string {
	return fmt.Sprintf("Grafana (%s)", c.config.URL)
}
