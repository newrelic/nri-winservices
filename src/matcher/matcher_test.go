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
		`regex "important.*$" #Comments`,
		`regex`,
		`regex .*`,
		`.*`,
		`"quoted"`,
	}

	m := New(filterList)

	assert.True(t, m.Match("customImportantService"))
	assert.True(t, m.Match("special.?^ServiceWithSpecialChars"))
	assert.True(t, m.Match("importantServiceSub"))
	assert.True(t, m.Match("quoted"))
	assert.True(t, m.Match("importantServiceSub"))
	assert.False(t, m.Match("randomService"))
}
