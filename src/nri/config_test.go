/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package nri

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	content := []byte(`
exporter_bind_address: 127.0.0.1
exporter_bind_port: 9182
scrape_interval: 30s
include_matching_entities:
  windowsService.name:
    - regex ".*"
    - "ServiceNameToBeIncluded"`)

	tmpfile, err := ioutil.TempFile("", "config")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name()) // clean up
	_, err = tmpfile.Write(content)
	require.NoError(t, err)

	_, err = NewConfig(tmpfile.Name())
	require.NoError(t, err)
}

func TestNewConfigWithExcludes(t *testing.T) {
	content := []byte(`
exporter_bind_address: 127.0.0.1
exporter_bind_port: 9182
scrape_interval: 30s
include_matching_entities:
  windowsService.name:
    - regex ".*"
exclude_matching_entities:
  windowsService.name:
    - "Windows Update"
    - regex "^(Themes|Spooler)$"`)

	tmpfile, err := ioutil.TempFile("", "config")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name()) // clean up
	_, err = tmpfile.Write(content)
	require.NoError(t, err)

	config, err := NewConfig(tmpfile.Name())
	require.NoError(t, err)

	// Test that the matcher works correctly with excludes
	require.True(t, config.Matcher.Match("newrelic-infra"))
	require.False(t, config.Matcher.Match("Windows Update"))
	require.False(t, config.Matcher.Match("Themes"))
}

func TestNewConfigBothIncludeAndExclude(t *testing.T) {
	content := []byte(`
exporter_bind_address: 127.0.0.1
exporter_bind_port: 9182
scrape_interval: 30s
include_matching_entities:
  windowsService.name:
    - regex "^Windows.*"
    - "CustomService"
exclude_matching_entities:
  windowsService.name:
    - "Windows Update"
    - regex ".*Audio.*"`)

	tmpfile, err := ioutil.TempFile("", "config")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name()) // clean up
	_, err = tmpfile.Write(content)
	require.NoError(t, err)

	config, err := NewConfig(tmpfile.Name())
	require.NoError(t, err)

	// Test comprehensive include/exclude logic
	// Should include: matches include and doesn't match exclude
	require.True(t, config.Matcher.Match("Windows Defender"))
	require.True(t, config.Matcher.Match("CustomService"))

	// Should exclude: matches include BUT also matches exclude
	require.False(t, config.Matcher.Match("Windows Update"))
	require.False(t, config.Matcher.Match("Windows Audio"))

	// Should exclude: doesn't match include
	require.False(t, config.Matcher.Match("Linux Service"))
}

func TestNewConfigExcludeOnlyNotSupported(t *testing.T) {
	content := []byte(`
exporter_bind_address: 127.0.0.1
exporter_bind_port: 9182
scrape_interval: 30s
exclude_matching_entities:
  windowsService.name:
    - "Windows Update"
    - regex "^(Themes|Spooler)$"`)

	tmpfile, err := ioutil.TempFile("", "config")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name()) // clean up
	_, err = tmpfile.Write(content)
	require.NoError(t, err)

	// Should fail because exclude-only is not supported
	_, err = NewConfig(tmpfile.Name())
	require.Error(t, err)
	require.Contains(t, err.Error(), "include_matching_entities is required")
}
