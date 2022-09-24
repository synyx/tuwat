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
	Tag    string
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
	return Aggregate{
		CheckTime: time.Now(),
		Alerts: []Alert{
			{
				Where:  "kubernetes/k8s-apps",
				Tag:    "synyx",
				What:   "MR !272",
				When:   1 * time.Minute,
				Status: "yellow",
			}, {
				Where:  "foo.synyx.coffee",
				Tag:    "prod",
				What:   "MR !272",
				When:   2 * time.Hour,
				Status: "gray",
			}, {
				Where:  "foo.contargo.net",
				Tag:    "RZ1",
				What:   "MR !272",
				When:   25 * time.Hour * 24,
				Status: "red",
			},
		},
	}
}
