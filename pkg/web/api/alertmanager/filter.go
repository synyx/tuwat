package alertmanager

import (
	"bufio"
	"strings"

	"github.com/synyx/tuwat/pkg/rules"
)

type fieldMatcher struct {
	field string
	m     rules.RuleMatcher
}

func parseFilter(val string) []fieldMatcher {
	var matchers []fieldMatcher
	s := newFilterScanner(val)
	for {
		matcher := s.scan()
		if matcher.field == "" {
			return matchers
		}
		matchers = append(matchers, matcher)
	}
}

type scanner struct {
	r *bufio.Reader
}

var invalid = rune(0)

func newFilterScanner(s string) *scanner {
	return &scanner{r: bufio.NewReader(strings.NewReader(s))}
}

// read reads the next rune from the buffered reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *scanner) read() rune {
	ch, _, err := s.r.ReadRune()
	if err != nil {
		return invalid
	}
	return ch
}

// unread places the previously read rune back on the reader.
func (s *scanner) unread() { _ = s.r.UnreadRune() }

func (s *scanner) scan() fieldMatcher {
top:
	ch := s.read()
	switch ch {
	case '{', ',', ' ':
		// we're just starting, go on
		goto top
	case '}':
		// last element
		return fieldMatcher{}
	default:
		s.unread()
		field := s.parseField()
		prefix := s.parsePrefix()
		str := s.parseString()
		return fieldMatcher{
			field: field,
			m:     rules.ParseRuleMatcher(prefix + " " + str),
		}
	}
}

func (s *scanner) parseField() string {
	var field []rune
	for {
		ch := s.read()
		if ch == '~' || ch == '!' || ch == '=' {
			s.unread()
			return string(field)
		}
		field = append(field, ch)
	}
}
func (s *scanner) parsePrefix() string {
	var prefix []rune
	for {
		ch := s.read()
		if ch == '~' || ch == '!' || ch == '=' {
			prefix = append(prefix, ch)
		} else {
			s.unread()
			return string(prefix)
		}
	}
}

func (s *scanner) parseString() string {
	var prefix []rune
	_ = s.read() // "
	for {
		ch := s.read()
		if ch == '"' {
			return string(prefix)
		} else {
			prefix = append(prefix, ch)
		}
	}
}
