package connectors

import (
	"context"
	"time"

	"github.com/synyx/tuwat/pkg/ruleengine"
)

type DowntimeCollector interface {
	CollectDowntimes(context.Context) ([]Downtime, error)
}

type Downtime struct {
	Author    string
	Comment   string
	StartTime time.Time
	EndTime   time.Time
	Matchers  map[string]ruleengine.RuleMatcher
}
