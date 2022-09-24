package nagiosapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/synyx/gonagdash/pkg/connectors"
)

func TestNagiosCollector(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write([]byte(nagiosApiMockResponse))
	}))
	defer func() { testServer.Close() }()

	cfg := Config{
		Name: "test",
		URL:  testServer.URL,
	}

	var collector connectors.Connector = NewCollector(cfg)
	alerts, err := collector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if alerts == nil || len(alerts) != 3 {
		t.Error("There should be alerts")
	}
}

const nagiosApiMockResponse = `
{
  "content": {
    "example-test.example.com": {
      "active_checks_enabled": "1",
      "current_attempt": "1",
      "performance_data": {
        "rta": "5.907000ms",
        "pl": "0%"
      },
      "last_hard_state": "0",
      "notifications_enabled": "1",
      "current_state": "0",
      "downtimes": {},
      "plugin_output": "PING OK - Packet loss = 0%, RTA = 5.91 ms",
      "last_check": "1664018740",
      "problem_has_been_acknowledged": "0",
      "last_state_change": "1661957235",
      "scheduled_downtime_depth": "0",
      "services": {
        "http status example-http.example.com Port 8092": {
          "active_checks_enabled": "1",
          "current_attempt": "1",
          "performance_data": {
            "size": "210B",
            "time": "0,015777s"
          },
          "last_hard_state": "0",
          "notifications_enabled": "1",
          "current_state": "1",
          "downtimes": {},
          "plugin_output": "HTTP NOT OK: Status line output matched \" 2, 1, 3,400,401,403,404\" - 210 bytes in 0,016 second response time",
          "last_check": "1664018785",
          "problem_has_been_acknowledged": "0",
          "last_state_change": "1661957475",
          "scheduled_downtime_depth": "0",
          "comments": {},
          "last_notification": "0",
          "max_attempts": "4"
        },
        "CPU Load": {
          "active_checks_enabled": "1",
          "current_attempt": "1",
          "performance_data": {
            "load1": 0.01,
            "load15": 0.05,
            "load5": 0.03
          },
          "last_hard_state": "0",
          "notifications_enabled": "1",
          "current_state": "2",
          "downtimes": {},
          "plugin_output": "OK - load average: 0.01, 0.03, 0.05",
          "last_check": "1664017781",
          "problem_has_been_acknowledged": "0",
          "last_state_change": "1506339511",
          "scheduled_downtime_depth": "0",
          "comments": {},
          "last_notification": "0",
          "max_attempts": "4"
        },
        "SSL certificate httpd contargo-elevation.synyx.coffee:443": {
          "active_checks_enabled": "1",
          "current_attempt": "1",
          "performance_data": {
            "days_chain_elem1": 293,
            "days_chain_elem3": 1719,
            "days_chain_elem2": 1718
          },
          "last_hard_state": "0",
          "notifications_enabled": "1",
          "current_state": "3",
          "downtimes": {},
          "plugin_output": "SSL_CERT OK - x509 certificate '*.example.com' from 'Certum Domain Validation CA SHA2' valid until Jul 15 09:05:40 2023 GMT (expires in 293 days)",
          "last_check": "1664018916",
          "problem_has_been_acknowledged": "0",
          "last_state_change": "1661871772",
          "scheduled_downtime_depth": "0",
          "comments": {},
          "last_notification": "0",
          "max_attempts": "4"
        }
      }
    }
  },
  "success": true
}
`
