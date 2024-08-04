package connectors

import (
	"time"

	"github.com/synyx/tuwat/pkg/rules"
)

type Downtime struct {
	Author    string
	Comment   string
	StartTime time.Time
	EndTime   time.Time
	Matchers  map[string]rules.RuleMatcher
}
