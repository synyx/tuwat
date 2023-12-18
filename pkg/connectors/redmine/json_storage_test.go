package redmine

import (
	"os"
	"path"
	"testing"
)

func TestStoringAndLoading(t *testing.T) {
	data := map[string]labels{
		"foo": map[string]string{"a": "1", "b": "2"},
		"bar": map[string]string{"c": "3", "d": "4"},
	}

	tmpdir, err := os.MkdirTemp("", "tuwat_test")
	if err != nil {
		t.Fail()
	}
	statefile := path.Join(tmpdir, "state")

	if err := storeSilences(statefile, data); err != nil {
		t.Fail()
	}

	data2 := make(map[string]labels)
	if err := loadSilences(statefile, &data2); err != nil {
		t.Fail()
	} else if len(data2) != len(data) {
		t.Fail()
	}

	if err := storeSilences(statefile, data); err != nil {
		t.Fail()
	}
}
