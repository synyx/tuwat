package icinga2

import (
	"context"
	"testing"
)

func TestIcinga2Downtime(t *testing.T) {
	connector, mockServer := testConnector(map[string]string{
		"/host":     icinga2MockHostResponse,
		"/service":  icinga2MockServiceResponse,
		"/downtime": icinga2MockDowntimeResponse,
		"/comment":  icinga2MockCommentResponse,
	})
	defer func() { mockServer.Close() }()

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

const icinga2MockDowntimeResponse = `
{
  "results": [
    {
      "attrs": {
        "__name": "noodle-air.example.com!client!7e82a9c1-077c-4e1d-9aa2-93403a079db7",
        "active": true,
        "author": "buch",
        "authoritative_zone": "",
        "comment": "Braucht noch ein bisschen",
        "config_owner": "",
        "config_owner_hash": "",
        "duration": 0,
        "end_time": 1727094293,
        "entry_time": 1721733933.9597,
        "fixed": true,
        "ha_mode": 0,
        "host_name": "noodle-air.example.com",
        "legacy_id": 1,
        "name": "7e82a9c1-077c-4e1d-9aa2-93403a079db7",
        "original_attributes": null,
        "package": "_api",
        "parent": "",
        "paused": false,
        "remove_time": 0,
        "scheduled_by": "",
        "service_name": "client",
        "source_location": {
          "first_column": 0,
          "first_line": 1,
          "last_column": 69,
          "last_line": 1,
          "path": "/var/lib/icinga2/api/packages/_api/bca7507f-ab76-4439-b24a-c1b96c42c498/conf.d/downtimes/noodle-air.example.com!client!7e82a9c1-077c-4e1d-9aa2-93403a079db7.conf"
        },
        "start_time": 1721733893,
        "templates": [
          "7e82a9c1-077c-4e1d-9aa2-93403a079db7"
        ],
        "trigger_time": 1721733933.9597,
        "triggered_by": "",
        "triggers": [],
        "type": "Downtime",
        "version": 1721733933.959751,
        "was_cancelled": false,
        "zone": "example.com"
      },
      "joins": {},
      "meta": {},
      "name": "noodle-air.example.com!client!7e82a9c1-077c-4e1d-9aa2-93403a079db7",
      "type": "Downtime"
    }
  ]
}
`

const icinga2MockCommentResponse = `
{
  "results": [
    {
      "attrs": {
        "__name": "balkon.example.com!7a037357-00fd-4aa9-8099-4083ebd2682b",
        "active": true,
        "author": "buch",
        "entry_time": 1722245791.013065,
        "entry_type": 4,
        "expire_time": 1730755200,
        "ha_mode": 0,
        "host_name": "balkon.example.com",
        "legacy_id": 1,
        "name": "7a037357-00fd-4aa9-8099-4083ebd2682b",
        "original_attributes": null,
        "package": "_api",
        "paused": true,
        "persistent": false,
        "service_name": "",
        "source_location": {
          "first_column": 0,
          "first_line": 1,
          "last_column": 68,
          "last_line": 1,
          "path": "/var/lib/icinga2/api/packages/_api/bca7507f-ab76-4439-b24a-c1b96c42c498/conf.d/comments/balkon.example.com!7a037357-00fd-4aa9-8099-4083ebd2682b.conf"
        },
        "templates": [
          "7a037357-00fd-4aa9-8099-4083ebd2682b"
        ],
        "text": "der azubi hats immer noch nicht installiert, seine ausrede ist dass die ports nicht gehen.",
        "type": "Comment",
        "version": 1722245791.013109,
        "zone": "example.com"
      },
      "joins": {},
      "meta": {},
      "name": "balkon.example.com!7a037357-00fd-4aa9-8099-4083ebd2682b",
      "type": "Comment"
    }
  ]
}
`
