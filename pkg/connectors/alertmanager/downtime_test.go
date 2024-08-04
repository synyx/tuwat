package alertmanager

import (
	"context"
	"testing"
)

func TestIcinga2Downtime(t *testing.T) {
	connector, closer := testConnector(map[string]string{
		"/silences": mockSilenceResponse,
	})
	defer closer()

	downtimes, err := connector.CollectDowntimes(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(downtimes) == 0 {
		t.Fatal("expected non-zero downtime")
	}

	if downtimes[0].Author != "buch" {
		t.Error("author mismatch")
	}
}

const mockSilenceResponse = `
[
  {
    "id": "be3e0ba5-224a-4c7b-985f-d2a8bd7ff7c2",
    "status": {
      "state": "active"
    },
    "updatedAt": "2024-07-31T08:29:31.192Z",
    "comment": "buch kümmert sich die tage",
    "createdBy": "buch",
    "endsAt": "2024-08-05T10:00:00.000Z",
    "matchers": [
      {
        "isEqual": true,
        "isRegex": false,
        "name": "job",
        "value": "metrics"
      }
    ],
    "startsAt": "2024-07-31T08:29:31.192Z"
  }
]
`
