package config

import (
	"math"
	"regexp"
	"strconv"
)

const (
	gt = iota
	eq
	ge
	lt
	le
)

type RuleMatcher interface {
	MatchString(s string) bool
}

var prefixMatcher = regexp.MustCompile(`^(~=|>|<|=|<=|>=)\s+(.*)`)

func ParseRuleMatcher(value string) RuleMatcher {
	matches := prefixMatcher.FindStringSubmatch(value)
	if matches != nil {
		prefix := matches[1]
		value := matches[2]
		switch prefix {
		case "~=":
			return regexpMatcher{regexp.MustCompile(value)}
		case ">":
			return newNumberMatcher(gt, value)
		case "=":
			if _, err := strconv.ParseFloat(value[2:], 64); err == nil {
				return newNumberMatcher(eq, value[2:])
			} else {
				return equalityMatcher{value[2:]}
			}
		case "<":
			return newNumberMatcher(lt, value)
		case "<=":
			return newNumberMatcher(le, value)
		case ">=":
			return newNumberMatcher(ge, value)
		}
	}

	return regexpMatcher{regexp.MustCompile(value)}
}

type regexpMatcher struct {
	r *regexp.Regexp
}

func (m regexpMatcher) MatchString(s string) bool {
	return m.r.MatchString(s)
}

type numberMatcher struct {
	operation int
	number    float64
}

func newNumberMatcher(op int, s string) numberMatcher {
	number, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic("config parsing error")
	}
	return numberMatcher{op, number}
}

func (m numberMatcher) MatchString(s string) bool {
	number, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// an invalid number cannot be compared, thus returns as non-matching
		return false
	}

	switch m.operation {
	case gt:
		return !floatEqualEnough(m.number, number) && number > m.number
	case eq:
		return floatEqualEnough(m.number, number)
	case ge:
		return floatEqualEnough(m.number, number) || number > m.number
	case lt:
		return !floatEqualEnough(m.number, number) && number < m.number
	case le:
		return floatEqualEnough(m.number, number) || number < m.number
	}

	return false
}

const epsilon = 1e-8

func floatEqualEnough(a, b float64) bool {
	return math.Abs(a-b) <= epsilon
}

type equalityMatcher struct {
	s string
}

func (m equalityMatcher) MatchString(s string) bool {
	return m.s == s
}