package patchman

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/synyx/tuwat/pkg/connectors"
)

func TestNagiosConnector(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write([]byte(patchmanApiMockResponse))
	}))
	defer func() { testServer.Close() }()

	cfg := Config{
		Tag: "test",
		HTTPConfig: connectors.HTTPConfig{
			URL: testServer.URL,
		},
	}

	var connector connectors.Connector = NewConnector(cfg)
	alerts, err := connector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if alerts == nil || len(alerts) != 2 {
		t.Error("There should be alerts")
	}
}

const patchmanApiMockResponse = `
[
  {
    "id": 98,
    "hostname": "example-host.example.com",
    "ipaddress": "192.168.0.37",
    "reversedns": "192-168-0-37.local",
    "check_dns": true,
    "os": "http://patchman.example.com/api/os/17/",
    "kernel": "5.20.0-107-generic",
    "arch": "http://patchman.example.com/api/machine-architecture/1/",
    "domain": "http://patchman.example.com/api/domain/1/",
    "lastreport": "2022-05-03T02:12:38.101661",
    "repos": [
      "http://patchman.example.com/api/repo/63/",
      "http://patchman.example.com/api/repo/55/",
      "http://patchman.example.com/api/repo/106/"
    ],
    "updates": [],
    "reboot_required": true,
    "host_repos_only": true,
    "tags": "Server prod example_support",
    "updated_at": "2022-05-23T13:29:14",
    "bugfix_update_count": 0,
    "security_update_count": 0
  },
  {
    "id": 102,
    "hostname": "test-host-test.example.com",
    "ipaddress": "192.168.0.37",
    "reversedns": "192-168-0-37.local",
    "check_dns": true,
    "os": "http://patchman.example.com/api/os/20/",
    "kernel": "4.0.0-20-amd64",
    "arch": "http://patchman.example.com/api/machine-architecture/1/",
    "domain": "http://patchman.example.com/api/domain/1/",
    "lastreport": "2022-05-03T03:45:31.785462",
    "repos": [
      "http://patchman.example.com/api/repo/113/",
      "http://patchman.example.com/api/repo/114/"
    ],
    "updates": [],
    "reboot_required": false,
    "host_repos_only": true,
    "tags": "Server test goo_ops cpatch",
    "updated_at": "2022-05-23T13:29:14",
    "bugfix_update_count": 0,
    "security_update_count": 10
  }
]
`
