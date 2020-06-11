/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package matcher

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMatcherMatch(t *testing.T) {
	var filterList = []string{
		`customImportantService`,
		`"special.?^ServiceWithSpecialChars" #Comments`,
		`regex "important.*$" #Comments`,
		`not "importantServiceToExclude"`,
		` not "importantServiceToExcludeSpacePrefix"`,
		`not regex "notImportant.*"`,
	}

	m := New(filterList)
	// filtersCount := strings.Count(filterList, "\"") / 2
	// assert.Len(t, m.patterns, filtersCount)
	for _, p := range m.patterns {
		fmt.Printf("exclude:%v regex:%v\n", p.exclude, p.regex)
	}
	assert.True(t, m.Match("customImportantService"))
	assert.True(t, m.Match("special.?^ServiceWithSpecialChars"))
	assert.True(t, m.Match("importantServiceSub"))
	assert.False(t, m.Match("importantServiceToExclude"))
	assert.False(t, m.Match("importantServiceToExcludeSpacePrefix"))
	assert.False(t, m.Match("notImportantService"))
	assert.False(t, m.Match("randomService"))
}
func TestPatternMatch(t *testing.T) {
	regex, _ := regexp.Compile("^importantService$")
	i := pattern{
		exclude: false,
		regex:   regex,
	}
	require.Equal(t, include, i.match("importantService"))
	require.Equal(t, noMatch, i.match("importantServiceTest"))

	regex, _ = regexp.Compile("^notImportantService$")
	e := pattern{
		exclude: true,
		regex:   regex,
	}
	require.Equal(t, exclude, e.match("notImportantService"))
	require.Equal(t, noMatch, e.match("importantService"))
	require.Equal(t, noMatch, e.match("importantServiceTest"))

	regex, _ = regexp.Compile("notImportant.*")
	r := pattern{
		exclude: true,
		regex:   regex,
	}
	require.Equal(t, exclude, r.match("notImportantService"))
	require.Equal(t, exclude, r.match("notImportantServiceTest"))
	require.Equal(t, noMatch, r.match("importantService"))
}
