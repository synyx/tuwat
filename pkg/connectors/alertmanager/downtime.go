package alertmanager

import (
	"context"
	"encoding/json"

	"github.com/synyx/tuwat/pkg/connectors"
)

func (c *Connector) CollectDowntimes(ctx context.Context) ([]connectors.Downtime, error) {
	res, err := c.get(ctx, "/silences")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)

	var response []silence
	if err := decoder.Decode(&response); err != nil {
		return nil, err
	}

	downtimes := make([]connectors.Downtime, 0, len(response))
	for _, dt := range response {
		if dt.Status.State != SilenceActive {
			continue
		}

		downtime := connectors.Downtime{
			Author:    dt.CreatedBy,
			Comment:   dt.Comment,
			StartTime: parseTime(ctx, dt.StartsAt),
			EndTime:   parseTime(ctx, dt.EndsAt),
		}
		downtimes = append(downtimes, downtime)
	}

	return downtimes, nil
}
