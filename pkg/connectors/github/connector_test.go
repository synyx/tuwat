package github

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/connectors/common"
)

func TestConnector(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write([]byte(mockResponse))
	}))
	defer testServer.Close()

	cfg := Config{
		Tag:   "test",
		Repos: []string{"test/test"},
		HTTPConfig: common.HTTPConfig{
			URL: testServer.URL,
		},
	}

	var connector connectors.Connector = NewConnector(&cfg)
	alerts, err := connector.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if len(alerts) == 0 {
		t.Error("There should be alerts")
	}
}

func TestDecode(t *testing.T) {
	var foo []issue
	err := json.Unmarshal([]byte(mockResponse), &foo)
	if err != nil {
		t.Fatal(err)
	}
}

const mockResponse = `
[
  {
    "url": "https://api.github.com/repos/synyx/tuwat/issues/18",
    "repository_url": "https://api.github.com/repos/synyx/tuwat",
    "labels_url": "https://api.github.com/repos/synyx/tuwat/issues/18/labels{/name}",
    "comments_url": "https://api.github.com/repos/synyx/tuwat/issues/18/comments",
    "events_url": "https://api.github.com/repos/synyx/tuwat/issues/18/events",
    "html_url": "https://github.com/synyx/tuwat/pull/18",
    "id": 1429893506,
    "node_id": "PR_kwDOIS4PyM5B4v3d",
    "number": 18,
    "title": "Use OPA rego for filtering alerts",
    "user": {
      "login": "BuJo",
      "id": 4713,
      "node_id": "MDQ6VXNlcjQ3MTM=",
      "avatar_url": "https://avatars.githubusercontent.com/u/4713?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/BuJo",
      "html_url": "https://github.com/BuJo",
      "followers_url": "https://api.github.com/users/BuJo/followers",
      "following_url": "https://api.github.com/users/BuJo/following{/other_user}",
      "gists_url": "https://api.github.com/users/BuJo/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/BuJo/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/BuJo/subscriptions",
      "organizations_url": "https://api.github.com/users/BuJo/orgs",
      "repos_url": "https://api.github.com/users/BuJo/repos",
      "events_url": "https://api.github.com/users/BuJo/events{/privacy}",
      "received_events_url": "https://api.github.com/users/BuJo/received_events",
      "type": "User",
      "site_admin": false
    },
    "labels": [
      {
        "id": 4714772125,
        "node_id": "LA_kwDOIS4PyM8AAAABGQW2nQ",
        "url": "https://api.github.com/repos/synyx/tuwat/labels/help%20wanted",
        "name": "help wanted",
        "color": "008672",
        "default": true,
        "description": "Extra attention is needed"
      }
    ],
    "state": "open",
    "locked": false,
    "assignee": {
      "login": "BuJo",
      "id": 4713,
      "node_id": "MDQ6VXNlcjQ3MTM=",
      "avatar_url": "https://avatars.githubusercontent.com/u/4713?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/BuJo",
      "html_url": "https://github.com/BuJo",
      "followers_url": "https://api.github.com/users/BuJo/followers",
      "following_url": "https://api.github.com/users/BuJo/following{/other_user}",
      "gists_url": "https://api.github.com/users/BuJo/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/BuJo/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/BuJo/subscriptions",
      "organizations_url": "https://api.github.com/users/BuJo/orgs",
      "repos_url": "https://api.github.com/users/BuJo/repos",
      "events_url": "https://api.github.com/users/BuJo/events{/privacy}",
      "received_events_url": "https://api.github.com/users/BuJo/received_events",
      "type": "User",
      "site_admin": false
    },
    "assignees": [
      {
        "login": "BuJo",
        "id": 4713,
        "node_id": "MDQ6VXNlcjQ3MTM=",
        "avatar_url": "https://avatars.githubusercontent.com/u/4713?v=4",
        "gravatar_id": "",
        "url": "https://api.github.com/users/BuJo",
        "html_url": "https://github.com/BuJo",
        "followers_url": "https://api.github.com/users/BuJo/followers",
        "following_url": "https://api.github.com/users/BuJo/following{/other_user}",
        "gists_url": "https://api.github.com/users/BuJo/gists{/gist_id}",
        "starred_url": "https://api.github.com/users/BuJo/starred{/owner}{/repo}",
        "subscriptions_url": "https://api.github.com/users/BuJo/subscriptions",
        "organizations_url": "https://api.github.com/users/BuJo/orgs",
        "repos_url": "https://api.github.com/users/BuJo/repos",
        "events_url": "https://api.github.com/users/BuJo/events{/privacy}",
        "received_events_url": "https://api.github.com/users/BuJo/received_events",
        "type": "User",
        "site_admin": false
      }
    ],
    "milestone": null,
    "comments": 0,
    "created_at": "2022-10-31T13:56:38Z",
    "updated_at": "2023-05-06T15:59:00Z",
    "closed_at": null,
    "author_association": "MEMBER",
    "active_lock_reason": null,
    "draft": false,
    "pull_request": {
      "url": "https://api.github.com/repos/synyx/tuwat/pulls/18",
      "html_url": "https://github.com/synyx/tuwat/pull/18",
      "diff_url": "https://github.com/synyx/tuwat/pull/18.diff",
      "patch_url": "https://github.com/synyx/tuwat/pull/18.patch",
      "merged_at": null
    },
    "body": "* Use the rego language from Open Policy Agent\r\n* see https://www.openpolicyagent.org/docs/latest/integration/#integrating-with-the-go-api",
    "reactions": {
      "url": "https://api.github.com/repos/synyx/tuwat/issues/18/reactions",
      "total_count": 0,
      "+1": 0,
      "-1": 0,
      "laugh": 0,
      "hooray": 0,
      "confused": 0,
      "heart": 0,
      "rocket": 0,
      "eyes": 0
    },
    "timeline_url": "https://api.github.com/repos/synyx/tuwat/issues/18/timeline",
    "performed_via_github_app": null,
    "state_reason": null
  },
  {
    "url": "https://api.github.com/repos/synyx/tuwat/issues/2",
    "repository_url": "https://api.github.com/repos/synyx/tuwat",
    "labels_url": "https://api.github.com/repos/synyx/tuwat/issues/2/labels{/name}",
    "comments_url": "https://api.github.com/repos/synyx/tuwat/issues/2/comments",
    "events_url": "https://api.github.com/repos/synyx/tuwat/issues/2/events",
    "html_url": "https://github.com/synyx/tuwat/issues/2",
    "id": 1420552266,
    "node_id": "I_kwDOIS4PyM5Uq-hK",
    "number": 2,
    "title": "Hide filtered alerts",
    "user": {
      "login": "BuJo",
      "id": 4713,
      "node_id": "MDQ6VXNlcjQ3MTM=",
      "avatar_url": "https://avatars.githubusercontent.com/u/4713?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/BuJo",
      "html_url": "https://github.com/BuJo",
      "followers_url": "https://api.github.com/users/BuJo/followers",
      "following_url": "https://api.github.com/users/BuJo/following{/other_user}",
      "gists_url": "https://api.github.com/users/BuJo/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/BuJo/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/BuJo/subscriptions",
      "organizations_url": "https://api.github.com/users/BuJo/orgs",
      "repos_url": "https://api.github.com/users/BuJo/repos",
      "events_url": "https://api.github.com/users/BuJo/events{/privacy}",
      "received_events_url": "https://api.github.com/users/BuJo/received_events",
      "type": "User",
      "site_admin": false
    },
    "labels": [],
    "state": "open",
    "locked": false,
    "assignee": null,
    "assignees": [],
    "milestone": null,
    "comments": 0,
    "created_at": "2022-10-24T09:49:19Z",
    "updated_at": "2022-10-24T09:49:19Z",
    "closed_at": null,
    "author_association": "MEMBER",
    "active_lock_reason": null,
    "body": "Currently the filtered alerts are shown with information which filter had been applied.  This helps the development of the code and development of the configuration, but overloads the interface.\r\n\r\nThe filtered list could be hidden behind a configuration item or personal configuration or some other method.",
    "reactions": {
      "url": "https://api.github.com/repos/synyx/tuwat/issues/2/reactions",
      "total_count": 0,
      "+1": 0,
      "-1": 0,
      "laugh": 0,
      "hooray": 0,
      "confused": 0,
      "heart": 0,
      "rocket": 0,
      "eyes": 0
    },
    "timeline_url": "https://api.github.com/repos/synyx/tuwat/issues/2/timeline",
    "performed_via_github_app": null,
    "state_reason": null
  }
]
`
