package clock

import "time"

type Clock struct {
}

func NewClock() *Clock {
	return &Clock{}
}

func (c *Clock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (c *Clock) NewTicker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}

func (c *Clock) Now() time.Time {
	return time.Now()
}
