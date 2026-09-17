package icingaweb2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
)

func TestIcingaWeb2Connector(t *testing.T) {
	connector, mockServer := testConnector(mockHostResponse, mockServiceResponse)
	defer func() { mockServer.Close() }()

	alerts, err := connector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if alerts == nil || len(alerts) != 2 {
		t.Error("There should be alerts")
	}
}

func TestAckPropagation(t *testing.T) {
	hostJson := strings.ReplaceAll(mockHostResponse, "\"host_acknowledged\": \"0\"", "\"host_acknowledged\": \"1\"")
	connector, mockServer := testConnector(hostJson, mockServiceResponse)
	defer func() { mockServer.Close() }()

	alerts, err := connector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(alerts) > 0 {
		t.Error("There should be no alerts, as host is acked:", len(alerts), "alerts")
	}
}

func testConnector(hostJson, serviceJson string) (connectors.Connector, *httptest.Server) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		if strings.Contains(req.URL.Path, "/hosts") {
			_, _ = res.Write([]byte(hostJson))
		} else if strings.Contains(req.URL.Path, "/services") {
			_, _ = res.Write([]byte(serviceJson))
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

const mockHostResponse = `
[
  {
    "host_icon_image": "/icons-rhenus/idrac.jpg",
    "host_icon_image_alt": "n/a",
    "host_name": "example.com",
    "host_display_name": "example.com",
    "host_state": "1",
    "host_acknowledged": "0",
    "host_output": "CRITICAL - 172.25.2.115: Host unreachable @ 172.25.8.249. rta nan, lost 100%",
    "host_attempt": "1/3",
    "host_in_downtime": "0",
    "host_is_flapping": "0",
    "host_state_type": "1",
    "host_handled": "1",
    "host_last_state_change": "1787241186",
    "host_notifications_enabled": "1",
    "host_active_checks_enabled": "1",
    "host_passive_checks_enabled": "1",
    "host_check_command": "icmp",
    "host_next_update": "1789544834"
  }
]
`

const mockServiceResponse = `
[
  {
    "host_name": "example.com",
    "host_display_name": "example.com",
    "host_state": "1",
    "service_description": "Disk Space: /",
    "service_display_name": "Disk Space: /",
    "service_state": "1",
    "service_in_downtime": "0",
    "service_acknowledged": "0",
    "service_handled": "0",
    "service_output": "DISK OK - free space: / 5496 MiB (52.11% inode=83%);",
    "service_perfdata": "/=5051MiB;10022;10579;0;11136",
    "service_attempt": "1/6",
    "service_last_state_change": "1788642498",
    "service_icon_image": "",
    "service_icon_image_alt": "",
    "service_is_flapping": "0",
    "service_state_type": "1",
    "service_severity": "4",
    "service_notifications_enabled": "1",
    "service_active_checks_enabled": "1",
    "service_passive_checks_enabled": "1",
    "service_check_command": "disk",
    "service_next_update": "1789548592"
  }
]
`
