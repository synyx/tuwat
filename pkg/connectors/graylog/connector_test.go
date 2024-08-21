package graylog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/synyx/tuwat/pkg/connectors/common"
)

func TestConnector(t *testing.T) {
	connector, mockServer := testConnector(map[string]string{
		"/api/streams/alerts/paginated": mockAlertsResult,
	})
	defer func() { mockServer.Close() }()

	alerts, err := connector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if alerts == nil || len(alerts) != 1 {
		t.Error("There should be alerts")
	}

	if alerts[0].Start.IsZero() {
		t.Error("alert start is zero")
	}
}

func testConnector(endpoints map[string]string) (*Connector, *httptest.Server) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		for endpoint, body := range endpoints {
			if strings.Contains(req.URL.Path, endpoint) {
				_, _ = res.Write([]byte(body))
			}
		}
	}))

	cfg := Config{
		Tag: "test",
		HTTPConfig: common.HTTPConfig{
			URL: mockServer.URL,
		},
	}

	return NewConnector(&cfg), mockServer
}

const mockAlertsResult = `
{
  "total": 1,
  "alerts": [
    {
      "id": "62e0dfe9f9610d7126e4c0c4",
      "description": "Stream had 1 messages in the last 5 minutes with trigger condition more than 0 messages. (Current grace time: 5 minutes)",
      "condition_id": "e77371ef-9ae6-4dc3-8514-8afa041c882e",
      "stream_id": "59ca447e94afad03ad598f02",
      "condition_parameters": {
        "backlog": 5,
        "repeat_notifications": false,
        "grace": 5,
        "threshold_type": "MORE",
        "threshold": 0,
        "time": 5
      },
 	  "triggered_at": "2022-07-27T06:49:13.231Z",
      "resolved_at": "2022-07-27T06:53:13.240Z",
      "is_interval": true
    }
  ]
}
`
