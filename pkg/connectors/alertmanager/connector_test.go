package alertmanager

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/synyx/tuwat/pkg/connectors"
)

func TestConnector(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write([]byte(mockResponse))
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

	if alerts == nil || len(alerts) != 3 {
		t.Error("There should be alerts")
	}
}

func TestDecode(t *testing.T) {
	var foo []alert
	err := json.Unmarshal([]byte(mockResponse), &foo)
	if err != nil {
		t.Fatal(err)
	}
}

const mockResponse = `
[
  {
    "annotations": {
      "description": "There were ` + "`" + `1` + "`" + ` messages in queue ` + "`" + `meister` + "`" + ` vhost ` + "`" + `/` + "`" + `\nfor the last 15 minutes in app cluster ` + "`" + `` + "`" + ` in namespace ` + "`" + `app-stage` + "`" + `.\n",
      "runbook": "https://gitlab.example.com/infra/k8s/app/-/tree/master/documentation/runbook.md#app-consumerless-queues",
      "summary": "app cluster has queues with messages but no consumers."
    },
    "endsAt": "2022-09-25T16:55:08.801Z",
    "fingerprint": "006eeab040ebbfb6",
    "receivers": [
      {
        "name": "ops-mails"
      }
    ],
    "startsAt": "2022-04-11T08:01:38.801Z",
    "status": {
      "inhibitedBy": [],
      "silencedBy": [],
      "state": "active"
    },
    "updatedAt": "2022-09-25T16:51:09.004Z",
    "generatorURL": "https://prometheus.example.com/graph?g0.expr=app_queue_messages%7Bqueue%21~%22.%2A%28dead%7Cdlq%7Cdebug%7Cmobile.sync%7Ctest%7Cmarvin%7Cresend%7Cmobile.events%29.%2A%22%7D+%3E+0+unless+app_queue_consumer_capacity+%3E+0&g0.tab=1",
    "labels": {
      "alertname": "appQueueMessagesNoConsumer",
      "cluster": "prometheus-apps",
      "container": "app",
      "endpoint": "prometheus",
      "instance": "100.111.11.198:15692",
      "job": "app-stage/app",
      "namespace": "app-stage",
      "pod": "app-server-0",
      "prometheus": "apps-monitoring/apps",
      "queue": "meister",
      "severity": "warning",
      "vhost": "/"
    }
  },
  {
    "annotations": {
      "description": "constraint violation of kind ContainerLimits in Pod gitlab-agent-7f967f9945-9k8p4 in namespace api-gateway-prod",
      "summary": "container <agent> has no resource limits"
    },
    "endsAt": "2022-09-25T16:54:42.994Z",
    "fingerprint": "00ae4411e137c417",
    "receivers": [
      {
        "name": "ops-mails"
      }
    ],
    "startsAt": "2022-08-30T09:40:12.994Z",
    "status": {
      "inhibitedBy": [],
      "silencedBy": [],
      "state": "active"
    },
    "updatedAt": "2022-09-25T16:50:43.008Z",
    "generatorURL": "https://prometheus.example.com/graph?g0.expr=opa_scorecard_constraint_violations+%3E+0&g0.tab=1",
    "labels": {
      "alertname": "GatekeeperConstraintViolations",
      "cluster": "prometheus-apps",
      "context": "gatekeeper",
      "endpoint": "9141-9141",
      "instance": "100.111.12.32:9141",
      "job": "opa-exporter",
      "kind": "ContainerLimits",
      "name": "pod-container-limits",
      "namespace": "monitoring",
      "pod": "opa-exporter-6fc88b44f4-6xgnn",
      "prometheus": "apps-monitoring/apps",
      "service": "opa-exporter",
      "severity": "warning",
      "violating_kind": "Pod",
      "violating_name": "gitlab-agent-7f967f9945-9k8p4",
      "violating_namespace": "api-gateway-prod",
      "violation_enforcement": "warn",
      "violation_msg": "container <agent> has no resource limits"
    }
  },
  {
    "annotations": {
      "description": "There were ` + "`" + `1` + "`" + ` dead lettered messages in queue ` + "`" + `selfcheckin.mot.checkin.command.deadletter` + "`" + ` vhost ` + "`" + `/` + "`" + `\nfor the last 15 minutes in app cluster ` + "`" + `` + "`" + ` in namespace ` + "`" + `app-stage` + "`" + `.\n",
      "runbook": "https://gitlab.example.com/infra/k8s/app/-/tree/master/documentation/runbook.md#app-deadletter-queue",
      "summary": "app cluster has dead letter messages"
    },
    "endsAt": "2022-09-25T16:55:08.801Z",
    "fingerprint": "01b99423f38362e5",
    "receivers": [
      {
        "name": "ops-mails"
      }
    ],
    "startsAt": "2022-06-01T14:10:08.801Z",
    "status": {
      "inhibitedBy": [],
      "silencedBy": [],
      "state": "active"
    },
    "updatedAt": "2022-09-25T16:51:08.900Z",
    "generatorURL": "https://prometheus.example.com/graph?g0.expr=app_queue_messages%7Bqueue%3D~%22.%2A%28dead%7Cdlq%29.%2A%22%7D+%3E+0&g0.tab=1",
    "labels": {
      "alertname": "appDeadletterQueueMessages",
      "cluster": "prometheus-apps",
      "container": "app",
      "endpoint": "prometheus",
      "instance": "100.111.11.198:15692",
      "job": "app-stage/app",
      "namespace": "app-stage",
      "pod": "app-server-0",
      "prometheus": "apps-monitoring/apps",
      "queue": "selfcheckin.mot.checkin.command.deadletter",
      "severity": "warning",
      "vhost": "/"
    }
  }
]
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
