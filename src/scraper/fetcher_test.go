//go:build windows && amd64
// +build windows,amd64

/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package scraper

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetReal(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "testdata/actualOutput")
	}))
	defer ts.Close()
	mfs, err := Get(http.DefaultClient, ts.URL)
	var actual []string
	for k := range mfs {
		actual = append(actual, k)
	}
	assert.NoError(t, err)
}
