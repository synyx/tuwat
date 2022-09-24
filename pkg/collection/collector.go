package collection

import "time"

type Collector interface {
	Collect() []Alert
}

type Alert struct {
	Tags  map[string]string
	Start time.Time
	State State
}

type State int

const (
	OK       State = 0
	Warning        = 1
	Critical       = 2
	Unknown        = 3
)
