package grafana

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
		"/api/prometheus/grafana/api/v1/rules": mockResponse,
	})
	defer func() { mockServer.Close() }()

	alerts, err := connector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(alerts) == 0 {
		t.Error("There should be alerts")
	}
}

func testConnector(endpoints map[string]string) (*Connector, *httptest.Server) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)

		for endpoint, body := range endpoints {
			if strings.HasPrefix(req.URL.Path, endpoint) {
				if _, err := res.Write([]byte(body)); err != nil {
					panic(err)
				}
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

const mockResponse = `
{
  "status": "success",
  "data": {
    "groups": [
      {
        "name": "failed authentications alert",
        "file": "Folder",
        "rules": [
          {
            "state": "alerting",
            "name": "Consumed no things alert",
            "query": "",
            "annotations": {
              "__alertId__": "81",
              "__dashboardUid__": "UlpdFLWMz",
              "__panelId__": "7",
              "message": "Long Message"
            },
            "alerts": [
              {
                "labels": {
                  "__contacts__": "\"Team\",\"jbuch mail\"",
                  "alertname": "Consumed no things alert",
                  "grafana_folder": "Folder",
                  "rule_uid": "kbMKlW04z"
                },
                "annotations": {
                  "__alertId__": "81",
                  "__dashboardUid__": "UlpdFLWMz",
                  "__panelId__": "7",
                  "message": "Long Message"
                },
                "state": "Alerting",
                "activeAt": "2024-08-13T12:41:40+02:00",
                "value": ""
              }
            ],
            "totals": {
              "normal": 1
            },
            "totalsFiltered": {
              "normal": 1
            },
            "labels": {
              "__contacts__": "\"Team\",\"jbuch mail\"",
              "rule_uid": "kbMKlW04z"
            },
            "health": "nodata",
            "type": "alerting",
            "lastEvaluation": "2024-08-30T15:18:40+02:00",
            "evaluationTime": 6.723146319
          }
        ],
        "totals": {
          "inactive": 1
        },
        "interval": 60,
        "lastEvaluation": "2024-08-30T15:18:40+02:00",
        "evaluationTime": 6.723146319
      }
    ],
    "totals": {
      "inactive": 10,
      "nodata": 5
    }
  }
}
`
