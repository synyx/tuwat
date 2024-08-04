package icinga2

import (
	"context"
	"encoding/json"

	"github.com/synyx/tuwat/pkg/connectors"
)

func (c *Connector) CollectDowntime(ctx context.Context) ([]connectors.Downtime, error) {
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

		downtime := connectors.Downtime{
			Author:    dt.Downtime.Author,
			Comment:   dt.Downtime.Comment,
			StartTime: floatToTime(dt.Downtime.StartTime),
			EndTime:   floatToTime(dt.Downtime.EndTime),
		}
		downtimes = append(downtimes, downtime)
	}

	return downtimes, nil
}
