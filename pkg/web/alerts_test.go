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

	renderer := wh.baseRenderer(req, "test", "_base.gohtml", "alerts.gohtml")
	alerts := []aggregation.Alert{
		{
			Id:      "asdf",
			Where:   "where",
			Tag:     "tag",
			What:    "what",
			Details: "details",
			When:    0,
			Status:  connectors.Warning.String(),
		},
	}
	aggregate := aggregation.Aggregate{
		CheckTime: clock.New().Now(),
		Alerts:    alerts,
	}
	renderer(w, 200, webContent{Content: aggregate})
}
