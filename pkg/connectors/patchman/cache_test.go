package patchman

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

func TestCache(t *testing.T) {
	clockMock := clock.NewMock()

	cache := newCache[int](clockMock, 10*time.Minute)
	called := 0

	f := func() (int, error) {
		called += 1
		return 1, nil
	}

	clockMock.Add(1 * time.Minute)

	i, err := cache.get(f)
	if err != nil {
		t.Fatal(err)
	}

	if i != 1 {
		t.Error("expected 1, got ", i)
	}

	if cache.data != 1 {
		t.Error("expected 1, got ", cache.data)
	}

	clockMock.Add(1 * time.Minute)

	_, err = cache.get(f)
	if err != nil {
		t.Fatal(err)
	}

	if called != 1 {
		t.Error("expected to be called only once, was called", called, "times")
	}

	clockMock.Add(20 * time.Minute)

	_, err = cache.get(f)
	if err != nil {
		t.Fatal(err)
	}

	if called != 2 {
		t.Error("expected to be called once more, was called", called, "times")
	}
}
