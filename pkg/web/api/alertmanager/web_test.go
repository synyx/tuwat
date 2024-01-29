package alertmanager

import (
	"testing"
)

func Test_newFilterScanner(t *testing.T) {
	s := newFilterScanner(`{a="b"}`)
	f := s.scan()
	if f.field == "" {
		t.Error("should be a valid field")
	}
}

func Test_parseFilter(t *testing.T) {
	filters := parseFilter(`{a="b"}`)
	if len(filters) != 1 {
		t.Error("should be a valid field")
	}
	if filters[0].field != "a" {
		t.Error("should be valid field")
	}
}

func Test_parseFilter2(t *testing.T) {
	filters := parseFilter(`{a="b", c!="d", eeee=~"ffff"}`)
	if len(filters) != 3 {
		t.Error("should be a valid fields")
	}
	if filters[1].field != "c" {
		t.Error("should be valid field")
	}
}
