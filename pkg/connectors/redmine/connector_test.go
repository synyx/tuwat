package redmine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
)

func TestRedmineConnector(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write([]byte(redmineApiMockResponse))
	}))
	defer func() { testServer.Close() }()

	cfg := Config{
		Tag: "test",
		HTTPConfig: common.HTTPConfig{
			URL: testServer.URL,
		},
	}

	var connector connectors.Connector = NewConnector(&cfg)
	alerts, err := connector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if alerts == nil || len(alerts) != 1 {
		t.Error("There should be alerts")
	}
}

const redmineApiMockResponse = `
{
  "issues": [
    {
      "id": 7,
      "project": {
        "id": 5,
        "name": "Infrastruktur"
      },
      "tracker": {
        "id": 6,
        "name": "User Story"
      },
      "status": {
        "id": 2,
        "name": "work in progress"
      },
      "priority": {
        "id": 4,
        "name": "Normal"
      },
      "author": {
        "id": 1,
        "name": "Foo"
      },
      "assigned_to": {
        "id": 2,
        "name": "Bar"
      },
      "subject": "Security things",
      "description": "",
      "start_date": "2023-06-19",
      "due_date": "2023-06-19",
      "done_ratio": 54,
      "is_private": false,
      "estimated_hours": null,
      "created_on": "2023-06-12T10:08:19+02:00",
      "updated_on": "2023-06-28T16:53:24+02:00",
      "closed_on": null
    }
  ],
  "total_count": 1,
  "offset": 0,
  "limit": 25
}
`
