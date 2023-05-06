package clock

import (
	"sort"
	"time"
)

type MockClock struct {
	now    time.Time
	timers []*struct {
		t  time.Time
		ch chan<- time.Time
	}
}

func NewMockClock(t time.Time) *MockClock {
	return &MockClock{
		now: t,
		timers: make([]*struct {
			t  time.Time
			ch chan<- time.Time
		}, 0),
	}
}

func (c *MockClock) Progress(d time.Duration) {
	i := 0
	then := c.now.Add(d)
	for _, s := range c.timers {
		if s.t.Before(then) {
			s.ch <- s.t
			i++
		} else {
			break
		}
	}
	c.timers = c.timers[i:]
	c.now = then
}

func (c *MockClock) After(d time.Duration) <-chan time.Time {

	ch := make(chan time.Time)
	c.insertTimer(c.now.Add(d), ch)

	return ch
}

func (c *MockClock) insertTimer(t time.Time, ch chan<- time.Time) {
	i := sort.Search(len(c.timers), func(i int) bool {
		return c.timers[i].t.After(t)
	})
	c.timers = append(c.timers, nil)
	copy(c.timers[i+1:], c.timers[i:])

	c.timers[i] = &struct {
		t  time.Time
		ch chan<- time.Time
	}{
		t,
		ch,
	}
}

// NewTicker ticks only once
func (c *MockClock) NewTicker(d time.Duration) *time.Ticker {
	ch := make(chan time.Time)
	c.insertTimer(c.now.Add(d), ch)

	return &time.Ticker{C: ch}
}

func (c *MockClock) Now() time.Time {
	return c.now
}
