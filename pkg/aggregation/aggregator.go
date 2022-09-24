package aggregation

import (
	"time"

	"github.com/synyx/gonagdash/pkg/config"
)

type Aggregate struct {
	CheckTime time.Time
	Alerts    []Alert
}

type Alert struct {
	Where  string
	What   string
	When   time.Duration
	Status string
}

type Aggregator struct {
}

func NewAggregator(cfg *config.Config) *Aggregator {
	return &Aggregator{}
}

func (a *Aggregator) Alerts() Aggregate {
	return Aggregate{}
}
