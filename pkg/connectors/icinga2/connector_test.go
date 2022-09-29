package icinga2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/synyx/gonagdash/pkg/connectors"
)

func TestIcinga2Connector(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		if strings.Contains(req.URL.Path, "/host") {
			_, _ = res.Write([]byte(icinga2MockHostResponse))
		} else if strings.Contains(req.URL.Path, "/service") {
			_, _ = res.Write([]byte(icinga2MockServiceResponse))
		}
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

const icinga2MockHostResponse = `
{
  "results": [
    {
      "attrs": {
        "__name": "test-host.example.com",
        "acknowledgement": 0,
        "acknowledgement_expiry": 0,
        "acknowledgement_last_change": 0,
        "action_url": "",
        "active": true,
        "address": "10.0.5.22",
        "address6": "",
        "check_attempt": 1,
        "check_command": "hostalive",
        "check_interval": 300,
        "check_period": "",
        "check_timeout": null,
        "command_endpoint": "",
        "display_name": "test-host.example.com",
        "downtime_depth": 0,
        "enable_active_checks": true,
        "enable_event_handler": true,
        "enable_flapping": false,
        "enable_notifications": true,
        "enable_passive_checks": true,
        "enable_perfdata": true,
        "event_command": "",
        "executions": null,
        "flapping": false,
        "flapping_current": 0,
        "flapping_ignore_states": null,
        "flapping_last_change": 0,
        "flapping_threshold": 0,
        "flapping_threshold_high": 30,
        "flapping_threshold_low": 25,
        "force_next_check": false,
        "force_next_notification": false,
        "groups": [
          "non-puppet-hosts"
        ],
        "ha_mode": 0,
        "handled": false,
        "icon_image": "",
        "icon_image_alt": "",
        "last_check": 1664024025.466623,
        "last_check_result": {
          "active": true,
          "check_source": "test-icinga2.example.com",
          "command": [
            "/usr/lib/nagios/plugins/check_ping",
            "-H",
            "10.0.5.22",
            "-c",
            "5000,100%",
            "-w",
            "3000,80%"
          ],
          "execution_end": 1664024025.466566,
          "execution_start": 1664024021.44423,
          "exit_status": 0,
          "output": "PING OK - Packet loss = 0%, RTA = 3.01 ms",
          "performance_data": [
            "rta=3.005000ms;3000.000000;5000.000000;0.000000",
            "pl=0%;80;100;0"
          ],
          "previous_hard_state": 0,
          "schedule_end": 1664024025.466623,
          "schedule_start": 1664024021.443825,
          "scheduling_source": "test-icinga2.example.com",
          "state": 1,
          "ttl": 0,
          "type": "CheckResult",
          "vars_after": {
            "attempt": 1,
            "reachable": true,
            "state": 0,
            "state_type": 1
          },
          "vars_before": {
            "attempt": 1,
            "reachable": true,
            "state": 0,
            "state_type": 1
          }
        },
        "last_hard_state": 0,
        "last_hard_state_change": 1657865500.550281,
        "last_reachable": true,
        "last_state": 0,
        "last_state_change": 1657865500.550281,
        "last_state_down": 1642064653.064841,
        "last_state_type": 1,
        "last_state_unreachable": 0,
        "last_state_up": 1664024025.466566,
        "max_check_attempts": 3,
        "name": "test-host.example.com",
        "next_check": 1664024321.956668,
        "next_update": 1664024630.00215,
        "notes": "",
        "notes_url": "",
        "original_attributes": null,
        "package": "_etc",
        "paused": false,
        "previous_state_change": 1657865500.550281,
        "problem": false,
        "retry_interval": 60,
        "severity": 0,
        "source_location": {
          "first_column": 1,
          "first_line": 62,
          "last_column": 34,
          "last_line": 62,
          "path": "/etc/icinga2/conf.d/../netbox/zones.d/example.com/hosts.conf"
        },
        "state": 0,
        "state_type": 1,
        "templates": [
          "test-host.example.com",
          "generic-host"
        ],
        "type": "Host",
        "vars": {
          "platform": "unknown",
          "platform_id": 11,
          "role": "dect-acess-point",
          "role_id": 24
        },
        "version": 0,
        "volatile": false,
        "zone": "example.com"
      },
      "joins": {},
      "meta": {},
      "name": "test-host.example.com",
      "type": "Host"
    }
  ]
}
`

const icinga2MockServiceResponse = `
{
  "results": [
    {
      "attrs": {
        "__name": "test-host.example.com!puppet-dns-test-host.example.com",
        "acknowledgement": 0,
        "acknowledgement_expiry": 0,
        "acknowledgement_last_change": 0,
        "action_url": "",
        "active": true,
        "check_attempt": 1,
        "check_command": "dns",
        "check_interval": 300,
        "check_period": "",
        "check_timeout": null,
        "command_endpoint": "",
        "display_name": "DNS: test-host.example.com",
        "downtime_depth": 0,
        "enable_active_checks": true,
        "enable_event_handler": true,
        "enable_flapping": false,
        "enable_notifications": true,
        "enable_passive_checks": true,
        "enable_perfdata": true,
        "event_command": "",
        "executions": null,
        "flapping": false,
        "flapping_current": 0,
        "flapping_ignore_states": null,
        "flapping_last_change": 0,
        "flapping_threshold": 0,
        "flapping_threshold_high": 30,
        "flapping_threshold_low": 25,
        "force_next_check": false,
        "force_next_notification": false,
        "groups": [
          "puppet-services"
        ],
        "ha_mode": 0,
        "handled": false,
        "host_name": "test-host.example.com",
        "icon_image": "",
        "icon_image_alt": "",
        "last_check": 1664023939.063637,
        "last_check_result": {
          "active": true,
          "check_source": "test-icinga2.example.com",
          "command": [
            "/usr/lib/nagios/plugins/check_dns",
            "-H",
            "test-host.example.com",
            "-t",
            "10"
          ],
          "execution_end": 1664023939.06359,
          "execution_start": 1664023939.043969,
          "exit_status": 0,
          "output": "DNS OK: 0.016 seconds response time. test-host.example.com returns 10.0.80.24",
          "performance_data": [
            "time=0.016055s;;;0.000000"
          ],
          "previous_hard_state": 0,
          "schedule_end": 1664023939.063637,
          "schedule_start": 1664023939.0434449,
          "scheduling_source": "test-icinga2.example.com",
          "state": 1,
          "ttl": 0,
          "type": "CheckResult",
          "vars_after": {
            "attempt": 1,
            "reachable": true,
            "state": 0,
            "state_type": 1
          },
          "vars_before": {
            "attempt": 1,
            "reachable": true,
            "state": 0,
            "state_type": 1
          }
        },
        "last_hard_state": 0,
        "last_hard_state_change": 1661729905.760361,
        "last_reachable": true,
        "last_state": 0,
        "last_state_change": 1661729905.760361,
        "last_state_critical": 1661729846.327839,
        "last_state_ok": 1664023939.06359,
        "last_state_type": 1,
        "last_state_unknown": 0,
        "last_state_unreachable": 1645104485.517008,
        "last_state_warning": 0,
        "max_check_attempts": 3,
        "name": "puppet-dns-test-host.example.com",
        "next_check": 1664024227.333698,
        "next_update": 1664024527.3739884,
        "notes": "",
        "notes_url": "",
        "original_attributes": null,
        "package": "_etc",
        "paused": false,
        "previous_state_change": 1661729905.760361,
        "problem": false,
        "retry_interval": 60,
        "severity": 0,
        "source_location": {
          "first_column": 1,
          "first_line": 49,
          "last_column": 46,
          "last_line": 49,
          "path": "/etc/icinga2/conf.d/../puppet/zones.d/test-host.example.com/services.conf"
        },
        "state": 1,
        "state_type": 1,
        "templates": [
          "puppet-dns-test-host.example.com",
          "puppet-service",
          "generic-service"
        ],
        "type": "Service",
        "vars": {
          "dns_lookup": "test-host.example.com"
        },
        "version": 0,
        "volatile": false,
        "zone": "example.com"
      },
      "joins": {},
      "meta": {},
      "name": "test-host.example.com!puppet-dns-test-host.example.com",
      "type": "Service"
    },
    {
      "attrs": {
        "__name": "test-host.example.com!puppet-mailq",
        "acknowledgement": 0,
        "acknowledgement_expiry": 0,
        "acknowledgement_last_change": 0,
        "action_url": "",
        "active": true,
        "check_attempt": 1,
        "check_command": "mailq-sudo",
        "check_interval": 300,
        "check_period": "",
        "check_timeout": null,
        "command_endpoint": "",
        "display_name": "Mailq length",
        "downtime_depth": 0,
        "enable_active_checks": true,
        "enable_event_handler": true,
        "enable_flapping": false,
        "enable_notifications": true,
        "enable_passive_checks": true,
        "enable_perfdata": true,
        "event_command": "",
        "executions": null,
        "flapping": false,
        "flapping_current": 0,
        "flapping_ignore_states": null,
        "flapping_last_change": 1661738000.648272,
        "flapping_threshold": 0,
        "flapping_threshold_high": 30,
        "flapping_threshold_low": 25,
        "force_next_check": false,
        "force_next_notification": false,
        "groups": [
          "puppet-services"
        ],
        "ha_mode": 0,
        "handled": false,
        "host_name": "test-host.example.com",
        "icon_image": "",
        "icon_image_alt": "",
        "last_check": 1664023875.574268,
        "last_check_result": {
          "active": true,
          "check_source": "test-host.example.com",
          "command": [
            "/usr/bin/sudo",
            "/usr/lib/nagios/plugins/check_mailq",
            "-c",
            "7",
            "-w",
            "3"
          ],
          "execution_end": 1664023875.574218,
          "execution_start": 1664023875.480481,
          "exit_status": 0,
          "output": "OK: exim mailq (0) is below threshold (3/7)",
          "performance_data": [
            "unsent=0;3;7;0"
          ],
          "previous_hard_state": 0,
          "schedule_end": 1664023875.574268,
          "schedule_start": 1664023875.48,
          "scheduling_source": "test-host.example.com",
          "state": 2,
          "ttl": 0,
          "type": "CheckResult",
          "vars_after": {
            "attempt": 1,
            "reachable": true,
            "state": 0,
            "state_type": 1
          },
          "vars_before": {
            "attempt": 1,
            "reachable": true,
            "state": 0,
            "state_type": 1
          }
        },
        "last_hard_state": 0,
        "last_hard_state_change": 1661734732.062849,
        "last_reachable": true,
        "last_state": 0,
        "last_state_change": 1661734732.062849,
        "last_state_critical": 0,
        "last_state_ok": 1664023875.574218,
        "last_state_type": 1,
        "last_state_unknown": 0,
        "last_state_unreachable": 1652699084.542508,
        "last_state_warning": 1661734674.919644,
        "max_check_attempts": 3,
        "name": "puppet-mailq",
        "next_check": 1664024171.514693,
        "next_update": 1664024471.703129,
        "notes": "",
        "notes_url": "",
        "original_attributes": null,
        "package": "_etc",
        "paused": false,
        "previous_state_change": 1661734732.062849,
        "problem": false,
        "retry_interval": 60,
        "severity": 0,
        "source_location": {
          "first_column": 1,
          "first_line": 138,
          "last_column": 29,
          "last_line": 138,
          "path": "/etc/icinga2/conf.d/../puppet/zones.d/test-host.example.com/services.conf"
        },
        "state": 2,
        "state_type": 1,
        "templates": [
          "puppet-mailq",
          "puppet-service",
          "generic-service"
        ],
        "type": "Service",
        "vars": {
          "mailq_critical": 7,
          "mailq_warning": 3
        },
        "version": 0,
        "volatile": false,
        "zone": "test-host.example.com"
      },
      "joins": {},
      "meta": {},
      "name": "test-host.example.com!puppet-mailq",
      "type": "Service"
    }
  ]
}
`
