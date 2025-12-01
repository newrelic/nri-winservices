//go:build windows && amd64
// +build windows,amd64

/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package matcher

import (
	"regexp"

	"github.com/newrelic/infra-integrations-sdk/v4/log"
)

// Matcher groups the rules to validate the service name
type Matcher struct {
	includePatterns []pattern
	excludePatterns []pattern
}
type pattern struct {
	regex *regexp.Regexp
}

// Match returns true if the string matches include patterns and doesn't match exclude patterns
// Include patterns are required - this matcher does not support exclude-only filtering
func (m *Matcher) Match(s string) bool {
	// Must match at least one include pattern first
	includeMatch := false
	for _, p := range m.includePatterns {
		if p.match(s) {
			includeMatch = true
			break
		}
	}

	// If no include patterns match, return false
	if !includeMatch {
		return false
	}

	// Check if it matches any exclude patterns (exclude takes precedence)
	for _, p := range m.excludePatterns {
		if p.match(s) {
			return false
		}
	}

	return true
}

// IsEmpty returns true if the Matcher has no include patterns
// (exclude patterns alone are not sufficient for a valid matcher)
func (m *Matcher) IsEmpty() bool {
	return len(m.includePatterns) == 0
}

func (p pattern) match(s string) bool {
	return p.regex.MatchString(s)
}

// New create a new Matcher instance from slices of include and exclude filters
func New(includeFilters []string) Matcher {
	return NewWithIncludesExcludes(includeFilters, nil)
}

// NewWithExcludes creates a new Matcher instance with both include and exclude filters
// (regex) "<filter>"
func NewWithIncludesExcludes(includeFilters, excludeFilters []string) Matcher {
	var m Matcher

	m.includePatterns = buildPatterns(includeFilters)
	m.excludePatterns = buildPatterns(excludeFilters)

	return m
}

// buildPatterns creates patterns from filter strings
func buildPatterns(filters []string) []pattern {
	var patterns []pattern
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
		patterns = append(patterns, p)
	}
	return patterns
}
