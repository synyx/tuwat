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
		"/api/events/search": mockEventSearchResult,
	})
	defer func() { mockServer.Close() }()

	alerts, err := connector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if alerts == nil || len(alerts) != 1 {
		t.Fatal("There should be alerts")
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

const mockEventSearchResult = `
{
  "events": [
    {
      "event": {
        "id": "01J5TR06DQQEXXQEY7T449GAFJ",
        "event_definition_type": "aggregation-v1",
        "event_definition_id": "62e11a902f23dc2537db9efd",
        "origin_context": null,
        "timestamp": "2024-08-21T15:13:28.969Z",
        "timestamp_processing": "2024-08-21T15:13:35.159Z",
        "timerange_start": "2024-08-21T15:08:28.969Z",
        "timerange_end": "2024-08-21T15:13:28.969Z",
        "streams": [
          "000000000000000000000002"
        ],
        "source_streams": [],
        "message": "Error(s) occured: count()=131036.0",
        "source": "source.example.net",
        "key_tuple": [],
        "key": null,
        "priority": 2,
        "alert": true,
        "fields": {},
        "group_by_fields": {}
      },
      "index_name": "gl-events_24",
      "index_type": "message"
    }
  ],
  "used_indices": [
    "gl-events_24",
    "gl-system-events_1"
  ],
  "parameters": {
    "page": 1,
    "per_page": 25,
    "timerange": {
      "type": "relative",
      "range": 60
    },
    "query": "",
    "filter": {
      "alerts": "only",
      "event_definitions": []
    },
    "sort_by": "timestamp",
    "sort_direction": "desc"
  },
  "total_events": 1,
  "duration": 3,
  "context": {
    "event_definitions": {
      "62e11a902f23dc2537db9efd": {
        "id": "62e11a902f23dc2537db9efd",
        "title": "Error(s) occured",
        "description": "Migrated message count alert condition"
      }
    },
    "streams": {
      "000000000000000000000002": {
        "id": "000000000000000000000002",
        "title": "All events",
        "description": "Stream containing all events created by Graylog"
      }
    }
  }
}
`
