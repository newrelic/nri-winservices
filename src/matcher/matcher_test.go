package matcher

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func TestMatcherMatch(t *testing.T) {
	var filterList = `windowsService.name:
"customImportantService"
"important.?^ServiceWithSpecialChars" #Comments
regex "important.*$" #Comments

! "importantServiceToExclude"
 ! "importantServiceToExcludeSpacePrefix"
! regex "notImportant.*"`

	v := New(filterList)
	filtersCount := strings.Count(filterList, "\"") / 2
	assert.Len(t, v.patterns, filtersCount)
	for _, p := range v.patterns {
		fmt.Printf("exclude:%v regex:%v\n", p.exclude, p.regex)
	}
	assert.True(t, v.Match("customImportantService"))
	assert.True(t, v.Match("important.?^ServiceWithSpecialChars"))
	assert.True(t, v.Match("importantServiceSub"))
	assert.False(t, v.Match("importantServiceToExclude"))
	assert.False(t, v.Match("importantServiceToExcludeSpacePrefix"))
	assert.False(t, v.Match("notImportantService"))
	assert.False(t, v.Match("randomService"))
}
func TestPatternMatch(t *testing.T) {
	i := pattern{
		exclude: false,
		regex:   "^importantService$",
	}
	require.Equal(t, include, i.match("importantService"))
	require.Equal(t, noMatch, i.match("importantServiceTest"))
	e := pattern{
		exclude: true,
		regex:   "^notImportantService$",
	}
	require.Equal(t, exclude, e.match("notImportantService"))
	require.Equal(t, noMatch, e.match("importantService"))
	require.Equal(t, noMatch, e.match("importantServiceTest"))

	r := pattern{
		exclude: true,
		regex:   "notImportant.*",
	}
	require.Equal(t, exclude, r.match("notImportantService"))
	require.Equal(t, exclude, r.match("notImportantServiceTest"))
	require.Equal(t, noMatch, r.match("importantService"))
}
