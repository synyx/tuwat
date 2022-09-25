package connectors

import (
	"context"
	"time"
)

type Connector interface {
	Name() string
	Collect(ctx context.Context) ([]Alert, error)
}

type Alert struct {
	Tags        map[string]string
	Start       time.Time
	State       State
	Description string
	Details     string
	Links       map[string]string
}

type State int

const (
	OK       State = 0
	Warning        = 1
	Critical       = 2
	Unknown        = 3
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
