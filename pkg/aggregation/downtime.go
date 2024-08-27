package aggregation

import (
	"fmt"
	"time"

	"github.com/synyx/tuwat/pkg/connectors"
	"github.com/synyx/tuwat/pkg/ruleengine"
)

func (a *Aggregator) downtimeRules(downtimes []connectors.Downtime) []ruleengine.Rule {
	rules := make([]ruleengine.Rule, 0, len(downtimes))
	for _, dt := range downtimes {
		rule := ruleengine.Rule{
			Description: a.downtimeDescription(dt),
			Labels:      dt.Matchers,
		}
		rules = append(rules, rule)
	}

	return rules
}

func (a *Aggregator) downtimeDescription(dt connectors.Downtime) string {
	description := fmt.Sprintf("Downtimed %s: %s", a.niceDate(dt.EndTime), dt.Comment)
	if len(description) > 100 {
		description = description[:99] + "…"
	}
	return description
}

func (a *Aggregator) niceDate(t time.Time) string {
	d := a.clock.Now().Sub(t)
	if d > 2*time.Hour*24 {
		return t.Format("until 2006-01-02")
	} else if d > 2*time.Hour {
		return fmt.Sprintf("for %.0fh", d.Hours())
	} else if d > 2*time.Minute {
		return fmt.Sprintf("for %.0fm", d.Minutes())
	} else if d > 0 {
		return fmt.Sprintf("for %.0fs", d.Seconds())
	} else {
		return t.Format("ended 2006-01-02 15:04")
	}
}
