package patchman

import (
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

func (c *cache[T]) get(f func() (T, error)) (T, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.t.IsZero() && c.clock.Now().Before(c.t.Add(c.timeout)) {
		return c.data, c.lastErr
	}

	data, err := f()
	c.t = c.clock.Now()
	c.data = data
	c.lastErr = err

	return data, err
}
