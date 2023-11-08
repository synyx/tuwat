package config

import "regexp"

type RuleMatcher interface {
	MatchString(s string) bool
}

type regexpMatcher struct {
	r *regexp.Regexp
}

func (r regexpMatcher) MatchString(s string) bool {
	return r.r.MatchString(s)
}

func ParseRuleMatcher(k, v string) RuleMatcher {
	return regexpMatcher{regexp.MustCompile(v)}
}
