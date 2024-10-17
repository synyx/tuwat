package patchman

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
)

type cache[T any] struct {
	data    T
	mu      sync.RWMutex
	t       time.Time
	lastErr error
	timeout time.Duration
	clock   clock.Clock
}

func newCache[T any](c clock.Clock, timeout time.Duration) *cache[T] {
	return &cache[T]{
		timeout: timeout,
		clock:   c,
	}
}

func (c *cache[T]) get(ctx context.Context, f func(context.Context) (T, error)) (T, error) {
	if c.t.IsZero() {
		slog.InfoContext(ctx, "initial")
		// this is the first run, try to get it real quick
		quickCtx, cancel := context.WithTimeout(ctx, 5*time.Millisecond)
		defer cancel()
		data, err := f(quickCtx)
		if err != nil && errors.Is(err, context.DeadlineExceeded) {
			slog.InfoContext(ctx, "quick failed, backgrounding job")
			go c.getBackground(ctx, f)
			return c.data, nil
		}
		slog.InfoContext(ctx, "quick succeeded")
		return data, err
	} else if c.clock.Now().Before(c.t.Add(c.timeout)) {
		slog.InfoContext(ctx, "cached")
		// the cache is populated and new enough, return last execution
		return c.data, c.lastErr
	}

	slog.InfoContext(ctx, "refresh")

	// need to refresh the cache
	go c.getBackground(ctx, f)
	return c.data, nil
}

func (c *cache[T]) getBackground(ctx context.Context, f func(context.Context) (T, error)) {
	if !c.mu.TryLock() {
		// do not queue too many background jobs
		return
	}
	defer c.mu.Unlock()

	data, err := f(ctx)
	if err != nil {
		c.lastErr = err
		slog.InfoContext(ctx, "background job failed", slog.Any("error", err))
		return
	}

	c.t = c.clock.Now()
	c.data = data

	slog.InfoContext(ctx, "background job succeeded")
}
