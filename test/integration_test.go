//go:build integration
// +build integration

/*
* Copyright 2020 New Relic Corporation. All rights reserved.
* SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/newrelic/nri-winservices/src/exporter"
	"github.com/newrelic/nri-winservices/test/jsonschema"
	"github.com/stretchr/testify/require"
)

//This can set whn running the test as -ldflags "-X github.com/newrelic/nri-winservices/test.integrationPath="
var (
	integrationPath = "../target/bin/windows_amd64/"
)

func isProcessRunning(name string) bool {
	out, err := exec.Command("tasklist").Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.Contains(string(out), name)
}

func runIntegration() (string, string, error) {

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(
		ctx,

		integrationPath+"nri-winservices.exe",
		"-config_path", "./config.yml",
		"-verbose",
	)
	defer cmd.Wait()
	defer cancel()

	var outbuf, errbuf bytes.Buffer
	cmd.Stdout = &outbuf
	cmd.Stderr = &errbuf

	err := cmd.Start()
	if err != nil {
		path, _ := os.Getwd()
		return "", "", fmt.Errorf("fail to start cmd: %v, currentPath: %s", err, path)
	}

	time.Sleep(60 * time.Second)
	stdout := outbuf.String()
	stderr := errbuf.String()

	stdout = strings.ReplaceAll(stdout, "{}\n", "")
	stdout = strings.ReplaceAll(stdout, "{}\r\n", "")
	stdout = strings.ReplaceAll(stdout, "\n", "")
	stdout = strings.ReplaceAll(stdout, "\r\n", "")

	return stdout, stderr, nil
}
func TestIntegration(t *testing.T) {

	require.False(t, isProcessRunning(exporter.ExporterName))

	stdout, stderr, err := runIntegration()
	//Notice that stdErr contains as well normal logs of the integration
	require.NotNil(t, stderr, "unexpected stderr")
	require.NoError(t, err, "Unexpected error")
	fmt.Println(stdout)
	fmt.Println(err)
	fmt.Println(stderr)

	schemaPath := filepath.Join("json-schema-files", "winservices-schema.json")
	err = jsonschema.Validate(schemaPath, stdout)
	require.NoError(t, err, "The output of WinServices integration doesn't have expected format")
	require.Less(t, 0, len(stdout), "The output should be longer than 0")
	require.False(t, isProcessRunning(exporter.ExporterName))
}
