/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package matcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatcherMatch(t *testing.T) {
	var filterList = []string{
		`customImportantService`,
		`"special.?^ServiceWithSpecialChars" #Comments`,
		`regex "^Important.*$" #Comments`,
		`regex`,
		`regex .*`,
		`.*`,
		`"quoted"`,
	}

	m := New(filterList)

	assert.True(t, m.Match("customimportantservice"))
	assert.True(t, m.Match("special.?^ServiceWithSpecialChars"))
	assert.True(t, m.Match("importantServiceSub"))
	assert.True(t, m.Match("quoted"))
	assert.False(t, m.Match("notimportantService"))
	assert.False(t, m.Match("randomService"))
}

func TestMatcherWithExcludes(t *testing.T) {
	var includeFilters = []string{
		`regex ".*"`, // Include all services
	}

	var excludeFilters = []string{
		`"Windows Update"`,
		`regex "^(Themes|Spooler)$"`,
	}

	m := NewWithIncludesExcludes(includeFilters, excludeFilters)

	// Should include services that match include but not exclude
	assert.True(t, m.Match("newrelic-infra"))
	assert.True(t, m.Match("CustomService"))

	// Should exclude services that match exclude filters
	assert.False(t, m.Match("windows update"))
	assert.False(t, m.Match("Themes"))
	assert.False(t, m.Match("Spooler"))

	// Should exclude even if matches include
	assert.False(t, m.Match("Windows Update"))
}

func TestMatcherWithExcludesOnly(t *testing.T) {
	var includeFilters = []string{
		`"ServiceA"`,
		`"ServiceB"`,
	}

	var excludeFilters = []string{
		`"ServiceA"`,
	}

	m := NewWithIncludesExcludes(includeFilters, excludeFilters)

	// ServiceA should be excluded even though it's in include list
	assert.False(t, m.Match("ServiceA"))

	// ServiceB should be included since it's not in exclude list
	assert.True(t, m.Match("ServiceB"))

	// ServiceC should not match since it's not in include list
	assert.False(t, m.Match("ServiceC"))
}

func TestMatcherBothIncludeAndExclude(t *testing.T) {
	var includeFilters = []string{
		`regex "^Windows.*"`, // Include all Windows services
		`"CustomService"`,    // Include specific custom service
	}

	var excludeFilters = []string{
		`"Windows Update"`,  // Exclude Windows Update specifically
		`regex ".*Audio.*"`, // Exclude any audio-related services
	}

	m := NewWithIncludesExcludes(includeFilters, excludeFilters)

	// Should include: matches include pattern and doesn't match exclude
	assert.True(t, m.Match("Windows Defender"))
	assert.True(t, m.Match("Windows Time"))
	assert.True(t, m.Match("CustomService"))

	// Should exclude: matches include pattern BUT also matches exclude pattern
	assert.False(t, m.Match("Windows Update"))       // Explicitly excluded
	assert.False(t, m.Match("Windows Audio"))        // Matches audio exclude pattern
	assert.False(t, m.Match("Custom Audio Service")) // Matches audio exclude pattern

	// Should exclude: doesn't match any include pattern
	assert.False(t, m.Match("Linux Service"))
	assert.False(t, m.Match("RandomService"))
	assert.False(t, m.Match("SomeOtherService"))
}

func TestMatcherOnlyIncludeFilters(t *testing.T) {
	// Test scenario 1: Only include_matching_entities provided
	var includeFilters = []string{
		`"newrelic-infra"`,
		`regex "^CustomService.*$"`,
	}

	m := NewWithIncludesExcludes(includeFilters, nil)

	// Should match services in the include list
	assert.True(t, m.Match("newrelic-infra"))
	assert.True(t, m.Match("CustomService123"))
	assert.True(t, m.Match("CustomServiceABC"))

	// Should not match services not in the include list
	assert.False(t, m.Match("Windows Update"))
	assert.False(t, m.Match("Spooler"))
	assert.False(t, m.Match("RandomService"))
}

// Removed TestMatcherOnlyExcludeFilters because exclude-only filtering is not supported
// Include patterns are always required for proper filtering behavior
