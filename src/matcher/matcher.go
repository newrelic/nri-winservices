package matcher

import (
	"regexp"
	"strings"

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
	//TODO i think this could be improved adding groups to detect exclude and regex
	r, _ := regexp.Compile("\"(.+)\"")

	for _, line := range filters {
		var p pattern
		var filter string
		isRegex := false
		line := strings.TrimSpace(line)

		if line == "" {
			log.Debug("filter line empty")
		}

		if strings.HasPrefix(line, "not") {
			p.exclude = true
		}

		if strings.Contains(line, "regex ") {
			isRegex = true
		}

		if !isRegex && !p.exclude {
			// double quotes are remove when unmarshal yml like: - "filter"
			filter = line
		} else {
			s := r.FindAllString(line, -1)
			if len(s) != 1 {
				log.Warn("wrong syntax of filter in line: %s", line)
				continue
			}
			filter = strings.ReplaceAll(s[0], "\"", "")
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
		log.Debug("pattern added regex: %v exclude: %v", filter, p.exclude)
		p.regex = reg
		m.patterns = append(m.patterns, p)
	}
	return m
}
