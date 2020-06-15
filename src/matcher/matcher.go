/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package matcher

import (
	"regexp"

	"github.com/newrelic/infra-integrations-sdk/log"
)

//Matcher groups the rules to validate the service name
type Matcher struct {
	patterns []pattern
}
type pattern struct {
	regex *regexp.Regexp
}

// Match returns true if the string matches one of the patterns
func (m *Matcher) Match(s string) bool {
	for _, p := range m.patterns {
		if p.match(s) {
			return true
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
func (p pattern) match(s string) bool {
	return p.regex.MatchString(s)
}

// New create a new Matcher instance from slices of filters
// (regex) "<filter>"
func New(filters []string) Matcher {
	var m Matcher

	r, _ := regexp.Compile("(regex)?.?\"(.+)\"")

	for _, line := range filters {
		var p pattern
		var filter string
		var isRegex bool

		if line == "" {
			log.Debug("filter line empty")
			continue
		}

		if s := r.FindStringSubmatch(line); s != nil {
			// s[1] -> (regex)
			if s[1] != "" {
				isRegex = true
			}
			// s[2] -> \"(.+)\"
			if s[2] != "" {
				filter = s[2]
			}
		} else {
			filter = line
		}

		// if the filter is not a regex all special regex characters are escaped
		if !isRegex {
			filter = "^" + regexp.QuoteMeta(filter) + "$"
		}
		// windows services names are collected in lower case, (?i) is the case insensitive flag
		reg, err := regexp.Compile("(?i)" + filter)
		if err != nil {
			log.Warn("failed to compile regex:%s err:%v", reg, err)
			continue
		}
		log.Debug("pattern added regex: %v ", filter)
		p.regex = reg
		m.patterns = append(m.patterns, p)
	}
	return m
}
