package web

import (
	"net/http/httptest"
	"testing"

	"github.com/benbjohnson/clock"

	"github.com/synyx/tuwat/pkg/aggregation"
	"github.com/synyx/tuwat/pkg/config"
	"github.com/synyx/tuwat/pkg/connectors"
)

func TestRendering(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()

	cfg := &config.Config{}
	agg := aggregation.NewAggregator(cfg, clock.NewMock())
	wh := NewWebHandler(cfg, agg)

	renderer := wh.baseRenderer(req, "test", "alerts.gohtml")
	alerts := []aggregation.KnownAlert{
		{Alert: aggregation.Alert{
			Id:      "id",
			Where:   "where",
			Tag:     "tag",
			What:    "wha",
			Details: "detail",
			When:    0,
			Status:  connectors.Warning.String(),
			Labels:  nil,
		}, Downtime: "comment",
		},
	}
	aggregate := aggregation.Aggregate{
		CheckTime: clock.New().Now(),
		Downtimes: alerts,
	}
	renderer(w, 200, webContent{Content: aggregate})
}
