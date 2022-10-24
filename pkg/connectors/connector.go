package connectors

import (
	"context"
	html "html/template"
	"time"
)

type Connector interface {
	Tag() string
	Collector
}

type Collector interface {
	Collect(ctx context.Context) ([]Alert, error)
}

type SilencerFunc func(ctx context.Context, duration time.Duration, user string) error

type Alert struct {
	Labels      map[string]string
	Start       time.Time
	State       State
	Description string
	Details     string
	Links       []html.HTML
	Silence     SilencerFunc
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
