package redmine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/synyx/tuwat/pkg/connectors/common"
)

func TestSilence(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		res.WriteHeader(http.StatusOK)
		_, _ = res.Write([]byte(redmineApiMockResponse))
	}))
	defer func() { testServer.Close() }()

	cfg := &Config{
		Tag: "test",
		HTTPConfig: common.HTTPConfig{
			URL: testServer.URL,
		},
	}

	s := NewSilencer(cfg)
	s.SetSilence("7", map[string]string{"Hostname": "wiki-test.synyx.coffee"})
	if err := s.Refresh(context.Background()); err != nil {
		t.Error(err)
	}
	silence := s.Silenced(map[string]string{"Hostname": "wiki-test.synyx.coffee", "Foo": "Bar"})
	if silence.Silenced != true {
		t.Fail()
	}
}
