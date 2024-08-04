package icinga2

import (
	"context"
	"encoding/json"

	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/connectors"
)

func (c *Connector) CollectDowntimes(ctx context.Context) ([]connectors.Downtime, error) {
	body, err := c.get("/v1/objects/downtimes", ctx)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	decoder := json.NewDecoder(body)

	var response DowntimeResponse
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}

	downtimes := make([]connectors.Downtime, 0, len(response.Results))
	for _, dt := range response.Results {
		if !dt.Downtime.Active {
			continue
		}

		matchers := make(map[string]config.RuleMatcher)
		matchers["Hostname"] = config.ParseRuleMatcher("= " + dt.Downtime.HostName)
		if dt.Downtime.ServiceName != "" {
			matchers["Service"] = config.ParseRuleMatcher("= " + dt.Downtime.ServiceName)
		}

		downtime := connectors.Downtime{
			Author:    dt.Downtime.Author,
			Comment:   dt.Downtime.Comment,
			StartTime: parseTime(dt.Downtime.StartTime),
			EndTime:   parseTime(dt.Downtime.EndTime),
			Matchers:  matchers,
		}
		downtimes = append(downtimes, downtime)
	}

	return downtimes, nil
}
