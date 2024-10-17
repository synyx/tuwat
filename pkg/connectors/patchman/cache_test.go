package patchman

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
)

func TestCache(t *testing.T) {
	clockMock := clock.NewMock()
	ctx := context.Background()

	cache := newCache[int](clockMock, 10*time.Minute)
	var called atomic.Uint64

	f := func(ctx context.Context) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(20 * time.Millisecond):
			called.Add(1)
			return int(called.Load()), nil
		}
	}

	// the first go around, we expect that the func hasn't been called yet
	i, err := cache.get(ctx, f)
	if err != nil {
		t.Fatal(err)
	}
	if i != 0 {
		t.Error("expected 0, got ", i)
	}

	clockMock.Add(1 * time.Minute)
	time.Sleep(20 * time.Millisecond)

	// it has been enough time, that the func has been called once now.
	i, err = cache.get(ctx, f)
	if err != nil {
		t.Fatal(err)
	}

	if i != 1 {
		t.Error("expected 1, got ", i)
	}
	if called.Load() != 1 {
		t.Error("expected to be called only once, was called", called.Load(), "times")
	}

	clockMock.Add(20 * time.Minute)
	_, err = cache.get(ctx, f)
	time.Sleep(20 * time.Millisecond)

	_, err = cache.get(ctx, f)
	if err != nil {
		t.Fatal(err)
	}

	if called.Load() != 2 {
		t.Error("expected to be called once more, was called", called.Load(), "times")
	}
}
