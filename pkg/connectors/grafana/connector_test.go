package grafana

import (
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
)

func TestConnector(t *testing.T) {
	connector, closer := mockConnector()
	defer closer()
	alerts, err := connector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if alerts == nil || len(alerts) != 3 {
		t.Error("There should be alerts")
	}
}

func mockConnector() (connectors.Connector, func()) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write([]byte(mockResponse))
	}))

	cfg := Config{
		Tag: "test",
		HTTPConfig: common.HTTPConfig{
			URL: testServer.URL,
		},
	}

	return NewConnector(&cfg), func() { testServer.Close() }
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
            "state": "inactive",
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
                  "__contacts__": "\"Trucking Team\",\"jbuch mail\"",
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
                "state": "Normal",
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
              "__contacts__": "\"Trucking Team\",\"jbuch mail\"",
              "rule_uid": "kbMKlW04z"
            },
            "health": "ok",
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

func TestConnector_Collect(t *testing.T) {
	r := regexp.MustCompile(`in namespace\W+([a-zA-Z-0-9_-]+)`)
	details := "constraint violation of kind ContainerLimits in Pod gitlab-agent-landingpage-659cf9567d-kkxsl in namespace api-gateway-stage\n\t\t"
	where := ""
	if s := r.FindAllStringSubmatch(details, 1); len(s) > 0 {
		where = s[0][1]
	}
	if where != "api-gateway-stage" {
		t.Fail()
	}
}
