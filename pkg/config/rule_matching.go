package config

import (
	"math"
	"regexp"
	"strconv"
	"strings"
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

func ParseRuleMatcher(label, s string) RuleMatcher {
	if strings.HasPrefix(s, "~= ") {
		return regexpMatcher{regexp.MustCompile(s[3:])}
	} else if strings.HasPrefix(s, "> ") {
		return newNumberMatcher(gt, s[2:])
	} else if strings.HasPrefix(s, "= ") {
		if _, err := strconv.ParseFloat(s[2:], 64); err == nil {
			return newNumberMatcher(eq, s[2:])
		} else {
			return equalityMatcher{s[2:]}
		}
	} else if strings.HasPrefix(s, ">= ") {
		return newNumberMatcher(ge, s[3:])
	} else if strings.HasPrefix(s, "< ") {
		return newNumberMatcher(lt, s[2:])
	} else if strings.HasPrefix(s, "<= ") {
		return newNumberMatcher(le, s[3:])
	} else {
		return regexpMatcher{regexp.MustCompile(s)}
	}
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
