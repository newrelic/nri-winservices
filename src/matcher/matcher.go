package matcher

import (
	"bufio"
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
	regex   string
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
func (p pattern) match(s string) matchResult {
	match, err := regexp.MatchString(p.regex, s)
	if err != nil {
		log.Warn(err.Error())
		return noMatch
	}

	if p.exclude && match {
		return exclude
	}

	if match {
		return include
	}
	return noMatch
}

// New create a new Matcher instance from a multiline filter
// - |
//   windowsService.name:
//   regex "^win.*$"
//   "newrelic-infra"
//   ! "winmgmt"
func New(filterList string) Matcher {
	var m Matcher
	scanner := bufio.NewScanner(strings.NewReader(filterList))
	r, _ := regexp.Compile("\"(.+)\"")

	for scanner.Scan() {
		var p pattern
		isRegex := false
		line := strings.TrimSpace(scanner.Text())

		if strings.HasSuffix(line, ":") {
			// "windowsService.name":
			// first line of the filter represents the attribute were the filter is applied
			// we discard this since currently the filter apply only to service name.
			continue
		}

		if strings.HasPrefix(line, "!") {
			p.exclude = true
		}

		if strings.Contains(line, "regex ") {
			isRegex = true
		}

		s := r.FindAllString(line, -1)
		if len(s) != 1 {
			log.Warn("wrong syntax of filter in line: %s", line)
			continue
		}

		regex := strings.ReplaceAll(s[0], "\"", "")

		p.regex = regex
		// if the filter is not a regex all special regex characters are escaped
		if !isRegex {
			p.regex = "^" + regexp.QuoteMeta(p.regex) + "$"
		}
		log.Debug("pattern added regex: %v exclude: %v", p.regex, p.exclude)
		m.patterns = append(m.patterns, p)
	}
	return m
}
