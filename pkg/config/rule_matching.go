package config

import (
	"fmt"
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
	fmt.Stringer
}

var prefixMatcher = regexp.MustCompile(`^(~=|=~|!~|>|<|=|!=|<=|>=)\s+(.+)$`)

func ParseRuleMatcher(value string) RuleMatcher {
	matches := prefixMatcher.FindStringSubmatch(value)
	if matches != nil {
		prefix := matches[1]
		value := matches[2]
		switch prefix {
		case "=~", "~=":
			return newRegexpMatcher(value)
		case "!~":
			return not(newRegexpMatcher(value))
		case ">":
			return newNumberMatcher(gt, value)
		case "=":
			if _, err := strconv.ParseFloat(value, 64); err == nil {
				return newNumberMatcher(eq, value)
			} else {
				return newEqualityMatcher(value)
			}
		case "!=":
			if _, err := strconv.ParseFloat(value, 64); err == nil {
				return not(newNumberMatcher(eq, value))
			} else {
				return not(newEqualityMatcher(value))
			}
		case "<":
			return newNumberMatcher(lt, value)
		case "<=":
			return newNumberMatcher(le, value)
		case ">=":
			return newNumberMatcher(ge, value)
		}
	}

	return newRegexpMatcher(value)
}

// regexpMatcher matches a string if given regular expression matches anywhere in the string
type regexpMatcher struct {
	r *regexp.Regexp
}

func newRegexpMatcher(value string) regexpMatcher {
	return regexpMatcher{r: regexp.MustCompile(value)}
}

func (m regexpMatcher) MatchString(s string) bool {
	return m.r.MatchString(s)
}

func (m regexpMatcher) String() string {
	return fmt.Sprintf("Regexp[/%s/]", m.r.String())
}

// numberMatcher matches a given numerical value when the difference is within epsilon
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

func (m numberMatcher) String() string {
	var op string
	switch m.operation {
	case gt:
		op = ">"
	case eq:
		op = "=="
	case ge:
		op = ">="
	case lt:
		op = "<"
	case le:
		op = "<="
	}
	return fmt.Sprintf("Number[%s, %s]", op, strconv.FormatFloat(m.number, 'f', -1, 64))
}

// equalityMatcher matches the exact string
type equalityMatcher struct {
	s string
}

func newEqualityMatcher(s string) equalityMatcher {
	return equalityMatcher{s: s}
}

func (m equalityMatcher) MatchString(s string) bool {
	return m.s == s
}

func (m equalityMatcher) String() string {
	return fmt.Sprintf("String[\"%s\"]", m.s)
}

// notMatcher inverts a given rule matcher
type notMatcher struct {
	m RuleMatcher
}

func not(m RuleMatcher) notMatcher {
	return notMatcher{m}
}

func (n notMatcher) MatchString(s string) bool {
	return !n.m.MatchString(s)
}

func (n notMatcher) String() string {
	return fmt.Sprintf("!%s", n.m.String())
}

func FalseMatcher() *falseMatcher {
	return &falseMatcher{}
}

type falseMatcher struct{}

func (f falseMatcher) MatchString(_ string) bool {
	return false
}

func (f falseMatcher) String() string {
	return fmt.Sprintf("!")
}
