package matcher

import (
	"regexp"

	"github.com/newrelic/infra-integrations-sdk/log"
)

type matchResult int

const (
	noMatch matchResult = iota
	exclude
	include
)

//Matcher groups the rules to validate the service name
type Matcher struct {
	patterns []pattern
}
type pattern struct {
	exclude bool
	regex   *regexp.Regexp
}

// Match returns true if the string matches one of the patterns
// within one level of precedence, the last matching pattern decides the outcome
func (m *Matcher) Match(s string) bool {
	n := len(m.patterns)
	for i := n - 1; i >= 0; i-- {
		if match := m.patterns[i].match(s); match > noMatch {
			return match == include
		}
	}
	return false
}

// IsEmpty returns true if the Matcher has no patterns
func (m *Matcher) IsEmpty() bool {
	if len(m.patterns) != 0 {
		return false
	}
	return true
}
func (p pattern) match(s string) matchResult {
	match := p.regex.MatchString(s)

	if p.exclude && match {
		return exclude
	}

	if match {
		return include
	}
	return noMatch
}

// New create a new Matcher instance from slices of filters
// (not) (regex) "<filter>"
func New(filters []string) Matcher {
	var m Matcher

	r, _ := regexp.Compile("(not)?.?(regex)?.?\"(.+)\"")

	for _, line := range filters {
		var p pattern
		var filter string
		var isRegex, exclude bool

		if line == "" {
			log.Debug("filter line empty")
			continue
		}

		if s := r.FindStringSubmatch(line); s != nil {
			// s[1] -> (not)
			if s[1] != "" {
				exclude = true
			}
			// s[2] -> (regex)
			if s[2] != "" {
				isRegex = true
			}
			// s[3] -> \"(.+)\"
			if s[3] != "" {
				filter = s[3]
			}
		} else {
			filter = line
		}

		// if the filter is not a regex all special regex characters are escaped
		if !isRegex {
			filter = "^" + regexp.QuoteMeta(filter) + "$"
		}
		reg, err := regexp.Compile(filter)
		if err != nil {
			log.Warn("failed to compile regex:%s err:%v", reg, err)
			continue
		}
		log.Debug("pattern added regex: %v exclude: %v", filter, exclude)
		p.regex = reg
		p.exclude = exclude
		m.patterns = append(m.patterns, p)
	}
	return m
}
