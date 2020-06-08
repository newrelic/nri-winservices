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

// New create a new Matcher instance from a multiline filter
// es:
// - |
//   windowsService.name:
//   regex "^win.*$"
//   "newrelic-infra"
//   ! "winmgmt"
func New(filterList string) Matcher {
	var m Matcher
	scanner := bufio.NewScanner(strings.NewReader(filterList))
	r, _ := regexp.Compile("\"(.+)\"")

	// On the current implementation "windowsService.name" is the only attribute available for filter
	// this attribute should be specified at the top of the filter.
	if scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "windowsService.name:" {
			log.Error("filter attribute not supported: %s", line)
			return m
		}
	}

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

		filter := strings.ReplaceAll(s[0], "\"", "")
		// if the filter is not a regex all special regex characters are escaped
		if !isRegex {
			filter = "^" + regexp.QuoteMeta(filter) + "$"
		}
		reg, err := regexp.Compile(filter)
		if err != nil {
			log.Warn("failed to compile regex:%s err:%v", reg, err)
			continue
		}
		log.Debug("pattern added regex: %v exclude: %v", p.regex, p.exclude)
		p.regex = reg
		m.patterns = append(m.patterns, p)
	}
	return m
}
