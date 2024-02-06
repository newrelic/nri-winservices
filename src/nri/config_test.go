/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package nri

import (
	"io/fs"
	"os"
	"path"
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

	configPath := path.Join(t.TempDir(), "config.yml")

	readOnly := 0o444
	err := os.WriteFile(configPath, content, fs.FileMode(readOnly))
	require.NoError(t, err)

	_, err = NewConfig(configPath)
	require.NoError(t, err)
}
