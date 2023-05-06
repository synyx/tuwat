package clock

import "time"

type RealClock struct {
}

func NewClock() *RealClock {
	return &RealClock{}
}

func (c *RealClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (c *RealClock) NewTicker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}

func (c *RealClock) Now() time.Time {
	return time.Now()
}
