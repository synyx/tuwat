package ruleengine

type Rule struct {
	Description string
	What        RuleMatcher
	When        RuleMatcher
	Labels      map[string]RuleMatcher
}
