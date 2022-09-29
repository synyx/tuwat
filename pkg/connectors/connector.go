package connectors

import (
	"context"
	"time"
)

type Connector interface {
	Tag() string
	Collector
}

type Collector interface {
	Collect(ctx context.Context) ([]Alert, error)
}

type Alert struct {
	Labels      map[string]string
	Start       time.Time
	State       State
	Description string
	Details     string
	Links       map[string]string
}

type State int

const (
	OK       State = 0
	Warning  State = 1
	Critical State = 2
	Unknown  State = 3
)

func (s State) String() string {
	switch s {
	case OK:
		return "green"
	case Warning:
		return "yellow"
	case Critical:
		return "red"
	case Unknown:
		return "grey"
	}
	return "grey"
}
